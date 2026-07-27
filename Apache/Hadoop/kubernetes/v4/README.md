# Hadoop 3.1.1 on Kubernetes — 生产级 HA 部署与运维指南

> 版本：**v4（NameNode HA + YARN HA + 全栈生产验证）**
> 日期：2026-07-24
> 开发者：腾讯workbuddy+腾讯hy3大模型
> 实际部署命名空间：**`hadoop`**（本文命令统一用 `-n hadoop`；**要部署到其他命名空间，必须改 configmap 里的 FQDN + web-ui-access 的 namespace 字段，详见 §16，不能只改 `-n`**）
> 架构：NN HA（Active/Standby + JournalNode + ZooKeeper + ZKFC） + **YARN HA（双 ResourceManager + ZKRMStateStore）** + 3× DataNode + 3× NodeManager
> **验证状态**：NN HA、YARN HA、K8s 三必改坑、Web UI 访问、以及「杀 active RM 运行中作业零中断」的真 failover 均已实测通过（见第 9 节）。
> **缺陷**：resourcemanager和nodemanager的web界面偶发异常，受限主角色随机分配pod，仅web界面展示功能偶发无法正常访问。

---

## 1. 架构概述

### 1.1 组件清单

| 组件 | 数量 | 存储 | 作用 |
|------|------|------|------|
| **NameNode** | 2 Pod（Active + Standby） | openebs-hostpath 50Gi | HDFS 元数据，ZKFC 自动故障切换 |
| **JournalNode** | 3 Pod | openebs-hostpath 20Gi | 共享编辑日志仲裁 |
| **ZooKeeper** | 3 Pod | openebs-hostpath 10Gi | ZKFC leader 选举 + RM 状态存储 |
| **ZKFC** | 2 sidecar（NN Pod 内） | 无 | 监控 NN 健康，触发切换 |
| **DataNode** | 3 Pod | openebs-hostpath 100Gi | HDFS 数据，3 副本 |
| **ResourceManager** | **2 Pod（HA）** | 无 PVC | YARN 调度，ZKRMStateStore 状态恢复 |
| **NodeManager** | 3 Pod | emptyDir | YARN 任务执行 |

### 1.2 HA 核心原理

**存储层（NN HA）**
```
Active NN 写 edits → JournalNode 仲裁(3/3 多数)
Standby NN 从 JN 实时 tail → 元数据一致
ZKFC 经 ZooKeeper 抢 Active 锁 → 节点故障 Active 释放锁 → Standby 秒级接管（~30s，零数据丢失）
```

**计算层（YARN HA）**
```
两个 RM（rm1=resourcemanager-0, rm2=resourcemanager-1）
  → 经 ZooKeeper ActiveStandbyElector 选主，仅一个为 Active
  → RM 状态（app/attempt/令牌）存 ZKRMStateStore（与选主同源=单一真相）
  → 杀 Active RM Pod → 原 Standby 秒级升 Active → 运行中作业靠 ZKRMStateStore 恢复 attempt 继续（实测通过，见第 9 节）
```

### 1.3 存储策略

全部使用 `openebs-hostpath`（local 存储），**无需 Longhorn 等网络存储**。HA 在应用层解决节点故障，不依赖存储层跨节点复制。

---

## 2. 前置条件

- Kubernetes ≥ 1.20
- 至少 3 个 K8s 节点（NN 硬反亲和需 2 否、DN/NM 硬反亲和需 3 节点）
- `openebs-hostpath` StorageClass 已安装并设为默认
- 镜像：
  - Hadoop：`swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/zhuyifeiruichuang/hadoop:3.1.1`
  - ZooKeeper：`swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/library/zookeeper:3.4.14`

---

## 3. 文件清单

