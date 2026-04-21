# 生产变更单模板（Change Request Template）

> 用途：用于 Hadoop Operator 生产变更审批与执行归档。
> 
> 适用对象：平台团队 / 大数据团队 / SRE。

---

## 1. 基本信息

- **变更单 ID**：
- **变更标题**：
- **变更类型**：`标准变更 / 重大变更 / 紧急变更`
- **系统/服务**：`Hadoop Operator / HadoopCluster`
- **环境**：`prod / staging`
- **命名空间**：
- **目标集群（HadoopCluster 名称）**：
- **发起人**：
- **执行人**：
- **复核人**：
- **批准人（CAB/负责人）**：
- **计划开始时间**：
- **计划结束时间**：
- **维护窗口**：

---

## 2. 变更目标与背景

### 2.1 业务目标

- 本次变更希望解决的问题：
- 预期收益（容量、稳定性、性能、成本等）：

### 2.2 背景说明

- 当前现状（版本、拓扑、痛点）：
- 相关缺陷/需求单号：
- 预发验证结论：

---

## 3. 变更范围（Scope）

- **影响范围**：
  - [ ] 单集群
  - [ ] 多集群
  - [ ] 单命名空间
  - [ ] 多命名空间
- **涉及组件**：
  - [ ] Operator Deployment
  - [ ] NameNode
  - [ ] DataNode
  - [ ] ResourceManager
  - [ ] NodeManager
  - [ ] JournalNode
  - [ ] ZooKeeper
  - [ ] Ingress/NodePort
- **是否涉及数据路径**：`是 / 否`
- **是否涉及架构变化（如非 HA -> HA）**：`是 / 否`

---

## 4. 变更内容明细

### 4.1 配置变更（CR diff）

> 建议附 `before`/`after` YAML 或 `kubectl diff` 输出

- 变更前：

```yaml
# 粘贴关键片段
```

- 变更后：

```yaml
# 粘贴关键片段
```

### 4.2 镜像/版本变更

- Operator 镜像（旧 -> 新）：
- Hadoop 镜像（旧 -> 新）：
- ZooKeeper 镜像（旧 -> 新）：
- 是否固定 digest：`是 / 否`

### 4.3 资源变更

- CPU/MEM requests/limits：
- 副本调整：
- PVC/StorageClass：
- 暴露方式（NodePort / Ingress）：

---

## 5. 风险评估

### 5.1 风险列表

| 风险项 | 触发条件 | 影响 | 概率 | 等级 | 缓解措施 |
|---|---|---|---|---|---|
| 示例：NN 不可用 | 切换失败 | 读写中断 | 中 | 高 | 预演+回滚 |

### 5.2 依赖与前置条件

- [ ] CRD 已存在且版本正确
- [ ] Operator 健康
- [ ] 存储类可用
- [ ] 网络策略/防火墙已确认
- [ ] Ingress/TLS 准备完成（如适用）
- [ ] 变更窗口内值班到位

### 5.3 影响评估

- 用户影响：`无感 / 轻微抖动 / 计划中断`
- SLA 风险：
- 预计影响时长：

---

## 6. 执行计划（Step-by-step）

> 每一步都应可审计、可复现。

### 6.1 变更前快照

```bash
kubectl -n <ns> get hadoopcluster <name> -o yaml > before-<name>.yaml
kubectl -n <ns> get sts,svc,pvc,ing -l data.hadoop.org/cluster=<name> -o wide > before-<name>-resources.txt
kubectl -n hadoop-operator-system logs deploy/hadoop-operator --tail=300 > before-operator.log
```

### 6.2 执行步骤

1. Step-1：
2. Step-2：
3. Step-3：

关键命令（示例）：

```bash
kubectl apply -f <your-cr-or-manifest>.yaml
kubectl -n <ns> get pod,sts,svc,pvc,ing
kubectl -n hadoop-operator-system logs deploy/hadoop-operator -f
```

### 6.3 执行暂停点（Gate）

- Gate-1（进入下一步条件）：
- Gate-2（进入下一步条件）：

---

## 7. 验证计划（Validation）

### 7.1 技术验证

- [ ] 所有目标 Pod Ready
- [ ] 无异常重启
- [ ] PVC 正常绑定
- [ ] NameNode 状态正常（HA 场景主备健康）
- [ ] YARN 关键接口可用
- [ ] 暴露链路可达（NodePort/Ingress）

### 7.2 业务验证

- [ ] 关键任务提交成功
- [ ] 关键任务运行成功
- [ ] 业务方验收签字

### 7.3 观察窗口

- 观察时长：
- 观察指标：

---

## 8. 回滚方案（必须可执行）

### 8.1 回滚触发条件

- 连续 `X` 分钟不可恢复错误
- 核心功能不可用
- 业务验收失败

### 8.2 回滚步骤

1. 应用回滚清单/快照
2. 验证基础可用性
3. 恢复观察

示例命令：

```bash
kubectl apply -f before-<name>.yaml
kubectl -n <ns> get pod,sts,svc,pvc,ing
kubectl -n hadoop-operator-system logs deploy/hadoop-operator --tail=300
```

### 8.3 回滚责任人

- 主责：
- 备份：

---

## 9. 通知与沟通

- 变更前通知对象：
- 变更中同步频率：
- 变更后公告渠道：
- 升级路径（Escalation）：

---

## 10. 审批记录

- 技术审批：`通过 / 驳回`
- 业务审批：`通过 / 驳回`
- 风险审批：`通过 / 驳回`
- 最终批准：`通过 / 驳回`

审批意见：

---

## 11. 执行记录（现场填写）

| 时间 | 操作人 | 执行内容 | 结果 | 证据链接 |
|---|---|---|---|---|

---

## 12. 变更总结（收尾）

- 实际开始/结束时间：
- 执行结果：`成功 / 部分成功 / 失败`
- 遗留问题：
- 后续行动项（owner + 截止时间）：

---

## 13. 附件清单

- [ ] 变更前后 CR 文件
- [ ] 关键日志
- [ ] 验证截图/报告
- [ ] 回滚记录（如发生）

