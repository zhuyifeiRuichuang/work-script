# Hadoop Operator 生产运维手册（README-ops）

> 适用范围：基于本仓库 `data.hadoop.org/v1alpha1` 的 Hadoop Operator 集群。
> 
> 目标：给出可落地的生产运维流程，覆盖发布、扩容、升级、回滚、故障演练与应急。

---

## 1. 生产基线与职责边界

### 1.1 建议的环境分层

- **dev**：开发自测，允许快速迭代与不稳定配置
- **staging**：预发回归，配置与生产尽量一致
- **prod**：生产，仅允许走变更流程

### 1.2 职责划分（建议）

- **平台团队**：K8s、存储类、Ingress、节点容量、镜像仓库
- **大数据团队**：Hadoop 参数、容量规划、任务 SLA
- **SRE/值班**：发布窗口、巡检、应急与复盘

### 1.3 强制基线（上线前必须满足）

1. 所有镜像固定版本（建议固定 digest）
2. 存储类经过压测并验证 IO 稳定
3. Namespace 安全策略允许本项目 init 权限需求
4. 已定义告警（Pod 重启、PVC 异常、节点资源水位、NameNode 健康）
5. 已完成至少 1 次预发“全流程演练”

---

## 2. 变更管理流程（建议按 CAB 标准）

### 2.1 变更类型

- **标准变更**：副本扩容、资源调优、暴露方式切换
- **重大变更**：非 HA -> HA、存储类迁移、版本大升级
- **紧急变更**：故障止血，先恢复服务后补审批

### 2.2 变更单最小内容

- 变更目标与范围（集群名、命名空间）
- 变更前后参数对比（建议附 CR diff）
- 风险评估与回滚条件
- 验证项与观察时长
- 执行人、复核人、回滚责任人

### 2.3 发布窗口建议

- 避开任务高峰期
- 保留完整观察窗口（至少 30~60 分钟）
- 严禁并行多个高风险变更

---

## 3. 上线前检查清单（Pre-flight）

### 3.1 K8s 与 Operator 健康

```bash
kubectl -n hadoop-operator-system get deploy,pod
kubectl -n hadoop-operator-system logs deploy/hadoop-operator --tail=200
kubectl get crd hadoopclusters.data.hadoop.org
```

### 3.2 目标命名空间与资源配额

```bash
kubectl get ns hadoop
kubectl -n hadoop get resourcequota,limitrange
kubectl -n hadoop auth can-i create statefulsets --as system:serviceaccount:hadoop-operator-system:hadoop-operator
```

### 3.3 存储可用性

```bash
kubectl get sc
kubectl -n hadoop get pvc
```

检查点：

- `hdfs`/`zookeeper`/`journalnode` 所需 PVC 均能绑定
- 节点磁盘容量与 IOPS 满足预估负载

### 3.4 网络与暴露

- NodePort：安全组/防火墙已放通
- Ingress：域名解析、IngressClass、TLS Secret 已就绪

---

## 4. 标准发布流程（生产）

### 4.1 准备变更前快照

```bash
kubectl -n hadoop get hadoopcluster demoha -o yaml > before-demoha.yaml
kubectl -n hadoop get sts,svc,pvc,ing -l data.hadoop.org/cluster=demoha -o wide > before-resources.txt
```

### 4.2 应用变更

```bash
kubectl apply -f operator/samples/hadoopcluster-ha.yaml
```

或对线上 CR 做最小补丁：

```bash
kubectl -n hadoop patch hadoopcluster demoha --type merge -p '{"spec":{"hdfs":{"datanodeReplicas":4}}}'
```

### 4.3 观察与验收

```bash
kubectl -n hadoop get hadoopcluster
kubectl -n hadoop get pod,sts,svc,pvc,ing
kubectl -n hadoop describe hadoopcluster demoha
kubectl -n hadoop logs deploy/hadoop-operator -f
```

验收条件（示例）：

- 所有 Pod Ready
- 无持续重启
- NameNode Web / RM Web 可访问（按你启用的暴露方式）
- 关键任务可正常提交和完成

---

## 5. 扩容与缩容操作手册

### 5.1 DataNode 扩容（推荐在线）

- 修改：`spec.hdfs.datanodeReplicas`
- 观察：新增 DataNode 就绪，HDFS 报告节点数增加

```bash
kubectl -n hadoop patch hadoopcluster demoha --type merge -p '{"spec":{"hdfs":{"datanodeReplicas":5}}}'
```

### 5.2 NodeManager 扩容

- 修改：`spec.yarn.nodemanagerReplicas`

```bash
kubectl -n hadoop patch hadoopcluster demoha --type merge -p '{"spec":{"yarn":{"enabled":true,"nodemanagerReplicas":4}}}'
```

### 5.3 缩容注意事项

- 缩容前先确认数据副本安全与任务排空
- 缩容后观察至少 30 分钟
- 严禁在故障状态下直接缩容

---

## 6. 暴露方式切换（NodePort / Ingress）

### 6.1 NodePort 启用

将对应组件字段改成 `NodePort`，例如：

```yaml
spec:
  expose:
    namenodeWeb: NodePort
```

### 6.2 Ingress 启用

```yaml
spec:
  expose:
    ingress:
      enabled: true
      className: nginx
      tlsSecretName: hadoop-ui-tls
      namenodeNn1Host: nn1.hadoop.example.com
      namenodeNn2Host: nn2.hadoop.example.com
      resourcemanagerHost: rm.hadoop.example.com
```