| 文件 | 作用 |
|------|------|
| `namespace.yaml` | 命名空间（本集群实际为 `hadoop`） |
| `configmap.yaml` | 全部 Hadoop 配置（NN HA + YARN HA + 三必改坑配置） |
| `zookeeper.yaml` | ZK 3 节点 StatefulSet + PDB |
| `journalnode.yaml` | JN 3 节点 StatefulSet + PDB |
| `namenode.yaml` | NN 2 节点 HA + ZKFC sidecar + PDB |
| `datanode.yaml` | DN 3 节点 + 反亲和 + PDB |
| `resourcemanager.yaml` | **RM 2 Pod HA**（ZKRMStateStore）+ PDB |
| `nodemanager.yaml` | NM 3 节点 + 反亲和 + PDB（tcpSocket 探针） |
| `web-ui-access.yaml` | 每 Pod 独立 NodePort 服务（随机端口，确定性访问 Web UI） |

---

## 4. 部署顺序与步骤

```
① namespace → ② configmap → ③ zookeeper → ④ journalnode → ⑤ namenode → ⑥ datanode → ⑦ resourcemanager → ⑧ nodemanager
```
> ZK 和 JN **必须在 NN 之前**。本集群已部署于 `hadoop`，无需重复建 namespace。

```bash
kubectl apply -f namespace.yaml                                  # 仅首次 / 切换到新 ns 时
kubectl apply -f configmap.yaml -n hadoop
kubectl apply -f zookeeper.yaml -n hadoop
kubectl apply -f journalnode.yaml -n hadoop
kubectl apply -f namenode.yaml -n hadoop
kubectl apply -f datanode.yaml -n hadoop
kubectl apply -f resourcemanager.yaml -n hadoop                 # RM 2 Pod HA
kubectl apply -f nodemanager.yaml -n hadoop
kubectl apply -f web-ui-access.yaml -n hadoop                   # 可选：Web UI 确定性访问
```

等待就绪：
```bash
kubectl wait --for=condition=ready pod -l app=hadoop-namenode -n hadoop --timeout=300s
kubectl wait --for=condition=ready pod -l app=hadoop-datanode -n hadoop --timeout=180s
kubectl wait --for=condition=ready pod -l app=hadoop-resourcemanager -n hadoop --timeout=180s
kubectl wait --for=condition=ready pod -l app=hadoop-nodemanager -n hadoop --timeout=180s
```

---

## 5. NameNode HA 初始化序列

### 5.1 namenode-0（nn1/Active）
```
initContainer：mkdir 数据目录 → 等 ZK 3/3 → 等 JN ≥2/3 → formatZK → format NN → initializeSharedEdits
主容器：NameNode daemon → ZKFC 获 Active 锁 → nn1 成为 Active
```

### 5.2 namenode-1（nn2/Standby）
```
initContainer：等 ZK/JN → 等 namenode-0 RPC → bootstrapStandby（须用 root，见踩坑）
主容器：NameNode daemon 从 JN 同步 → ZKFC 发现锁被持 → nn2 保持 Standby
```

### 5.3 首次 vs 后续重启
| 步骤 | 首次 | 后续重启 |
|------|------|---------|
| formatZK / format NN / initializeSharedEdits / bootstrapStandby | 执行 | **跳过**（initContainer 靠 VERSION 与 znode 判断，幂等） |

---

## 6. 基础验证

```bash
# NN HA 状态
kubectl exec -n hadoop namenode-0 -c namenode -- hdfs haadmin -getServiceState nn1   # active
kubectl exec -n hadoop namenode-1 -c namenode -- hdfs haadmin -getServiceState nn2   # standby
# HDFS
kubectl exec -n hadoop namenode-0 -c namenode -- hdfs dfsadmin -report
# YARN 节点
kubectl exec -n hadoop resourcemanager-0 -c resourcemanager -- yarn node -list        # 3 个 NM
```

---

## 7. YARN HA 专属三必改（K8s 部署踩坑，缺一不可）

Hadoop 3.1.1 在 K8s 上跑通 MapReduce，必须解决以下三道 K8s 特有坑。任一道不修，作业都会失败且报错极具迷惑性。

### 7.1 NM 注册必须用 FQDN（去掉 `127.0.0.1` 绑定）

