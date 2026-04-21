#!/bin/bash
# Hadoop HA Cluster Deployment Script
# Usage: ./deploy.sh <namespace>

set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <namespace>"
    echo "Example: $0 hadoop-cluster"
    exit 1
fi

NAMESPACE=$1
echo "Deploying Hadoop HA cluster to namespace: $NAMESPACE"

# Create namespace if not exists
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# Update namespace in yaml files
for f in *.yaml; do
    sed -i "s/^  namespace:.*/  namespace: $NAMESPACE/" "$f" 2>/dev/null || true
done

# Deploy in order
echo "=== Deploying RBAC ==="
kubectl apply -f 00-rbac.yaml -n $NAMESPACE

echo "=== Deploying ConfigMap ==="
kubectl apply -f 00-configmap.yaml -n $NAMESPACE

echo "=== Deploying Zookeeper ==="
kubectl apply -f 01-zookeeper.yaml -n $NAMESPACE

echo "=== Waiting for Zookeeper to be ready ==="
kubectl wait --for=condition=ready pod/zookeeper-0 -n $NAMESPACE --timeout=300s || true
kubectl wait --for=condition=ready pod/zookeeper-1 -n $NAMESPACE --timeout=300s || true
kubectl wait --for=condition=ready pod/zookeeper-2 -n $NAMESPACE --timeout=300s || true

echo "=== Deploying JournalNode ==="
kubectl apply -f 02-journalnode.yaml -n $NAMESPACE

echo "=== Waiting for JournalNode to be ready ==="
kubectl wait --for=condition=ready pod/journalnode-0 -n $NAMESPACE --timeout=300s || true
kubectl wait --for=condition=ready pod/journalnode-1 -n $NAMESPACE --timeout=300s || true
kubectl wait --for=condition=ready pod/journalnode-2 -n $NAMESPACE --timeout=300s || true

echo "=== Deploying NameNode ==="
kubectl apply -f 03-namenode.yaml -n $NAMESPACE

echo "=== Waiting for NameNode to be ready ==="
kubectl wait --for=condition=ready pod/namenode-0 -n $NAMESPACE --timeout=300s || true
kubectl wait --for=condition=ready pod/namenode-1 -n $NAMESPACE --timeout=300s || true

echo "=== Deploying DataNode ==="
kubectl apply -f 04-datanode.yaml -n $NAMESPACE

echo "=== Deploying ResourceManager ==="
kubectl apply -f 05-resourcemanager.yaml -n $NAMESPACE

echo "=== Deploying NodeManager ==="
kubectl apply -f 06-nodemanager.yaml -n $NAMESPACE

echo "=== Deploying External Services ==="
kubectl apply -f 07-external-services.yaml -n $NAMESPACE

echo "=== Deployment complete ==="
echo ""
echo "Services:"
kubectl get svc -n $NAMESPACE
echo ""
echo "Pods:"
kubectl get pods -n $NAMESPACE