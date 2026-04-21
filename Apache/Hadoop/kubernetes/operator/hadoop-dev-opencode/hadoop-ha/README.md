# Hadoop HA Cluster Kubernetes Operator

生产环境可用的Hadoop HA集群Kubernetes部署方案，参考Apache Doris优秀实践进行生产级别优化。

## 架构概览

```
┌─────────────────────────────────────────────────────────────┐
│                     Kubernetes Cluster                      │
├─────────────────────────────────────────────────────────────┤
│  Zookeeper (3)  │  JournalNode (3)  │  NameNode (2 HA)     │
├─────────────────────────────────────────────────────────────┤
│  DataNode (3)   │  ResourceManager (2 HA)                  │
├─────────────────────────────────────────────────────────────┤
│  NodeManager (DaemonSet - 每个Worker节点)                   │
└─────────────────────────────────────────────────────────────┘
```

## 新增生产环境特性

本方案基于Apache Doris的生产实践，新增以下生产级特性：

| 特性 | 说明 |
|------|------|
| 亲和性调度 | podAntiAffinity、nodeAffinity，支持高可用和资源优化 |
| 安全上下文 | runAsNonRoot、fsGroup、runAsGroup，确保容器安全运行 |
| 健康检查 | startupProbe + readinessProbe + livenessProbe |
| 生命周期管理 | preStop优雅停机，确保服务平滑关闭 |
| 容忍调度 | tolerations支持节点故障时的服务迁移 |
| 标签与注解 | prometheus.io监控集成 |
| PVC存储 | 所有有状态组件均配置持久化存储 |

## 组件清单

| 组件 | 副本数 | 类型 | 存储 | ServiceAccount |
|------|--------|------|------|----------------|
| Zookeeper | 3 | StatefulSet | PVC (local) | - |
| JournalNode | 3 | StatefulSet | PVC (local) | hadoop |
| NameNode | 2 | StatefulSet (HA) | PVC (local) | hadoop |
| DataNode | 3 | StatefulSet | PVC (local) | hadoop |
| ResourceManager | 2 | StatefulSet (HA) | PVC | hadoop |
| NodeManager | N | DaemonSet | - | hadoop |

## 快速开始

### 1. 部署集群

```bash
# 创建命名空间并部署
kubectl create namespace hadoop-cluster
kubectl apply -f . -n hadoop-cluster

# 或使用部署脚本
./deploy.sh hadoop-cluster
```

### 2. 初始化HDFS HA (首次部署必须执行)

```bash
# 等待所有Pod就绪
kubectl get pods -n hadoop-cluster -w

# 格式化并启动第一个NameNode
kubectl exec -it namenode-0 -n hadoop-cluster -- hdfs namenode -format -force -nonInteractive

# 启动namenode-0
kubectl exec -it namenode-0 -n hadoop-cluster -- hdfs namenode

# 在另一个终端，初始化namenode-1为Standby
kubectl exec -it namenode-1 -n hadoop-cluster -- hdfs namenode -bootstrapStandby

# 初始化Zookeeper Failover Controller
kubectl exec -it namenode-0 -n hadoop-cluster -- hdfs zkfc -formatZK

# 将namenode-0设为Active
kubectl exec -it namenode-0 -n hadoop-cluster -- hdfs haadmin -transitionToActive namenode-0
```

## 配置说明

### 生产环境配置

#### 亲和性调度

各组件默认配置了podAntiAffinity，确保同一组件的Pod分布在不同节点：

```yaml
affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        labelSelector:
          matchExpressions:
          - key: app
            operator: In
            values:
            - hadoop-datanode
        topologyKey: kubernetes.io/hostname
```

Master节点（NameNode/ResourceManager）额外配置nodeAffinity，优先调度到非控制平面节点：

```yaml
nodeSelector:
  hadoop-role: master  # 或 worker
```

#### 安全上下文

所有容器配置了安全上下文，确保符合安全最佳实践：

```yaml
securityContext:
  runAsUser: 1000
  runAsGroup: 1000
  fsGroup: 1000
  runAsNonRoot: true
```

#### 健康检查

各组件配置了三层健康检查：

