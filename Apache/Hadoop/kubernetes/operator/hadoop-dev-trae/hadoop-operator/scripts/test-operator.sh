#!/bin/bash

set -e

# 解析参数
NAMESPACE="default"
while getopts "n:" opt; do
  case $opt in
    n)
      NAMESPACE="$OPTARG"
      ;;
    *)
      echo "用法: $0 [-n namespace]"
      exit 1
      ;;
  esac
done

echo "=== 测试 Hadoop Operator (namespace: $NAMESPACE) ==="

# 检查kubectl是否可用
if ! command -v kubectl &> /dev/null; then
    echo "错误: kubectl 命令不可用"
    exit 1
fi

# 检查集群状态
echo "1. 检查 Kubernetes 集群状态..."
kubectl cluster-info

# 部署CRD
echo "2. 部署 HadoopCluster CRD..."
kubectl apply -f deploy/crd.yaml

# 确保namespace存在
echo "3. 确保 namespace $NAMESPACE 存在..."
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# 部署Operator
echo "4. 部署 Hadoop Operator..."
kubectl apply -n $NAMESPACE -f deploy/operator.yaml

# 等待Operator就绪
echo "5. 等待 Operator 就绪..."
kubectl wait deployment hadoop-operator --for=condition=available --timeout=60s -n $NAMESPACE

# 部署示例集群
echo "6. 部署示例 Hadoop 集群..."
kubectl apply -n $NAMESPACE -f deploy/example-hadoopcluster.yaml

# 等待集群创建完成
echo "7. 等待 Hadoop 集群创建完成..."
sleep 60

# 检查集群状态
echo "8. 检查 Hadoop 集群状态..."
kubectl get hadoopclusters -n $NAMESPACE
echo "9. 检查组件状态..."
kubectl get pods -l cluster=example-hadoop -n $NAMESPACE
echo "10. 检查服务状态..."
kubectl get services -l cluster=example-hadoop -n $NAMESPACE

echo "=== 测试完成 ==="
echo "可以通过以下命令查看集群详细信息:"
echo "kubectl describe hadoopcluster example-hadoop -n $NAMESPACE"
echo "kubectl logs deployment/hadoop-operator -n $NAMESPACE"
echo ""
echo "使用示例:"
echo "  ./test-operator.sh -n my-namespace  # 在指定namespace中测试"

