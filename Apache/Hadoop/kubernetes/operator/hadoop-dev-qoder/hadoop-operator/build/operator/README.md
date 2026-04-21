# Hadoop Operator 镜像构建

## 用途说明

本目录包含 **Hadoop Operator 控制器** 的镜像构建文件。

### 给谁使用？

| 角色 | 使用场景 |
|------|----------|
| **运维人员** | 构建自定义 Operator 镜像部署到私有仓库 |
| **开发者** | 发布新版本的 Operator 镜像 |
| **CI/CD 系统** | 自动化构建和发布流程 |

### 构建产物

构建产物是包含 `hadoop-operator` 二进制文件的容器镜像，用于在 Kubernetes 集群中运行 Operator 控制器。

### 镜像名称对应关系

| 项目 | 默认值 | 说明 |
|------|--------|------|
| **构建默认名称** | `apache/hadoop-operator:latest` | Makefile 中 IMG 变量的默认值 |
| **部署配置位置** | [config/manager/manager.yaml](../../config/manager/manager.yaml) | Operator Deployment 的镜像字段 |
| **部署默认名称** | `apache/hadoop-operator:latest` | manager.yaml 中的 image 字段 |

**重要**：构建的镜像名称必须与部署配置中的镜像名称一致。使用私有仓库时，需要同时修改构建命令和部署配置。

## 目录结构

```
build/operator/
├── Dockerfile          # Operator 镜像构建文件
├── README.md           # 本文件
└── build.sh            # 构建脚本（可选）
```

## 快速开始

### 方式一：使用 Makefile（推荐）

```bash
# 从项目根目录执行
cd hadoop-operator

# 构建 Operator 镜像
make docker-build IMG=myregistry/hadoop-operator:v1.0.0

# 构建多平台镜像
make docker-buildx IMG=myregistry/hadoop-operator:v1.0.0

# 推送镜像
make docker-push IMG=myregistry/hadoop-operator:v1.0.0
```

### 方式二：使用 Docker 直接构建

```bash
# 进入构建目录
cd hadoop-operator/build/operator

# 构建镜像（需要在项目根目录执行，因为需要访问源码）
docker build -t myregistry/hadoop-operator:v1.0.0 -f Dockerfile ../../

# 多平台构建
docker buildx build --platform linux/amd64,linux/arm64 \
  -t myregistry/hadoop-operator:v1.0.0 \
  --build-arg VERSION=1.0.0 \
  --build-arg BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
  --build-arg GIT_COMMIT=$(git rev-parse --short HEAD) \
  -f Dockerfile ../../
```

## 构建参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `TARGETOS` | 目标操作系统 | `linux` |
| `TARGETARCH` | 目标架构 | `amd64` |
| `VERSION` | 镜像版本号 | - |
| `BUILD_DATE` | 构建时间（ISO 8601格式） | - |
| `GIT_COMMIT` | Git 提交哈希 | - |
| `GIT_BRANCH` | Git 分支名 | - |

## 镜像特点

- **多阶段构建**：减小最终镜像体积
- **Distroless 基础镜像**：最小化攻击面，无 shell
- **非 root 运行**：使用 UID 65532 运行
- **静态编译**：CGO 禁用，无外部依赖
- **版本信息注入**：二进制包含构建信息

## 运行容器

```bash
# 本地运行（需要 kubeconfig）
docker run -v ~/.kube/config:/kubeconfig \
  -e KUBECONFIG=/kubeconfig \
  myregistry/hadoop-operator:v1.0.0 \
  --leader-elect=true \
  --metrics-bind-address=:8080
```

## 注意事项

1. 构建上下文必须是项目根目录（`hadoop-operator/`），因为 Dockerfile 需要访问 `go.mod`、`cmd/`、`api/`、`internal/` 等目录
2. 多平台构建需要启用 Docker buildx
3. 推送镜像前需要登录镜像仓库
