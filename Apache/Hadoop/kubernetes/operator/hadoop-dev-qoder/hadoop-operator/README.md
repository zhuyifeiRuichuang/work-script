# Hadoop Operator for Kubernetes

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/apache/hadoop-operator)](https://goreportcard.com/report/github.com/apache/hadoop-operator)

一个用于在 Kubernetes 上部署和管理 Hadoop 集群的生产级 Operator。支持 HDFS 和 YARN 的高可用部署，以及离线环境部署。

## 功能特性

- **完整组件支持**：NameNode、DataNode、ResourceManager、NodeManager
- **高可用 (HA)**：支持 NameNode HA 和 ResourceManager HA
- **配置管理**：通过 CRD 灵活配置 Hadoop 参数（core-site.xml、hdfs-site.xml、yarn-site.xml、mapred-site.xml、capacity-scheduler.xml）
- **离线部署**：支持私有镜像仓库和离线环境部署
- **监控集成**：支持 Prometheus 监控和 Grafana 仪表盘
- **存储管理**：支持动态 PVC 和多种 StorageClass
- **安全支持**：预留 Kerberos、TLS、Ranger 集成接口
- **生产就绪**：Pod 反亲和性、资源限制、健康检查、Leader Election

## 目录

- [快速开始](#快速开始)
- [镜像构建指南](#镜像构建指南)
- [离线部署指南](#离线部署指南)
- [CRD 配置说明](#crd-配置说明)
- [部署配置调整指南](#部署配置调整指南)
- [开发指南](#开发指南)
- [故障排查](#故障排查)
- [架构图](#架构图)
- [API 参考](#api-参考)
- [相关文档](#相关文档)

## 快速开始

### 前置要求

- Kubernetes 1.24+
- kubectl 已配置
- Helm 3.x（可选）
- 至少 3 个工作节点的集群（用于高可用部署）
- StorageClass 已配置（用于动态 PVC 供应）

### 安装 Operator

#### 方式一：使用 kubectl 部署（推荐）

```bash
# 1. 创建命名空间
kubectl apply -f config/namespace/namespace.yaml

# 2. 安装 CRD
kubectl apply -f config/crd/bases/hadoop.apache.org_hadoopclusters.yaml

# 3. 安装 RBAC（包含 Leader Election 权限）
kubectl apply -f config/rbac/service_account.yaml
kubectl apply -f config/rbac/role.yaml
kubectl apply -f config/rbac/role_binding.yaml
kubectl apply -f config/rbac/leader_election_role.yaml
kubectl apply -f config/rbac/leader_election_role_binding.yaml

# 4. 部署 Operator
kubectl apply -f config/manager/manager.yaml

# 5. 验证 Operator 部署
kubectl get pods -n hadoop
kubectl logs -n hadoop deployment/hadoop-operator
```

#### 方式二：使用 kustomize 部署

```bash
# 一键部署所有资源
kubectl apply -k config/

# 或使用特定 overlay
kubectl apply -k config/default/
```

#### 方式三：本地运行（开发模式）

```bash
# 克隆仓库
git clone https://github.com/apache/hadoop-operator.git
cd hadoop-operator

# 安装依赖
go mod download

# 运行 Operator（需要本地有 kubeconfig）
make run

# 或者指定配置
make run ARGS="--leader-elect=false --metrics-bind-address=:8080"
```

### 部署 Hadoop 集群

#### 基础部署（单实例，适合测试）

```bash
# 创建集群命名空间
kubectl create namespace hadoop

# 部署基础集群
kubectl apply -f config/samples/hadoop_v1_hadoopcluster.yaml
```

#### 高可用部署（双 NameNode + 双 ResourceManager）

```bash
# 创建集群命名空间
kubectl create namespace hadoop

# 部署高可用集群
kubectl apply -f config/samples/hadoop_v1_hadoopcluster_ha.yaml
```

#### 生产环境部署（推荐）

```bash
# 创建集群命名空间
kubectl create namespace hadoop

# 部署生产级高可用集群
kubectl apply -f config/samples/hadoop_v1_hadoopcluster_production.yaml
```

**生产环境配置特点：**
- NameNode HA（2 副本）+ JournalNode（3 副本）
- ResourceManager HA（2 副本）
- DataNode（3 副本）+ NodeManager（3 副本）
- Pod 反亲和性配置，确保组件分布在不同节点
- 合理的资源限制（CPU/Memory）
- 持久化存储配置
- Prometheus 监控集成

### 验证部署

```bash
# 查看集群状态
kubectl get hadoopcluster -n hadoop

# 查看 Pod 状态
kubectl get pods -n hadoop

# 查看服务
kubectl get svc -n hadoop

# 查看 NameNode UI
kubectl port-forward -n hadoop svc/hadoop-sample-namenode-external 9870:9870
# 访问 http://localhost:9870

# 查看 ResourceManager UI
kubectl port-forward -n hadoop svc/hadoop-sample-resourcemanager-external 8088:8088
# 访问 http://localhost:8088

# 验证 HDFS 状态
kubectl exec -it -n hadoop hadoop-sample-namenode-0 -- hdfs dfsadmin -report

# 验证 YARN 状态
kubectl exec -it -n hadoop hadoop-sample-resourcemanager-0 -- yarn node -list
```

## 镜像构建指南

本章节详细介绍如何构建 Hadoop Operator 和 Hadoop 集群组件的容器镜像。

### 镜像名称对应关系

构建的镜像名称需要与部署配置中的镜像名称保持一致：

| 镜像类型 | 构建默认名称 | 部署配置位置 | 部署默认名称 |
|---------|-------------|-------------|-------------|
| **Operator** | `apache/hadoop-operator:latest` | `config/manager/manager.yaml` | `apache/hadoop-operator:latest` |
| **Hadoop** | `apache/hadoop:3.3.6` | `config/samples/hadoop_v1_hadoopcluster.yaml` | `apache/hadoop:3.3.6` |

**重要提示**：
- 构建时通过 `IMG` 变量指定 Operator 镜像名称
- 构建时通过 `HADOOP_VERSION` 变量指定 Hadoop 版本标签
- 部署前请确保构建的镜像名称与部署配置中的 `image` 字段一致
- 使用私有仓库时，需要同时修改构建命令和部署配置中的镜像地址

### 构建目录结构

```
build/
├── operator/          # Operator 镜像构建资源
│   ├── Dockerfile     # Operator Dockerfile
│   └── README.md      # Operator 镜像构建说明
└── hadoop/            # Hadoop 组件镜像构建资源
    ├── Dockerfile     # Hadoop Dockerfile
    ├── README.md      # Hadoop 镜像构建说明
    ├── conf/          # Hadoop 配置文件
    └── scripts/       # 启动脚本
```

### 快速构建

#### 一键构建所有镜像

```bash
# 构建 Operator 和 Hadoop 镜像（使用默认名称）
make build-images

# 构建并推送到私有仓库
make push-images DOCKER_REGISTRY=registry.example.com

# 自定义 Operator 镜像名称和 Hadoop 版本
make build-images IMG=registry.example.com/myoperator:v1.0.0 HADOOP_VERSION=3.3.5
make push-images IMG=registry.example.com/myoperator:v1.0.0 DOCKER_REGISTRY=registry.example.com
```

#### 单独构建 Operator 镜像

```bash
# 构建 Operator 镜像
make docker-build IMG=myregistry/hadoop-operator:v1.0.0

# 多平台构建
make docker-buildx IMG=myregistry/hadoop-operator:v1.0.0
```

详细说明参见 [build/operator/README.md](build/operator/README.md)。

#### 单独构建 Hadoop 镜像

```bash
# 构建 Hadoop 镜像
make build-hadoop-image HADOOP_VERSION=3.3.6

# 多平台构建
make build-hadoop-multiarch PLATFORMS=linux/amd64,linux/arm64
```

详细说明参见 [build/hadoop/README.md](build/hadoop/README.md)。

### 代码编译方式

#### 本地编译（推荐用于开发）

```bash
# 编译二进制到 bin/manager
make build

# 本地运行
make run
```

#### 容器编译（用于生产构建）

```bash
# 构建镜像
make docker-build IMG=myregistry/hadoop-operator:v1.0.0

# 提取编译产物到本地
docker build --target builder -t temp:builder -f build/operator/Dockerfile .
docker create --name extract temp:builder
docker cp extract:/workspace/manager ./bin/manager
docker rm extract
```

详细说明参见 [build/README.md](build/README.md)。

## 离线部署指南

离线部署需要提前准备容器镜像并传输到离线环境。确保构建的镜像名称与部署配置中的镜像名称一致。

### 镜像名称对应关系

离线部署时，需要确保以下镜像名称对应关系：

| 镜像 | 构建/拉取名称 | 离线部署配置 |
|------|--------------|-------------|
| **Operator** | `apache/hadoop-operator:latest` | `config/manager/manager.yaml` 中的 `image` 字段 |
| **Hadoop** | `apache/hadoop:3.3.6` | `config/samples/offline-deployment.yaml` 中的 `spec.image.repository` 和 `tag` |

### 1. 在有网络的环境中准备镜像

#### 方式一：直接拉取官方镜像（推荐）

```bash
# 拉取官方镜像
docker pull apache/hadoop-operator:latest
docker pull apache/hadoop:3.3.6

# 保存镜像
docker save apache/hadoop-operator:latest > hadoop-operator.tar
docker save apache/hadoop:3.3.6 > hadoop.tar

# 传输到离线环境
tar czvf hadoop-images.tar.gz hadoop-operator.tar hadoop.tar
```

#### 方式二：自定义构建后导出

```bash
# 构建自定义镜像
make docker-build IMG=myregistry.example.com:5000/apache/hadoop-operator:v1.0.0
make build-hadoop-image HADOOP_VERSION=3.3.6

# 标记镜像（如果需要）
docker tag apache/hadoop:3.3.6 myregistry.example.com:5000/apache/hadoop:3.3.6

# 保存镜像
docker save myregistry.example.com:5000/apache/hadoop-operator:v1.0.0 > hadoop-operator.tar
docker save myregistry.example.com:5000/apache/hadoop:3.3.6 > hadoop.tar
```

#### 方式三：使用脚本批量保存

```bash
# 保存镜像到本地
cd hack/offline
./save-images.sh --output-dir ./offline-images

# 传输到离线环境
tar czvf hadoop-images.tar.gz offline-images/
```

### 2. 在离线环境加载镜像

```bash
# 解压
tar xzvf hadoop-images.tar.gz

# 加载镜像
./load-images.sh --input-dir ./offline-images

# 或者推送到私有仓库
./load-images.sh --input-dir ./offline-images --target-registry myregistry.example.com:5000
```

### 3. 配置私有镜像仓库

```bash
# 创建镜像拉取 Secret
kubectl create secret docker-registry regcred \
  --docker-server=myregistry.example.com:5000 \
  --docker-username=<username> \
  --docker-password=<password> \
  --docker-email=<email> \
  -n <namespace>
```

### 4. 配置离线部署

#### 配置 Operator 镜像（如使用私有仓库）

编辑 `config/manager/manager.yaml`，修改 Operator 镜像地址：

```yaml
spec:
  template:
    spec:
      containers:
      - name: manager
        image: myregistry.example.com:5000/apache/hadoop-operator:v1.0.0  # 私有镜像地址
        imagePullPolicy: IfNotPresent
      imagePullSecrets:
        - name: regcred  # 镜像拉取 Secret
```

#### 配置 Hadoop 集群镜像

编辑 `config/samples/offline-deployment.yaml`，修改以下配置：

```yaml
spec:
  image:
    repository: myregistry.example.com:5000/apache/hadoop  # 私有镜像仓库地址
    tag: "3.3.6"
    pullPolicy: IfNotPresent
    pullSecrets:
      - name: regcred  # 镜像拉取 Secret
  
  hdfs:
    nameNode:
      storage:
        storageClassName: local  # 离线环境的 StorageClass
    dataNode:
      storage:
        storageClassName: local
```

**重要提示**：确保 `repository` 和 `tag` 与步骤 1 中保存的镜像名称完全一致。

### 5. 部署离线集群

```bash
# 创建命名空间
kubectl create namespace hadoop

# 创建镜像拉取 Secret
kubectl create secret docker-registry regcred \
  --docker-server=myregistry.example.com:5000 \
  --docker-username=<username> \
  --docker-password=<password> \
  --docker-email=<email> \
  -n hadoop

# 部署集群
kubectl apply -f config/samples/offline-deployment.yaml
```

## CRD 配置说明

### HadoopCluster Spec

```yaml
apiVersion: hadoop.apache.org/v1
kind: HadoopCluster
metadata:
  name: my-cluster
spec:
  # 镜像配置
  image:
    repository: apache/hadoop      # 镜像仓库
    tag: "3.3.6"                   # 镜像标签
    pullPolicy: IfNotPresent       # 拉取策略
    pullSecrets:                   # 镜像拉取密钥
      - name: regcred

  # HDFS 配置
  hdfs:
    nameNode:
      replicas: 2                  # 副本数，HA 模式下至少为 2
      ha:
        enabled: true              # 启用 HA
        journalNode:
          replicas: 3              # JournalNode 副本数
      resources:                   # 资源限制
        requests:
          memory: "4Gi"
          cpu: "1000m"
        limits:
          memory: "8Gi"
          cpu: "2000m"
      storage:                     # 存储配置
        size: "100Gi"
        storageClassName: standard
      affinity:                    # 亲和性配置
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchLabels:
                app: hadoop-namenode
            topologyKey: kubernetes.io/hostname

    dataNode:
      replicas: 3
      resources:
        requests:
          memory: "4Gi"
          cpu: "1000m"
        limits:
          memory: "8Gi"
          cpu: "2000m"
      storage:
        size: "500Gi"
        storageClassName: standard

  # YARN 配置
  yarn:
    resourceManager:
      replicas: 2
      ha:
        enabled: true
      resources:
        requests:
          memory: "4Gi"
          cpu: "1000m"
        limits:
          memory: "8Gi"
          cpu: "2000m"

    nodeManager:
      replicas: 3
      resources:
        requests:
          memory: "4Gi"
          cpu: "1000m"
        limits:
          memory: "8Gi"
          cpu: "2000m"

  # Hadoop 配置覆盖
  config:
    coreSite:
      "hadoop.tmp.dir": "/opt/hadoop/tmpdata"
    hdfsSite:
      "dfs.replication": "3"
      "dfs.permissions.enabled": "false"
    yarnSite:
      "yarn.nodemanager.pmem-check-enabled": "false"
      "yarn.nodemanager.vmem-check-enabled": "false"
    mapredSite:
      "mapreduce.framework.name": "yarn"

  # 监控配置
  metrics:
    enabled: true
    serviceMonitor:
      enabled: true
      labels:
        release: prometheus
```

### 配置项详细说明

#### core-site.xml 配置

通过 `spec.config.coreSite` 覆盖 core-site.xml 配置：

```yaml
config:
  coreSite:
    "hadoop.tmp.dir": "/opt/hadoop/tmp"
    "fs.defaultFS": "hdfs://my-cluster-namenode:9000"
    "io.file.buffer.size": "131072"
    "hadoop.security.authentication": "simple"
    "hadoop.security.authorization": "false"
```

常用配置项：
- `hadoop.tmp.dir`: Hadoop 临时目录
- `fs.defaultFS`: 默认文件系统 URI
- `io.file.buffer.size`: IO 缓冲区大小
- `hadoop.security.authentication`: 认证方式（simple/kerberos）
- `hadoop.security.authorization`: 是否启用授权

#### hdfs-site.xml 配置

通过 `spec.config.hdfsSite` 覆盖 hdfs-site.xml 配置：

```yaml
config:
  hdfsSite:
    "dfs.replication": "3"
    "dfs.blocksize": "134217728"
    "dfs.namenode.handler.count": "100"
    "dfs.datanode.handler.count": "10"
    "dfs.permissions.enabled": "true"
    "dfs.webhdfs.enabled": "true"
    "dfs.namenode.acls.enabled": "true"
```

常用配置项：
- `dfs.replication`: 数据块副本数
- `dfs.blocksize`: 数据块大小（字节）
- `dfs.namenode.handler.count`: NameNode RPC 处理线程数
- `dfs.datanode.handler.count`: DataNode 传输线程数
- `dfs.permissions.enabled`: 是否启用权限检查
- `dfs.webhdfs.enabled`: 是否启用 WebHDFS

#### yarn-site.xml 配置

通过 `spec.config.yarnSite` 覆盖 yarn-site.xml 配置：

```yaml
config:
  yarnSite:
    "yarn.resourcemanager.hostname": "my-cluster-resourcemanager"
    "yarn.scheduler.minimum-allocation-mb": "1024"
    "yarn.scheduler.maximum-allocation-mb": "8192"
    "yarn.nodemanager.resource.memory-mb": "8192"
    "yarn.nodemanager.resource.cpu-vcores": "4"
    "yarn.nodemanager.aux-services": "mapreduce_shuffle"
```

常用配置项：
- `yarn.resourcemanager.hostname`: ResourceManager 主机名
- `yarn.scheduler.minimum/maximum-allocation-mb`: 容器内存分配范围
- `yarn.nodemanager.resource.memory-mb`: NodeManager 可用内存
- `yarn.nodemanager.resource.cpu-vcores`: NodeManager 可用 CPU 核数
- `yarn.nodemanager.aux-services`: 辅助服务

#### mapred-site.xml 配置

通过 `spec.config.mapredSite` 覆盖 mapred-site.xml 配置：

```yaml
config:
  mapredSite:
    "mapreduce.framework.name": "yarn"
    "mapreduce.job.maps": "10"
    "mapreduce.job.reduces": "5"
    "mapreduce.map.memory.mb": "2048"
    "mapreduce.reduce.memory.mb": "4096"
```

#### capacity-scheduler.xml 配置

通过 `spec.config.capacityScheduler` 覆盖 capacity-scheduler.xml 配置：

```yaml
config:
  capacityScheduler:
    "yarn.scheduler.capacity.maximum-applications": "10000"
    "yarn.scheduler.capacity.maximum-am-resource-percent": "0.1"
    "yarn.scheduler.capacity.root.queues": "default"
    "yarn.scheduler.capacity.root.default.capacity": "100"
```

## 开发指南

### 项目结构

```
hadoop-operator/
├── api/v1/                    # CRD Go 类型定义
│   ├── hadoopcluster_types.go # CRD 类型定义
│   ├── groupversion_info.go   # API 版本信息
│   └── zz_generated.deepcopy.go # 自动生成的 DeepCopy 方法
├── build/                     # 镜像构建资源
│   ├── operator/              # Operator 镜像构建
│   │   ├── Dockerfile         # Operator Dockerfile
│   │   └── README.md          # Operator 镜像构建说明
│   └── hadoop/                # Hadoop 组件镜像构建
│       ├── Dockerfile         # Hadoop Dockerfile
│       ├── README.md          # Hadoop 镜像构建说明
│       ├── conf/              # Hadoop 配置文件模板
│       │   ├── core-site.xml
│       │   ├── hdfs-site.xml
│       │   ├── yarn-site.xml
│       │   └── mapred-site.xml
│       └── scripts/           # 启动脚本
│           ├── entrypoint.sh
│           └── healthcheck.sh
├── cmd/
│   └── main.go                # Operator 入口
├── internal/
│   ├── controller/            # 主控制器
│   │   └── hadoopcluster_controller.go
│   └── reconciler/            # 组件协调器
│       ├── configmap.go       # ConfigMap 协调
│       ├── namenode.go        # NameNode 协调
│       ├── datanode.go        # DataNode 协调
│       ├── yarn.go            # YARN 协调
│       └── ha.go              # 高可用组件协调
├── config/
│   ├── crd/bases/             # CRD YAML
│   ├── manager/               # Operator 部署配置
│   ├── rbac/                  # RBAC 配置
│   │   ├── service_account.yaml
│   │   ├── role.yaml          # ClusterRole
│   │   ├── role_binding.yaml  # ClusterRoleBinding
│   │   ├── leader_election_role.yaml      # Leader Election Role
│   │   └── leader_election_role_binding.yaml # Leader Election RoleBinding
│   ├── namespace/             # 命名空间配置
│   └── samples/               # 示例配置
│       ├── hadoop_v1_hadoopcluster.yaml           # 基础配置
│       ├── hadoop_v1_hadoopcluster_ha.yaml        # 高可用配置
│       └── hadoop_v1_hadoopcluster_production.yaml # 生产配置
├── hack/offline/              # 离线部署工具
├── Makefile
└── go.mod
```

### 开发环境搭建

```bash
# 1. 克隆仓库
git clone https://github.com/apache/hadoop-operator.git
cd hadoop-operator

# 2. 安装依赖
go mod download

# 3. 安装开发工具
# 安装 controller-gen（用于生成 CRD）
make controller-gen

# 安装 kustomize（用于配置管理）
make kustomize
```

### 构建

#### 构建二进制

```bash
# 构建二进制
make build
```

#### 构建 Operator 镜像

Operator 镜像用于部署 Hadoop Operator 控制器到 Kubernetes 集群。

```bash
# 构建镜像（开发版本）
make docker-build IMG=hadoop-operator:dev

# 构建镜像（发布版本）
make docker-build IMG=apache/hadoop-operator:v0.1.0

# 推送镜像
make docker-push IMG=apache/hadoop-operator:v0.1.0

# 构建多平台镜像
make docker-buildx IMG=apache/hadoop-operator:v0.1.0
```

详细说明参见 [build/operator/README.md](build/operator/README.md)。

#### 构建 Hadoop 集群组件镜像

Hadoop 组件镜像是一个统一的基础镜像，支持运行 NameNode、DataNode、ResourceManager、NodeManager 等服务。

```bash
# 构建 Hadoop 镜像（默认版本 3.3.6）
make build-hadoop-image

# 指定 Hadoop 版本构建
make build-hadoop-image HADOOP_VERSION=3.3.6

# 多平台构建
make build-hadoop-multiarch PLATFORMS=linux/amd64,linux/arm64

# 构建并推送到私有仓库
make push-images DOCKER_REGISTRY=registry.example.com
```

详细说明参见 [build/hadoop/README.md](build/hadoop/README.md)。

#### 一键构建所有镜像

```bash
# 构建 Operator 和 Hadoop 镜像（使用默认名称）
make build-images

# 构建并推送所有镜像到私有仓库
make push-images DOCKER_REGISTRY=registry.example.com

# 自定义版本构建
make build-images IMG=apache/hadoop-operator:v1.0.0 HADOOP_VERSION=3.3.6
make push-images IMG=apache/hadoop-operator:v1.0.0 DOCKER_REGISTRY=registry.example.com
```

#### 本地编译与容器编译

**本地编译**（推荐用于开发）：

```bash
# 编译二进制到 bin/manager
make build

# 本地运行
make run
```

**容器编译**（用于生产构建）：

```bash
# 使用 Docker 编译并打包成镜像
make docker-build IMG=apache/hadoop-operator:v1.0.0

# 如需提取编译产物到本地
docker build --target builder -t temp:builder -f build/operator/Dockerfile .
docker create --name extract temp:builder
docker cp extract:/workspace/manager ./bin/manager
docker rm extract
```

详细说明参见 [build/README.md](build/README.md)。

#### 镜像安全扫描

```bash
# 扫描 Operator 镜像
make image-scan

# 使用 Trivy 扫描（需要安装 Trivy）
trivy image apache/hadoop-operator:latest
trivy image apache/hadoop:3.3.6
```

### 代码生成

```bash
# 生成 CRD 和 DeepCopy 方法
make generate

# 生成 manifests（CRD、RBAC、Webhook）
make manifests

# 同时运行 generate 和 manifests
make generate manifests
```

### 本地测试

```bash
# 运行单元测试
make test

# 运行特定测试
make test TEST_ARGS="-run TestReconcile"

# 运行端到端测试（需要 Kubernetes 集群）
make test-e2e

# 使用现有集群运行 e2e 测试
make test-e2e USE_EXISTING_CLUSTER=true
```

### 本地运行 Operator

```bash
# 方式一：直接运行（需要 kubeconfig）
make run

# 方式二：带参数运行
make run ARGS="--leader-elect=false --metrics-bind-address=:8080"

# 方式三：调试模式
make run ARGS="--zap-devel=true --zap-log-level=debug"

# 方式四：指定命名空间
make run ARGS="--leader-elect=false" NAMESPACE=hadoop
```

### 部署到集群

```bash
# 方式一：使用 Makefile 部署
make deploy IMG=apache/hadoop-operator:v0.1.0

# 方式二：使用 kustomize 部署
kubectl apply -k config/

# 方式三：手动部署（适合自定义配置）
kubectl apply -f config/namespace/namespace.yaml
kubectl apply -f config/crd/bases/
kubectl apply -f config/rbac/
kubectl apply -f config/manager/manager.yaml
```

### 清理部署

```bash
# 删除 Operator（保留 CRD 和集群数据）
make undeploy

# 完全清理（包括 CRD 和所有集群）
kubectl delete -f config/crd/bases/
kubectl delete namespace hadoop
```

## 部署配置调整指南

### 资源配置调整

#### 调整 Operator 资源限制

编辑 `config/manager/manager.yaml`：

```yaml
resources:
  limits:
    cpu: "2"          # 根据集群规模调整
    memory: 1Gi       # 大规模集群需要更多内存
  requests:
    cpu: 200m
    memory: 256Mi
```

#### 调整 Hadoop 组件资源

编辑集群配置文件（如 `hadoop_v1_hadoopcluster_production.yaml`）：

```yaml
hdfs:
  nameNode:
    resources:
      requests:
        memory: "16Gi"   # 大规模集群建议 16-32Gi
        cpu: "4000m"
      limits:
        memory: "32Gi"
        cpu: "8000m"
    storage:
      size: "500Gi"      # 根据元数据量调整
```

### 存储配置调整

#### 使用特定 StorageClass

```yaml
hdfs:
  nameNode:
    storage:
      size: "200Gi"
      storageClassName: "fast-ssd"  # 指定高性能存储
      accessMode: ReadWriteOnce
```

#### 多存储类型配置

```yaml
hdfs:
  dataNode:
    storage:
      size: "2Ti"
      storageClassName: "standard-hdd"  # DataNode 可使用大容量存储
```

### 高可用配置调整

#### 调整 JournalNode 副本数

```yaml
hdfs:
  nameNode:
    ha:
      enabled: true
      journalNode:
        replicas: 5  # 大规模集群可增加至 5 个
        resources:
          requests:
            memory: "4Gi"
            cpu: "1000m"
```

#### 调整 ZooKeeper 配置（使用外部 ZooKeeper）

```yaml
hdfs:
  nameNode:
    ha:
      enabled: true
      zooKeeper:
        connectionString: "zk1:2181,zk2:2181,zk3:2181"
```

### 网络配置调整

#### 配置 NodePort 服务

```yaml
hdfs:
  nameNode:
    service:
      type: NodePort
      nodePorts:
        web: 30070    # NameNode Web UI
        rpc: 30090    # NameNode RPC
```

#### 配置 LoadBalancer

```yaml
hdfs:
  nameNode:
    service:
      type: LoadBalancer
      annotations:
        service.beta.kubernetes.io/aws-load-balancer-type: nlb
```

### 监控配置调整

#### 启用 Prometheus 监控

```yaml
metrics:
  enabled: true
  exporterImage: "prometheus/jmx-exporter:0.20.0"
  serviceMonitor:
    enabled: true
    labels:
      release: prometheus
    interval: "30s"
```

#### 自定义 JMX Exporter 配置

创建 ConfigMap 并挂载到 Operator：

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: jmx-exporter-config
data:
  hadoop.yml: |
    rules:
      - pattern: "hadoop<service=NameNode, name=FSNamesystem><>(.*)"
        name: "hadoop_namenode_$1"
```

### 安全配置调整

#### 启用 Kerberos（预留接口）

```yaml
security:
  kerberos:
    enabled: true
    realm: "EXAMPLE.COM"
    kdc: "kdc.example.com"
    adminServer: "kdc.example.com"
    keytabSecret: "hadoop-keytab"
```

#### 启用 TLS（预留接口）

```yaml
security:
  tls:
    enabled: true
    certificateSecret: "hadoop-tls-certs"
```

### 亲和性与容忍配置

#### Pod 反亲和性（确保组件分布在不同节点）

```yaml
hdfs:
  nameNode:
    affinity:
      podAntiAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
        - labelSelector:
            matchExpressions:
            - key: app.kubernetes.io/component
              operator: In
              values:
              - namenode
          topologyKey: kubernetes.io/hostname
```

#### 节点亲和性（指定特定节点）

```yaml
hdfs:
  dataNode:
    affinity:
      nodeAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
          nodeSelectorTerms:
          - matchExpressions:
            - key: node-type
              operator: In
              values:
              - data-node
```

#### 容忍配置（容忍特定污点）

```yaml
hdfs:
  dataNode:
    tolerations:
    - key: "dedicated"
      operator: "Equal"
      value: "hadoop"
      effect: "NoSchedule"
```

## 故障排查

### 常见问题

#### 1. Operator 启动失败

```bash
# 查看 Operator Pod 状态
kubectl get pods -n hadoop

# 查看 Operator 日志
kubectl logs -n hadoop deployment/hadoop-operator

# 检查 RBAC 权限
kubectl auth can-i create leases --as=system:serviceaccount:hadoop:controller-manager -n hadoop
```

**常见原因：**
- Leader Election RBAC 配置错误
- 镜像拉取失败
- 资源限制过低

#### 2. Hadoop 集群创建失败

```bash
# 查看 HadoopCluster 状态
kubectl get hadoopcluster -n hadoop
kubectl describe hadoopcluster <cluster-name> -n hadoop

# 查看 Operator 日志
kubectl logs -n hadoop deployment/hadoop-operator | grep -i error

# 检查事件
kubectl get events -n hadoop --field-selector involvedObject.kind=HadoopCluster
```

#### 3. Pod 启动失败

```bash
# 查看 Pod 事件
kubectl describe pod <pod-name> -n hadoop

# 查看日志
kubectl logs <pod-name> -n hadoop

# 查看之前的容器日志（如果 Pod 已重启）
kubectl logs <pod-name> -n hadoop --previous
```

#### 4. NameNode 格式化失败

```bash
# 检查 PVC 状态
kubectl get pvc -n hadoop

# 检查存储类
kubectl get storageclass

# 查看 NameNode 日志
kubectl logs <namenode-pod> -n hadoop

# 手动格式化（谨慎操作，仅在新集群）
kubectl exec -it <namenode-pod> -n hadoop -- hdfs namenode -format -force
```

#### 5. NameNode HA 故障转移失败

```bash
# 检查 ZooKeeper 连接
kubectl exec -it <namenode-pod> -n hadoop -- zkCli.sh -server zookeeper:2181 ls /hadoop-ha

# 查看 JournalNode 状态
kubectl logs <journalnode-pod> -n hadoop

# 手动触发故障转移
kubectl exec -it <namenode-pod> -n hadoop -- hdfs haadmin -transitionToActive nn1
```

#### 6. DataNode 无法连接到 NameNode

```bash
# 检查网络连通性
kubectl exec -it <datanode-pod> -n hadoop -- nc -zv <namenode-service> 9000

# 检查配置
kubectl get configmap <cluster-name>-config -o yaml -n hadoop

# 检查 DataNode 注册状态
kubectl exec -it <namenode-pod> -n hadoop -- hdfs dfsadmin -report
```

#### 7. ResourceManager HA 问题

```bash
# 查看 ResourceManager 状态
kubectl exec -it <resourcemanager-pod> -n hadoop -- yarn rmadmin -getServiceState rm1

# 检查 ZooKeeper 中的 RM 状态
kubectl exec -it <resourcemanager-pod> -n hadoop -- zkCli.sh -server zookeeper:2181 ls /yarn-leader-election
```

#### 8. 镜像拉取失败

```bash
# 检查镜像是否存在
docker pull <image>

# 检查 Secret 配置
kubectl get secret regcred -n hadoop
kubectl get secret regcred -o yaml -n hadoop

# 检查镜像拉取权限
kubectl auth can-i get secrets --as=system:serviceaccount:hadoop:default -n hadoop
```

#### 9. 存储问题

```bash
# 检查 PVC 绑定状态
kubectl get pvc -n hadoop

# 检查 PV 状态
kubectl get pv

# 检查 StorageClass
kubectl get storageclass

# 查看 PVC 事件
kubectl describe pvc <pvc-name> -n hadoop
```

#### 10. 性能问题

```bash
# 检查资源使用
kubectl top pod -n hadoop

# 检查节点资源
kubectl top node

# 查看 NameNode 内存使用
kubectl exec -it <namenode-pod> -n hadoop -- jstat -gc 1

# 查看 HDFS 性能指标
kubectl exec -it <namenode-pod> -n hadoop -- hdfs dfsadmin -report
```

## 架构图

### 整体架构

```
┌─────────────────────────────────────────────────────────────────┐
│                     Kubernetes Cluster                          │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                     hadoop                              │   │
│  │  ┌─────────────────────────────────────────────────┐   │   │
│  │  │           Hadoop Operator                       │   │   │
│  │  │  ┌─────────────┐    ┌───────────────────────┐  │   │   │
│  │  │  │ Controller  │───▶│   Reconciler          │  │   │   │
│  │  │  │             │    │  - NameNode           │  │   │   │
│  │  │  │  Leader     │    │  - DataNode           │  │   │   │
│  │  │  │  Election   │    │  - ResourceManager    │  │   │   │
│  │  │  │  (HA)       │    │  - NodeManager        │  │   │   │
│  │  │  └─────────────┘    │  - HA Components      │  │   │   │
│  │  │                     └───────────────────────┘  │   │   │
│  │  └─────────────────────────────────────────────────┘   │   │
│  └─────────────────────────────────────────────────────────┘   │
│                              │                                  │
│                              ▼                                  │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                     hadoop                              │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │   │
│  │  │  NameNode    │  │   DataNode   │  │ ResourceManager│  │   │
│  │  │  (HA: 2)     │  │   (3+)       │  │  (HA: 2)     │  │   │
│  │  │  + JournalNode│  │              │  │              │  │   │
│  │  │  (3)         │  │              │  │              │  │   │
│  │  └──────────────┘  └──────────────┘  └──────┬───────┘  │   │
│  │                                              │         │   │
│  │                                       ┌──────┴───────┐ │   │
│  │                                       │ NodeManager  │ │   │
│  │                                       │ (3+)         │ │   │
│  │                                       └──────────────┘ │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

### 控制流程

```
User ──▶ kubectl apply -f hadoopcluster.yaml
           │
           ▼
    ┌──────────────┐
    │  Kubernetes  │
    │    API       │
    └──────┬───────┘
           │
           ▼
    ┌──────────────┐
    │   Hadoop     │◀── Watch
    │  Operator    │
    │              │
    │ ┌──────────┐ │
    │ │Reconcile │ │
    │ │  Loop    │ │
    │ └────┬─────┘ │
    └──────┼───────┘
           │
    ┌──────┼──────┐
    │      │      │
    ▼      ▼      ▼
┌──────┐┌──────┐┌──────┐
│Config││Service││Stateful│
│Map   ││      ││Set    │
└──┬───┘└──┬───┘└───┬───┘
   │       │        │
   ▼       ▼        ▼
┌─────────────────────────┐
│     Hadoop Cluster      │
│  ┌─────┐ ┌─────┐       │
│  │HDFS │ │YARN │       │
│  │NN/DN│ │RM/NM│       │
│  └─────┘ └─────┘       │
└─────────────────────────┘
```

### 高可用架构

```
┌─────────────────────────────────────────────────────────────┐
│                      HDFS HA Architecture                   │
│                                                             │
│   ┌─────────────┐         ┌─────────────┐                  │
│   │  NameNode   │◀───────▶│  NameNode   │                  │
│   │   Active    │         │   Standby   │                  │
│   └──────┬──────┘         └──────┬──────┘                  │
│          │                       │                          │
│          ▼                       ▼                          │
│   ┌─────────────────────────────────────┐                  │
│   │        JournalNode Quorum           │                  │
│   │    ┌─────┐   ┌─────┐   ┌─────┐     │                  │
│   │    │ JN1 │◀─▶│ JN2 │◀─▶│ JN3 │     │                  │
│   │    └─────┘   └─────┘   └─────┘     │                  │
│   └─────────────────────────────────────┘                  │
│          │                       │                          │
│          ▼                       ▼                          │
│   ┌─────────────────────────────────────┐                  │
│   │           DataNodes (3+)            │                  │
│   └─────────────────────────────────────┘                  │
│                                                             │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                     YARN HA Architecture                    │
│                                                             │
│   ┌─────────────┐         ┌─────────────┐                  │
│   │ResourceManager│◀─────▶│ResourceManager│                │
│   │   Active    │         │   Standby   │                  │
│   └──────┬──────┘         └──────┬──────┘                  │
│          │                       │                          │
│          ▼                       ▼                          │
│   ┌─────────────────────────────────────┐                  │
│   │         NodeManagers (3+)           │                  │
│   └─────────────────────────────────────┘                  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## API 参考

### HadoopCluster CRD 字段说明

#### 顶层字段

| 字段 | 类型 | 必需 | 说明 |
|------|------|------|------|
| `spec.image` | Object | 否 | 镜像配置 |
| `spec.hdfs` | Object | 是 | HDFS 配置 |
| `spec.yarn` | Object | 是 | YARN 配置 |
| `spec.config` | Object | 否 | Hadoop 配置覆盖 |
| `spec.security` | Object | 否 | 安全配置 |
| `spec.metrics` | Object | 否 | 监控配置 |

#### ImageSpec

```yaml
image:
  repository: apache/hadoop    # 镜像仓库
  tag: "3.3.6"                 # 镜像标签
  pullPolicy: IfNotPresent     # 拉取策略：Always/Never/IfNotPresent
  pullSecrets:                 # 镜像拉取密钥列表
    - name: regcred
```

#### NameNodeSpec

```yaml
hdfs:
  nameNode:
    replicas: 2                  # 副本数，HA 模式下至少为 2
    ha:
      enabled: true              # 启用 HA
      journalNode:
        replicas: 3              # JournalNode 副本数，建议 3 或 5
        resources:               # JournalNode 资源限制
          requests:
            memory: "2Gi"
            cpu: "500m"
        storage:                 # JournalNode 存储
          size: "50Gi"
          storageClassName: standard
    resources:                   # NameNode 资源限制
      requests:
        memory: "4Gi"
        cpu: "1000m"
      limits:
        memory: "8Gi"
        cpu: "2000m"
    storage:                     # NameNode 存储
      size: "100Gi"
      storageClassName: standard
      accessMode: ReadWriteOnce
    service:                     # 服务配置
      type: ClusterIP            # 服务类型：ClusterIP/NodePort/LoadBalancer
      nodePorts:                 # NodePort 映射（当 type 为 NodePort 时）
        web: 30070
        rpc: 30090
    affinity:                    # 亲和性配置
      podAntiAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
        - labelSelector:
            matchLabels:
              app: hadoop-namenode
          topologyKey: kubernetes.io/hostname
    tolerations: []              # 容忍配置
```

#### DataNodeSpec

```yaml
hdfs:
  dataNode:
    replicas: 3                  # 副本数
    resources:
      requests:
        memory: "4Gi"
        cpu: "1000m"
      limits:
        memory: "8Gi"
        cpu: "2000m"
    storage:
      size: "500Gi"
      storageClassName: standard
    service:
      type: ClusterIP
    affinity: {}                 # 亲和性配置
    tolerations: []              # 容忍配置
```

#### ResourceManagerSpec

```yaml
yarn:
  resourceManager:
    replicas: 2                  # 副本数，HA 模式下至少为 2
    ha:
      enabled: true              # 启用 HA
    resources:
      requests:
        memory: "4Gi"
        cpu: "1000m"
      limits:
        memory: "8Gi"
        cpu: "2000m"
    service:
      type: ClusterIP
      nodePorts:
        web: 30088
    affinity: {}
    tolerations: []
```

#### NodeManagerSpec

```yaml
yarn:
  nodeManager:
    replicas: 3
    resources:
      requests:
        memory: "4Gi"
        cpu: "1000m"
      limits:
        memory: "8Gi"
        cpu: "2000m"
    service:
      type: ClusterIP
    affinity: {}
    tolerations: []
```

#### HadoopConfig

```yaml
config:
  coreSite:                      # core-site.xml 配置
    "hadoop.tmp.dir": "/opt/hadoop/tmp"
    "fs.defaultFS": "hdfs://my-cluster-namenode:9000"
  hdfsSite:                      # hdfs-site.xml 配置
    "dfs.replication": "3"
    "dfs.permissions.enabled": "true"
    "dfs.webhdfs.enabled": "true"
  yarnSite:                      # yarn-site.xml 配置
    "yarn.resourcemanager.hostname": "my-cluster-resourcemanager"
    "yarn.nodemanager.aux-services": "mapreduce_shuffle"
  mapredSite:                    # mapred-site.xml 配置
    "mapreduce.framework.name": "yarn"
  capacityScheduler:             # capacity-scheduler.xml 配置
    "yarn.scheduler.capacity.maximum-applications": "10000"
```

#### SecuritySpec

```yaml
security:
  kerberos:                      # Kerberos 配置（预留）
    enabled: true
    realm: "EXAMPLE.COM"
    kdc: "kdc.example.com"
    adminServer: "kdc.example.com"
    keytabSecret: "hadoop-keytab"
  tls:                           # TLS 配置（预留）
    enabled: true
    certificateSecret: "hadoop-tls-certs"
  ranger:                        # Ranger 集成（预留）
    enabled: true
    adminURL: "http://ranger-admin:6080"
```

#### MetricsSpec

```yaml
metrics:
  enabled: true
  exporterImage: "prometheus/jmx-exporter:0.20.0"
  serviceMonitor:
    enabled: true
    labels:
      release: prometheus
    interval: "30s"
```

### 状态字段

```yaml
status:
  phase: Running                 # 集群阶段：Pending/Creating/Running/Failed/Deleting/Upgrading
  conditions:                    # 条件列表
    - type: Ready
      status: "True"
      lastTransitionTime: "2024-01-01T00:00:00Z"
  nameNode:
    readyReplicas: 2
    replicas: 2
    active: "my-cluster-namenode-0"
    standby:
      - "my-cluster-namenode-1"
  dataNode:
    readyReplicas: 3
    replicas: 3
    liveNodes: 3
    deadNodes: 0
  resourceManager:
    readyReplicas: 2
    replicas: 2
    active: "my-cluster-resourcemanager-0"
    standby:
      - "my-cluster-resourcemanager-1"
  nodeManager:
    readyReplicas: 3
    replicas: 3
```

## 相关文档

### 部署指南

- [kubectl 命令行部署指南](.qoder/repowiki/zh/content/安装与部署/在线部署/kubectl%20命令行部署.md) - 详细的 kubectl 部署步骤
- [Helm Chart 部署](.qoder/repowiki/zh/content/安装与部署/在线部署/Helm%20Chart%20部署.md) - 使用 Helm 部署
- [本地开发模式部署](.qoder/repowiki/zh/content/安装与部署/在线部署/本地开发模式部署.md) - 本地开发和调试
- [离线部署](.qoder/repowiki/zh/content/安装与部署/离线部署.md) - 私有镜像仓库和离线环境
- [部署验证与测试](.qoder/repowiki/zh/content/安装与部署/在线部署/部署验证与测试.md) - 部署后的验证步骤

### 配置参考

- [高可用配置](.qoder/repowiki/zh/content/配置参考/高可用配置.md) - NameNode/ResourceManager HA 配置
- [HDFS 配置](.qoder/repowiki/zh/content/配置参考/HDFS%20配置.md) - HDFS 详细配置选项
- [YARN 配置](.qoder/repowiki/zh/content/配置参考/YARN%20配置.md) - YARN 详细配置选项
- [监控配置](.qoder/repowiki/zh/content/配置参考/监控配置.md) - Prometheus 监控集成
- [安全配置](.qoder/repowiki/zh/content/配置参考/安全配置.md) - Kerberos/TLS 配置

### 开发指南

- [项目结构说明](.qoder/repowiki/zh/content/开发指南/项目结构说明.md) - 代码组织结构
- [代码架构设计](.qoder/repowiki/zh/content/开发指南/代码架构设计/代码架构设计.md) - 架构设计文档
- [Reconciler 模式架构](.qoder/repowiki/zh/content/开发指南/代码架构设计/Reconciler%20模式架构.md) - 协调器模式
- [构建与打包](.qoder/repowiki/zh/content/开发指南/构建与打包/构建与打包.md) - 构建流程
- [测试与调试](.qoder/repowiki/zh/content/开发指南/测试与调试.md) - 测试方法

### 组件管理

- [NameNode 组件管理](.qoder/repowiki/zh/content/组件管理/NameNode%20组件管理.md)
- [DataNode 组件管理](.qoder/repowiki/zh/content/组件管理/DataNode%20组件管理.md)
- [YARN 组件管理](.qoder/repowiki/zh/content/组件管理/YARN%20组件管理.md)
- [高可用模式管理](.qoder/repowiki/zh/content/组件管理/高可用模式管理.md)

### 其他

- [项目 Wiki 首页](.qoder/repowiki/zh/content/)
- [故障排查](.qoder/repowiki/zh/content/故障排查.md)
- [API 参考](.qoder/repowiki/zh/content/API%20参考.md)

## 贡献

欢迎提交 Issue 和 Pull Request！

### 开发流程

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

### 代码规范

- 遵循 Go 代码规范
- 添加必要的单元测试
- 更新相关文档
- 确保所有测试通过

## 许可证

Apache License 2.0

Copyright 2024 Apache Software Foundation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
