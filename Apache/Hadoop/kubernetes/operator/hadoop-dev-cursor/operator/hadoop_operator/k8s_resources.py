"""Build Kubernetes API objects for a HadoopCluster."""

from __future__ import annotations

from typing import Any, Dict, List, Optional

from kubernetes import client

from hadoop_operator.config_xml import (
    build_capacity_scheduler,
    build_core_site,
    build_core_site_ha,
    build_hdfs_site,
    build_hdfs_site_ha,
    build_mapred_site,
    build_yarn_site,
    build_zookeeper_zoo_cfg,
)
from hadoop_operator.constants import DEFAULT_IMAGE, DEFAULT_PULL_POLICY


def nameservice_id(cluster: str) -> str:
    base = "".join(c for c in cluster if c.isalnum())
    return f"{base}ns" if base else "hadoopns"


def qjournal_uri(cluster: str, jn_replicas: int, nameservice: str) -> str:
    svc = f"{cluster}-journalnode"
    hosts = [f"{cluster}-journalnode-{i}.{svc}:8485" for i in range(jn_replicas)]
    return f"qjournal://{';'.join(hosts)}/{nameservice}"


def zk_connect_string(cluster: str, zk_replicas: int) -> str:
    svc = f"{cluster}-zookeeper"
    return ",".join(f"{cluster}-zookeeper-{i}.{svc}:2181" for i in range(zk_replicas))


def _labels(cluster: str, component: str) -> Dict[str, str]:
    return {
        "app.kubernetes.io/name": "hadoop",
        "app.kubernetes.io/instance": cluster,
        "app.kubernetes.io/managed-by": "hadoop-operator",
        "app.kubernetes.io/component": component,
        "data.hadoop.org/cluster": cluster,
    }


def owner_refs(api_version: str, kind: str, name: str, uid: str) -> List[client.V1OwnerReference]:
    return [
        client.V1OwnerReference(
            api_version=api_version,
            kind=kind,
            name=name,
            uid=uid,
            controller=True,
            block_owner_deletion=True,
        )
    ]


def meta(
    *,
    name: str,
    namespace: str,
    cr_name: str,
    cr_uid: str,
    cr_api_version: str,
    cr_kind: str,
    cluster: str,
    component: str,
) -> client.V1ObjectMeta:
    return client.V1ObjectMeta(
        name=name,
        namespace=namespace,
        labels=_labels(cluster, component),
        owner_references=owner_refs(cr_api_version, cr_kind, cr_name, cr_uid),
    )


def normalize_spec(spec: Optional[Dict[str, Any]]) -> Dict[str, Any]:
    s = spec or {}
    hdfs = s.get("hdfs") or {}
    ha = hdfs.get("ha") or {}
    yarn = s.get("yarn") or {}
    expose = s.get("expose") or {}
    ingress = expose.get("ingress") or {}
    resources = s.get("resources") or {}
    zk_in = s.get("zookeeper") or {}

    dn_rep = int(hdfs.get("datanodeReplicas", 1))
    replication = int(hdfs.get("replication", min(3, max(1, dn_rep))))

    def _expose_mode(v: Any) -> str:
        return "NodePort" if v == "NodePort" else "ClusterIP"

    ha_enabled = bool(ha.get("enabled", False))
    jn_rep = int(ha.get("journalnodeReplicas", 3))
    nn_ha_rep = 2 if ha_enabled else 1
    if ha_enabled:
        jn_rep = max(3, jn_rep)
        if jn_rep % 2 == 0:
            jn_rep += 1

    zk_rep = int(zk_in.get("replicas", 3))
    if ha_enabled:
        zk_rep = max(3, zk_rep)
        if zk_rep % 2 == 0:
            zk_rep += 1

    nameservice = ha.get("nameservice") or None

    return {
        "image": s.get("image", DEFAULT_IMAGE),
        "imagePullPolicy": s.get("imagePullPolicy", DEFAULT_PULL_POLICY),
        "imagePullSecrets": s.get("imagePullSecrets") or [],
        "hdfs": {
            "namenodeStorageClass": hdfs.get("namenodeStorageClass", "standard"),
            "namenodeStorageSize": hdfs.get("namenodeStorageSize", "20Gi"),
            "datanodeReplicas": dn_rep,
            "datanodeStorageClass": hdfs.get("datanodeStorageClass", "standard"),
            "datanodeStorageSize": hdfs.get("datanodeStorageSize", "20Gi"),
            "replication": replication,
            "ha": {
                "enabled": ha_enabled,
                "nameservice": nameservice,
                "journalnodeReplicas": jn_rep,
                "namenodeReplicas": nn_ha_rep,
            },
        },
        "zookeeper": {
            "image": zk_in.get("image", "zookeeper:3.8.4"),
            "replicas": zk_rep,
            "storageClass": zk_in.get("storageClass") or hdfs.get("namenodeStorageClass", "standard"),
            "storageSize": zk_in.get("storageSize", "8Gi"),
        },
        "yarn": {
            "enabled": bool(yarn.get("enabled", True)),
            "nodemanagerReplicas": int(yarn.get("nodemanagerReplicas", 1)),
        },
        "expose": {
            "namenodeWeb": _expose_mode(expose.get("namenodeWeb", "ClusterIP")),
            "datanodeWeb": _expose_mode(expose.get("datanodeWeb", "ClusterIP")),
            "resourcemanagerWeb": _expose_mode(expose.get("resourcemanagerWeb", "ClusterIP")),
            "nodemanagerWeb": _expose_mode(expose.get("nodemanagerWeb", "ClusterIP")),
            "ingress": {
                "enabled": bool(ingress.get("enabled", False)),
                "className": ingress.get("className", "nginx"),
                "tlsSecretName": ingress.get("tlsSecretName") or "",
                "namenodeNn1Host": ingress.get("namenodeNn1Host") or "",
                "namenodeNn2Host": ingress.get("namenodeNn2Host") or "",
                "namenodeSingleHost": ingress.get("namenodeSingleHost") or "",
                "resourcemanagerHost": ingress.get("resourcemanagerHost") or "",
            },
        },
        "resources": {
            "namenode": resources.get("namenode") or _default_nn_res(),
            "datanode": resources.get("datanode") or _default_dn_res(),
            "resourcemanager": resources.get("resourcemanager") or _default_rm_res(),
            "nodemanager": resources.get("nodemanager") or _default_nm_res(),
            "journalnode": resources.get("journalnode") or _default_jn_res(),
            "zookeeper": resources.get("zookeeper") or _default_zk_res(),
            "zkfc": resources.get("zkfc") or _default_zkfc_res(),
        },
    }