**现象**：NM 注册成 `localhost:<port>`，AM 无法路由 → 作业卡住或失败。
**根因**：NM 的 `InetAddress.getLocalHost()` 把 `127.0.0.1` 反解成 `localhost`。若在 `hadoop-env`/`yarn-env` 里写 `127.0.0.1 $(hostname)`，NM 就以 localhost 注册。
**修复**（`nodemanager.yaml`）：
- 启动命令中**去掉** `127.0.0.1 $(hostname` 片段；
- 容器 `nsswitch.conf` 保持 `hosts: files dns`，依赖 K8s 注入的 `<PodIP> <hostname>.<svc>` 让 NM 注册 FQDN（如 `nodemanager-2.nodemanager.hadoop.svc.cluster.local:43167`）。
- RM 主容器**可保留** `127.0.0.1 $(hostname)`（RM 用配置的 `yarn.resourcemanager.hostname.rmX` FQDN，不依赖 getLocalHost，无害）。

### 7.2 AM 的 log4j 必须进 classpath

**现象**：AM 崩溃，client 的 Diagnostics 只显示 `log4j:WARN No appenders`，真实异常被静默吞。
**根因**：AM 子进程 `HADOOP_CONF_DIR` 缺少 `log4j.properties`，异常无 appender。
**修复**（三处协同）：
1. `configmap.yaml` 增加 `log4j.properties` key（带 ConsoleAppender 打到 stderr）；
2. `nodemanager.yaml` 增加 volumeMount（subPath 挂载 `log4j.properties` 到 `/opt/hadoop/etc/hadoop/log4j.properties`）；
3. `mapred-site.xml` 增加 `yarn.app.mapreduce.am.command-opts = -Xmx1024m -Dlog4j.configuration=file:///opt/hadoop/etc/hadoop/log4j.properties`。
> **subPath 不热更新**：改 configmap 后必须 `kubectl rollout restart statefulset/nodemanager` 才生效。

### 7.3 AM Web 服务在 HA 下必须补 per-RM-id web 地址（最隐蔽的 NPE）

**现象**：AM 启动即 `java.lang.NullPointerException` → `RMCommunicator.register` → `MRClientService.getHttpPort()`，作业 FAILED。
**源码级根因**（Hadoop 3.1.1）：
```
ERROR client.MRClientService: Webapps failed to start. Ignoring for now:
java.lang.NullPointerException
  at org.apache.hadoop.util.StringUtils.join(StringUtils.java:941)
  at org.apache.hadoop.yarn.server.webproxy.amfilter.AmFilterInitializer.initFilter(AmFilterInitializer.java:74)
  at org.apache.hadoop.http.HttpServer2.initializeWebServer(HttpServer2.java:587)
```
AM 的 `AmFilterInitializer` 取 **RM 代理主机列表**时，在 HA 模式下走 per-RM-id 地址 `yarn.resourcemanager.webapp.address.rm1/.rm2`。**只设 `hostname.rm1/.rm2` 而没设 `webapp.address.rm1/.rm2` → 列表返回 null → `StringUtils.join(null)` NPE → `webApp` 留 null → `getHttpPort()` 崩溃**。
> 注：`yarn.app.mapreduce.am.webapp.address=0.0.0.0:0`（AM 自身绑 0.0.0.0 临时端口）**方向正确但非本 NPE 解药**，保留即可；真正的修复是补 per-RM-id web 地址。

**修复**（`configmap.yaml` → `yarn-site.xml`）：
```xml
<property>
  <name>yarn.resourcemanager.webapp.address</name>
  <value>resourcemanager-0.resourcemanager.hadoop.svc.cluster.local:8088,resourcemanager-1.resourcemanager.hadoop.svc.cluster.local:8088</value>
  <description>base 读取路径，逗号分隔两 RM，保非 null</description>
</property>
<property>
  <name>yarn.resourcemanager.webapp.address.rm1</name>
  <value>resourcemanager-0.resourcemanager.hadoop.svc.cluster.local:8088</value>
</property>
<property>
  <name>yarn.resourcemanager.webapp.address.rm2</name>
  <value>resourcemanager-1.resourcemanager.hadoop.svc.cluster.local:8088</value>
</property>
```

### 7.4 配套坑（同样会卡作业）

