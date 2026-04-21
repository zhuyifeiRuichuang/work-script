# MEMORY.md — Hadoop Operator 项目长期记忆

## 项目基本信息

- **项目**：Hadoop Kubernetes Operator
- **目标**：在 K8s 集群中部署和管理 Hadoop 集群（HDFS/YARN/HBase）
- **模块路径**：`github.com/hadoop-operator/hadoop-k8s-operator/operator`
- **API Group**：`hadoop.kubedoop.dev`（注意：不是 hadoop.apache.org！）
- **CRD**：HadoopCluster (shortName: hc)、HadoopApplication (shortName: ha)
- **框架**：`sigs.k8s.io/controller-runtime` v0.17.0

## 目录结构

```
operator/
├── cmd/main.go                          # 入口
├── pkg/apis/hadoop/v1/                  # CRD 类型定义
│   ├── hadoopcluster_types.go
│   ├── hadoopcluster_webhook.go
│   ├── groupversion_info.go             # GroupVersion = hadoop.kubedoop.dev/v1
│   └── zz_generated.deepcopy.go
├── pkg/builder/                         # 各组件 Builder
│   ├── builder.go                       # Builder 接口 + BuilderFactory + computePhase()
│   ├── configmap_builder.go             # ConfigMap + entrypoint.sh 生成
│   ├── nn_dn_builder.go                 # NameNode/DataNode
│   ├── journalnode_builder.go           # JournalNode（HDFS HA 必需）
│   ├── rm_nm_builder.go                 # ResourceManager/NodeManager
│   └── hbase_builder.go                 # HBase Master/RegionServer
├── pkg/controller/
│   ├── hadoopcluster_controller.go
│   └── hadoopapplication_controller.go
└── config/webhook/webhook.yaml
```

## 关键技术决策

### controller-runtime v0.17.0 API 变更
- `ctrl.Options.MetricsBindAddress` → `ctrl.Options.Metrics.BindAddress`（需 import `sigs.k8s.io/controller-runtime/pkg/metrics/server`）
- `ctrl.Options.Namespace` → `ctrl.Options.Cache.DefaultNamespaces`（需 import `sigs.k8s.io/controller-runtime/pkg/cache`）
- `ctrl.Options.LeaderElectionNamespace` 保持在顶层

### 状态推断
- StatefulSet/Deployment 没有 `.Status.Phase` 字段（只有 Pod 才有）
- 使用 `computePhase(readyReplicas, replicas int32) string` 辅助函数推断
- 返回值：`"Running"` / `"Degraded"` / `"Pending"`

### Log 字段类型
- Controller 中 `Log` 字段应为 `logr.Logger`（来自 `github.com/go-logr/logr`）
- 不能用 `sigs.k8s.io/controller-runtime/pkg/log` 包的 `Logger` 类型

### Webhook 配置
- 代码注解生成路径：`/mutate-hadoop-kubedoop-dev-v1-hadoopcluster`
- webhook.yaml 必须与代码 API Group `hadoop.kubedoop.dev` 一致
- failurePolicy 统一为 `Fail`

### entrypoint.sh 脚本
- 生成在 ConfigMap 的 `entrypoint.sh` 键中
- ConfigMap volume 设置 `defaultMode: 0755` 使脚本可执行
- 容器通过 subPath 挂载到 `/entrypoint.sh`
- NameNode 自动判断是否需要格式化（检查 `/data/hadoop/namenode/current` 目录）

## 已知待改进点

1. Docker 镜像内置 entrypoint.sh 比 ConfigMap 挂载更可靠
2. cert-manager 管理 webhook 证书（当前 webhook.yaml 是占位符证书）
3. CRD YAML 文件需要用 controller-gen 重新生成（当前可能过时）
4. JournalNode 的 DataNode 等待逻辑需要实际测试验证
5. NodeManager DaemonSet 的状态字段不同（用 NumberReady/NumberAvailable/DesiredNumberScheduled）

## 最后更新
2026-04-16：完成所有文档同步更新，统一 API Group 为 hadoop.kubedoop.dev（7个Go文件36处 + 4个YAML配置 + 4个示例YAML + 2个文档文件）。全面重写 README.md/README_zh.md。