def _default_nn_res() -> Dict[str, Any]:
    return {
        "requests": {"cpu": "500m", "memory": "2Gi"},
        "limits": {"cpu": "2000m", "memory": "4Gi"},
    }


def _default_dn_res() -> Dict[str, Any]:
    return {
        "requests": {"cpu": "500m", "memory": "2Gi"},
        "limits": {"cpu": "2000m", "memory": "4Gi"},
    }


def _default_rm_res() -> Dict[str, Any]:
    return {
        "requests": {"cpu": "250m", "memory": "1Gi"},
        "limits": {"cpu": "1000m", "memory": "2Gi"},
    }


def _default_nm_res() -> Dict[str, Any]:
    return {
        "requests": {"cpu": "250m", "memory": "1Gi"},
        "limits": {"cpu": "1000m", "memory": "2Gi"},
    }


def _default_jn_res() -> Dict[str, Any]:
    return {
        "requests": {"cpu": "250m", "memory": "512Mi"},
        "limits": {"cpu": "1000m", "memory": "1Gi"},
    }


def _default_zk_res() -> Dict[str, Any]:
    return {
        "requests": {"cpu": "100m", "memory": "256Mi"},
        "limits": {"cpu": "500m", "memory": "512Mi"},
    }


def _default_zkfc_res() -> Dict[str, Any]:
    return {
        "requests": {"cpu": "100m", "memory": "256Mi"},
        "limits": {"cpu": "500m", "memory": "512Mi"},
    }


def _res_requirements(block: Dict[str, Any]) -> client.V1ResourceRequirements:
    return client.V1ResourceRequirements(
        requests=block.get("requests"),
        limits=block.get("limits"),
    )


def build_configmap(
    *,
    cluster: str,
    namespace: str,
    cr_api_version: str,
    cr_kind: str,
    cr_name: str,
    cr_uid: str,
    norm: Dict[str, Any],
) -> client.V1ConfigMap:
    rm_svc = f"{cluster}-resourcemanager"
    rm_host = f"{cluster}-resourcemanager-0.{rm_svc}"

    if norm["hdfs"]["ha"]["enabled"]:
        ns_id = norm["hdfs"]["ha"]["nameservice"] or nameservice_id(cluster)
        nn_svc = f"{cluster}-namenode"
        nn1 = f"{cluster}-namenode-0.{nn_svc}"
        nn2 = f"{cluster}-namenode-1.{nn_svc}"
        fs_default = f"hdfs://{ns_id}"
        zk_q = zk_connect_string(cluster, norm["zookeeper"]["replicas"])
        qj = qjournal_uri(cluster, norm["hdfs"]["ha"]["journalnodeReplicas"], ns_id)
        hdfs_xml = build_hdfs_site_ha(
            nameservice=ns_id,
            nn1_rpc=f"{nn1}:9000",
            nn2_rpc=f"{nn2}:9000",
            nn1_http=f"{nn1}:9870",
            nn2_http=f"{nn2}:9870",
            qjournal_uri=qj,
            replication=norm["hdfs"]["replication"],
        )
        core_xml = build_core_site_ha(fs_default, zk_q)
    else:
        nn_svc = f"{cluster}-namenode"
        nn_pod = f"{cluster}-namenode-0"
        nn_host = f"{nn_pod}.{nn_svc}"
        fs_default = f"hdfs://{nn_host}:9000"
        namenode_rpc = f"{nn_host}:9000"
        core_xml = build_core_site(fs_default)
        hdfs_xml = build_hdfs_site(
            namenode_rpc,
            norm["hdfs"]["replication"],
        )

    data: Dict[str, str] = {
        "core-site.xml": core_xml,
        "hdfs-site.xml": hdfs_xml,
    }
    if norm["yarn"]["enabled"]:
        data["yarn-site.xml"] = build_yarn_site(rm_host)
        data["mapred-site.xml"] = build_mapred_site()
        data["capacity-scheduler.xml"] = build_capacity_scheduler()

    return client.V1ConfigMap(
        api_version="v1",
        kind="ConfigMap",
        metadata=meta(
            name=f"{cluster}-hadoop-config",
            namespace=namespace,
            cr_name=cr_name,
            cr_uid=cr_uid,
            cr_api_version=cr_api_version,
            cr_kind=cr_kind,
            cluster=cluster,
            component="config",
        ),
        data=data,
    )


def _pod_security() -> client.V1PodSecurityContext:
    return client.V1PodSecurityContext(
        fs_group=1000,
        seccomp_profile=client.V1SeccompProfile(type="RuntimeDefault"),
    )


def _container_security() -> client.V1SecurityContext:
    return client.V1SecurityContext(
        allow_privilege_escalation=False,
        read_only_root_filesystem=False,
    )


def _config_volumes(cm_name: str) -> tuple[List[client.V1Volume], List[client.V1VolumeMount]]:
    vol = client.V1Volume(
        name="hadoop-config-volume",
        config_map=client.V1ConfigMapVolumeSource(name=cm_name),
    )
    mounts = [
        client.V1VolumeMount(
            name="hadoop-config-volume",
            mount_path="/opt/hadoop/etc/hadoop/core-site.xml",
            sub_path="core-site.xml",
        ),
        client.V1VolumeMount(
            name="hadoop-config-volume",
            mount_path="/opt/hadoop/etc/hadoop/hdfs-site.xml",
            sub_path="hdfs-site.xml",
        ),
    ]
    return [vol], mounts


def _yarn_config_mounts() -> List[client.V1VolumeMount]:
    return [
        client.V1VolumeMount(
            name="hadoop-config-volume",
            mount_path="/opt/hadoop/etc/hadoop/yarn-site.xml",
            sub_path="yarn-site.xml",
        ),
        client.V1VolumeMount(
            name="hadoop-config-volume",
            mount_path="/opt/hadoop/etc/hadoop/mapred-site.xml",
            sub_path="mapred-site.xml",
        ),
        client.V1VolumeMount(
            name="hadoop-config-volume",
            mount_path="/opt/hadoop/etc/hadoop/capacity-scheduler.xml",
            sub_path="capacity-scheduler.xml",
        ),
    ]


def headless_service(
    *,
    cluster: str,
    name: str,
    namespace: str,
    cr_api_version: str,
    cr_kind: str,
    cr_name: str,
    cr_uid: str,
    component: str,
    ports: List[client.V1ServicePort],
) -> client.V1Service:
    return client.V1Service(
        api_version="v1",
        kind="Service",
        metadata=meta(
            name=name,
            namespace=namespace,
            cr_name=cr_name,
            cr_uid=cr_uid,
            cr_api_version=cr_api_version,
            cr_kind=cr_kind,
            cluster=cluster,
            component=component,
        ),
        spec=client.V1ServiceSpec(
            cluster_ip="None",
            publish_not_ready_addresses=True,
            selector=_labels(cluster, component),
            ports=ports,
        ),
    )