- **NM 探针必须 `tcpSocket` 不是 `httpGet`**：Hadoop 3.1.1 NM REST `/ws/v1/nodeinfo` 返回 404，kubelet 会把 NM 杀进 restart loop。所有 NM 探针用 `tcpSocket:{port:8042}`。RM `/ws/v1/cluster/info:8088` httpGet 正常（active/standby 都返回 200）。
- **改 configmap 后必须 `rollout restart` RM + NM**：subPath 挂载不随 ConfigMap 变更热更新。修改 `yarn-site.xml`/`mapred-site.xml` 后：`kubectl apply -f configmap.yaml -n hadoop && kubectl rollout restart statefulset/resourcemanager -n hadoop && kubectl rollout restart statefulset/nodemanager -n hadoop`。
- **`yarn jar` glob 需 `sh -c` 包裹**：`kubectl exec pod -- yarn jar /p/*.jar` 不生成 shell，`*.jar` 被字面传递 → 用 `sh -c 'yarn jar /opt/hadoop/share/hadoop/mapreduce/hadoop-mapreduce-examples-*.jar pi 2 10'`。
- **AM 真因定位用 `yarn logs` 而非 Diagnostics**：client Diagnostics 只贴 "Last 4096 bytes of stderr"，会把更早的 `Webapps failed to start` 截断藏掉。`yarn logs -applicationId <id>` 取全量，再 `grep -A40 'Webapps failed to start'` 看完整 Caused by。

---

## 8. Web UI 访问（确定性）

### 8.1 问题
默认的 `resourcemanager-external`(30088) / `nodemanager-external`(31799) 用 StatefulSet 级 selector（`app: hadoop-*`），**负载均衡到全部 Pod**。HA 下访问 30088 有一半概率命中 standby（节点表空），访问 31799 每次刷新跳不同 NM → "看不到 node"。

### 8.2 方案 A：per-pod NodePort（随机端口）— `web-ui-access.yaml`
每个 Pod 用 `statefulset.kubernetes.io/pod-name` 标签建独立 NodePort 服务，**NodePort 不指定由 K8s 随机分配**（30000-32767）。应用后查端口：
```bash
kubectl apply -f web-ui-access.yaml -n hadoop
kubectl get svc -n hadoop -l app=hadoop-webui        # 看自动分配的 NodePort
# 浏览器：http://<worker-ip>:<rm-port>/cluster/nodes   （用 `yarn rmadmin -getServiceState` 选 active 对应端口）
#        http://<worker-ip>:<nm-port>/node
```

### 8.3 方案 B：port-forward（推荐，零改动）
```bash
kubectl port-forward -n hadoop pod/resourcemanager-<active> 8088:8088   # 本地开 http://localhost:8088
kubectl port-forward -n hadoop pod/nodemanager-0 8042:8042             # http://localhost:8042/node
```

---

## 9. YARN 真 failover 验证（HA 毕业考）

**目标**：杀掉 active RM，证 standby 秒级接管 + 运行中作业零中断。

```bash
# 1. 先确认谁是 active（一次只一个 serviceId）
kubectl exec -n hadoop resourcemanager-0 -c resourcemanager -- yarn rmadmin -getServiceState rm1
kubectl exec -n hadoop resourcemanager-0 -c resourcemanager -- yarn rmadmin -getServiceState rm2

# 2. 后台启一个长作业（留时间杀 RM）
kubectl exec -n hadoop resourcemanager-0 -c resourcemanager -- \
  sh -c 'yarn jar /opt/hadoop/share/hadoop/mapreduce/hadoop-mapreduce-examples-*.jar pi 16 10000' &

# 3. 作业跑到 map 阶段时，删【active】那个 Pod（下例假设 rm2 active）
kubectl delete pod resourcemanager-1 -n hadoop

# 4. 观察翻转 + 作业存活
kubectl exec -n hadoop resourcemanager-0 -c resourcemanager -- yarn rmadmin -getServiceState rm1   # → active
kubectl exec -n hadoop resourcemanager-0 -c resourcemanager -- yarn rmadmin -getServiceState rm2   # → standby
kubectl exec -n hadoop resourcemanager-0 -c resourcemanager -- yarn application -status application_xxxx_xxxx
```

