# Hadoop Operator

Hadoop Operator 是一个用于在 Kubernetes 集群中部署和管理 Hadoop 集群的 Operator。它支持在离线环境和联网环境中使用，符合生产环境的标准要求。

## 功能特性

- **自动化部署**: 通过自定义资源定义 (CRD) 实现 Hadoop 集群的自动化部署
- **组件管理**: 支持 NameNode、DataNode、ResourceManager 和 NodeManager 的管理
- **配置管理**: 自动生成和管理 Hadoop 配置文件
- **存储管理**: 支持持久化存储配置
- **健康检查**: 内置健康检查机制，确保集群稳定运行
- **离线支持**: 支持在离线环境中部署和使用
- **命名空间管理**: 自动创建和管理专用命名空间
- **RBAC权限管理**: 细粒度的权限控制，包括 leader election 权限
- **Webhook支持**: 支持验证和 mutating webhook
- **安全配置**: 支持非 root 用户运行、权限控制和安全上下文
- **多架构支持**: 支持 amd64 和 arm64 架构
- **更完善的状态管理**: 实时监控组件状态，提供详细的状态信息

## 系统要求

- Kubernetes 1.20+ 集群
- Go 1.20+ (用于构建)
- Docker (用于构建镜像)

## 目录结构

```
hadoop-operator/
├── cmd/                    # 命令行工具
│   └── manager/            # Operator 管理器
├── pkg/                    # 核心代码
│   ├── apis/               # API 定义
│   ├── controller/         # 控制器
│   └── resources/          # 资源生成
├── deploy/                 # 部署配置
│   ├── crd.yaml            # CRD 定义
│   ├── operator.yaml       # Operator 部署
│   └── example-hadoopcluster.yaml  # 示例集群配置
├── build/                  # 构建配置
│   └── Dockerfile          # Docker 构建文件
├── scripts/                # 脚本
│   ├── build-operator.sh   # 构建脚本
│   └── test-operator.sh    # 测试脚本
├── go.mod                  # Go 模块定义
└── README.md               # 文档
```

## 部署步骤

### 1. 构建 Operator

```bash
# 克隆仓库
git clone <repository-url>
cd hadoop-operator

# 构建 Operator 镜像（默认架构：amd64）
./scripts/build-operator.sh

# 构建指定版本的镜像
./scripts/build-operator.sh -v v1.0.1

# 构建并推送镜像到仓库
./scripts/build-operator.sh -v v1.0.1 -r my-registry -p

# 构建 arm64 架构镜像
./scripts/build-operator.sh -a arm64

# 构建多架构镜像并推送
./scripts/build-operator.sh -v v1.0.1 -r my-registry -p
```

### 2. 部署 CRD

```bash
kubectl apply -f deploy/crd.yaml
```

### 3. 部署 Operator

Operator 会自动创建 `hadoop` 命名空间并在其中部署：

```bash
kubectl apply -f deploy/operator.yaml
```

### 4. 部署 Hadoop 集群

在 `hadoop` 命名空间中部署 Hadoop 集群：

```bash
kubectl apply -n hadoop -f deploy/example-hadoopcluster.yaml
```

### 5. 启用 Webhook（可选）

如果需要启用 Webhook 功能，可以修改 `operator.yaml` 中的环境变量：

```yaml
env:
  - name: ENABLE_WEBHOOK
    value: "true"
```

### 6. 验证部署

```bash
# 检查 Hadoop 集群状态
kubectl get hadoopclusters -n hadoop

# 检查 Pod 状态
kubectl get pods -l cluster=example-hadoop -n hadoop

# 检查服务状态
kubectl get services -l cluster=example-hadoop -n hadoop
```

## 离线环境部署

### 1. 准备离线镜像

```bash
# 在联网环境中拉取镜像
docker pull zhuyifeiruichuang/hadoop:3.1.1
docker pull hadoop-operator:v1.0.0

# 如果需要 arm64 架构镜像
docker pull --platform linux/arm64 zhuyifeiruichuang/hadoop:3.1.1
docker pull --platform linux/arm64 hadoop-operator:v1.0.0-arm64

# 保存镜像
docker save -o hadoop-images.tar zhuyifeiruichuang/hadoop:3.1.1 hadoop-operator:v1.0.0

# 在离线环境中加载镜像
docker load -i hadoop-images.tar
```

### 2. 离线部署

```bash
# 部署 CRD
kubectl apply -f deploy/crd.yaml

# 部署 Operator
kubectl apply -f deploy/operator.yaml

# 部署 Hadoop 集群
kubectl apply -f deploy/example-hadoopcluster.yaml
```

## 配置说明

### HadoopCluster 资源配置

| 字段 | 描述 | 默认值 |
|------|------|--------|
| `spec.nameNode.replicas` | NameNode 副本数 | 1 |
| `spec.nameNode.resources` | NameNode 资源配置 | 见示例 |
| `spec.nameNode.storage` | NameNode 存储配置 | 见示例 |
| `spec.dataNode.replicas` | DataNode 副本数 | 1 |
| `spec.dataNode.resources` | DataNode 资源配置 | 见示例 |
| `spec.dataNode.storage` | DataNode 存储配置 | 见示例 |
| `spec.resourceManager.replicas` | ResourceManager 副本数 | 1 |
| `spec.resourceManager.resources` | ResourceManager 资源配置 | 见示例 |
| `spec.nodeManager.replicas` | NodeManager 副本数 | 1 |
| `spec.nodeManager.resources` | NodeManager 资源配置 | 见示例 |
| `spec.hadoopConfig` | Hadoop 配置 | 见示例 |

### 示例配置