def optional_nodeport_service(
    *,
    cluster: str,
    name: str,
    namespace: str,
    cr_api_version: str,
    cr_kind: str,
    cr_name: str,
    cr_uid: str,
    component: str,
    svc_type: str,
    ports: List[client.V1ServicePort],
) -> Optional[client.V1Service]:
    if svc_type != "NodePort":
        return None
    return client.V1Service(
        api_version="v1",
        kind="Service",
        metadata=meta(
            name=name,
            namespace=namespace,
            cr_name=cr_name,
            cr_uid=cr_uid,
            cr_api_version=cr_api_version,
            cr_kind=cr_kind,
            cluster=cluster,
            component=component,
        ),
        spec=client.V1ServiceSpec(
            type="NodePort",
            selector=_labels(cluster, component),
            ports=ports,
        ),
    )


def clusterip_service_for_pod(
    *,
    cluster: str,
    name: str,
    namespace: str,
    cr_api_version: str,
    cr_kind: str,
    cr_name: str,
    cr_uid: str,
    pod_name: str,
    component: str,
    ports: List[client.V1ServicePort],
) -> client.V1Service:
    return client.V1Service(
        api_version="v1",
        kind="Service",
        metadata=meta(
            name=name,
            namespace=namespace,
            cr_name=cr_name,
            cr_uid=cr_uid,
            cr_api_version=cr_api_version,
            cr_kind=cr_kind,
            cluster=cluster,
            component=component,
        ),
        spec=client.V1ServiceSpec(
            type="ClusterIP",
            selector={"statefulset.kubernetes.io/pod-name": pod_name},
            ports=ports,
        ),
    )


def nodeport_service_for_pod(
    *,
    cluster: str,
    name: str,
    namespace: str,
    cr_api_version: str,
    cr_kind: str,
    cr_name: str,
    cr_uid: str,
    pod_name: str,
    component: str,
    ports: List[client.V1ServicePort],
) -> client.V1Service:
    return client.V1Service(
        api_version="v1",
        kind="Service",
        metadata=meta(
            name=name,
            namespace=namespace,
            cr_name=cr_name,
            cr_uid=cr_uid,
            cr_api_version=cr_api_version,
            cr_kind=cr_kind,
            cluster=cluster,
            component=component,
        ),
        spec=client.V1ServiceSpec(
            type="NodePort",
            selector={"statefulset.kubernetes.io/pod-name": pod_name},
            ports=ports,
        ),
    )


def build_zookeeper_configmap(
    *,
    cluster: str,
    namespace: str,
    cr_api_version: str,
    cr_kind: str,
    cr_name: str,
    cr_uid: str,
    norm: Dict[str, Any],
) -> client.V1ConfigMap:
    zkc = build_zookeeper_zoo_cfg(cluster=cluster, replicas=norm["zookeeper"]["replicas"])
    return client.V1ConfigMap(
        api_version="v1",
        kind="ConfigMap",
        metadata=meta(
            name=f"{cluster}-zookeeper-config",
            namespace=namespace,
            cr_name=cr_name,
            cr_uid=cr_uid,
            cr_api_version=cr_api_version,
            cr_kind=cr_kind,
            cluster=cluster,
            component="zookeeper-config",
        ),
        data={"zoo.cfg": zkc},
    )


def zookeeper_statefulset(
    *,
    cluster: str,
    namespace: str,
    cr_api_version: str,
    cr_kind: str,
    cr_name: str,
    cr_uid: str,
    norm: Dict[str, Any],
    cm_name: str,
    config_revision: str,
) -> client.V1StatefulSet:
    zkm = client.V1VolumeMount(
        name="zoo-cfg",
        mount_path="/conf/zoo.cfg",
        sub_path="zoo.cfg",
    )
    main = client.V1Container(
        name="zookeeper",
        image=norm["zookeeper"]["image"],
        image_pull_policy=norm["imagePullPolicy"],
        command=["/bin/bash", "-c"],
        args=[
            """
set -e
mkdir -p /data
ORD=${HOSTNAME##*-}
echo $((ORD + 1)) > /data/myid
ZK_HOME=""
for d in /apache-zookeeper-* /opt/apache-zookeeper-*; do
  if [ -x "$d/bin/zkServer.sh" ]; then ZK_HOME="$d"; break; fi
done
if [ -z "$ZK_HOME" ]; then
  echo "Could not find zkServer.sh"; exit 1
fi
exec "$ZK_HOME/bin/zkServer.sh" start-foreground /conf/zoo.cfg
""".strip()
        ],
        ports=[
            client.V1ContainerPort(container_port=2181, name="client"),
            client.V1ContainerPort(container_port=2888, name="peer"),
            client.V1ContainerPort(container_port=3888, name="election"),
        ],
        resources=_res_requirements(norm["resources"]["zookeeper"]),
        volume_mounts=[zkm],
    )
    return client.V1StatefulSet(
        api_version="apps/v1",
        kind="StatefulSet",
        metadata=meta(
            name=f"{cluster}-zookeeper",
            namespace=namespace,
            cr_name=cr_name,
            cr_uid=cr_uid,
            cr_api_version=cr_api_version,
            cr_kind=cr_kind,
            cluster=cluster,
            component="zookeeper",
        ),
        spec=client.V1StatefulSetSpec(
            service_name=f"{cluster}-zookeeper",
            replicas=norm["zookeeper"]["replicas"],
            selector=client.V1LabelSelector(match_labels=_labels(cluster, "zookeeper")),
            template=client.V1PodTemplateSpec(
                metadata=client.V1ObjectMeta(
                    labels=_labels(cluster, "zookeeper"),
                    annotations={"data.hadoop.org/config-revision": config_revision},
                ),
                spec=client.V1PodSpec(
                    security_context=_pod_security(),
                    image_pull_secrets=[
                        client.V1LocalObjectReference(name=s["name"])
                        for s in norm["imagePullSecrets"]
                        if s.get("name")
                    ],
                    containers=[main],
                    volumes=[
                        client.V1Volume(
                            name="zoo-cfg",
                            config_map=client.V1ConfigMapVolumeSource(name=cm_name),
                        ),
                    ],
                ),
            ),
            volume_claim_templates=[
                client.V1PersistentVolumeClaim(
                    metadata=client.V1ObjectMeta(name="zookeeper-data"),
                    spec=client.V1PersistentVolumeClaimSpec(
                        access_modes=["ReadWriteOnce"],
                        storage_class_name=norm["zookeeper"]["storageClass"],
                        resources=client.V1ResourceRequirements(
                            requests={"storage": norm["zookeeper"]["storageSize"]}
                        ),
                    ),
                )
            ],
        ),
    )


