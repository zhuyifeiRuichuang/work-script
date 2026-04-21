# Hadoop Kubernetes Operator

[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go%20version-1.21+-blue.svg)](https://golang.org/)
[![Kubernetes](https://img.shields.io/badge/kubernetes-1.24+-blue.svg)](https://kubernetes.io/)
[![Operator](https://img.shields.io/badge/operator-kubedoop-blue.svg)](https://kubedoop.dev/)

用于在 Kubernetes 上管理 Apache Hadoop 集群的 Kubernetes Operator。本 Operator 提供了一种声明式方式来部署和管理 Hadoop 集群（HDFS、YARN、HBase），使用 Kubernetes 自定义资源。

## 📖 目录

- [特性](#-特性)
- [架构](#-架构)
- [前置条件](#-前置条件)
- [快速开始](#-快速开始)
- [安装](#-安装)
- [配置](#-配置)
- [CRD 参考](#-crd-参考)
- [开发](#-开发)
- [部署](#-部署)
- [示例](#-示例)
- [故障排除](#-故障排除)
- [贡献](#-贡献)
- [许可证](#-许可证)

## ✨ 特性

### 核心功能
- **HDFS 管理**：部署和管理 NameNode、DataNode 和 JournalNode
- **YARN 管理**：部署和管理 ResourceManager 和 NodeManager
- **HBase 集成**：可选的 HBase 集群支持（Master 和 RegionServers）
- **高可用性**：原生支持 HDFS HA（基于 JournalNode QJM）和 YARN HA（基于 ZooKeeper 自动故障转移）
- **HDFS 联邦**：支持多个 HDFS 命名空间
- **Kerberos 安全**：可选的 Kerberos 认证支持
- **HadoopApplication CRD**：以 Kubernetes 自定义资源方式提交和管理 Hadoop 应用（MapReduce、Spark、Hive 等）

### 高级功能
- **角色组**：灵活的角色组配置，适应不同工作负载
- **日志**：每个组件的日志配置（控制台 + 文件，支持按 logger 级别控制）
- **资源管理**：每个组件的 CPU 和内存限制
- **存储**：使用 PVC 的可配置持久化存储（支持 EmptyDir 用于测试）
- **健康检查**：内置存活探针和就绪探针
- **滚动更新**：支持滚动升级
- **ConfigMap 配置管理**：自动生成 `core-site.xml`、`hdfs-site.xml`、`yarn-site.xml`、`mapred-site.xml` 和 `entrypoint.sh`
- **自定义 ConfigMap 合并**：将用户提供的 ConfigMap 与 Operator 生成的配置合并

### Operator 功能
- **Webhook 验证**：CRD 验证和默认值设置（准入 Webhook）
- **指标**：Prometheus 指标端点
- **领导者选举**：Operator 自身的高可用性
- **多命名空间支持**：监视单个或所有命名空间
- **Finalizer 清理**：CRD 删除时的正确资源清理

## 🏗️ 架构

```
┌─────────────────────────────────────────────────────────────┐
│                     Hadoop Operator                         │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              HadoopCluster Controller                │   │
│  │  ┌───────────┐ ┌───────────┐ ┌───────────┐          │   │
│  │  │ NameNode  │ │ DataNode  │ │JournalNode│          │   │
│  │  │ Builder   │ │ Builder   │ │ Builder   │          │   │
│  │  └───────────┘ └───────────┘ └───────────┘          │   │
│  │  ┌───────────┐ ┌───────────┐ ┌───────────┐          │   │
│  │  │ResourceMgr│ │NodeManager│ │ConfigMap  │          │   │
│  │  │ Builder   │ │ Builder   │ │ Builder   │          │   │
│  │  └───────────┘ └───────────┘ └───────────┘          │   │
│  │  ┌───────────┐ ┌───────────┐                         │   │
│  │  │HBaseMaster│ │HBaseRS    │  HadoopApplication      │   │
│  │  │ Builder   │ │ Builder   │  Controller             │   │
│  │  └───────────┘ └───────────┘                         │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   StatefulSet │  │   StatefulSet │  │   StatefulSet │     │
│  │   (NameNode)  │  │   (DataNode)  │  │ (JournalNode) │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │  Deployment   │  │   StatefulSet │  │   StatefulSet│     │
│  │(ResourceMgr)  │  │ (HBaseMaster) │  │(HBaseRegionS)│     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│  ┌──────────────┐  ┌──────────────┐                        │
│  │  DaemonSet   │  │   ConfigMap   │                        │
│  │ (NodeManager) │  │ (Hadoop Config)│                       │
│  └──────────────┘  └──────────────┘                        │
└─────────────────────────────────────────────────────────────┘
```

**API Group**：`hadoop.kubedoop.dev/v1`

**自定义资源**：
- `HadoopCluster`（缩写：`hc`）— Hadoop 集群生命周期管理
- `HadoopApplication`（缩写：`ha`）— Hadoop 应用提交

## 📋 前置条件

- Kubernetes 1.24 或更高版本
- kubectl 1.24 或更高版本
- Go 1.21 或更高版本（用于开发）

### 可选依赖

如需 HDFS 高可用：
- ZooKeeper 集群（或使用 [zookeeper-operator](https://github.com/zncdatadev/zookeeper-operator)）

## 🚀 快速开始

### 1. 部署 Operator

```bash
# 克隆仓库
git clone https://github.com/hadoop-operator/hadoop-k8s-operator.git
cd hadoop-k8s-operator

# 安装 CRD
kubectl apply -f operator/config/crd/hadoopcluster-crd.yaml

# 创建命名空间
kubectl create namespace hadoop-system

# 应用 RBAC
kubectl apply -f operator/config/rbac/role.yaml

# 应用 Webhook 配置
kubectl apply -f operator/config/webhook/webhook.yaml

# 部署 Operator
kubectl apply -f operator/config/deploy/operator-deployment.yaml
```

### 2. 部署 Hadoop 集群

```yaml
# hadoopcluster.yaml
apiVersion: hadoop.kubedoop.dev/v1
kind: HadoopCluster
metadata:
  name: my-hadoop-cluster
spec:
  image: apache/hadoop:3.4.1
  nameNodeSpec:
    replicas: 1
    storage:
      useEmptyDir: true
  dataNodeSpec:
    replicas: 3
    storage:
      useEmptyDir: true
  resourceManagerSpec:
    replicas: 1
  nodeManagerSpec:
    replicas: 3
```

```bash
kubectl apply -f hadoopcluster.yaml -n hadoop-system
```

### 3. 检查集群状态

```bash
# 查看集群
kubectl get hadoopcluster -n hadoop-system

# 查看所有资源
kubectl get all -n hadoop-system -l hadoop.kubedoop.dev/cluster=my-hadoop-cluster

# 查看详细状态
kubectl describe hadoopcluster my-hadoop-cluster -n hadoop-system
```

## 📥 安装

### 方式一：手动安装（推荐用于开发）

```bash
# 克隆仓库
git clone https://github.com/hadoop-operator/hadoop-k8s-operator.git
cd hadoop-k8s-operator

# 安装 CRD
kubectl apply -f operator/config/crd/hadoopcluster-crd.yaml

# 创建命名空间
kubectl create namespace hadoop-system

# 应用 RBAC
kubectl apply -f operator/config/rbac/role.yaml

# 部署 Operator
kubectl apply -f operator/config/deploy/operator-deployment.yaml
```

### 方式二：从源码构建和部署

```bash
cd operator

# 安装依赖
go mod download

# 构建 Operator 二进制文件
go build -o bin/hadoop-operator ./cmd/main.go

# 构建 Docker 镜像
docker build -t hadoop-operator:latest -f ../docker/hadoop/Dockerfile .
```

## ⚙️ 配置

### 基础配置

```yaml
apiVersion: hadoop.kubedoop.dev/v1
kind: HadoopCluster
metadata:
  name: my-cluster
spec:
  image: apache/hadoop:3.4.1
  imagePullPolicy: IfNotPresent

  # 集群全局配置
  clusterConfig:
    replicationFactor: 3
    blockSize: 134217728
    zooKeeperConfigMapName: my-zk-config

  # Hadoop XML 和环境变量配置
  configSpec:
    logDir: /opt/hadoop/logs
    dataDir: /data/hadoop
    hadoopEnv:
      HDFS_HEAPSIZE: "4096"
      YARN_HEAPSIZE: "2048"
    coreSite:
      fs.trash.interval: "360"
    hdfsSite:
      dfs.replication: "3"
      dfs.blocksize: "134217728"
    yarnSite:
      yarn.nodemanager.resource.memory-mb: "16384"
    mapredSite:
      mapreduce.framework.name: "yarn"

  # 组件规格
  nameNodeSpec:
    replicas: 1
    resources:
      limits:
        cpu: "2"
        memory: 4Gi
      requests:
        cpu: "1"
        memory: 2Gi
    storage:
      storageClassName: standard
      resources:
        requests:
          storage: 50Gi

  dataNodeSpec:
    replicas: 3
    volumesPerNode: 1

  resourceManagerSpec:
    replicas: 1

  nodeManagerSpec:
    replicas: 3
```

### 高可用配置

```yaml
apiVersion: hadoop.kubedoop.dev/v1
kind: HadoopCluster
metadata:
  name: my-ha-cluster
spec:
  image: apache/hadoop:3.4.1

  # ZooKeeper 配置
  clusterConfig:
    zooKeeperConfigMapName: my-zk-config

  # HA 配置
  ha:
    nameNodeHA:
      enabled: true
      nameServiceId: ns1
      journalClusterId: jc1
      replicas: 2
    resourceManagerHA:
      enabled: true
      clusterId: rm-cluster
      replicas: 2

  # 组件
  nameNodeSpec:
    replicas: 2
  journalNodeSpec:
    replicas: 3
  dataNodeSpec:
    replicas: 3
  resourceManagerSpec:
    replicas: 2
  nodeManagerSpec:
    replicas: 3
```

### HBase 配置

```yaml
apiVersion: hadoop.kubedoop.dev/v1
kind: HadoopCluster
metadata:
  name: my-cluster-with-hbase
spec:
  image: apache/hadoop:3.4.1

  # ... 其他组件配置 ...

  hbaseSpec:
    enabled: true
    masterSpec:
      replicas: 2
      resources:
        limits:
          cpu: "2"
          memory: 4Gi
    regionServerSpec:
      replicas: 3
      resources:
        limits:
          cpu: "2"
          memory: 4Gi
    config:
      hbaseSite:
        hbase.cluster.distributed: "true"
        hbase.rootdir: "hdfs://my-cluster-namenode:9000/hbase"
      hbaseEnv:
        HBASE_HEAPSIZE: "4096"
```

### 角色组（高级）

```yaml
apiVersion: hadoop.kubedoop.dev/v1
kind: HadoopCluster
metadata:
  name: my-cluster
spec:
  dataNodeSpec:
    roleGroups:
      default:
        replicas: 3
        resources:
          limits:
            cpu: "4"
            memory: 8Gi
        config:
          hdfsSite:
            dfs.datanode.du.reserved: "10737418240"
      high-io:
        replicas: 2
        resources:
          limits:
            cpu: "8"
            memory: 16Gi
        nodeSelector:
          node-type: high-io
```

## 📚 CRD 参考

### HadoopCluster

| 字段 | 类型 | 必需 | 说明 |
|------|------|------|------|
| `spec.image` | string | 否 | Hadoop 镜像（默认：`apache/hadoop:3.4.1`） |
| `spec.imagePullPolicy` | PullPolicy | 否 | 镜像拉取策略（默认：`IfNotPresent`） |
| `spec.imagePullSecrets` | []LocalObjectReference | 否 | 镜像拉取密钥 |
| `spec.serviceAccountName` | string | 否 | ServiceAccount 名称（默认：`hadoop-operator`） |
| `spec.clusterConfig` | ClusterConfigSpec | 否 | 集群全局配置（复制因子、块大小、ZooKeeper） |
| `spec.configSpec` | ConfigSpec | 否 | Hadoop XML 和环境变量配置 |
| `spec.nameNodeSpec` | NameNodeSpec | 否 | NameNode 配置 |
| `spec.dataNodeSpec` | DataNodeSpec | 否 | DataNode 配置 |
| `spec.journalNodeSpec` | JournalNodeSpec | 否 | JournalNode 配置（HDFS HA 必需） |
| `spec.resourceManagerSpec` | ResourceManagerSpec | 否 | ResourceManager 配置 |
| `spec.nodeManagerSpec` | NodeManagerSpec | 否 | NodeManager 配置 |
| `spec.hbaseSpec` | HBaseSpec | 否 | HBase 配置 |
| `spec.ha` | HAConfig | 否 | 高可用配置 |
| `spec.authentication` | AuthenticationSpec | 否 | 认证配置（TLS、Kerberos、OIDC） |
| `spec.federation` | FederationConfig | 否 | HDFS 联邦配置 |
| `spec.clusterOperation` | ClusterOperationSpec | 否 | 集群操作设置（自动格式化、升级） |

### 组件规格通用字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `replicas` | int32 | 副本数量 |
| `resources` | ResourceRequirements | CPU 和内存限制/请求 |
| `storage` | StorageSpec | 持久化存储配置（PVC 或 EmptyDir） |
| `affinity` | Affinity | Pod 亲和规则 |
| `nodeSelector` | map[string]string | 节点选择器标签 |
| `tolerations` | []Toleration | Pod 容忍配置 |
| `image` | string | 覆盖默认镜像 |
| `imagePullPolicy` | PullPolicy | 镜像拉取策略 |
| `ports` | *Ports | 端口配置（组件特定） |
| `logging` | LoggingSpec | 日志配置（控制台 + 文件） |
| `roleGroups` | map[string]RoleGroupSpec | 角色组配置，用于异构工作负载 |
| `annotations` | map[string]string | Pod 注解 |
| `labels` | map[string]string | Pod 标签 |

### 集群状态

| 字段 | 类型 | 说明 |
|------|------|------|
| `status.phase` | ClusterPhase | 集群阶段：Pending、Creating、Running、Upgrading、Deleting、Failed、Unknown |
| `status.nameNodeStatus` | ComponentStatus | NameNode 组件状态 |
| `status.dataNodeStatus` | ComponentStatus | DataNode 组件状态 |
| `status.journalNodeStatus` | ComponentStatus | JournalNode 组件状态 |
| `status.resourceManagerStatus` | ComponentStatus | ResourceManager 组件状态 |
| `status.nodeManagerStatus` | ComponentStatus | NodeManager 组件状态 |
| `status.hbaseStatus` | HBaseStatus | HBase Master 和 RegionServer 状态 |
| `status.conditions` | []ClusterCondition | 集群条件（Ready、ConfigReady、ComponentReady 等） |

### HadoopApplication

| 字段 | 类型 | 必需 | 说明 |
|------|------|------|------|
| `spec.clusterRef` | ClusterRef | 是 | 目标 HadoopCluster 引用 |
| `spec.type` | ApplicationType | 是 | 应用类型：mapreduce、spark、hive、hbase、pig、sqoop |
| `spec.jarFile` | string | 否 | 应用 JAR 文件路径 |
| `spec.mainClass` | string | 否 | 主类 |
| `spec.args` | []string | 否 | 命令行参数 |
| `spec.env` | []EnvVar | 否 | 环境变量 |
| `spec.resources` | ResourceRequirements | 否 | 资源需求 |
| `spec.config` | ApplicationConfig | 否 | 应用特定配置 |

## 🔧 开发

### 前置条件

- Go 1.21+
- Docker 或 Podman
- kubectl
- kind（用于本地测试）

### 设置开发环境

```bash
# 克隆仓库
git clone https://github.com/hadoop-operator/hadoop-k8s-operator.git
cd hadoop-k8s-operator/operator

# 安装依赖
go mod download

# 构建 Operator
go build -o bin/hadoop-operator ./cmd/main.go

# 运行测试
go test ./...
```

### 本地运行 Operator

```bash
# 创建 kind 集群
kind create cluster --name=hadoop-operator-dev

# 安装 CRD
kubectl apply -f config/crd/hadoopcluster-crd.yaml

# 应用 RBAC
kubectl apply -f config/rbac/role.yaml

# 运行 Operator
go run ./cmd/main.go
```

### 构建 Docker 镜像

```bash
# 构建镜像
docker build -t <your-registry>/hadoop-operator:latest -f ../docker/hadoop/Dockerfile .

# 推送镜像
docker push <your-registry>/hadoop-operator:latest
```

### 项目结构

```
operator/
├── cmd/main.go                          # 入口
├── pkg/apis/hadoop/v1/                  # CRD 类型定义
│   ├── hadoopcluster_types.go           # HadoopCluster + HadoopApplication 类型
│   ├── hadoopcluster_webhook.go         # 准入 Webhook（验证/默认值）
│   └── groupversion_info.go             # GroupVersion: hadoop.kubedoop.dev/v1
├── pkg/builder/                         # 组件 Builder
│   ├── builder.go                       # Builder 接口 + BuilderFactory + computePhase()
│   ├── configmap_builder.go             # ConfigMap + entrypoint.sh 生成
│   ├── nn_dn_builder.go                 # NameNode/DataNode StatefulSet + Service
│   ├── journalnode_builder.go           # JournalNode StatefulSet + Service（HDFS HA）
│   ├── rm_nm_builder.go                 # ResourceManager Deployment + NodeManager DaemonSet
│   └── hbase_builder.go                 # HBase Master/RegionServer StatefulSet + Service
├── pkg/controller/
│   ├── hadoopcluster_controller.go      # HadoopCluster 调谐循环
│   └── hadoopapplication_controller.go  # HadoopApplication 生命周期管理
└── config/
    ├── crd/hadoopcluster-crd.yaml       # CRD 定义
    ├── rbac/role.yaml                   # ClusterRole、ServiceAccount、绑定
    ├── deploy/operator-deployment.yaml  # Operator Deployment
    ├── webhook/webhook.yaml             # 准入 Webhook 配置
    └── samples/                         # 示例 CRD 清单
```

## 🚢 部署

### 生产部署

```bash
# 安装 CRD
kubectl apply -f operator/config/crd/hadoopcluster-crd.yaml

# 创建命名空间
kubectl create namespace hadoop-system

# 应用 RBAC
kubectl apply -f operator/config/rbac/role.yaml

# 部署 Operator
kubectl apply -f operator/config/deploy/operator-deployment.yaml -n hadoop-system

# 验证安装
kubectl get pods -n hadoop-system
```

### 卸载

```bash
# 先删除所有 HadoopCluster
kubectl delete hadoopcluster --all --all-namespaces

# 删除 Operator
kubectl delete -f operator/config/deploy/operator-deployment.yaml -n hadoop-system

# 删除 CRD（注意：这将删除所有 HadoopCluster）
kubectl delete crd hadoopclusters.hadoop.kubedoop.dev
```

## 📝 示例

示例可在 [examples](examples/) 目录中找到：

- [简单集群](examples/hadoopcluster-simple.yaml) - 用于开发的基本集群（EmptyDir 存储）
- [HA 集群](examples/hadoopcluster-ha.yaml) - 基于 JournalNode 和 ZooKeeper 故障转移的高可用集群
- [带 HBase 的集群](examples/hadoopcluster-with-hbase.yaml) - 启用 HBase 的集群
- [自定义配置](examples/hadoopcluster-custom-config.yaml) - 带亲和性、节点选择器和完整 XML 覆盖的高级配置

[operator/config/samples/](operator/config/samples/) 中的其他示例：
- [示例集群](operator/config/samples/hadoop_v1alpha1_hadoopcluster.yaml)
- [示例 HA 集群](operator/config/samples/hadoop_v1alpha1_hadoopcluster-ha.yaml)
- [示例带 HBase 集群](operator/config/samples/hadoop_v1alpha1_hadoopcluster-with-hbase.yaml)

## 🔍 故障排除

### 常见问题

#### Pod 无法启动

```bash
# 检查 Pod 状态
kubectl get pods -n <namespace>

# 查看 Pod 事件
kubectl describe pod <pod-name> -n <namespace>

# 查看 Pod 日志
kubectl logs <pod-name> -n <namespace>
```

#### 存储问题

```bash
# 检查 PVC 状态
kubectl get pvc -n <namespace>

# 检查存储类
kubectl get storageclass
```

#### Operator 问题

```bash
# 检查 Operator 日志
kubectl logs -n hadoop-system deployment/hadoop-operator

# 检查 Operator 状态
kubectl get deployment -n hadoop-system
```

#### 验证组件状态

```bash
# 查看特定集群的所有组件
kubectl get pods -n <namespace> -l hadoop.kubedoop.dev/cluster=<cluster-name>

# 查看特定组件
kubectl get pods -n <namespace> -l hadoop.kubedoop.dev/component=namenode
kubectl get pods -n <namespace> -l hadoop.kubedoop.dev/component=datanode
kubectl get pods -n <namespace> -l hadoop.kubedoop.dev/component=journalnode
kubectl get pods -n <namespace> -l hadoop.kubedoop.dev/component=resourcemanager
```

## 🤝 贡献

欢迎贡献！请参阅我们的 [贡献指南](CONTRIBUTING.md) 了解更多详情。

1. Fork 本仓库
2. 创建您的功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交您的更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 Apache License 2.0 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 🙏 致谢

- [Kubedoop](https://kubedoop.dev/) - 本 Operator 所属的数据平台
- [zncdatadev/hdfs-operator](https://github.com/zncdatadev/hdfs-operator) - Operator 设计的参考
- [chriskery/hadoop-operator](https://github.com/chriskery/hadoop-operator) - 最初灵感来源
- [Apache Hadoop](https://hadoop.apache.org/) - 分布式计算框架

## 📬 联系方式

- GitHub Issues: [https://github.com/hadoop-operator/hadoop-k8s-operator/issues](https://github.com/hadoop-operator/hadoop-k8s-operator/issues)

## 🔗 相关项目

- [zookeeper-operator](https://github.com/zncdatadev/zookeeper-operator) - ZooKeeper operator
- [commons-operator](https://github.com/zncdatadev/commons-operator) - 通用 operator 工具
- [listener-operator](https://github.com/zncdatadev/listener-operator) - 监听器 operator
- [secret-operator](https://github.com/zncdatadev/secret-operator) - 密钥 operator
