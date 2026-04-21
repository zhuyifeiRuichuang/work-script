# Hadoop 高可用集群部署指南

本文档描述如何部署支持高可用性（HA）的生产级 Hadoop 集群。

## 架构概览

```
                      ┌─────────────────────────────────────────────────────────┐
                      │                   ZooKeeper 集群                       │
                      │         (外部部署或嵌入式 - 3 节点)                       │
                      └─────────────────────────────────────────────────────────┘
                                          │
          ┌────────────────────────────────┼────────────────────────────────┐
          │                                │                                │
          ▼                                ▼                                ▼
┌─────────────────┐              ┌─────────────────┐              ┌─────────────────┐
│   NameNode 1    │◄────────────►│  JournalNode    │◄────────────►│   NameNode 2    │
│   (主节点)       │   Edits QJM  │   (3 节点)      │   Edits QJM  │   (备节点)       │
└─────────────────┘              └─────────────────┘              └─────────────────┘
          │                                                                │
          │                        HDFS 客户端                             │
          └────────────────────────────────────────────────────────────────┘

┌─────────────────┐              ┌─────────────────┐              ┌─────────────────┐
│ResourceManager 1│◄────────────►│     ZooKeeper   │◄────────────►│ResourceManager 2│
│   (主节点)       │  ZK 故障转移  │    (领导者)      │  ZK 故障转移  │   (备节点)       │
└─────────────────┘              └─────────────────┘              └─────────────────┘
          │                                                                │
          │                        YARN 客户端                             │
          └────────────────────────────────────────────────────────────────┘
```

## 前置条件

### 1. Kubernetes 集群要求

- Kubernetes 1.24+
- 至少 6 个工作节点（推荐）
- 每个节点最低资源：4 CPU 核心，8GB 内存
- 支持动态供给的 StorageClass

### 2. ZooKeeper 集群（HA 必需）

在部署 Hadoop HA 之前，必须先部署一个可用的 ZooKeeper 集群。

#### 方式一：使用 Helm 部署 ZooKeeper（推荐）

```bash
# 添加 Bitnami 仓库
helm repo add bitnami https://charts.bitnami.com/bitnami

# 安装 ZooKeeper
helm install zookeeper bitnami/zookeeper \
  --set replicaCount=3 \
  --set persistence.enabled=true \
  --set persistence.storageClass=standard \
  --set persistence.size=2Gi
```

#### 方式二：创建 ZooKeeper ConfigMap

ZooKeeper 部署完成后，创建 ConfigMap 存储 ZooKeeper 连接串：

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: hadoopcluster-ha-zk
  namespace: default
data:
  ZOOKEEPER_SERVERS: "zookeeper-0.zookeeper-headless:2181,zookeeper-1.zookeeper-headless:2181,zookeeper-2.zookeeper-headless:2181"
```

## 部署 HA Hadoop 集群

### 方式一：使用 CRD

```bash
# 应用 HA 集群配置
kubectl apply -f config/samples/hadoop_v1alpha1_hadoopcluster-ha.yaml
```

### 方式二：使用 Helm Chart

```bash
# 使用 HA 配置安装
helm install hadoop-ha deploy/helm/hadoop-operator \
  --values deploy/helm/hadoop-operator/examples/values-ha.yaml \
  --set ha.nameNodeHA.enabled=true \
  --set ha.resourceManagerHA.enabled=true \
  --set zookeeper.servers="zookeeper-0.zookeeper-headless:2181,zookeeper-1.zookeeper-headless:2181,zookeeper-2.zookeeper-headless:2181"
```

## HA 配置详解

### NameNode HA

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `spec.ha.nameNodeHA.enabled` | 启用 NameNode HA | `false` |
| `spec.ha.nameNodeHA.nameServiceId` | HDFS 名称服务 ID | 必填 |
| `spec.ha.nameNodeHA.journalClusterId` | Journal 集群 ID | 可选 |
| `spec.ha.nameNodeHA.replicas` | NameNode 数量 | `2` |

生成的配置 (`hdfs-site.xml`):

```xml
<property>
    <name>dfs.nameservices</name>
    <value>ns</value>
</property>
<property>
    <name>dfs.ha.namenodes.ns</name>
    <value>nn1,nn2</value>
</property>
<property>
    <name>dfs.namenode.rpc-address.ns.nn1</name>
    <value>hadoopcluster-ha-namenode-0.hadoopcluster-ha-namenode:9000</value>
</property>
<property>
    <name>dfs.namenode.shared.edits.dir</name>
    <value>qjournal://hadoopcluster-ha-journalnode-0.hadoopcluster-ha-journalnode:8485;.../ns</value>
</property>
<property>
    <name>dfs.ha.automatic-failover.enabled</name>
    <value>true</value>
</property>
```

### ResourceManager HA

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `spec.ha.resourceManagerHA.enabled` | 启用 ResourceManager HA | `false` |
| `spec.ha.resourceManagerHA.clusterId` | YARN 集群 ID | `rm-cluster` |
| `spec.ha.resourceManagerHA.replicas` | ResourceManager 数量 | `2` |

生成的配置 (`yarn-site.xml`):

```xml
<property>
    <name>yarn.resourcemanager.ha.enabled</name>
    <value>true</value>