def journalnode_statefulset(
    *,
    cluster: str,
    namespace: str,
    cr_api_version: str,
    cr_kind: str,
    cr_name: str,
    cr_uid: str,
    norm: Dict[str, Any],
    cm_name: str,
    config_revision: str,
) -> client.V1StatefulSet:
    vols, cm_mounts = _config_volumes(cm_name)
    init = client.V1Container(
        name="init-jn",
        image=norm["image"],
        image_pull_policy=norm["imagePullPolicy"],
        security_context=client.V1SecurityContext(run_as_user=0, privileged=True),
        command=["/bin/bash", "-c"],
        args=[
            """
set -e
mkdir -p /opt/hadoop/data/jn
chown -R hadoop:hadoop /opt/hadoop/data || chown -R 1000:1000 /opt/hadoop/data
chmod -R 775 /opt/hadoop/data
""".strip()
        ],
        volume_mounts=[
            client.V1VolumeMount(
                name="hadoop-jn-data",
                mount_path="/opt/hadoop/data/jn",
                sub_path="jn",
            ),
        ],
    )
    main = client.V1Container(
        name="journalnode",
        image=norm["image"],
        image_pull_policy=norm["imagePullPolicy"],
        security_context=_container_security(),
        command=["/bin/bash", "-c"],
        args=["exec hdfs journalnode"],
        env=[
            client.V1EnvVar(name="HADOOP_CONF_DIR", value="/opt/hadoop/etc/hadoop"),
        ],
        ports=[
            client.V1ContainerPort(container_port=8485, name="rpc"),
            client.V1ContainerPort(container_port=8480, name="web"),
        ],
        resources=_res_requirements(norm["resources"]["journalnode"]),
        volume_mounts=[
            client.V1VolumeMount(
                name="hadoop-jn-data",
                mount_path="/opt/hadoop/data/jn",
                sub_path="jn",
            ),
            *cm_mounts,
        ],
    )
    return client.V1StatefulSet(
        api_version="apps/v1",
        kind="StatefulSet",
        metadata=meta(
            name=f"{cluster}-journalnode",
            namespace=namespace,
            cr_name=cr_name,
            cr_uid=cr_uid,
            cr_api_version=cr_api_version,
            cr_kind=cr_kind,
            cluster=cluster,
            component="journalnode",
        ),
        spec=client.V1StatefulSetSpec(
            service_name=f"{cluster}-journalnode",
            replicas=norm["hdfs"]["ha"]["journalnodeReplicas"],
            selector=client.V1LabelSelector(match_labels=_labels(cluster, "journalnode")),
            template=client.V1PodTemplateSpec(
                metadata=client.V1ObjectMeta(
                    labels=_labels(cluster, "journalnode"),
                    annotations={"data.hadoop.org/config-revision": config_revision},
                ),
                spec=client.V1PodSpec(
                    security_context=_pod_security(),
                    image_pull_secrets=[
                        client.V1LocalObjectReference(name=s["name"])
                        for s in norm["imagePullSecrets"]
                        if s.get("name")
                    ],
                    init_containers=[init],
                    containers=[main],
                    volumes=vols,
                ),
            ),
            volume_claim_templates=[
                client.V1PersistentVolumeClaim(
                    metadata=client.V1ObjectMeta(name="hadoop-jn-data"),
                    spec=client.V1PersistentVolumeClaimSpec(
                        access_modes=["ReadWriteOnce"],
                        storage_class_name=norm["hdfs"]["namenodeStorageClass"],
                        resources=client.V1ResourceRequirements(
                            requests={"storage": "10Gi"},
                        ),
                    ),
                )
            ],
        ),
    )