**实测结果（2026-07-24）**：
- 杀 active(rm2) 时作业 `map 25%`；
- `rm1` 翻为 `active`，`rm2` 重建后以 `standby` 归队（`RESTARTS 0`）；
- 作业 `State: RUNNING` → 继续跑到 `State: FINISHED / Final-State: SUCCEEDED`，`Diagnostics: Attempt recovered after RM restart`；
- `Estimated value of Pi is 3.14127500000000000000`（16 万样本，证明蒙特卡洛正确，早先 3.8 仅为 20 样本噪声）。

**关键结论**：早先"RM 重启后集群空、需手动 `restart nodemanager`"是 **同时冷重启两个 RM** 的副作用；**干净 failover（standby 是热的、只杀一个 active）下 NM 自动向新 active 重注册，作业零中断**。HA 本身完全正确。

---

## 10. 节点故障恢复

| 组件 | 自动恢复 | 备注 |
|------|---------|------|
| NameNode | ZKFC 释放锁 → Standby 升 Active(~30s)；PVC 在故障节点则重建后从 JN 同步 | 数据零丢失 |
| DataNode | 3 副本保护，NN 触发补齐 | 永久故障：删 PVC + 缩/扩容 StatefulSet |
| JournalNode | 仲裁 2/3 多数可用 | 永久故障：删 PVC + 重建 |
| ZooKeeper | 仲裁 2/3 多数可用 | 同上 |
| ResourceManager | **HA 自动 failover**（见第 9 节） | 运行中作业不中断 |
| NodeManager | NM 向新 active RM 自动重注册 | 干净 failover 无需手动重启 |

---

## 11. 遗留写死地址清理（HA 修正）

`configmap.yaml` 中原写死 `resourcemanager-0` 单 Pod 的两项，failover 后会失效，已改为 RM headless 服务 DNS：
```xml
<!-- yarn-site.xml -->
<property>
  <name>yarn.log.server.url</name>
  <value>http://resourcemanager.hadoop.svc.cluster.local:19888/jobhistory</value>
</property>
<!-- mapred-site.xml -->
<property>
  <name>mapreduce.jobtracker.address</name>
  <value>resourcemanager.hadoop.svc.cluster.local:8032</value>
</property>
```
- `yarn.log.server.url`：RM/NM Web UI 的"历史"链接目标（JobHistoryServer，标准端口 19888）。原写死单 Pod → 变 standby/被删后链接失效；改 headless 服务 DNS 后稳定。完整历史需单独部署 JobHistoryServer。
- `mapreduce.jobtracker.address`：MRv1 遗留属性，YARN 下普通作业不依赖；改 headless 服务 DNS 避免 legacy 路径在单 Pod 死亡后连不上。
- **生效**：`kubectl apply -f configmap.yaml -n hadoop && kubectl rollout restart statefulset/resourcemanager -n hadoop && kubectl rollout restart statefulset/nodemanager -n hadoop`。

---

## 12. 运维

- **扩缩容**：`kubectl scale statefulset datanode -n hadoop --replicas=N`（≥3）；`kubectl scale statefulset nodemanager -n hadoop --replicas=N`。
- **滚动更新**：NN 用 `OrderedReady`（nn0 先 HA 初始化）；ZK/JN/DN/NM/RM 用 `Parallel`。
- **NN 元数据备份**：`kubectl exec -n hadoop namenode-0 -c namenode -- hdfs dfsadmin -fetchImage /tmp/fsimage.img`。
- **查看 HA 状态**：`kubectl exec -n hadoop namenode-0 -c namenode -- hdfs haadmin -getAllServiceState`。

---

## 13. 故障排查

