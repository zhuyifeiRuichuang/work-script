"""Create or replace namespaced objects (idempotent reconcile)."""

from __future__ import annotations

from kubernetes import client
from kubernetes.client.rest import ApiException


def replace_or_create_configmap(
    api: client.CoreV1Api,
    namespace: str,
    obj: client.V1ConfigMap,
) -> None:
    name = obj.metadata.name
    try:
        cur = api.read_namespaced_config_map(name, namespace)
        obj.metadata.resource_version = cur.metadata.resource_version
        api.replace_namespaced_config_map(name, namespace, obj)
    except ApiException as e:
        if e.status == 404:
            api.create_namespaced_config_map(namespace, obj)
        else:
            raise


def ensure_service(
    api: client.CoreV1Api,
    namespace: str,
    obj: client.V1Service,
) -> None:
    """Services keep immutable fields (clusterIP / NodePort); create if missing only."""
    name = obj.metadata.name
    try:
        api.read_namespaced_service(name, namespace)
    except ApiException as e:
        if e.status == 404:
            api.create_namespaced_service(namespace, obj)
        else:
            raise


def replace_or_create_statefulset(
    api: client.AppsV1Api,
    namespace: str,
    obj: client.V1StatefulSet,
) -> None:
    name = obj.metadata.name
    try:
        cur = api.read_namespaced_stateful_set(name, namespace)
        obj.metadata.resource_version = cur.metadata.resource_version
        api.replace_namespaced_stateful_set(name, namespace, obj)
    except ApiException as e:
        if e.status == 404:
            api.create_namespaced_stateful_set(namespace, obj)
        else:
            raise


def replace_or_create_pdb(
    api: client.PolicyV1Api,
    namespace: str,
    obj: client.V1PodDisruptionBudget,
) -> None:
    name = obj.metadata.name
    try:
        cur = api.read_namespaced_pod_disruption_budget(name, namespace)
        obj.metadata.resource_version = cur.metadata.resource_version
        api.replace_namespaced_pod_disruption_budget(name, namespace, obj)
    except ApiException as e:
        if e.status == 404:
            api.create_namespaced_pod_disruption_budget(namespace, obj)
        else:
            raise


def delete_service(api: client.CoreV1Api, namespace: str, name: str) -> None:
    try:
        api.delete_namespaced_service(name, namespace)
    except ApiException as e:
        if e.status != 404:
            raise


def delete_statefulset(api: client.AppsV1Api, namespace: str, name: str) -> None:
    try:
        api.delete_namespaced_stateful_set(name, namespace)
    except ApiException as e:
        if e.status != 404:
            raise


def delete_pdb(api: client.PolicyV1Api, namespace: str, name: str) -> None:
    try:
        api.delete_namespaced_pod_disruption_budget(name, namespace)
    except ApiException as e:
        if e.status != 404:
            raise


def delete_config_map(api: client.CoreV1Api, namespace: str, name: str) -> None:
    try:
        api.delete_namespaced_config_map(name, namespace)
    except ApiException as e:
        if e.status != 404:
            raise


def replace_or_create_ingress(
    api: client.NetworkingV1Api,
    namespace: str,
    obj: client.V1Ingress,
) -> None:
    name = obj.metadata.name
    try:
        cur = api.read_namespaced_ingress(name, namespace)
        obj.metadata.resource_version = cur.metadata.resource_version
        api.replace_namespaced_ingress(name, namespace, obj)
    except ApiException as e:
        if e.status == 404:
            api.create_namespaced_ingress(namespace, obj)
        else:
            raise


def delete_ingress(api: client.NetworkingV1Api, namespace: str, name: str) -> None:
    try:
        api.delete_namespaced_ingress(name, namespace)
    except ApiException as e:
        if e.status != 404:
            raise
