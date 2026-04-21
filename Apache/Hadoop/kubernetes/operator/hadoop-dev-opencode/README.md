# Hadoop Kubernetes Operator

生产环境可用的Hadoop集群Kubernetes部署方案，支持单机和HA模式。

## 目录结构

```
├── hadoop-crd.yaml          # CRD定义文件
├── hadoop-ha/               # HA模式配置
│   ├── 00-rbac.yaml         # RBAC配置
│   ├── 00-configmap.yaml    # 配置映射
│   ├── 01-zookeeper.yaml   # Zookeeper
│   ├── 02-journalnode.yaml # JournalNode
│   ├── 03-namenode.yaml    # NameNode (HA)
│   ├── 04-datanode.yaml    # DataNode
│   ├── 05-resourcemanager.yaml # ResourceManager (HA)
│   ├── 06-nodemanager.yaml # NodeManager
│   └── README.md           # HA模式说明
├── namenode.yaml            # 单机NameNode
├── datanode.yaml            # DataNode
├── resourcemanager.yaml     # ResourceManager
├── nodemanager.yaml         # NodeManager
└── configmap.yaml           # Hadoop配置
```

## 部署模式

### 1. 单机模式

适用于开发和测试环境：

```bash
# 部署ConfigMap
kubectl apply -f configmap.yaml

# 部署NameNode
kubectl apply -f namenode.yaml

# 部署DataNode
kubectl apply -f datanode.yaml

# 部署ResourceManager
kubectl apply -f resourcemanager.yaml

# 部署NodeManager
kubectl apply -f nodemanager.yaml
```

### 2. HA模式

适用于生产环境：

```bash
cd hadoop-cluster
kubectl apply -f . -n hadoop-cluster
```

详见 [hadoop-ha/README.md](hadoop-ha/README.md)

## 生产环境特性

基于Apache Doris优秀实践进行优化：

| 特性 | 说明 |
|------|------|
| 亲和性调度 | podAntiAffinity、nodeAffinity |
| 安全上下文 | runAsNonRoot、fsGroup |
| 健康检查 | startupProbe + readinessProbe + livenessProbe |
| 生命周期管理 | preStop优雅停机 |
| 容忍调度 | tolerations支持节点故障迁移 |
| 监控集成 | prometheus.io注解 |
| PVC存储 | 持久化存储配置 |

## 快速开始

### 单机模式

```bash
# 创建命名空间
kubectl create namespace hadoop

# 部署所有组件
kubectl apply -f . -n hadoop

# 查看状态
kubectl get pods -n hadoop
```

### 访问服务

| 服务 | NodePort | 地址 |
|------|----------|------|
| NameNode Web | 30070 | http://node-ip:30070 |
| NameNode RPC | 30090 | - |
| ResourceManager Web | 30088 | http://node-ip:30088 |

## 组件说明

### 核心组件

| 组件 | 文件 | 说明 |
|------|------|------|
| NameNode | namenode.yaml | HDFS元数据管理 |
| DataNode | datanode.yaml | HDFS数据存储 |
| ResourceManager | resourcemanager.yaml | YARN资源调度 |
| NodeManager | nodemanager.yaml | YARN计算节点 |

### 配置文件

- configmap.yaml: Hadoop核心配置文件
  - core-site.xml
  - hdfs-site.xml
  - yarn-site.xml
  - mapred-site.xml
  - capacity-scheduler.xml

## 资源规格

### 单机模式

| 组件 | CPU Request | CPU Limit | Memory Request | Memory Limit | 存储 |
|------|-------------|-----------|----------------|--------------|------|
| NameNode | 500m | 1 | 2Gi | 4Gi | 20Gi |
| DataNode | 500m | 1 | 2Gi | 4Gi | 20Gi |
| ResourceManager | 500m | 1 | 2Gi | 4Gi | 10Gi |
| NodeManager | 250m | 500m | 1Gi | 2Gi | 10Gi |

### HA模式

详见 [hadoop-ha/README.md](hadoop-ha/README.md)

## 运维命令

```bash
# 查看Pod状态
kubectl get pods -n hadoop

# 查看日志
kubectl logs -f namenode-0 -n hadoop

# 进入容器
kubectl exec -it namenode-0 -n hadoop -- /bin/bash

# 扩容
kubectl scale sts datanode --replicas=3 -n hadoop

# 滚动更新
kubectl rollout restart sts namenode -n hadoop
```

## CRD定义

应用CRD定义后可使用声明式API：

```bash
kubectl apply -f hadoop-crd.yaml

# 查看CRD
kubectl get crd hadoopclusters.hadoop.apache.org
```

## 注意事项

1. 单机模式仅用于开发测试
2. 生产环境请使用HA模式
3. 确保配置足够的存储资源
4. 建议提前导入镜像到私有仓库