### 6.3 冲突规避建议

- 若只想保留一种访问路径：
  - 关闭 Ingress：`ingress.enabled: false`
  - 或将对应 `*Web` 改回 `ClusterIP`
- 切换后做端到端访问测试

---

## 7. HA 专项运维（强制演练）

### 7.1 查看主备状态

```bash
kubectl -n hadoop exec -it demoha-namenode-0 -- hdfs haadmin -getServiceState nn1
kubectl -n hadoop exec -it demoha-namenode-0 -- hdfs haadmin -getServiceState nn2
```

### 7.2 手动故障切换演练

```bash
kubectl -n hadoop exec -it demoha-namenode-0 -- hdfs haadmin -failover nn1 nn2
```

### 7.3 强制故障演练（建议预发先做）

```bash
kubectl -n hadoop delete pod demoha-namenode-0
```

观察点：

- ZKFC 是否推动主备切换
- 客户端读写是否在 SLA 内恢复
- YARN 作业是否出现大面积失败

---

## 8. 版本升级策略（Operator 与 Hadoop）

### 8.1 Operator 升级

1. 新镜像先在 staging 验证
2. 更新 `operator/deploy/operator-deployment.yaml`
3. 滚动发布 operator
4. 观察 reconcile 日志与线上集群稳定性

### 8.2 Hadoop 组件镜像升级

通过 CR 修改 `spec.image`，建议：

- 小步升级（先 staging，再 prod）
- 每次只改一个关键变量（版本 or 参数）
- 保留回滚镜像标签

---

## 9. 回滚流程（必须预先定义）

### 9.1 回滚触发条件（示例）

- 关键服务不可用超过 5 分钟
- NameNode 无法选主
- 大面积任务失败且无快速修复路径

### 9.2 回滚步骤

1. 应用变更前快照（`before-demoha.yaml`）
2. 恢复到前一版本 CR
3. 若为 operator 引入问题，回滚 operator 镜像
4. 验证核心路径恢复

```bash
kubectl apply -f before-demoha.yaml
```

---

## 10. 日常巡检（建议自动化）

### 10.1 每日巡检项

- Pod Ready 与重启次数
- PVC 状态
- Node 资源水位（CPU/内存/磁盘）
- NameNode 活跃状态
- ZooKeeper/JournalNode 存活数

### 10.2 每周巡检项

- 配置漂移（CR 与 Git 声明一致性）
- 备份可恢复性抽检
- 告警噪音治理

---

## 11. 监控与告警建议（Prometheus 方向）

至少覆盖：

- Pod：`restarts`, `not ready`, `OOMKilled`
- Node：磁盘可用率、inode、网络丢包
- HDFS：活跃 NameNode 状态、DataNode 数、容量水位
- YARN：RM 健康、队列积压
- ZooKeeper：follower/leader 数、会话异常

告警分级建议：

- P1：服务不可用 / NameNode 无主
- P2：容量高水位 / 组件持续重启
- P3：单点异常、可容忍降级

---

## 12. 备份与恢复建议

### 12.1 必备备份

- `HadoopCluster` CR YAML
- NameNode 元数据快照（依据你们现网备份方案）
- 关键配置与证书（Ingress TLS 等）

### 12.2 恢复演练

- 至少季度一次全流程恢复演练
- 输出 RTO/RPO 实测数据
- 把演练问题纳入改进清单

---

## 13. 应急响应 Runbook（简版）

### 13.1 5 分钟内动作

1. 锁定影响范围（集群/命名空间/租户）
2. 停止高风险变更
3. 收集现场信息（pod 状态、operator 日志、事件）
4. 决策：修复还是回滚

### 13.2 快速采样命令

```bash
kubectl -n hadoop get pod -o wide
kubectl -n hadoop get events --sort-by=.lastTimestamp | tail -n 50
kubectl -n hadoop logs deploy/hadoop-operator --tail=300
kubectl -n hadoop describe hadoopcluster demoha
```

### 13.3 事后复盘模板（建议）

- 发生时间线
- 直接原因 / 根因
- 影响范围与时长
- 临时处置与长期修复
- 防再发行动项与 owner

---

## 14. 不建议直接做的高风险操作

- 在生产直接做“非 HA -> HA”架构切换（建议新建集群迁移）
- 在故障期间同时改镜像+参数+副本
- 无快照直接改关键存储与网络配置
- 未经预发验证直接上线

---

## 15. 推荐发布节奏

- **日常小改**：每周固定窗口
- **重大升级**：月度窗口，提前演练
- **紧急修复**：走应急流程，事后补齐审计

---

## 16. 可直接复用的发布检查模板

### 16.1 变更前

- [ ] 变更单审批完成
- [ ] 已导出当前 CR 与关键资源清单
- [ ] 回滚文件与回滚条件明确
- [ ] 监控/告警值班在线

### 16.2 变更中

- [ ] 单变量变更（避免叠加）
- [ ] 持续观察 Operator 日志
- [ ] 记录关键时间点

### 16.3 变更后

- [ ] 业务验收通过
- [ ] 观察窗口结束无异常
- [ ] 关闭变更并归档证据

---

如需，我可以继续补两份模板文件：

1. `change-request-template.md`（变更单模板）
2. `incident-postmortem-template.md`（故障复盘模板）
