# Hadoop Kubernetes Operator（生产可用版）

本项目提供一个面向生产环境的 Hadoop Operator（Kopf/Python 实现），支持：

- 单 NameNode 与 NameNode HA（QJM + ZooKeeper + ZKFC）
- HDFS + YARN（可关闭 YARN）
- 在线环境与离线环境部署
- 双暴露方式：`NodePort` 与 `Ingress`（可同时启用，也可注释其一）
- 配置变更自动滚动（基于 ConfigMap 内容哈希）

---

## 1. 目录结构

```text
.
├─ operator/
│  ├─ deploy/
│  │  ├─ crd.yaml
│  │  ├─ rbac.yaml
│  │  └─ operator-deployment.yaml
│  ├─ hadoop_operator/
│  │  ├─ controller.py
│  │  ├─ k8s_resources.py
│  │  ├─ config_xml.py
│  │  ├─ apply.py
│  │  └─ constants.py
│  ├─ samples/
│  │  ├─ hadoopcluster-minimal.yaml
│  │  └─ hadoopcluster-ha.yaml
│  ├─ docker/
│  │  └─ hadoop-reference.Dockerfile
│  ├─ Dockerfile
│  └─ requirements.txt
├─ namenode.yaml
├─ datanode.yaml
├─ resourcemanager.yaml
├─ nodemanager.yaml
└─ configmap.yaml
```

> 根目录下旧的 `namenode.yaml`/`datanode.yaml` 等为早期静态清单，建议以 Operator 方案为主。

---

## 2. 关键能力说明

### 2.1 CRD 定义

- 资源组：`data.hadoop.org`
- 版本：`v1alpha1`
- Kind：`HadoopCluster`
- 作用域：`Namespaced`

### 2.2 Operator 会自动协调的资源

按每个 `HadoopCluster` 创建/更新：

- `ConfigMap`（Hadoop XML 配置）
- `Service`（Headless + 可选 NodePort + 可选 Ingress 辅助 Service）
- `StatefulSet`（NameNode/DataNode/ResourceManager/NodeManager）
- HA 模式下额外：`JournalNode`、`ZooKeeper`、`ZKFC`
- `PodDisruptionBudget`（NameNode）

删除 `HadoopCluster` 后，通过 OwnerReference 级联清理子资源。

### 2.3 NameNode HA 机制

HA 打开后，自动切到：

- NameNode 副本数固定为 2（`nn1/nn2`）
- QJM 共享编辑日志（JournalNode，默认奇数副本 >= 3）
- ZooKeeper（默认奇数副本 >= 3）
- ZKFC 自动故障转移
- `fs.defaultFS` 改为 `hdfs://<nameservice>`

---

## 3. 前置条件（生产建议）

- Kubernetes 1.26+
- 集群内可用动态存储类（StorageClass）
- 允许目标命名空间运行带 `privileged` 的初始化容器
  - 当前 NameNode/DataNode/JournalNode 初始化涉及目录权限修正
- 建议启用 NTP/时钟同步
- 建议为存储、网络、节点资源预留足够容量

---

## 4. 开发环境配置

### 4.1 Python 依赖

在 `operator/` 目录：

```bash
python -m pip install -r requirements.txt
```

依赖为：

- `kopf>=1.37.0`
- `kubernetes>=29.0.0`
- `PyYAML>=6.0.1`

### 4.2 本地运行调试（连现有 kubeconfig）

```bash
python -m hadoop_operator
```

`controller.py` 在启动时会优先使用 InCluster 配置，失败后回落到本地 kubeconfig。

### 4.3 语法检查

```bash
python -m py_compile hadoop_operator/controller.py hadoop_operator/k8s_resources.py hadoop_operator/config_xml.py hadoop_operator/apply.py
```

---

## 5. 构建与部署 Operator

### 5.1 构建 Operator 镜像

在仓库根目录执行：

```bash
docker build -t <your-registry>/hadoop-operator:0.1.0 -f operator/Dockerfile operator
```

推送：

```bash
docker push <your-registry>/hadoop-operator:0.1.0
```

### 5.2 修改部署镜像地址

编辑：`operator/deploy/operator-deployment.yaml`

将：

