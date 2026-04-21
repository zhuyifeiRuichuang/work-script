"""Kopf reconciliation entrypoints."""

from __future__ import annotations

import hashlib
import logging
from typing import Any, Dict

import kopf
from kubernetes import client, config
from kubernetes.client.rest import ApiException

from hadoop_operator.apply import (
    delete_config_map,
    delete_ingress,
    delete_pdb,
    delete_service,
    delete_statefulset,
    ensure_service,
    replace_or_create_configmap,
    replace_or_create_ingress,
    replace_or_create_pdb,
    replace_or_create_statefulset,
)
from hadoop_operator.constants import GROUP, PLURAL, VERSION
from hadoop_operator.k8s_resources import (
    build_configmap,
    build_hadoop_ingress,
    build_zookeeper_configmap,
    clusterip_service_for_pod,
    datanode_statefulset,
    headless_service,
    journalnode_statefulset,
    namenode_ha_statefulset,
    namenode_pdb,
    namenode_statefulset,
    nodeport_service_for_pod,
    nodemanager_statefulset,
    normalize_spec,
    optional_nodeport_service,
    resourcemanager_statefulset,
    zookeeper_statefulset,
)

logger = logging.getLogger(__name__)


def _api_clients() -> tuple[
    client.CoreV1Api,
    client.AppsV1Api,
    client.PolicyV1Api,
    client.NetworkingV1Api,
]:
    return (
        client.CoreV1Api(),
        client.AppsV1Api(),
        client.PolicyV1Api(),
        client.NetworkingV1Api(),
    )


@kopf.on.startup()
def _on_startup(settings: kopf.Configuration, **_kwargs: Any) -> None:
    settings.posting.level = logging.INFO
    try:
        config.load_incluster_config()
        logger.info("Loaded in-cluster kubeconfig")
    except config.ConfigException:
        config.load_kube_config()
        logger.info("Loaded local kubeconfig")


def _delete_ha_infra(
    core: client.CoreV1Api,
    apps: client.AppsV1Api,
    namespace: str,
    cluster: str,
) -> None:
    for svc in (f"{cluster}-zookeeper", f"{cluster}-journalnode"):
        delete_service(core, namespace, svc)
    delete_statefulset(apps, namespace, f"{cluster}-zookeeper")
    delete_statefulset(apps, namespace, f"{cluster}-journalnode")
    delete_config_map(core, namespace, f"{cluster}-zookeeper-config")
    for i in (0, 1):
        delete_service(core, namespace, f"{cluster}-namenode-{i}-external")
        delete_service(core, namespace, f"{cluster}-namenode-{i}-web")


def _delete_ingress_ui(core: client.CoreV1Api, net: client.NetworkingV1Api, namespace: str, cluster: str) -> None:
    delete_ingress(net, namespace, f"{cluster}-hadoop-ui")
    for name in (
        f"{cluster}-namenode-0-web",
        f"{cluster}-namenode-1-web",
        f"{cluster}-resourcemanager-0-web",
    ):
        delete_service(core, namespace, name)