| 检查类型 | 说明 | 示例 |
|----------|------|------|
| startupProbe | 启动探针，确保应用启动完成 | 30s初始延迟，30次重试 |
| readinessProbe | 就绪探针，检测服务是否可接收流量 | 10-15s间隔 |
| livenessProbe |存活探针，检测应用是否存活 | 30-60s间隔 |

示例配置：
```yaml
startupProbe:
  exec:
    command:
    - sh
    - -c
    - nc -z localhost 9870 || exit 1
  initialDelaySeconds: 30
  periodSeconds: 10
  failureThreshold: 30

readinessProbe:
  httpGet:
    path: /jmx
    port: 9870
  initialDelaySeconds: 40
  periodSeconds: 10
  timeoutSeconds: 5

livenessProbe:
  httpGet:
    path: /jmx
    port: 9870
  initialDelaySeconds: 100
  periodSeconds: 30
```

#### 生命周期管理

配置preStop钩子实现优雅停机：

```yaml
lifecycle:
  preStop:
    exec:
      command:
      - sh
      - -c
      - |
        su hadoop -c "hdfs dfsadmin -safemode leave" || true
        sleep 10
```

#### 容忍调度

配置节点故障时的容忍策略：

```yaml
tolerations:
- key: "node.kubernetes.io/not-ready"
  operator: "Exists"
  effect: "NoSchedule"
  tolerationSeconds: 60
- key: "node.kubernetes.io/unreachable"
  operator: "Exists"
  effect: "NoSchedule"
  tolerationSeconds: 60
```

### 镜像配置

当前使用镜像：
- Hadoop: `zhuyifeiruichuang/hadoop:3.1.1`
- Zookeeper: `zookeeper:3.8.0`

如需修改镜像，编辑各yaml文件中的 `image` 字段。

### 离线环境部署

镜像策略已设为 `IfNotPresent`，支持离线部署。需提前导入镜像：

```bash
# 在有网络的机器上拉取镜像
docker pull zookeeper:3.8.0
docker pull zhuyifeiruichuang/hadoop:3.1.1

# 导出并导入到部署环境
docker save zookeeper:3.8.0 hadoop.tar | ssh target-node 'docker load'
```

### 存储配置

所有StatefulSet已配置 `storageClassName: "local"`。

如需使用其他存储，修改各文件的 `storageClassName`。

### 资源配置

| 组件 | CPU Request | CPU Limit | Memory Request | Memory Limit | 临时存储 | 存储 |
|------|-------------|-----------|----------------|--------------|----------|------|
| Zookeeper | 250m | 500m | 512Mi | 1Gi | - | 10Gi+5Gi |
| JournalNode | 250m | 500m | 512Mi | 1Gi | 500Mi-1Gi | 10Gi |
| NameNode | 1 | 2 | 4Gi | 8Gi | 2-4Gi | 20Gi+10Gi |
| DataNode | 1 | 2 | 4Gi | 8Gi | 2-4Gi | 50Gi |
| ResourceManager | 1 | 2 | 4Gi | 8Gi | 2-4Gi | 10Gi |
| NodeManager | 2 | 4 | 4Gi | 8Gi | 2-4Gi | - |

修改资源限制：编辑各文件中的 `resources.requests` 和 `resources.limits`。

### 副本数调整

| 文件 | 组件 | 调整方式 |
|------|------|----------|
| 01-zookeeper.yaml | Zookeeper | 修改 `replicas: 3` |
| 02-journalnode.yaml | JournalNode | 修改 `replicas: 3` |
| 03-namenode.yaml | NameNode | 修改 `replicas: 2` |
| 04-datanode.yaml | DataNode | 修改 `replicas: 3` |
| 05-resourcemanager.yaml | ResourceManager | 修改 `replicas: 2` |

### ServiceAccount

所有有状态组件配置了 `serviceAccountName: hadoop`，如需配置RBAC，参考以下YAML：

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: hadoop
  namespace: hadoop-cluster
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: hadoop
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: hadoop
subjects:
- kind: ServiceAccount
  name: hadoop
  namespace: hadoop-cluster