- `ghcr.io/REPLACE_ME/hadoop-operator:0.1.0`

替换为你的真实镜像地址。

### 5.3 部署顺序

```bash
kubectl apply -f operator/deploy/crd.yaml
kubectl apply -f operator/deploy/rbac.yaml
kubectl apply -f operator/deploy/operator-deployment.yaml
```

检查：

```bash
kubectl -n hadoop-operator-system get pod
kubectl -n hadoop-operator-system logs deploy/hadoop-operator -f
```

---

## 6. 联网环境与离线环境部署

### 6.1 联网环境

- `spec.image` 保持 `zhuyifeiruichuang/hadoop:3.1.1`
- HA 时 `spec.zookeeper.image` 可使用 `zookeeper:3.8.4`

### 6.2 离线环境

需要提前导入到私有仓库的镜像：

- Operator 镜像：`<your-registry>/hadoop-operator:0.1.0`
- Hadoop 镜像：`zhuyifeiruichuang/hadoop:3.1.1`（或你的等效地址）
- ZooKeeper 镜像：`zookeeper:3.8.4`（HA 时）

在 CR 中配置：

- `spec.image`（Hadoop 镜像地址）
- `spec.zookeeper.image`（HA 时 ZooKeeper 镜像地址）
- `spec.imagePullSecrets`

---

## 7. 创建 Hadoop 集群

### 7.1 最小化（非 HA）

使用：`operator/samples/hadoopcluster-minimal.yaml`

```bash
kubectl apply -f operator/samples/hadoopcluster-minimal.yaml
```

### 7.2 HA 模式

使用：`operator/samples/hadoopcluster-ha.yaml`

```bash
kubectl apply -f operator/samples/hadoopcluster-ha.yaml
```

检查状态：

```bash
kubectl -n hadoop get hadoopcluster
kubectl -n hadoop get sts,pvc,svc,ing
kubectl -n hadoop get pod -o wide
```

---

## 8. CR 字段说明（重点）

### 8.1 镜像与拉取

- `spec.image`：Hadoop 镜像
- `spec.imagePullPolicy`：默认 `IfNotPresent`
- `spec.imagePullSecrets`：私有仓库密钥

### 8.2 HDFS

- `spec.hdfs.namenodeStorageClass`
- `spec.hdfs.namenodeStorageSize`
- `spec.hdfs.datanodeReplicas`
- `spec.hdfs.datanodeStorageClass`
- `spec.hdfs.datanodeStorageSize`
- `spec.hdfs.replication`

### 8.3 HA 子项

- `spec.hdfs.ha.enabled`：`true/false`
- `spec.hdfs.ha.nameservice`：可选，不填自动生成
- `spec.hdfs.ha.journalnodeReplicas`：建议奇数，至少 3（operator 会自动修正到奇数）

### 8.4 ZooKeeper（HA 时）

- `spec.zookeeper.image`
- `spec.zookeeper.replicas`（建议奇数，至少 3）
- `spec.zookeeper.storageClass`
- `spec.zookeeper.storageSize`

### 8.5 YARN

- `spec.yarn.enabled`
- `spec.yarn.nodemanagerReplicas`

### 8.6 暴露方式

- `spec.expose.namenodeWeb`：`ClusterIP` 或 `NodePort`
- `spec.expose.datanodeWeb`：`ClusterIP` 或 `NodePort`
- `spec.expose.resourcemanagerWeb`：`ClusterIP` 或 `NodePort`
- `spec.expose.nodemanagerWeb`：`ClusterIP` 或 `NodePort`
- `spec.expose.ingress.*`：Ingress 配置

### 8.7 资源

可按组件覆盖请求/限制：

- `resources.namenode`
- `resources.datanode`
- `resources.resourcemanager`
- `resources.nodemanager`
- `resources.journalnode`
- `resources.zookeeper`
- `resources.zkfc`

---

## 9. NodePort 与 Ingress 同时支持（含冲突处理）

### 9.1 推荐原则

- 需要外部固定端口：优先 `NodePort`
- 需要域名/TLS/七层路由：优先 `Ingress`
- 可同时启用；若你希望更简单，注释掉其中一种

### 9.2 非 HA 场景

