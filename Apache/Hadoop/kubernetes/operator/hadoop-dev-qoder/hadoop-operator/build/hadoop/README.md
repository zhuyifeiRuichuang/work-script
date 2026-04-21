# Hadoop 集群组件镜像构建

## 用途说明

本目录包含 **Hadoop 集群各组件** 的统一基础镜像构建文件。

### 给谁使用？

| 角色 | 使用场景 |
|------|----------|
| **运维人员** | 构建自定义 Hadoop 镜像部署到私有仓库 |
| **开发者** | 测试 Hadoop 集群组件、开发新功能 |
| **数据工程师** | 在本地或测试环境运行 Hadoop 集群 |
| **CI/CD 系统** | 自动化构建和发布 Hadoop 镜像 |

### 构建产物

构建产物是一个**统一的基础镜像**，可用于运行以下 Hadoop 集群组件：

- **NameNode** - HDFS 元数据管理节点
- **DataNode** - HDFS 数据存储节点
- **ResourceManager** - YARN 资源管理器
- **NodeManager** - YARN 节点管理器
- **JournalNode** - HA 模式下的元数据同步节点
- **JobHistoryServer** - MapReduce 作业历史服务器

### 镜像名称对应关系

| 项目 | 默认值 | 说明 |
|------|--------|------|
| **构建默认名称** | `apache/hadoop:3.3.6` | Makefile 中 HADOOP_IMG 的默认值 |
| **部署配置位置** | [config/samples/hadoop_v1_hadoopcluster.yaml](../../config/samples/hadoop_v1_hadoopcluster.yaml) | HadoopCluster CR 的 image 字段 |
| **部署默认名称** | `apache/hadoop:3.3.6` | repository: apache/hadoop, tag: "3.3.6" |

**重要**：构建的镜像名称必须与部署配置中的镜像名称一致。使用私有仓库时，需要同时修改构建命令和部署配置中的 `repository` 和 `tag`。

## 目录结构

```
build/hadoop/
├── Dockerfile              # Hadoop 组件镜像构建文件
├── README.md               # 本文件
├── conf/                   # Hadoop 配置文件模板
│   ├── core-site.xml       # 核心配置（默认文件系统、临时目录等）
│   ├── hdfs-site.xml       # HDFS 配置（块大小、副本数、端口等）
│   ├── yarn-site.xml       # YARN 配置（资源管理、内存设置等）
│   └── mapred-site.xml     # MapReduce 配置
└── scripts/                # 启动脚本
    ├── entrypoint.sh       # 主入口脚本，根据组件类型执行初始化
    └── healthcheck.sh      # 健康检查脚本，支持各组件的健康检测
```

## 快速开始

### 方式一：使用 Makefile（推荐）

```bash
# 从项目根目录执行
cd hadoop-operator

# 构建 Hadoop 镜像（默认版本 3.3.6）
make build-hadoop-image

# 指定版本构建
make build-hadoop-image HADOOP_VERSION=3.3.6

# 多平台构建
make build-hadoop-multiarch PLATFORMS=linux/amd64,linux/arm64

# 构建并推送到私有仓库
make push-images REGISTRY=registry.example.com
```

### 方式二：使用 Docker 直接构建

```bash
# 进入构建目录
cd hadoop-operator/build/hadoop

# 构建本地镜像
docker build -t myregistry/hadoop:3.3.6 .

# 构建指定版本
docker build --build-arg HADOOP_VERSION=3.3.6 -t myregistry/hadoop:3.3.6 .

# 多平台构建并推送
docker buildx build --platform linux/amd64,linux/arm64 \
  -t myregistry/hadoop:3.3.6 \
  --push .
```

## 运行容器

### NameNode

```bash
docker run -d \
  --name namenode \
  -e HADOOP_COMPONENT=namenode \
  -p 9870:9870 -p 9000:9000 \
  -v namenode-data:/opt/hadoop/data/nn \
  myregistry/hadoop:3.3.6 \
  hdfs namenode
```

### DataNode

```bash
docker run -d \
  --name datanode \
  -e HADOOP_COMPONENT=datanode \
  -e NAMENODE_HOST=namenode \
  -p 9864:9864 \
  -v datanode-data:/opt/hadoop/data/dn \
  myregistry/hadoop:3.3.6 \
  hdfs datanode
```

### ResourceManager

```bash
docker run -d \
  --name resourcemanager \
  -e HADOOP_COMPONENT=resourcemanager \
  -p 8088:8088 \
  myregistry/hadoop:3.3.6 \
  yarn resourcemanager
```

### NodeManager

```bash
docker run -d \
  --name nodemanager \
  -e HADOOP_COMPONENT=nodemanager \
  -e RESOURCEMANAGER_HOST=resourcemanager \
  -p 8042:8042 \
  myregistry/hadoop:3.3.6 \
  yarn nodemanager
```

## 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `HADOOP_COMPONENT` | 组件类型（namenode/datanode/resourcemanager/nodemanager/journalnode/historyserver） | - |
| `NAMENODE_HOST` | NameNode 主机名（DataNode 使用） | namenode |
| `NAMENODE_PORT` | NameNode RPC 端口 | 9000 |
| `RESOURCEMANAGER_HOST` | ResourceManager 主机名 | resourcemanager |
| `HADOOP_HEAPSIZE` | JVM 堆内存大小 | 1024 |
| `HADOOP_LOG_DIR` | 日志目录 | /opt/hadoop/logs |

## 暴露端口

### HDFS 端口
- `8020` / `9000` - NameNode RPC
- `9870` - NameNode Web UI
- `9871` - NameNode HTTPS
- `9864` - DataNode Web UI
- `9866` - DataNode 数据传输
- `9867` - DataNode IPC

### YARN 端口
- `8030` - ResourceManager 调度器
- `8031` - ResourceManager 资源跟踪器
- `8032` - ResourceManager RPC
- `8033` - ResourceManager 管理员
- `8088` - ResourceManager Web UI
- `8040` - NodeManager 本地化
- `8041` - NodeManager RPC
- `8042` - NodeManager Web UI

### 其他端口
- `10020` / `19888` - JobHistory Server
- `8485` / `8480` / `8481` - JournalNode（HA 模式）

## 数据持久化

建议将以下目录挂载到持久化存储：

| 目录 | 用途 | 建议挂载组件 |
|------|------|-------------|
| `/opt/hadoop/data/nn` | NameNode 元数据 | NameNode |
| `/opt/hadoop/data/dn` | DataNode 数据 | DataNode |
| `/opt/hadoop/data/jn` | JournalNode 数据 | JournalNode |
| `/opt/hadoop/logs` | 日志文件 | 所有组件 |

## 健康检查

镜像内置健康检查脚本，根据 `HADOOP_COMPONENT` 自动检测对应组件的健康状态。

## 镜像特点

- **统一基础镜像**：一个镜像支持所有 Hadoop 组件
- **官方发行版**：基于 Apache Hadoop 官方发行版
- **多架构支持**：linux/amd64、linux/arm64 等
- **非 root 运行**：使用 hadoop 用户（UID 1000）运行
- **预置配置**：包含优化的 Hadoop 配置文件模板
- **健康检查**：内置各组件健康检测脚本

## 注意事项

1. 首次启动 NameNode 需要格式化，镜像会自动检测并执行
2. DataNode 启动时会等待 NameNode 就绪
3. 建议在生产环境中使用持久化存储挂载数据目录
4. 可以通过挂载自定义配置文件覆盖默认配置
