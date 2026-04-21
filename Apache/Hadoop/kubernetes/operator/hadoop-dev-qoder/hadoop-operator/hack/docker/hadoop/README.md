# Apache Hadoop Docker 镜像

本目录包含 Apache Hadoop 集群组件的 Docker 镜像构建文件。

## 概述

这是一个统一的基础镜像，支持运行 Hadoop 集群的所有组件：

- **NameNode** - HDFS 元数据管理
- **DataNode** - HDFS 数据存储
- **ResourceManager** - YARN 资源管理
- **NodeManager** - YARN 节点管理
- **JournalNode** - HA 模式下的元数据同步
- **JobHistoryServer** - MapReduce 作业历史服务

## 镜像特点

- 基于 Eclipse Temurin JRE 11（Alpine 版本）
- 使用 Apache Hadoop 官方发行版
- 支持多架构（amd64、arm64、s390x、ppc64le）
- 包含健康检查脚本
- 非 root 用户运行（hadoop，UID 1000）
- 预配置 Hadoop 配置文件

## 目录结构

```
hack/docker/hadoop/
├── Dockerfile              # 镜像构建文件
├── README.md               # 本文件
├── conf/                   # Hadoop 配置文件模板
│   ├── core-site.xml       # 核心配置
│   ├── hdfs-site.xml       # HDFS 配置
│   ├── yarn-site.xml       # YARN 配置
│   └── mapred-site.xml     # MapReduce 配置
└── scripts/                # 启动脚本
    ├── entrypoint.sh       # 主入口脚本
    └── healthcheck.sh      # 健康检查脚本
```

## 快速开始

### 构建镜像

```bash
# 从项目根目录执行
cd hadoop-operator

# 使用 Makefile 构建
make build-hadoop-image

# 指定版本构建
make build-hadoop-image HADOOP_VERSION=3.3.6

# 多平台构建
make build-hadoop-multiarch PLATFORMS=linux/amd64,linux/arm64

# 构建并推送
make push-images REGISTRY=registry.example.com
```

### 直接构建

```bash
# 构建本地镜像
docker build -t apache/hadoop:3.3.6 .

# 构建指定版本
docker build --build-arg HADOOP_VERSION=3.3.6 -t apache/hadoop:3.3.6 .
```

## 运行容器

### NameNode

```bash
docker run -d \
  --name namenode \
  -e HADOOP_COMPONENT=namenode \
  -p 9870:9870 \
  -p 9000:9000 \
  -v namenode-data:/opt/hadoop/data/nn \
  apache/hadoop:3.3.6 \
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
  apache/hadoop:3.3.6 \
  hdfs datanode
```

### ResourceManager

```bash
docker run -d \
  --name resourcemanager \
  -e HADOOP_COMPONENT=resourcemanager \
  -p 8088:8088 \
  -p 8032:8032 \
  apache/hadoop:3.3.6 \
  yarn resourcemanager
```

### NodeManager

```bash
docker run -d \
  --name nodemanager \
  -e HADOOP_COMPONENT=nodemanager \
  -e RESOURCEMANAGER_HOST=resourcemanager \
  -p 8042:8042 \
  apache/hadoop:3.3.6 \
  yarn nodemanager
```

## 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `HADOOP_COMPONENT` | 组件类型（namenode/datanode/resourcemanager/nodemanager/journalnode） | - |
| `NAMENODE_HOST` | NameNode 主机名（DataNode 使用） | namenode |
| `NAMENODE_PORT` | NameNode RPC 端口 | 9000 |
| `RESOURCEMANAGER_HOST` | ResourceManager 主机名 | resourcemanager |
| `HADOOP_HEAPSIZE` | JVM 堆内存大小 | 1024 |
| `HADOOP_LOG_DIR` | 日志目录 | /opt/hadoop/logs |

## 暴露端口

### HDFS 端口
- `8020` - NameNode RPC（默认）
- `9000` - NameNode RPC（常用）
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

### MapReduce 端口
- `10020` - JobHistory RPC
- `19888` - JobHistory Web UI

### JournalNode 端口（HA）
- `8485` - JournalNode RPC
- `8480` - JournalNode HTTP
- `8481` - JournalNode HTTPS

## 健康检查

镜像内置健康检查脚本，根据 `HADOOP_COMPONENT` 自动检测对应组件的健康状态：

```bash
# NameNode - 检查 Web UI 和 RPC 端口
curl -sf http://localhost:9870/jmx

# DataNode - 检查 Web UI
curl -sf http://localhost:9864/

# ResourceManager - 检查 Web UI
curl -sf http://localhost:8088/ws/v1/cluster/info

# NodeManager - 检查 Web UI
curl -sf http://localhost:8042/ws/v1/node/info
```

## 数据持久化

建议将以下目录挂载到持久化存储：

- `/opt/hadoop/data/nn` - NameNode 元数据
- `/opt/hadoop/data/dn` - DataNode 数据
- `/opt/hadoop/data/jn` - JournalNode 数据（HA 模式）
- `/opt/hadoop/logs` - 日志文件

## 配置文件

镜像包含预配置的 Hadoop 配置文件，位于 `/opt/hadoop/etc/hadoop/`：

- `core-site.xml` - 核心配置
- `hdfs-site.xml` - HDFS 配置
- `yarn-site.xml` - YARN 配置
- `mapred-site.xml` - MapReduce 配置

可以通过挂载自定义配置文件或环境变量覆盖配置。

## 安全说明

- 容器以非 root 用户 `hadoop`（UID 1000）运行
- 使用 distroless 风格的 Alpine 基础镜像
- 仅包含运行 Hadoop 所需的最小依赖
- 建议在生产环境中启用 Kerberos 认证

## 许可证

Apache License 2.0