- `namenodeWeb: NodePort` 会创建 `*-namenode-external`
- Ingress 可用 `namenodeSingleHost`

### 9.3 HA 场景

- `namenodeWeb: NodePort` 会为每个 NN 创建独立 Service：
  - `*-namenode-0-external`
  - `*-namenode-1-external`
- Ingress 可分别配置：
  - `namenodeNn1Host`
  - `namenodeNn2Host`

> 如果出现访问策略冲突，直接在 CR 中注释掉一组配置（如关闭 `ingress.enabled` 或将 `namenodeWeb` 改回 `ClusterIP`）。

---

## 10. 生产部署检查清单

上线前建议逐项确认：

- 存储类与容量：`NN/DN/JN/ZK` PVC 都能成功绑定
- Namespace 安全策略允许必要 init 权限
- 核心 Pod 探针稳定，无频繁重启
- HDFS 副本数与 DataNode 数量匹配
- NameNode HA 能切主（模拟故障测试）
- NodePort/Ingress 的网络策略与防火墙已放通
- 监控与日志采集已接入（至少 Pod/节点/存储指标）
- 做过一次 Operator 重启与 CR 变更回归验证

---

## 11. 常见运维命令

### 11.1 看 Operator

```bash
kubectl -n hadoop-operator-system get deploy,pod
kubectl -n hadoop-operator-system logs deploy/hadoop-operator -f
```

### 11.2 看集群资源

```bash
kubectl -n hadoop get hadoopcluster
kubectl -n hadoop get pod,svc,sts,pvc,ing
```

### 11.3 检查 NameNode HA

```bash
kubectl -n hadoop exec -it demoha-namenode-0 -- hdfs haadmin -getServiceState nn1
kubectl -n hadoop exec -it demoha-namenode-0 -- hdfs haadmin -getServiceState nn2
```

### 11.4 强制主备切换（测试）

```bash
kubectl -n hadoop exec -it demoha-namenode-0 -- hdfs haadmin -failover nn1 nn2
```

---

## 12. 故障排查建议

### 12.1 Pod 一直 Pending

- 检查 PVC 是否绑定失败：
  - `kubectl -n hadoop describe pvc <name>`
- 检查 StorageClass 名称是否存在、容量是否够

### 12.2 HA 初始化失败

- 检查 ZooKeeper 与 JournalNode 是否先就绪
- 查看 NameNode `ha-bootstrap` init 容器日志

### 12.3 Web 页面无法访问

- NodePort：检查安全组/防火墙
- Ingress：检查 `IngressClass`、域名解析、TLS Secret

### 12.4 CR 修改后未滚动

- Operator 已按 ConfigMap 内容计算 `config-revision` 注解
- 若未触发，先确认 CR 是否确实改动到配置内容字段

---

## 13. 参考 Hadoop 镜像 Dockerfile

路径：`operator/docker/hadoop-reference.Dockerfile`

用途：给你一个更可靠的 Hadoop 镜像制作模板（可选，不影响当前 `spec.image`）

特点：

- 基于 `eclipse-temurin:11-jdk-jammy`
- 从 Apache 归档下载 Hadoop 发行包
- 支持可选 sha512 校验（生产建议开启）
- `tini` 作为 entrypoint
- 默认非 root 用户 `hadoop`

---

## 14. 版本与兼容性说明

当前是 `v1alpha1`，Schema 允许未知字段（便于演进）。

建议：

- 生产中固定镜像版本与 digest
- 升级前先在预发命名空间做一次完整回归
- 避免直接对已运行集群做“架构切换式”变更（尤其非 HA -> HA）

---

## 15. 快速开始（5 分钟）

```bash
# 1) 部署 operator
kubectl apply -f operator/deploy/crd.yaml
kubectl apply -f operator/deploy/rbac.yaml
kubectl apply -f operator/deploy/operator-deployment.yaml

# 2) 创建 HA 集群（可按需编辑 storageClass / expose）
kubectl apply -f operator/samples/hadoopcluster-ha.yaml

# 3) 观察状态
kubectl -n hadoop get pod,svc,sts,pvc,ing
```

如果你希望，我可以下一步再补一份 `README-ops.md`，专门写“生产变更流程（扩容、滚动升级、回滚、应急）”。
