# Hadoop Operator 构建指南

本目录包含 Hadoop Operator 和 Hadoop 集群组件的镜像构建资源。

## 镜像名称对应关系

构建的镜像名称需要与部署配置中的镜像名称保持一致：

| 镜像类型 | 构建默认名称 | 部署配置位置 | 部署默认名称 |
|---------|-------------|-------------|-------------|
| **Operator** | `apache/hadoop-operator:latest` | [config/manager/manager.yaml](../config/manager/manager.yaml) | `apache/hadoop-operator:latest` |
| **Hadoop** | `apache/hadoop:3.3.6` | [config/samples/hadoop_v1_hadoopcluster.yaml](../config/samples/hadoop_v1_hadoopcluster.yaml) | `apache/hadoop:3.3.6` |

### 使用私有仓库时的配置

当使用私有镜像仓库时，需要同时修改：

1. **构建时**：通过 `IMG` 或 `DOCKER_REGISTRY` 变量指定镜像名称
2. **部署时**：修改对应 YAML 文件中的 `image` 字段

示例：
```bash
# 构建时指定私有仓库
make docker-build IMG=myregistry.com/apache/hadoop-operator:v1.0.0
make build-hadoop-image DOCKER_REGISTRY=myregistry.com

# 部署时修改 config/manager/manager.yaml
# image: myregistry.com/apache/hadoop-operator:v1.0.0

# 部署时修改 config/samples/hadoop_v1_hadoopcluster.yaml
# repository: myregistry.com/apache/hadoop
# tag: "3.3.6"
```

## 目录结构

```
build/
├── README.md              # 本文件 - 构建指南总览
├── operator/              # Operator 镜像构建
│   ├── Dockerfile         # Operator 镜像 Dockerfile
│   └── README.md          # Operator 镜像构建说明
└── hadoop/                # Hadoop 集群组件镜像构建
    ├── Dockerfile         # Hadoop 组件镜像 Dockerfile
    ├── README.md          # Hadoop 镜像构建说明
    ├── conf/              # Hadoop 配置文件模板
    │   ├── core-site.xml
    │   ├── hdfs-site.xml
    │   ├── yarn-site.xml
    │   └── mapred-site.xml
    └── scripts/           # 启动脚本
        ├── entrypoint.sh
        └── healthcheck.sh
```

## 镜像类型说明

### 1. Operator 镜像 (`build/operator/`)

**用途**: Hadoop Operator 控制器镜像

**给谁使用**:
- 运维人员部署 Operator 到 Kubernetes 集群
- 开发者发布新版本的 Operator
- CI/CD 系统自动化构建

**构建产物**: 包含 `hadoop-operator` 二进制文件的容器镜像

**详细说明**: 参见 [build/operator/README.md](operator/README.md)

### 2. Hadoop 集群组件镜像 (`build/hadoop/`)

**用途**: Hadoop 集群各组件的统一基础镜像

**给谁使用**:
- 运维人员部署 Hadoop 集群
- 开发者测试 Hadoop 组件
- 数据工程师本地开发测试

**支持组件**:
- NameNode (HDFS 元数据管理)
- DataNode (HDFS 数据存储)
- ResourceManager (YARN 资源管理)
- NodeManager (YARN 节点管理)
- JournalNode (HA 元数据同步)
- JobHistoryServer (MapReduce 历史服务)

**详细说明**: 参见 [build/hadoop/README.md](hadoop/README.md)

## 快速构建

### 构建所有镜像

```bash
cd hadoop-operator

# 构建 Operator 和 Hadoop 镜像（使用默认名称）
make build-images

# 构建并推送到私有仓库
make push-images DOCKER_REGISTRY=registry.example.com

# 自定义 Operator 镜像名称和 Hadoop 版本
make build-images IMG=registry.example.com/myoperator:v1.0.0 HADOOP_VERSION=3.3.5
make push-images IMG=registry.example.com/myoperator:v1.0.0 DOCKER_REGISTRY=registry.example.com
```

### 单独构建

```bash
# 仅构建 Operator 镜像
make build-operator-image IMG=myregistry/hadoop-operator:v1.0.0

# 仅构建 Hadoop 镜像
make build-hadoop-image HADOOP_VERSION=3.3.6
```

## 代码编译方式

### 方式一：本地编译

直接在宿主机编译 Go 代码，无需 Docker。

**适用场景**:
- 开发阶段快速迭代
- CI/CD 流水线
- 本地测试