```yaml
apiVersion: hadoop.apache.org/v1alpha1
kind: HadoopCluster
metadata:
  name: example-hadoop
  namespace: default
spec:
  nameNode:
    replicas: 1
    resources:
      cpuRequest: "500m"
      memoryRequest: "2Gi"
      cpuLimit: "1000m"
      memoryLimit: "4Gi"
    storage:
      size: "20Gi"
      storageClass: "local"
  dataNode:
    replicas: 1
    resources:
      cpuRequest: "500m"
      memoryRequest: "2Gi"
      cpuLimit: "1000m"
      memoryLimit: "4Gi"
    storage:
      size: "20Gi"
      storageClass: "local"
  resourceManager:
    replicas: 1
    resources:
      cpuRequest: "500m"
      memoryRequest: "1Gi"
      cpuLimit: "1000m"
      memoryLimit: "2Gi"
  nodeManager:
    replicas: 1
    resources:
      cpuRequest: "500m"
      memoryRequest: "1Gi"
      cpuLimit: "1000m"
      memoryLimit: "2Gi"
  hadoopConfig:
    coreSite: {}
    hdfSite: {}
    yarnSite: {}
    mapredSite: {}
```

## 监控和维护

### 查看集群状态

```bash
kubectl describe hadoopcluster <cluster-name> -n hadoop
```

### 查看 Operator 日志

```bash
kubectl logs deployment/hadoop-operator -n hadoop
```

### 查看健康检查状态

```bash
# 查看 Operator 健康状态
kubectl get pods -n hadoop -l app.kubernetes.io/name=deployment
kubectl describe pod <operator-pod-name> -n hadoop
```

### 扩展集群

修改 `HadoopCluster` 资源的副本数，Operator 会自动扩展集群：

```bash
kubectl edit hadoopcluster <cluster-name> -n hadoop
```

### 升级集群

修改 `HadoopCluster` 资源的配置，Operator 会自动更新集群：

```bash
kubectl edit hadoopcluster <cluster-name> -n hadoop
```

### 查看详细状态信息

Operator 提供了更完善的状态管理，包括各组件的实际就绪副本数和错误信息：

```bash
kubectl get hadoopcluster <cluster-name> -n hadoop -o yaml
```

## 故障排除

### 常见问题

1. **Operator 启动失败**
   - 检查 RBAC 权限是否正确
   - 检查 Kubernetes 集群版本是否符合要求
   - 查看 Operator 日志：`kubectl logs deployment/hadoop-operator -n hadoop`
   - 检查命名空间是否存在：`kubectl get namespace hadoop`

2. **Webhook 相关问题**
   - 检查 Webhook 服务是否正常：`kubectl get service hadoop-operator-service -n hadoop`
   - 检查证书是否正确配置：`kubectl get secret hadoop-operator-secret-cert -n hadoop`
   - 暂时禁用 Webhook：修改 `operator.yaml` 中的 `ENABLE_WEBHOOK` 环境变量为 `"false"`

3. **Hadoop 组件启动失败**
   - 检查存储配置是否正确
   - 检查网络连接是否正常
   - 查看组件日志：`kubectl logs <pod-name> -n hadoop`
   - 检查资源配置是否足够

4. **集群状态异常**
   - 查看 Operator 日志：`kubectl logs deployment/hadoop-operator -n hadoop`
   - 检查 Hadoop 集群状态：`kubectl describe hadoopcluster <cluster-name> -n hadoop`
   - 查看详细状态信息：`kubectl get hadoopcluster <cluster-name> -n hadoop -o yaml`

5. **安全相关问题**
   - 检查 Pod 安全上下文是否正确配置
   - 检查非 root 用户运行是否正常
   - 查看安全相关日志：`kubectl logs <pod-name> -n hadoop | grep security`

## 注意事项

1. **存储配置**：确保使用的存储类支持持久化存储，并且有足够的存储空间
2. **资源配置**：根据实际需求调整组件的资源配置，避免资源不足导致集群不稳定
3. **网络配置**：确保 Kubernetes 集群内的网络通信正常，特别是组件之间的通信
4. **镜像拉取**：在离线环境中，确保已提前加载所需的镜像
5. **安全配置**：在生产环境中，建议配置适当的安全措施，如 TLS 加密、认证等
6. **命名空间管理**：Operator 会自动创建 `hadoop` 命名空间，确保该命名空间不存在冲突
7. **Webhook配置**：启用 Webhook 时，确保证书配置正确，否则可能导致资源创建失败
8. **多架构支持**：在不同架构的集群中部署时，确保使用相应架构的镜像
9. **健康检查**：定期检查 Operator 和 Hadoop 组件的健康状态，确保集群稳定运行
10. **版本管理**：使用 `build-operator.sh` 脚本的 `-v` 参数管理版本，确保版本一致性

## 版本历史

- v1.0.0: 初始版本，支持基本的 Hadoop 集群部署和管理
- v1.1.0: 增强版本，添加了以下功能：
  - 命名空间管理：自动创建和管理专用命名空间
  - RBAC权限管理：细粒度的权限控制，包括 leader election 权限
  - Webhook支持：支持验证和 mutating webhook
  - 安全配置：支持非 root 用户运行、权限控制和安全上下文
  - 健康检查：添加了健康检查端点和健康检查机制
  - 多架构支持：支持 amd64 和 arm64 架构
  - 更完善的状态管理：实时监控组件状态，提供详细的状态信息
  - 构建优化：添加了版本控制、镜像仓库支持和多架构构建

## 贡献

欢迎提交 issue 和 pull request 来改进这个项目。

## 许可证

[Apache License 2.0](LICENSE)