def namenode_ha_statefulset(
    *,
    cluster: str,
    namespace: str,
    cr_api_version: str,
    cr_kind: str,
    cr_name: str,
    cr_uid: str,
    norm: Dict[str, Any],
    cm_name: str,
    config_revision: str,
) -> client.V1StatefulSet:
    vols, cm_mounts = _config_volumes(cm_name)
    logs = client.V1Volume(
        name="hadoop-logs",
        empty_dir=client.V1EmptyDirVolumeSource(),
    )
    vols.append(logs)

    zk_rep = norm["zookeeper"]["replicas"]
    jn_rep = norm["hdfs"]["ha"]["journalnodeReplicas"]
    zk_wait_lines = []
    for i in range(zk_rep):
        h = f"{cluster}-zookeeper-{i}.{cluster}-zookeeper"
        zk_wait_lines.append(
            f'until bash -c "echo > /dev/tcp/{h}/2181" 2>/dev/null; do echo waiting zk {i}; sleep 2; done'
        )
    jn_wait_lines = []
    for i in range(jn_rep):
        h = f"{cluster}-journalnode-{i}.{cluster}-journalnode"
        jn_wait_lines.append(
            f'until bash -c "echo > /dev/tcp/{h}/8485" 2>/dev/null; do echo waiting jn {i}; sleep 2; done'
        )
    nn_svc = f"{cluster}-namenode"
    nn0 = f"{cluster}-namenode-0.{nn_svc}"

    init_wait_zk_jn = client.V1Container(
        name="wait-zk-jn",
        image=norm["image"],
        image_pull_policy=norm["imagePullPolicy"],
        command=["/bin/bash", "-c"],
        args=["\n".join(["set -e", *zk_wait_lines, *jn_wait_lines, "echo ZK and JN ready"])],
    )

    ha_bootstrap = client.V1Container(
        name="ha-bootstrap",
        image=norm["image"],
        image_pull_policy=norm["imagePullPolicy"],
        security_context=client.V1SecurityContext(run_as_user=0, privileged=True),
        command=["/bin/bash", "-c"],
        args=[
            f"""
set -e
export HADOOP_CONF_DIR=/opt/hadoop/etc/hadoop
mkdir -p /opt/hadoop/data/nn
chown -R hadoop:hadoop /opt/hadoop/data || chown -R 1000:1000 /opt/hadoop/data
run_nn() {{ su -s /bin/bash hadoop -c "$1" || su -s /bin/bash -c "$1" nobody || bash -c "$1"; }}
ORD=${{HOSTNAME##*-}}
if [ "$ORD" = "0" ]; then
  if [ ! -f /opt/hadoop/data/nn/current/VERSION ]; then
    run_nn "hdfs namenode -format -nonInteractive"
    run_nn "hdfs namenode -initializeSharedEdits -nonInteractive -force"
  fi
else
  until bash -c "echo > /dev/tcp/{nn0}/9000" 2>/dev/null; do sleep 5; done
  if [ ! -f /opt/hadoop/data/nn/current/VERSION ]; then
    for i in {{1..30}}; do
      run_nn "hdfs namenode -bootstrapStandby -nonInteractive" && break
      sleep 15
    done
  fi
fi
""".strip()
        ],
        volume_mounts=[
            client.V1VolumeMount(
                name="hadoop-nn-data",
                mount_path="/opt/hadoop/data/nn",
                sub_path="nn",
            ),
            *cm_mounts,
        ],
    )

    nn_main = client.V1Container(
        name="namenode",
        image=norm["image"],
        image_pull_policy=norm["imagePullPolicy"],
        security_context=_container_security(),
        command=["/bin/bash", "-c"],
        args=[
            """
set -e
export HADOOP_CONF_DIR=/opt/hadoop/etc/hadoop
export HADOOP_LOG_DIR=/opt/hadoop/logs
export HADOOP_HEAPSIZE=1024
ORD=${HOSTNAME##*-}
if [ "$ORD" = "0" ]; then ID=nn1; else ID=nn2; fi
exec hdfs namenode -Ddfs.ha.namenode.id=$ID
""".strip()
        ],
        ports=[
            client.V1ContainerPort(container_port=9870, name="web"),
            client.V1ContainerPort(container_port=9000, name="rpc"),
        ],
        readiness_probe=client.V1Probe(
            http_get=client.V1HTTPGetAction(path="/jmx", port=9870),
            initial_delay_seconds=50,
            period_seconds=10,
        ),
        liveness_probe=client.V1Probe(
            http_get=client.V1HTTPGetAction(path="/jmx", port=9870),
            initial_delay_seconds=150,
            period_seconds=30,
        ),
        resources=_res_requirements(norm["resources"]["namenode"]),
        volume_mounts=[
            client.V1VolumeMount(
                name="hadoop-nn-data",
                mount_path="/opt/hadoop/data/nn",
                sub_path="nn",
            ),
            client.V1VolumeMount(name="hadoop-logs", mount_path="/opt/hadoop/logs"),
            *cm_mounts,
        ],
    )

    zkfc = client.V1Container(
        name="zkfc",
        image=norm["image"],
        image_pull_policy=norm["imagePullPolicy"],
        security_context=_container_security(),
        command=["/bin/bash", "-c"],
        args=[
            """
set -e
export HADOOP_CONF_DIR=/opt/hadoop/etc/hadoop
export HADOOP_LOG_DIR=/opt/hadoop/logs
ORD=${HOSTNAME##*-}
if [ "$ORD" = "0" ]; then ID=nn1; else ID=nn2; fi
for i in {1..120}; do
  bash -c "echo > /dev/tcp/127.0.0.1/9000" 2>/dev/null && break
  sleep 3
done
if [ "$ORD" = "0" ]; then
  hdfs zkfc -formatZK -force 2>/dev/null || true
fi
exec hdfs zkfc -Ddfs.ha.namenode.id=$ID
""".strip()
        ],
        resources=_res_requirements(norm["resources"]["zkfc"]),
        volume_mounts=[*cm_mounts],
    )

    return client.V1StatefulSet(
        api_version="apps/v1",
        kind="StatefulSet",
        metadata=meta(
            name=f"{cluster}-namenode",
            namespace=namespace,
            cr_name=cr_name,
            cr_uid=cr_uid,
            cr_api_version=cr_api_version,
            cr_kind=cr_kind,
            cluster=cluster,
            component="namenode",
        ),
        spec=client.V1StatefulSetSpec(
            service_name=f"{cluster}-namenode",
            replicas=norm["hdfs"]["ha"]["namenodeReplicas"],
            selector=client.V1LabelSelector(match_labels=_labels(cluster, "namenode")),
            template=client.V1PodTemplateSpec(
                metadata=client.V1ObjectMeta(
                    labels=_labels(cluster, "namenode"),
                    annotations={"data.hadoop.org/config-revision": config_revision},
                ),
                spec=client.V1PodSpec(
                    security_context=_pod_security(),
                    image_pull_secrets=[
                        client.V1LocalObjectReference(name=s["name"])
                        for s in norm["imagePullSecrets"]
                        if s.get("name")
                    ],
                    init_containers=[init_wait_zk_jn, ha_bootstrap],
                    containers=[nn_main, zkfc],
                    volumes=vols,
                    termination_grace_period_seconds=120,
                ),
            ),
            volume_claim_templates=[
                client.V1PersistentVolumeClaim(
                    metadata=client.V1ObjectMeta(name="hadoop-nn-data"),
                    spec=client.V1PersistentVolumeClaimSpec(
                        access_modes=["ReadWriteOnce"],
                        storage_class_name=norm["hdfs"]["namenodeStorageClass"],
                        resources=client.V1ResourceRequirements(
                            requests={"storage": norm["hdfs"]["namenodeStorageSize"]}
                        ),
                    ),
                )
            ],
        ),
    )


def build_hadoop_ingress(
    *,
    cluster: str,
    namespace: str,
    cr_api_version: str,
    cr_kind: str,
    cr_name: str,
    cr_uid: str,
    norm: Dict[str, Any],
) -> Optional[client.V1Ingress]:
    ing = norm["expose"]["ingress"]
    if not ing["enabled"]:
        return None
    rules: List[client.V1IngressRule] = []
    paths_nn = [
        client.V1HTTPIngressPath(
            path="/",
            path_type="Prefix",
            backend=client.V1IngressBackend(
                service=client.V1IngressServiceBackend(
                    name=f"{cluster}-namenode-0-web",
                    port=client.V1ServiceBackendPort(number=9870),
                ),
            ),
        ),
    ]
    if norm["hdfs"]["ha"]["enabled"]:
        if ing["namenodeNn1Host"]:
            rules.append(
                client.V1IngressRule(
                    host=ing["namenodeNn1Host"],
                    http=client.V1HTTPIngressRuleValue(paths=paths_nn),
                )
            )
        if ing["namenodeNn2Host"]:
            rules.append(
                client.V1IngressRule(
                    host=ing["namenodeNn2Host"],
                    http=client.V1HTTPIngressRuleValue(
                        paths=[
                            client.V1HTTPIngressPath(
                                path="/",
                                path_type="Prefix",
                                backend=client.V1IngressBackend(
                                    service=client.V1IngressServiceBackend(
                                        name=f"{cluster}-namenode-1-web",
                                        port=client.V1ServiceBackendPort(number=9870),
                                    ),
                                ),
                            ),
                        ],
                    ),
                )
            )
    else:
        if ing["namenodeSingleHost"]:
            rules.append(
                client.V1IngressRule(
                    host=ing["namenodeSingleHost"],
                    http=client.V1HTTPIngressRuleValue(
                        paths=[
                            client.V1HTTPIngressPath(
                                path="/",
                                path_type="Prefix",
                                backend=client.V1IngressBackend(
                                    service=client.V1IngressServiceBackend(
                                        name=f"{cluster}-namenode-0-web",
                                        port=client.V1ServiceBackendPort(number=9870),
                                    ),
                                ),
                            ),
                        ],
                    ),
                )
            )
    if norm["yarn"]["enabled"] and ing["resourcemanagerHost"]:
        rules.append(
            client.V1IngressRule(
                host=ing["resourcemanagerHost"],
                http=client.V1HTTPIngressRuleValue(
                    paths=[
                        client.V1HTTPIngressPath(
                            path="/",
                            path_type="Prefix",
                            backend=client.V1IngressBackend(
                                service=client.V1IngressServiceBackend(
                                    name=f"{cluster}-resourcemanager-0-web",
                                    port=client.V1ServiceBackendPort(number=8088),
                                ),
                            ),
                        ),
                    ],
                ),
            )
        )
    if not rules:
        return None
    tls_list: Optional[List[client.V1IngressTLS]] = None
    if ing["tlsSecretName"]:
        hosts = [r.host for r in rules if r.host]
        if hosts:
            tls_list = [
                client.V1IngressTLS(hosts=hosts, secret_name=ing["tlsSecretName"]),
            ]
    return client.V1Ingress(
        api_version="networking.k8s.io/v1",
        kind="Ingress",
        metadata=meta(
            name=f"{cluster}-hadoop-ui",
            namespace=namespace,
            cr_name=cr_name,
            cr_uid=cr_uid,
            cr_api_version=cr_api_version,
            cr_kind=cr_kind,
            cluster=cluster,
            component="ingress",
        ),
        spec=client.V1IngressSpec(
            ingress_class_name=ing["className"] or None,
            rules=rules,
            tls=tls_list,
        ),
    )