def _reconcile(body: Dict[str, Any], **_: Any) -> None:
    meta = body["metadata"]
    namespace = meta["namespace"]
    cluster = meta["name"]
    cr_uid = meta["uid"]
    cr_api_version = body["apiVersion"]
    cr_kind = body["kind"]
    cr_name = cluster

    norm = normalize_spec(body.get("spec"))
    core, apps, policy, net = _api_clients()

    cm = build_configmap(
        cluster=cluster,
        namespace=namespace,
        cr_api_version=cr_api_version,
        cr_kind=cr_kind,
        cr_name=cr_name,
        cr_uid=cr_uid,
        norm=norm,
    )
    cm_name = cm.metadata.name
    replace_or_create_configmap(core, namespace, cm)

    payload_parts = ["".join(f"{k}={v}" for k, v in sorted((cm.data or {}).items()))]
    zk_cm_name = f"{cluster}-zookeeper-config"
    if norm["hdfs"]["ha"]["enabled"]:
        zkcm = build_zookeeper_configmap(
            cluster=cluster,
            namespace=namespace,
            cr_api_version=cr_api_version,
            cr_kind=cr_kind,
            cr_name=cr_name,
            cr_uid=cr_uid,
            norm=norm,
        )
        replace_or_create_configmap(core, namespace, zkcm)
        payload_parts.append(
            "".join(f"{k}={v}" for k, v in sorted((zkcm.data or {}).items()))
        )
    else:
        _delete_ha_infra(core, apps, namespace, cluster)
        delete_config_map(core, namespace, zk_cm_name)

    config_revision = hashlib.sha256("".join(payload_parts).encode()).hexdigest()[:16]

    if norm["hdfs"]["ha"]["enabled"]:
        zk_hl = headless_service(
            cluster=cluster,
            name=f"{cluster}-zookeeper",
            namespace=namespace,
            cr_api_version=cr_api_version,
            cr_kind=cr_kind,
            cr_name=cr_name,
            cr_uid=cr_uid,
            component="zookeeper",
            ports=[
                client.V1ServicePort(name="client", port=2181, target_port=2181),
                client.V1ServicePort(name="peer", port=2888, target_port=2888),
                client.V1ServicePort(name="election", port=3888, target_port=3888),
            ],
        )
        ensure_service(core, namespace, zk_hl)
        zk_sts = zookeeper_statefulset(
            cluster=cluster,
            namespace=namespace,
            cr_api_version=cr_api_version,
            cr_kind=cr_kind,
            cr_name=cr_name,
            cr_uid=cr_uid,
            norm=norm,
            cm_name=zk_cm_name,
            config_revision=config_revision,
        )
        replace_or_create_statefulset(apps, namespace, zk_sts)

        jn_hl = headless_service(
            cluster=cluster,
            name=f"{cluster}-journalnode",
            namespace=namespace,
            cr_api_version=cr_api_version,
            cr_kind=cr_kind,
            cr_name=cr_name,
            cr_uid=cr_uid,
            component="journalnode",
            ports=[
                client.V1ServicePort(name="rpc", port=8485, target_port=8485),
                client.V1ServicePort(name="web", port=8480, target_port=8480),
            ],
        )
        ensure_service(core, namespace, jn_hl)
        jn_sts = journalnode_statefulset(
            cluster=cluster,
            namespace=namespace,
            cr_api_version=cr_api_version,
            cr_kind=cr_kind,
            cr_name=cr_name,
            cr_uid=cr_uid,
            norm=norm,
            cm_name=cm_name,
            config_revision=config_revision,
        )
        replace_or_create_statefulset(apps, namespace, jn_sts)

    nn_hl = headless_service(
        cluster=cluster,
        name=f"{cluster}-namenode",
        namespace=namespace,
        cr_api_version=cr_api_version,
        cr_kind=cr_kind,
        cr_name=cr_name,
        cr_uid=cr_uid,
        component="namenode",
        ports=[
            client.V1ServicePort(name="web", port=9870, target_port=9870),
            client.V1ServicePort(name="rpc", port=9000, target_port=9000),
        ],
    )
    ensure_service(core, namespace, nn_hl)

    if norm["expose"]["namenodeWeb"] == "NodePort":
        if norm["hdfs"]["ha"]["enabled"]:
            delete_service(core, namespace, f"{cluster}-namenode-external")
            for i in range(norm["hdfs"]["ha"]["namenodeReplicas"]):
                pod_name = f"{cluster}-namenode-{i}"
                np = nodeport_service_for_pod(
                    cluster=cluster,
                    name=f"{cluster}-namenode-{i}-external",
                    namespace=namespace,
                    cr_api_version=cr_api_version,
                    cr_kind=cr_kind,
                    cr_name=cr_name,
                    cr_uid=cr_uid,
                    pod_name=pod_name,
                    component="namenode",
                    ports=[
                        client.V1ServicePort(name="web", port=9870, target_port=9870),
                        client.V1ServicePort(name="rpc", port=9000, target_port=9000),
                    ],
                )
                ensure_service(core, namespace, np)
        else:
            for i in (0, 1):
                delete_service(core, namespace, f"{cluster}-namenode-{i}-external")
            nn_np = optional_nodeport_service(
                cluster=cluster,
                name=f"{cluster}-namenode-external",
                namespace=namespace,
                cr_api_version=cr_api_version,
                cr_kind=cr_kind,
                cr_name=cr_name,
                cr_uid=cr_uid,
                component="namenode",
                svc_type="NodePort",
                ports=[
                    client.V1ServicePort(name="web", port=9870, target_port=9870),
                    client.V1ServicePort(name="rpc", port=9000, target_port=9000),
                ],
            )
            if nn_np is not None:
                ensure_service(core, namespace, nn_np)
    else:
        delete_service(core, namespace, f"{cluster}-namenode-external")
        for i in (0, 1):
            delete_service(core, namespace, f"{cluster}-namenode-{i}-external")

    if norm["expose"]["ingress"]["enabled"]:
        if norm["hdfs"]["ha"]["enabled"]:
            for i in range(norm["hdfs"]["ha"]["namenodeReplicas"]):
                pod_name = f"{cluster}-namenode-{i}"
                wsvc = clusterip_service_for_pod(
                    cluster=cluster,
                    name=f"{cluster}-namenode-{i}-web",
                    namespace=namespace,
                    cr_api_version=cr_api_version,
                    cr_kind=cr_kind,
                    cr_name=cr_name,
                    cr_uid=cr_uid,
                    pod_name=pod_name,
                    component="namenode",
                    ports=[client.V1ServicePort(name="web", port=9870, target_port=9870)],
                )
                ensure_service(core, namespace, wsvc)
        else:
            delete_service(core, namespace, f"{cluster}-namenode-1-web")
            wsvc = clusterip_service_for_pod(
                cluster=cluster,
                name=f"{cluster}-namenode-0-web",
                namespace=namespace,
                cr_api_version=cr_api_version,
                cr_kind=cr_kind,
                cr_name=cr_name,
                cr_uid=cr_uid,
                pod_name=f"{cluster}-namenode-0",
                component="namenode",
                ports=[client.V1ServicePort(name="web", port=9870, target_port=9870)],
            )
            ensure_service(core, namespace, wsvc)
        if norm["yarn"]["enabled"] and norm["expose"]["ingress"]["resourcemanagerHost"]:
            rsvc = clusterip_service_for_pod(
                cluster=cluster,
                name=f"{cluster}-resourcemanager-0-web",
                namespace=namespace,
                cr_api_version=cr_api_version,
                cr_kind=cr_kind,
                cr_name=cr_name,
                cr_uid=cr_uid,
                pod_name=f"{cluster}-resourcemanager-0",
                component="resourcemanager",
                ports=[client.V1ServicePort(name="web", port=8088, target_port=8088)],
            )
            ensure_service(core, namespace, rsvc)
        else:
            delete_service(core, namespace, f"{cluster}-resourcemanager-0-web")
        ing = build_hadoop_ingress(
            cluster=cluster,
            namespace=namespace,
            cr_api_version=cr_api_version,
            cr_kind=cr_kind,
            cr_name=cr_name,
            cr_uid=cr_uid,
            norm=norm,
        )
        if ing is not None:
            replace_or_create_ingress(net, namespace, ing)
        else:
            delete_ingress(net, namespace, f"{cluster}-hadoop-ui")
    else:
        _delete_ingress_ui(core, net, namespace, cluster)

    if norm["expose"]["datanodeWeb"] == "NodePort":
        dn_np = optional_nodeport_service(
            cluster=cluster,
            name=f"{cluster}-datanode-external",
            namespace=namespace,
            cr_api_version=cr_api_version,
            cr_kind=cr_kind,
            cr_name=cr_name,
            cr_uid=cr_uid,
            component="datanode",
            svc_type="NodePort",
            ports=[
                client.V1ServicePort(name="web", port=9864, target_port=9864),
                client.V1ServicePort(name="data", port=9866, target_port=9866),
            ],
        )
        if dn_np is not None:
            ensure_service(core, namespace, dn_np)
    else:
        delete_service(core, namespace, f"{cluster}-datanode-external")

    dn_hl = headless_service(
        cluster=cluster,
        name=f"{cluster}-datanode",
        namespace=namespace,
        cr_api_version=cr_api_version,
        cr_kind=cr_kind,
        cr_name=cr_name,
        cr_uid=cr_uid,
        component="datanode",
        ports=[
            client.V1ServicePort(name="web", port=9864, target_port=9864),
            client.V1ServicePort(name="data", port=9866, target_port=9866),
        ],
    )
    ensure_service(core, namespace, dn_hl)

    if norm["hdfs"]["ha"]["enabled"]:
        nn_sts = namenode_ha_statefulset(
            cluster=cluster,
            namespace=namespace,
            cr_api_version=cr_api_version,
            cr_kind=cr_kind,
            cr_name=cr_name,
            cr_uid=cr_uid,
            norm=norm,
            cm_name=cm_name,
            config_revision=config_revision,
        )
    else:
        nn_sts = namenode_statefulset(
            cluster=cluster,
            namespace=namespace,
            cr_api_version=cr_api_version,
            cr_kind=cr_kind,
            cr_name=cr_name,
            cr_uid=cr_uid,
            norm=norm,
            cm_name=cm_name,
            config_revision=config_revision,
        )
    replace_or_create_statefulset(apps, namespace, nn_sts)

    dn_sts = datanode_statefulset(
        cluster=cluster,
        namespace=namespace,
        cr_api_version=cr_api_version,
        cr_kind=cr_kind,
        cr_name=cr_name,
        cr_uid=cr_uid,
        norm=norm,
        cm_name=cm_name,
        config_revision=config_revision,
    )
    replace_or_create_statefulset(apps, namespace, dn_sts)

    pdb = namenode_pdb(
        cluster=cluster,
        namespace=namespace,
        cr_api_version=cr_api_version,
        cr_kind=cr_kind,
        cr_name=cr_name,
        cr_uid=cr_uid,
    )
    replace_or_create_pdb(policy, namespace, pdb)

    if norm["yarn"]["enabled"]:
        rm_hl = headless_service(
            cluster=cluster,
            name=f"{cluster}-resourcemanager",
            namespace=namespace,
            cr_api_version=cr_api_version,
            cr_kind=cr_kind,
            cr_name=cr_name,
            cr_uid=cr_uid,
            component="resourcemanager",
            ports=[
                client.V1ServicePort(name="web", port=8088, target_port=8088),
                client.V1ServicePort(name="rpc", port=8032, target_port=8032),
            ],
        )
        ensure_service(core, namespace, rm_hl)

        if norm["expose"]["resourcemanagerWeb"] == "NodePort":
            rm_np = optional_nodeport_service(
                cluster=cluster,
                name=f"{cluster}-resourcemanager-external",
                namespace=namespace,
                cr_api_version=cr_api_version,
                cr_kind=cr_kind,
                cr_name=cr_name,
                cr_uid=cr_uid,
                component="resourcemanager",
                svc_type="NodePort",
                ports=[
                    client.V1ServicePort(name="web", port=8088, target_port=8088),
                    client.V1ServicePort(name="rpc", port=8032, target_port=8032),
                ],
            )
            if rm_np is not None:
                ensure_service(core, namespace, rm_np)
        else:
            delete_service(core, namespace, f"{cluster}-resourcemanager-external")

        nm_hl = headless_service(
            cluster=cluster,
            name=f"{cluster}-nodemanager",
            namespace=namespace,
            cr_api_version=cr_api_version,
            cr_kind=cr_kind,
            cr_name=cr_name,
            cr_uid=cr_uid,
            component="nodemanager",
            ports=[client.V1ServicePort(name="web", port=8042, target_port=8042)],
        )
        ensure_service(core, namespace, nm_hl)

        if norm["expose"]["nodemanagerWeb"] == "NodePort":
            nm_np = optional_nodeport_service(
                cluster=cluster,
                name=f"{cluster}-nodemanager-external",
                namespace=namespace,
                cr_api_version=cr_api_version,
                cr_kind=cr_kind,
                cr_name=cr_name,
                cr_uid=cr_uid,
                component="nodemanager",
                svc_type="NodePort",
                ports=[client.V1ServicePort(name="web", port=8042, target_port=8042)],
            )
            if nm_np is not None:
                ensure_service(core, namespace, nm_np)
        else:
            delete_service(core, namespace, f"{cluster}-nodemanager-external")

        rm_sts = resourcemanager_statefulset(
            cluster=cluster,
            namespace=namespace,
            cr_api_version=cr_api_version,
            cr_kind=cr_kind,
            cr_name=cr_name,
            cr_uid=cr_uid,
            norm=norm,
            cm_name=cm_name,
            config_revision=config_revision,
        )
        replace_or_create_statefulset(apps, namespace, rm_sts)

        nm_sts = nodemanager_statefulset(
            cluster=cluster,
            namespace=namespace,
            cr_api_version=cr_api_version,
            cr_kind=cr_kind,
            cr_name=cr_name,
            cr_uid=cr_uid,
            norm=norm,
            cm_name=cm_name,
            config_revision=config_revision,
        )
        replace_or_create_statefulset(apps, namespace, nm_sts)
    else:
        for svc in (
            f"{cluster}-resourcemanager-external",
            f"{cluster}-resourcemanager",
            f"{cluster}-nodemanager-external",
            f"{cluster}-nodemanager",
        ):
            delete_service(core, namespace, svc)
        for sts in (f"{cluster}-resourcemanager", f"{cluster}-nodemanager"):
            delete_statefulset(apps, namespace, sts)


@kopf.on.create(GROUP, VERSION, PLURAL)
@kopf.on.update(GROUP, VERSION, PLURAL)
def reconcile_hadoop(body: Dict[str, Any], **kwargs: Any) -> None:
    try:
        _reconcile(body, **kwargs)
    except ApiException as e:
        raise kopf.TemporaryError(f"Kubernetes API error: {e.reason}", delay=30) from e


@kopf.on.delete(GROUP, VERSION, PLURAL)
def delete_hadoop(**_: Any) -> None:
    """OwnerReferences cascade-delete workload; no extra steps required."""
    pass