| 症状 | 根因 | 解决 |
|------|------|------|
| NM 一直 Error/restart loop | NM 探针 `httpGet /ws/v1/nodeinfo` 返回 404 | 改用 `tcpSocket:8042` |
| 作业 AM `getHttpPort()` NPE | 缺 `yarn.resourcemanager.webapp.address.rm1/.rm2` | 补 per-RM-id web 地址（§7.3） |
| AM 崩溃只见 `No appenders` | AM 缺 log4j.properties | §7.2 三处协同修复 |
| 作业卡 ACCEPTED / cluster resource empty | RM 重启丢失 NM 注册（双 RM 冷重启才触发） | `rollout restart statefulset/nodemanager` |
| `JAR does not exist: *.jar` | `kubectl exec` 无 shell，glob 字面传参 | `sh -c 'yarn jar /p/*.jar ...'` |
| 两个 NN 都 standby | ZKFC 异常 | 看 zkfc 日志，`hdfs haadmin -transitionToActive nn1` |
| NN bootstrapStandby 报 superuser 拒绝 | init 用 `su hadoop` 降级 | init 全程用 root 跑 hdfs 命令 |

---

## 14. 安全加固建议

1. HDFS 权限 `dfs.permissions.enabled=true`，超管组 `hadoop`
2. ACL：`dfs.namenode.acls.enabled=true`、`yarn.acl.enable=true`
3. 容器安全：建议 `runAsUser:1000`（hadoop）替代 `runAsUser:0`，去 `privileged:true`
4. NetworkPolicy 限制 Pod 间访问
5. 生产建议启用 Kerberos、HDFS 传输加密

---

## 15. 验证清单

| # | 验证项 | 命令 | 期望 |
|---|--------|------|------|
| 1 | ZK 仲裁 | `kubectl get pods -l app=zookeeper -n hadoop` | 3/3 Running |
| 2 | JN 仲裁 | `kubectl get pods -l app=hadoop-journalnode -n hadoop` | 3/3 Running |
| 3 | NN HA | `kubectl get pods -l app=hadoop-namenode -n hadoop` | 2/2 Running |
| 4 | nn1 Active / nn2 Standby | `hdfs haadmin -getServiceState nn1/nn2` | active / standby |
| 5 | DN | `kubectl get pods -l app=hadoop-datanode -n hadoop` | 3/3 Running |
| 6 | HDFS 安全模式 | `hdfs dfsadmin -safemode get` | OFF |
| 7 | **RM HA** | `kubectl get pods -l app=hadoop-resourcemanager -n hadoop` | **2/2 Running** |
| 8 | **RM 角色** | `yarn rmadmin -getServiceState rm1/rm2` | 一 active 一 standby |
| 9 | NM | `kubectl get pods -l app=hadoop-nodemanager -n hadoop` | 3/3 Running |
| 10 | YARN 节点 | `yarn node -list` | 3 个 NM |
| 11 | **MapReduce 作业** | `yarn jar ... pi 2 10` | `Job Finished` + `SUCCEEDED` |
| 12 | **YARN 真 failover** | 杀 active RM + `yarn application -status` | standby→active，作业 `SUCCEEDED` |
| 13 | Web UI 可访问 | `kubectl get svc -n hadoop -l app=hadoop-webui` | 各 Pod 独立 NodePort |

---

## 16. 部署到不同命名空间（namespace）须知

### 16.1 核心原理（必读）

K8s 集群内 DNS 解析规则是 `<service>.<namespace>.svc.cluster.local`。**namespace 一旦改变，所有写死在配置里的 FQDN 都会指向一个不存在的 DNS 名字**，服务发现集体失效——这是最致命也最易被忽略的一点（也是本集群最早一版从 `hadoop1` 迁 `hadoop` 时踩过的坑）。

本套配置里 namespace 出现在**两个层面**，改动量天差地别：

| 层面 | 位置 | 改动难度 |
|------|------|----------|
| CLI / 文档层 | 所有 `kubectl ... -n hadoop` 命令、`namespace.yaml` 的 `name` | 易（全局替换） |
| **配置层（致命）** | `configmap.yaml` 内 14 处 `*.hadoop.svc.cluster.local` FQDN + `web-ui-access.yaml` 内 5 处 `namespace: hadoop` | **必须逐处改，否则集群假死** |

> ⚠️ 各 StatefulSet YAML（zookeeper / journalnode / namenode / datanode / resourcemanager / nodemanager）**不写死 namespace 和 FQDN**，靠 `kubectl apply -n <ns>` 注入 + 引用 configmap。所以只要 `-n` 正确，它们自动落在目标 ns，无需改文件内容。

### 16.2 必须修改的配置清单