def namenode_statefulset(
    *,
    cluster: str,
    namespace: str,
    cr_api_version: str,
    cr_kind: str,
    cr_name: str,
    cr_uid: str,
    norm: Dict[str, Any],
    cm_name: str,
    config_revision: str,
) -> client.V1StatefulSet:
    vols, cm_mounts = _config_volumes(cm_name)
    logs = client.V1Volume(
        name="hadoop-logs",
        empty_dir=client.V1EmptyDirVolumeSource(),
    )
    vols.append(logs)

    init = client.V1Container(
        name="init-namenode",
        image=norm["image"],
        image_pull_policy=norm["imagePullPolicy"],
        security_context=client.V1SecurityContext(
            run_as_user=0,
            privileged=True,
        ),
        command=["/bin/bash", "-c"],
        args=[
            """
set -e
mkdir -p /opt/hadoop/data/nn
chown -R hadoop:hadoop /opt/hadoop/data || chown -R 1000:1000 /opt/hadoop/data
if [ ! -f /opt/hadoop/data/nn/current/VERSION ]; then
  su -s /bin/bash hadoop -c "hdfs namenode -format -nonInteractive" || \\
    su -s /bin/bash -c "hdfs namenode -format -nonInteractive" nobody || \\
    hdfs namenode -format -nonInteractive
else
  echo "NameNode already formatted."
fi
""".strip()
        ],
        volume_mounts=[
            client.V1VolumeMount(
                name="hadoop-nn-data",
                mount_path="/opt/hadoop/data/nn",
                sub_path="nn",
            ),
            *cm_mounts,
        ],
    )

    main = client.V1Container(
        name="namenode",
        image=norm["image"],
        image_pull_policy=norm["imagePullPolicy"],
        security_context=_container_security(),
        command=["/bin/bash", "-c"],
        args=["exec hdfs namenode"],
        env=[
            client.V1EnvVar(name="HADOOP_CONF_DIR", value="/opt/hadoop/etc/hadoop"),
            client.V1EnvVar(name="HADOOP_LOG_DIR", value="/opt/hadoop/logs"),
            client.V1EnvVar(name="HADOOP_HEAPSIZE", value="1024"),
        ],
        ports=[
            client.V1ContainerPort(container_port=9870, name="web"),
            client.V1ContainerPort(container_port=9000, name="rpc"),
        ],
        readiness_probe=client.V1Probe(
            http_get=client.V1HTTPGetAction(path="/jmx", port=9870),
            initial_delay_seconds=40,
            period_seconds=10,
        ),
        liveness_probe=client.V1Probe(
            http_get=client.V1HTTPGetAction(path="/jmx", port=9870),
            initial_delay_seconds=120,
            period_seconds=30,
        ),
        resources=_res_requirements(norm["resources"]["namenode"]),
        volume_mounts=[
            client.V1VolumeMount(
                name="hadoop-nn-data",
                mount_path="/opt/hadoop/data/nn",
                sub_path="nn",
            ),
            client.V1VolumeMount(name="hadoop-logs", mount_path="/opt/hadoop/logs"),
            *cm_mounts,
        ],
    )

    return client.V1StatefulSet(
        api_version="apps/v1",
        kind="StatefulSet",
        metadata=meta(
            name=f"{cluster}-namenode",
            namespace=namespace,
            cr_name=cr_name,
            cr_uid=cr_uid,
            cr_api_version=cr_api_version,
            cr_kind=cr_kind,
            cluster=cluster,
            component="namenode",
        ),
        spec=client.V1StatefulSetSpec(
            service_name=f"{cluster}-namenode",
            replicas=1,
            selector=client.V1LabelSelector(
                match_labels=_labels(cluster, "namenode"),
            ),
            template=client.V1PodTemplateSpec(
                metadata=client.V1ObjectMeta(
                    labels=_labels(cluster, "namenode"),
                    annotations={"data.hadoop.org/config-revision": config_revision},
                ),
                spec=client.V1PodSpec(
                    security_context=_pod_security(),
                    image_pull_secrets=[
                        client.V1LocalObjectReference(name=s["name"])
                        for s in norm["imagePullSecrets"]
                        if s.get("name")
                    ],
                    init_containers=[init],
                    containers=[main],
                    volumes=vols,
                    termination_grace_period_seconds=120,
                ),
            ),
            volume_claim_templates=[
                client.V1PersistentVolumeClaim(
                    metadata=client.V1ObjectMeta(name="hadoop-nn-data"),
                    spec=client.V1PersistentVolumeClaimSpec(
                        access_modes=["ReadWriteOnce"],
                        storage_class_name=norm["hdfs"]["namenodeStorageClass"],
                        resources=client.V1ResourceRequirements(
                            requests={"storage": norm["hdfs"]["namenodeStorageSize"]}
                        ),
                    ),
                )
            ],
        ),
    )