**步骤**:

```bash
# 1. 进入项目目录
cd hadoop-operator

# 2. 安装依赖
go mod download

# 3. 编译二进制
make build

# 4. 编译产物位于 bin/manager
ls -la bin/manager
```

**本地运行**:

```bash
# 直接运行编译好的二进制
./bin/manager --leader-elect=false

# 或使用 Makefile
make run
```

### 方式二：容器编译

使用 Docker 容器编译代码，并将产物复制到本地。

**适用场景**:
- 需要与生产环境一致的编译环境
- 本地 Go 版本不匹配
- 交叉编译其他平台

**步骤**:

```bash
# 1. 进入项目目录
cd hadoop-operator

# 2. 使用 Docker 编译
docker run --rm \
  -v "$(pwd)":/workspace \
  -w /workspace \
  golang:1.21-alpine \
  sh -c "go mod download && go build -o bin/manager cmd/main.go"

# 3. 编译产物位于 bin/manager
ls -la bin/manager
```

**使用多阶段构建获取产物**:

```bash
# 构建并提取二进制
docker build --target builder -t hadoop-operator:builder -f build/operator/Dockerfile .

# 创建临时容器提取产物
docker create --name extract hadoop-operator:builder
docker cp extract:/workspace/manager ./bin/manager
docker rm extract
```

### 方式三：使用 Makefile 容器构建

```bash
# 构建镜像（自动处理编译）
make docker-build IMG=myregistry/hadoop-operator:v1.0.0

# 编译后的二进制在镜像中
# 如需提取，可以使用上面多阶段构建的方法
```

## 编译参数说明

### Operator 编译参数

| 参数 | 说明 | 示例 |
|------|------|------|
| `VERSION` | 版本号 | `1.0.0` |
| `BUILD_DATE` | 构建时间 | `2024-01-15T10:30:00Z` |
| `GIT_COMMIT` | Git 提交哈希 | `abc1234` |
| `GIT_BRANCH` | Git 分支 | `main` |

### 注入版本信息

```bash
# 本地编译时注入版本信息
go build -ldflags "-X main.Version=1.0.0 \
  -X main.BuildDate=$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
  -X main.GitCommit=$(git rev-parse --short HEAD)" \
  -o bin/manager cmd/main.go
```

## 多平台构建

### 本地交叉编译

```bash
# 编译 Linux AMD64 版本
GOOS=linux GOARCH=amd64 go build -o bin/manager-linux-amd64 cmd/main.go

# 编译 Linux ARM64 版本
GOOS=linux GOARCH=arm64 go build -o bin/manager-linux-arm64 cmd/main.go
```

### Docker 多平台构建

```bash
# 启用 buildx
docker buildx create --use

# 多平台构建 Operator 镜像
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t myregistry/hadoop-operator:v1.0.0 \
  -f build/operator/Dockerfile \
  --push .

# 多平台构建 Hadoop 镜像
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t myregistry/hadoop:3.3.6 \
  -f build/hadoop/Dockerfile \
  --push build/hadoop/
```

## 常见问题

### Q: 本地编译和容器编译有什么区别？

**本地编译**:
- 使用宿主机 Go 环境
- 速度快，适合开发迭代
- 需要本地 Go 版本与项目要求一致

**容器编译**:
- 使用容器内 Go 环境
- 环境一致性好，适合生产构建
- 支持交叉编译不同平台

### Q: 如何获取容器编译的产物？

使用多阶段构建的 `--target builder` 选项，然后创建临时容器复制文件：

```bash
docker build --target builder -t temp:builder -f build/operator/Dockerfile .
docker create --name extract temp:builder
docker cp extract:/workspace/manager ./bin/manager
docker rm extract
```

### Q: 构建 Hadoop 镜像时如何指定版本？

```bash
# 通过构建参数指定
docker build --build-arg HADOOP_VERSION=3.3.5 -t myregistry/hadoop:3.3.5 build/hadoop/

# 或通过 Makefile
make build-hadoop-image HADOOP_VERSION=3.3.5
```

## 相关文档

- [Operator 镜像构建说明](operator/README.md)
- [Hadoop 镜像构建说明](hadoop/README.md)
- [本地构建指南](../../.qoder/repowiki/zh/content/开发指南/构建与打包/本地构建.md)
- [Docker 打包指南](../../.qoder/repowiki/zh/content/开发指南/构建与打包/Docker 打包.md)