**① `namespace.yaml`**
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: <新ns>        # 原 hadoop → 改为目标命名空间
```

**② `configmap.yaml` — 14 处 FQDN 全部替换**
把 `.hadoop.svc.cluster.local` → `.<新ns>.svc.cluster.local`。覆盖以下服务（节选关键项）：

- **ZooKeeper**：`zookeeper-0/1/2.zookeeper.hadoop.svc.cluster.local:2181`（`ha.zookeeper.quorum`、`yarn.resourcemanager.zk-address`）
- **JournalNode**：`qjournal://journalnode-0/1/2.journalnode.hadoop.svc.cluster.local:8485/mycluster`（`dfs.namenode.shared.edits.dir`）
- **NameNode**：`namenode-0/1.namenode.hadoop.svc.cluster.local:9000`（rpc）、`:9870`（http）
- **ResourceManager**：`resourcemanager-0/1.resourcemanager.hadoop.svc.cluster.local`（hostname.rm1/rm2）、`:8088`（webapp.address 及其 .rm1/.rm2 变体）
- **RM headless DNS**：`resourcemanager.hadoop.svc.cluster.local`（`yarn.log.server.url`、`mapreduce.jobtracker.address`、`yarn.resourcemanager.address` 类）

**③ `web-ui-access.yaml` — 5 处 `namespace: hadoop` → `namespace: <新ns>`**

### 16.3 机械操作步骤（以目标 ns = `prod` 为例）

```bash
# 0. 本地批量替换（先备份原文件，防改错无法回滚）
cp configmap.yaml configmap.yaml.bak
cp web-ui-access.yaml web-ui-access.yaml.bak
cp namespace.yaml namespace.yaml.bak
sed -i 's/\.hadoop\.svc\.cluster\.local/.prod.svc.cluster.local/g' configmap.yaml
sed -i 's/namespace: hadoop/namespace: prod/g' web-ui-access.yaml
sed -i 's/name: hadoop/name: prod/g' namespace.yaml

# 1. 创建 ns + 全套 apply（所有 -n 换成 prod）
kubectl apply -f namespace.yaml
kubectl apply -f configmap.yaml -n prod
kubectl apply -f zookeeper.yaml -n prod
kubectl apply -f journalnode.yaml -n prod
kubectl apply -f namenode.yaml -n prod
kubectl apply -f datanode.yaml -n prod
kubectl apply -f resourcemanager.yaml -n prod
kubectl apply -f nodemanager.yaml -n prod
kubectl apply -f web-ui-access.yaml -n prod

# 2. 校验：进 pod 确认 FQDN 已生效（关键！漏改会立刻服务发现全断）
kubectl exec -n prod namenode-0 -c namenode -- \
  grep -o 'prod.svc.cluster.local' /opt/hadoop/etc/hadoop/hdfs-site.xml | head
# 应输出若干行 *.prod.svc.cluster.local，且不得有 *.hadoop.svc.cluster.local 残留

# 3. 复验（同 §15 验证清单，把 -n hadoop 换成 -n prod）
kubectl get all -n prod
kubectl exec -n prod resourcemanager-0 -c resourcemanager -- \
  sh -c 'yarn jar /opt/hadoop/share/hadoop/mapreduce/hadoop-mapreduce-examples-*.jar pi 2 10'
```

### 16.4 警告与边界

- ❌ **不要把新 ns 的 configmap apply 进旧 ns，也不要把旧 ns 的 configmap apply 进新 ns** —— FQDN 不匹配会立即导致服务发现全断（NM 注册不了、NN/JN 连不上 ZK）。
- 🔁 **新 ns 是全新空集群**，不会继承旧 ns 的数据（数据存在旧 ns 的 PVC 里，PVC 不跨 ns 共享）。需要迁移数据请用 `distcp` 跨集群拷贝。
- 💾 **存储与 ns 无关**：openebs-hostpath 是节点本地盘，新 ns 部署会在节点上新建空 PVC。
- ✅ 改完务必跑 §15 全量验证清单（重点：NN HA 角色、RM HA 角色、pi 作业、真 failover），确认全绿再投产。