def datanode_statefulset(
    *,
    cluster: str,
    namespace: str,
    cr_api_version: str,
    cr_kind: str,
    cr_name: str,
    cr_uid: str,
    norm: Dict[str, Any],
    cm_name: str,
    config_revision: str,
) -> client.V1StatefulSet:
    vols, cm_mounts = _config_volumes(cm_name)
    nn_svc = f"{cluster}-namenode"
    nn0 = f"{cluster}-namenode-0.{nn_svc}"
    if norm["hdfs"]["ha"]["enabled"]:
        nn1 = f"{cluster}-namenode-1.{nn_svc}"
        wait_script = f"""
set -e
echo "Waiting for NameNode HA RPC..."
for i in {{1..360}}; do
  if bash -c "echo > /dev/tcp/{nn0}/9000" 2>/dev/null || bash -c "echo > /dev/tcp/{nn1}/9000" 2>/dev/null; then
    exit 0
  fi
  sleep 5
done
exit 1
""".strip()
    else:
        wait_script = f"""
set -e
echo "Waiting for NameNode RPC at {nn0}:9000..."
for i in {{1..360}}; do
  if bash -c "echo > /dev/tcp/{nn0}/9000" 2>/dev/null; then
    exit 0
  fi
  sleep 5
done
exit 1
""".strip()

    init_wait = client.V1Container(
        name="wait-for-namenode",
        image=norm["image"],
        image_pull_policy=norm["imagePullPolicy"],
        command=["/bin/bash", "-c"],
        args=[wait_script],
    )

    init_perm = client.V1Container(
        name="init-permissions",
        image=norm["image"],
        image_pull_policy=norm["imagePullPolicy"],
        security_context=client.V1SecurityContext(run_as_user=0, privileged=True),
        command=["/bin/bash", "-c"],
        args=[
            """
set -e
mkdir -p /opt/hadoop/data/dn
if id "hadoop" >/dev/null 2>&1; then
  chown -R hadoop:hadoop /opt/hadoop/data
else
  chown -R 1000:1000 /opt/hadoop/data
fi
chmod -R 775 /opt/hadoop/data
""".strip()
        ],
        volume_mounts=[
            client.V1VolumeMount(
                name="hadoop-dn-data",
                mount_path="/opt/hadoop/data/dn",
                sub_path="dn",
            ),
        ],
    )

    yarn_mounts = _yarn_config_mounts() if norm["yarn"]["enabled"] else []

    main = client.V1Container(
        name="datanode",
        image=norm["image"],
        image_pull_policy=norm["imagePullPolicy"],
        security_context=client.V1SecurityContext(run_as_user=0),
        command=["/bin/bash", "-c"],
        args=["exec hdfs datanode"],
        env=[
            client.V1EnvVar(name="HADOOP_CONF_DIR", value="/opt/hadoop/etc/hadoop"),
        ],
        ports=[
            client.V1ContainerPort(container_port=9864, name="web"),
            client.V1ContainerPort(container_port=9866, name="data"),
        ],
        readiness_probe=client.V1Probe(
            http_get=client.V1HTTPGetAction(path="/", port=9864),
            initial_delay_seconds=30,
            period_seconds=10,
        ),
        liveness_probe=client.V1Probe(
            http_get=client.V1HTTPGetAction(path="/", port=9864),
            initial_delay_seconds=90,
            period_seconds=20,
        ),
        resources=_res_requirements(norm["resources"]["datanode"]),
        volume_mounts=[
            client.V1VolumeMount(
                name="hadoop-dn-data",
                mount_path="/opt/hadoop/data/dn",
                sub_path="dn",
            ),
            *cm_mounts,
            *yarn_mounts,
        ],
    )

    return client.V1StatefulSet(
        api_version="apps/v1",
        kind="StatefulSet",
        metadata=meta(
            name=f"{cluster}-datanode",
            namespace=namespace,
            cr_name=cr_name,
            cr_uid=cr_uid,
            cr_api_version=cr_api_version,
            cr_kind=cr_kind,
            cluster=cluster,
            component="datanode",
        ),
        spec=client.V1StatefulSetSpec(
            service_name=f"{cluster}-datanode",
            replicas=norm["hdfs"]["datanodeReplicas"],
            selector=client.V1LabelSelector(
                match_labels=_labels(cluster, "datanode"),
            ),
            template=client.V1PodTemplateSpec(
                metadata=client.V1ObjectMeta(
                    labels=_labels(cluster, "datanode"),
                    annotations={"data.hadoop.org/config-revision": config_revision},
                ),
                spec=client.V1PodSpec(
                    security_context=_pod_security(),
                    image_pull_secrets=[
                        client.V1LocalObjectReference(name=s["name"])
                        for s in norm["imagePullSecrets"]
                        if s.get("name")
                    ],
                    init_containers=[init_wait, init_perm],
                    containers=[main],
                    volumes=vols,
                    termination_grace_period_seconds=60,
                ),
            ),
            volume_claim_templates=[
                client.V1PersistentVolumeClaim(
                    metadata=client.V1ObjectMeta(name="hadoop-dn-data"),
                    spec=client.V1PersistentVolumeClaimSpec(
                        access_modes=["ReadWriteOnce"],
                        storage_class_name=norm["hdfs"]["datanodeStorageClass"],
                        resources=client.V1ResourceRequirements(
                            requests={"storage": norm["hdfs"]["datanodeStorageSize"]}
                        ),
                    ),
                )
            ],
        ),
    )