</property>
<property>
    <name>yarn.resourcemanager.ha.rm-ids</name>
    <value>rm1,rm2</value>
</property>
<property>
    <name>yarn.resourcemanager.cluster-id</name>
    <value>rm-cluster</value>
</property>
<property>
    <name>yarn.resourcemanager.zk-address</name>
    <value>zookeeper:2181</value>
</property>
<property>
    <name>yarn.resourcemanager.ha.automatic-failover.enabled</name>
    <value>true</value>
</property>
```

## 验证 HA 部署

### 1. 检查集群状态

```bash
kubectl get hadoopcluster
kubectl describe hadoopcluster hadoopcluster-ha
```

### 2. 验证 NameNode HA

```bash
# 获取 NameNode Pod
kubectl get pods -l hadoop.kubedoop.dev/component=namenode

# 检查 HDFS HA 状态（进入主 NameNode）
kubectl exec -it hadoopcluster-ha-namenode-0 -- hdfs haadmin -ns ns -getServiceState

# 预期输出:
# namenode-0 active
# namenode-1 standby
```

### 3. 验证 ResourceManager HA

```bash
# 获取 ResourceManager Pod
kubectl get pods -l hadoop.kubedoop.dev/component=resourcemanager

# 检查 YARN HA 状态（进入主 ResourceManager）
kubectl exec -it hadoopcluster-ha-resourcemanager-rm1 -- yarn rmadmin -getServiceState rm1

# 预期输出:
# rm1 active
# rm2 standby
```

### 4. 测试自动故障转移

#### NameNode 故障转移测试

```bash
# 模拟 NameNode 故障
kubectl exec -it hadoopcluster-ha-namenode-0 -- kill 1

# 等待故障转移并检查状态
sleep 30
kubectl exec -it hadoopcluster-ha-namenode-1 -- hdfs haadmin -ns ns -getServiceState
```

#### ResourceManager 故障转移测试

```bash
# 模拟 ResourceManager 故障
kubectl delete pod hadoopcluster-ha-resourcemanager-rm1

# 等待故障转移并检查状态
kubectl exec -it hadoopcluster-ha-resourcemanager-rm2 -- yarn rmadmin -getServiceState rm2
```

## 访问 HA 服务

### HDFS Web UI

```
http://hadoopcluster-ha-namenode-0.hadoopcluster-ha-namenode:9870
http://hadoopcluster-ha-namenode-1.hadoopcluster-ha-namenode:9870
```

### YARN Web UI

```
http://hadoopcluster-ha-resourcemanager-rm1.hadoopcluster-ha-resourcemanager:8088
http://hadoopcluster-ha-resourcemanager-rm2.hadoopcluster-ha-resourcemanager:8088
```

### 使用 HA 代理（推荐客户端使用）

创建服务代理到主 ResourceManager：

```yaml
apiVersion: v1
kind: Service
metadata:
  name: yarn-ui-proxy
spec:
  selector:
    hadoop.kubedoop.dev/component: resourcemanager
    hadoop.kubedoop.dev/ha: "true"
  ports:
  - port: 8088
    targetPort: 8088
```

## 故障排除

### 问题：ZooKeeper 连接失败

**症状**：NameNode 或 ResourceManager 启动失败，提示 ZooKeeper 连接错误。

**解决方案**：
1. 确认 ZooKeeper 正在运行：`kubectl get pods -l app.kubernetes.io/name=zookeeper`
2. 检查 ConfigMap 中的 ZooKeeper 地址是否正确
3. 更新 `spec.clusterConfig.zooKeeperConfigMapName`

### 问题：NameNode 无法退出安全模式

**症状**：HA 配置后 NameNode 一直停留在安全模式。

**解决方案**：
1. 确认 JournalNode 正在运行：`kubectl get pods -l hadoop.kubedoop.dev/component=journalnode`
2. 等待 JournalNode 同步 edits
3. 执行：`kubectl exec -it <active-namenode> -- hdfs dfsadmin -safemode leave`

### 问题：ResourceManager 无法发现其他 RM

**症状**：ResourceManager 无法转换为主/备状态。

**解决方案**：
1. 确认两个 ResourceManager Pod 都在运行
2. 检查 Pod 内的 ZooKeeper 访问权限
3. 确认 HA 模式下 Service 是 Headless（ClusterIP: None）

## HA 组件要求

| 组件 | HA 必须 | 最小副本数 | 备注 |
|------|---------|-----------|------|
| NameNode | 是 | 2 | 必须是偶数 |
| JournalNode | 是 | 3 | 必须是奇数 |
| DataNode | 否 | 1+ | 副本数应大于复制因子 |
| ResourceManager | 是 | 2 | 必须是偶数 |
| NodeManager | 否 | 1+ | 取决于工作负载 |

## 清理

```bash
# 删除 HA 集群
kubectl delete -f config/samples/hadoop_v1alpha1_hadoopcluster-ha.yaml

# 删除 ZooKeeper（如使用 Helm 安装）
helm uninstall zookeeper
```