roleRef:
  kind: Role
  name: hadoop
  apiGroup: rbac.authorization.k8s.io
```

## 服务访问

### NodePort服务

| 服务 | NodePort | 说明 |
|------|----------|------|
| NameNode Web UI | 30070 | http://node-ip:30070 |
| NameNode RPC | 30090 | - |
| ResourceManager | 30888 | http://node-ip:30888 |
| NodeManager | 30442 | http://node-ip:30442 |
| DataNode Web | 30064 | http://node-ip:30064 |

### 内部DNS

```
namenode-0.namenode.hadoop:9870  (Web)
namenode-0.namenode.hadoop:9000  (RPC)
resourcemanager-0.resourcemanager.hadoop:8088
zookeeper-0.zookeeper.hadoop:2181
journalnode-0.journalnode.hadoop:8485
```

## 运维命令

### 查看状态

```bash
# 查看所有Pod
kubectl get pods -n hadoop-cluster

# 查看服务
kubectl get svc -n hadoop-cluster

# 查看StatefulSet
kubectl get sts -n hadoop-cluster

# 查看DaemonSet (NodeManager)
kubectl get ds -n hadoop-cluster
```

### 故障排查

```bash
# 查看Pod日志
kubectl logs <pod-name> -n hadoop-cluster

# 进入Pod调试
kubectl exec -it <pod-name> -n hadoop-cluster -- /bin/bash

# 查看NameNode状态
kubectl exec -it namenode-0 -n hadoop-cluster -- hdfs haadmin -getServiceState

# 查看HDFS健康状态
kubectl exec -it namenode-0 -n hadoop-cluster -- hdfs dfsadmin -safemode get

# 查看YARN状态
kubectl exec -it resourcemanager-0 -n hadoop-cluster -- yarn rmadmin -getServiceState
```

### 扩缩容

```bash
# 扩容DataNode
kubectl scale sts datanode --replicas=5 -n hadoop-cluster

# 缩容DataNode
kubectl scale sts datanode --replicas=3 -n hadoop-cluster

# 滚动更新镜像
kubectl set image sts/nameNode namenode=newhadoop:3.1.1 -n hadoop-cluster
```

### 滚动更新

```bash
# 查看更新状态
kubectl rollout status sts/namenode -n hadoop-cluster

# 回滚
kubectl rollout undo sts/namenode -n hadoop-cluster
```

### 监控

各组件已配置prometheus.io注解，支持Prometheus监控：

```yaml
annotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "9870"  # 各组件对应端口
```

## 卸载

```bash
# 删除所有资源
kubectl delete -f . -n hadoop-cluster

# 删除命名空间
kubectl delete namespace hadoop-cluster
```

## 高可用说明

### HDFS HA
- 2个NameNode，一主一备
- 通过JournalNode同步元数据
- Zookeeper进行故障自动切换

### YARN HA
- 2个ResourceManager
- 通过Zookeeper进行状态同步和故障切换

### 故障恢复
1. Zookeeper/ZKFC自动检测NameNode故障
2. 自动切换到健康的NameNode
3. JournalNode保证元数据不丢失

## CRD定义

如需使用声明式API管理Hadoop集群，可应用CRD定义（需自行实现Operator）：

```bash
# 应用CRD
kubectl apply -f hadoop-crd.yaml

# 查看CRD
kubectl get crd hadoopclusters.hadoop.apache.org
```

CRD支持以下组件配置：
- nameNodeSpec: 副本数、镜像、资源、亲和性、存储
- dataNodeSpec: 副本数、镜像、资源、亲和性、存储
- resourceManagerSpec: 副本数、镜像、资源、亲和性
- nodeManagerSpec: 副本数、镜像、资源、亲和性

## 注意事项

1. 首次部署必须执行HA初始化步骤
2. 存储建议使用持久化存储，emptyDir在Pod重启后数据会丢失
3. 生产环境建议配置资源限制和请求
4. 确保节点有足够的资源运行所有组件
5. 离线环境需提前导入所有镜像
6. 生产环境建议配置监控告警系统
7. 建议使用Kubernetes 1.20+版本以获得更好的Pod管理能力