def resourcemanager_statefulset(
    *,
    cluster: str,
    namespace: str,
    cr_api_version: str,
    cr_kind: str,
    cr_name: str,
    cr_uid: str,
    norm: Dict[str, Any],
    cm_name: str,
    config_revision: str,
) -> client.V1StatefulSet:
    vols, cm_mounts = _config_volumes(cm_name)
    yarn_mounts = _yarn_config_mounts()

    nn_svc = f"{cluster}-namenode"
    nn0 = f"{cluster}-namenode-0.{nn_svc}"
    if norm["hdfs"]["ha"]["enabled"]:
        nn1 = f"{cluster}-namenode-1.{nn_svc}"
        rm_wait_script = f"""
set -e
for i in {{1..360}}; do
  if bash -c "echo > /dev/tcp/{nn0}/9000" 2>/dev/null || bash -c "echo > /dev/tcp/{nn1}/9000" 2>/dev/null; then
    exit 0
  fi
  sleep 5
done
exit 1
""".strip()
    else:
        rm_wait_script = f"""
set -e
for i in {{1..360}}; do
  if bash -c "echo > /dev/tcp/{nn0}/9000" 2>/dev/null; then
    exit 0
  fi
  sleep 5
done
exit 1
""".strip()

    init_wait = client.V1Container(
        name="wait-for-namenode",
        image=norm["image"],
        image_pull_policy=norm["imagePullPolicy"],
        command=["/bin/bash", "-c"],
        args=[rm_wait_script],
    )

    main = client.V1Container(
        name="resourcemanager",
        image=norm["image"],
        image_pull_policy=norm["imagePullPolicy"],
        security_context=_container_security(),
        command=["yarn", "resourcemanager"],
        env=[
            client.V1EnvVar(name="HADOOP_CONF_DIR", value="/opt/hadoop/etc/hadoop"),
        ],
        ports=[
            client.V1ContainerPort(container_port=8088, name="web"),
            client.V1ContainerPort(container_port=8032, name="rpc"),
        ],
        readiness_probe=client.V1Probe(
            http_get=client.V1HTTPGetAction(path="/ws/v1/cluster/info", port=8088),
            initial_delay_seconds=40,
            period_seconds=10,
        ),
        liveness_probe=client.V1Probe(
            http_get=client.V1HTTPGetAction(path="/ws/v1/cluster/info", port=8088),
            initial_delay_seconds=120,
            period_seconds=30,
        ),
        resources=_res_requirements(norm["resources"]["resourcemanager"]),
        volume_mounts=[*cm_mounts, *yarn_mounts],
    )

    return client.V1StatefulSet(
        api_version="apps/v1",
        kind="StatefulSet",
        metadata=meta(
            name=f"{cluster}-resourcemanager",
            namespace=namespace,
            cr_name=cr_name,
            cr_uid=cr_uid,
            cr_api_version=cr_api_version,
            cr_kind=cr_kind,
            cluster=cluster,
            component="resourcemanager",
        ),
        spec=client.V1StatefulSetSpec(
            service_name=f"{cluster}-resourcemanager",
            replicas=1,
            selector=client.V1LabelSelector(
                match_labels=_labels(cluster, "resourcemanager"),
            ),
            template=client.V1PodTemplateSpec(
                metadata=client.V1ObjectMeta(
                    labels=_labels(cluster, "resourcemanager"),
                    annotations={"data.hadoop.org/config-revision": config_revision},
                ),
                spec=client.V1PodSpec(
                    security_context=_pod_security(),
                    image_pull_secrets=[
                        client.V1LocalObjectReference(name=s["name"])
                        for s in norm["imagePullSecrets"]
                        if s.get("name")
                    ],
                    init_containers=[init_wait],
                    containers=[main],
                    volumes=vols,
                    termination_grace_period_seconds=60,
                ),
            ),
        ),
    )


def nodemanager_statefulset(
    *,
    cluster: str,
    namespace: str,
    cr_api_version: str,
    cr_kind: str,
    cr_name: str,
    cr_uid: str,
    norm: Dict[str, Any],
    cm_name: str,
    config_revision: str,
) -> client.V1StatefulSet:
    vols, cm_mounts = _config_volumes(cm_name)
    yarn_mounts = _yarn_config_mounts()

    rm_svc = f"{cluster}-resourcemanager"
    rm_wait_host = f"{cluster}-resourcemanager-0.{rm_svc}"

    init_wait = client.V1Container(
        name="wait-for-resourcemanager",
        image=norm["image"],
        image_pull_policy=norm["imagePullPolicy"],
        command=["/bin/bash", "-c"],
        args=[
            f"""
set -e
echo "Waiting for ResourceManager RPC at {rm_wait_host}:8032..."
for i in {{1..360}}; do
  if bash -c "echo > /dev/tcp/{rm_wait_host}/8032" 2>/dev/null; then
    exit 0
  fi
  sleep 5
done
exit 1
""".strip()
        ],
    )

    main = client.V1Container(
        name="nodemanager",
        image=norm["image"],
        image_pull_policy=norm["imagePullPolicy"],
        security_context=_container_security(),
        command=["yarn", "nodemanager"],
        env=[
            client.V1EnvVar(name="HADOOP_CONF_DIR", value="/opt/hadoop/etc/hadoop"),
        ],
        ports=[client.V1ContainerPort(container_port=8042, name="web")],
        readiness_probe=client.V1Probe(
            http_get=client.V1HTTPGetAction(path="/ws/v1/node/info", port=8042),
            initial_delay_seconds=40,
            period_seconds=10,
        ),
        liveness_probe=client.V1Probe(
            http_get=client.V1HTTPGetAction(path="/ws/v1/node/info", port=8042),
            initial_delay_seconds=120,
            period_seconds=30,
        ),
        resources=_res_requirements(norm["resources"]["nodemanager"]),
        volume_mounts=[*cm_mounts, *yarn_mounts],
    )

    return client.V1StatefulSet(
        api_version="apps/v1",
        kind="StatefulSet",
        metadata=meta(
            name=f"{cluster}-nodemanager",
            namespace=namespace,
            cr_name=cr_name,
            cr_uid=cr_uid,
            cr_api_version=cr_api_version,
            cr_kind=cr_kind,
            cluster=cluster,
            component="nodemanager",
        ),
        spec=client.V1StatefulSetSpec(
            service_name=f"{cluster}-nodemanager",
            replicas=norm["yarn"]["nodemanagerReplicas"],
            selector=client.V1LabelSelector(
                match_labels=_labels(cluster, "nodemanager"),
            ),
            template=client.V1PodTemplateSpec(
                metadata=client.V1ObjectMeta(
                    labels=_labels(cluster, "nodemanager"),
                    annotations={"data.hadoop.org/config-revision": config_revision},
                ),
                spec=client.V1PodSpec(
                    security_context=_pod_security(),
                    image_pull_secrets=[
                        client.V1LocalObjectReference(name=s["name"])
                        for s in norm["imagePullSecrets"]
                        if s.get("name")
                    ],
                    init_containers=[init_wait],
                    containers=[main],
                    volumes=vols,
                    termination_grace_period_seconds=60,
                ),
            ),
        ),
    )


def namenode_pdb(
    *,
    cluster: str,
    namespace: str,
    cr_api_version: str,
    cr_kind: str,
    cr_name: str,
    cr_uid: str,
) -> client.V1PodDisruptionBudget:
    return client.V1PodDisruptionBudget(
        api_version="policy/v1",
        kind="PodDisruptionBudget",
        metadata=meta(
            name=f"{cluster}-namenode-pdb",
            namespace=namespace,
            cr_name=cr_name,
            cr_uid=cr_uid,
            cr_api_version=cr_api_version,
            cr_kind=cr_kind,
            cluster=cluster,
            component="namenode",
        ),
        spec=client.V1PodDisruptionBudgetSpec(
            min_available=1,
            selector=client.V1LabelSelector(
                match_labels=_labels(cluster, "namenode"),
            ),
        ),
    )
