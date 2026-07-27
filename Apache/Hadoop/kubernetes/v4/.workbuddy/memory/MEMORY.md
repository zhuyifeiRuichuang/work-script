# Hadoop K8s Project Memory

## Project Overview
- Hadoop 3.1.1 on Kubernetes, production-grade configuration with **NameNode HA**
- Hadoop Image: `swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/zhuyifeiruichuang/hadoop:3.1.1`
- ZooKeeper Image: `swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/library/zookeeper:3.4.14`
- Namespace: canonical/default = `hadoop` (user-mandated 2026-07-24). The **currently-running production cluster is still in `hadoop1`** (legacy, pre-rename). All configs/docs (`configmap.yaml` FQDNs, `web-ui-access.yaml`, `deploy-guide.md`) now target `hadoop` — do NOT `kubectl apply` the new `hadoop`-targeted configmap into the live `hadoop1` cluster (DNS FQDN mismatch would break service discovery).

## Key Design Decisions — NameNode HA (Plan B)
- **NameNode: 2 Pods (Active + Standby) + ZKFC sidecar** — application-level HA, no storage dependency
- **No Longhorn needed** — all components use `openebs-hostpath` (local storage)
- HA mechanism (JournalNode + ZooKeeper + ZKFC) guarantees zero data loss on node failure
- Active NN writes edits to JournalNode quorum (3 nodes); Standby syncs from JN
- ZKFC monitors NN health, triggers automatic failover (~30s recovery)
- Hard podAntiAffinity: two NN Pods must be on different nodes
- OrderedReady podManagement: namenode-0 starts first, namenode-1 bootstraps after
- Single PVC per NN Pod (nn1/nn2 dirs within same volume for directory-level redundancy)

## Components
- ZooKeeper: 3 Pods, hard anti-affinity, PDB minAvailable=2, openebs-hostpath 10Gi, **Parallel** startup
- JournalNode: 3 Pods, hard anti-affinity, PDB minAvailable=2, openebs-hostpath 20Gi, **Parallel** startup
- NameNode: 2 Pods + ZKFC sidecar, hard anti-affinity, PDB minAvailable=1, openebs-hostpath 50Gi, **OrderedReady** startup (HA init sequence)
- DataNode: 3 Pods, hard anti-affinity, PDB minAvailable=2, openebs-hostpath 100Gi, **Parallel** startup
- ResourceManager: 2 Pods HA (replicas:2, hard anti-affinity across nodes, ZKRMStateStore for state + ActiveStandbyElector for leader election), PDB minAvailable=1
- NodeManager: 3 Pods, hard anti-affinity, PDB minAvailable=2, **Parallel** startup

## Deployment Order
ZK → ConfigMap → JN → NN → DN → RM → NM (ZK and JN MUST be before NN)

## Node Failure Strategy (HA Mode)
- NameNode node failure: ZKFC detects → Standby becomes Active (~30s). PVC on failed node is lost but data preserved via JN quorum + surviving NN. New Pod rebuilt on healthy node, syncs from JN, becomes Standby.
- DataNode node failure: Same as before — HDFS 3-replica protects data, NN triggers re-replication.
- JournalNode failure: Quorum still works (2/3 majority). Recovery: delete PVC + rebuild Pod.
- ZooKeeper failure: Quorum still works (2/3 majority). Recovery: delete PVC + rebuild Pod.

## Files
- `namespace.yaml` - Namespace definition
- `configmap.yaml` - All Hadoop configs (HA configs: nameservice, JN, ZK, failover proxy, fencing)
- `zookeeper.yaml` - ZK 3-node StatefulSet + PDB
- `journalnode.yaml` - JN 3-node StatefulSet + PDB
- `namenode.yaml` - NN 2-node HA + ZKFC sidecar + PDB (single PVC, openebs-hostpath)
- `datanode.yaml` - DN 3-node + PDB (HA-aware initContainer: wait nn1 OR nn2)
- `resourcemanager.yaml` - RM 2-Pod HA (ZKRMStateStore) + PDB + wait-for-hdfs init (nsswitch fix + safemode 2>&1 DEBUG)
- `nodemanager.yaml` - NM 3-node + PDB
- `deploy-guide.md` - HA 部署+运维 runbook（**已更新 v4 @2026-07-24**：补 YARN 双 RM HA、K8s 三必改坑源码级根因、Web UI 确定性访问、真 failover 测试、写死地址清理；RM 由 1→2 Pod）

## Important Notes
- **Timezone convention (mandated)**: All component logs MUST be Asia/Shanghai (GMT+8). Container default is UTC (8h behind local env) which makes log-based troubleshooting confusing. Fix = set `TZ=Asia/Shanghai` env on EVERY container + `-Duser.timezone=Asia/Shanghai` in the JVM flag envs (`JAVA_TOOL_OPTIONS` / `HADOOP_OPTS` / `JVMFLAGS`). `TZ` alone is sometimes ignored by the JVM, so both levers are required. Applied to zookeeper / journalnode / namenode(+zkfc) / datanode / resourcemanager / nodemanager yaml.
- No Longhorn required — HA is the reliability mechanism, not storage replication
- NN init sequence: nn0 → formatZK → format NN → initializeSharedEdits; nn1 → wait nn0 → bootstrapStandby
- dfs.permissions.enabled = true (production)
- Capacity scheduler: default + production dual queues
- fs.defaultFS = hdfs://mycluster (logical URI, not single NN hostname)

## YARN / NodeManager Production Runbook (HA cluster)
- **NM probe MUST be `tcpSocket` not `httpGet`**: Hadoop 3.1.1 NM REST `/ws/v1/nodeinfo` returns 404 → kubelet kills NM in a restart loop. Use `tcpSocket:{port:8042}` for liveness/readiness/startup. (RM `/ws/v1/cluster/info:8088` httpGet is fine — both active & standby RM serve 200.)
- **Never bind Pod hostname to `127.0.0.1` in NM**: its `InetAddress.getLocalHost()` reverse-resolves 127.0.0.1 → `localhost`, so NM registers as `localhost:<port>` and AMs can't route. Fix = keep `hosts: files dns` nsswitch + rely on K8s-injected `<PodIP> <hostname>.<svc>` so NM registers FQDN. (RM main container MAY keep `127.0.0.1 $(hostname)` — RM uses configured `yarn.resourcemanager.hostname.rmX` FQDNs, not getLocalHost, so it's harmless there.)
- **NM is always a SIGTERM *victim*, not a crash source**: its "crashes" were probe 404 kills / cluster jitter. Don't hunt for NM code bugs; check probes + registration address first.
- **Failover discipline**: RM HA only proves itself when the ACTIVE pod is killed. Always `yarn rmadmin -getServiceState rm1/rm2` FIRST, then `delete pod` of the **active**. Deleting the standby is a no-op. Expect standby→active in ~10-30s; running jobs continue via ZKRMStateStore.
- **`yarn jar` path pitfalls in `kubectl exec`**: (1) `$HADOOP_HOME` is NOT exported in a non-interactive `kubectl exec` shell → use absolute `/opt/hadoop/share/...`. (2) `kubectl exec pod -- yarn jar /path/*.jar` does NOT spawn a shell, so `*.jar` glob is passed literally → wrap in `sh -c 'yarn jar /path/*.jar ...'` OR `ls` the exact filename first.
- **StatefulSet apply ≠ auto-recreate for configmap volume mounts**: config changes need `rollout restart` (or delete pod) to take effect. Pod-template changes (probes/command) DO trigger RollingUpdate, but `rollout restart` is the safe explicit trigger.

## Critical Pitfall — NN init container MUST run hdfs as root (NOT `su hadoop`)
- **Symptom**: Standby NN (namenode-1) CrashLoopBackOff, main NN log: `java.io.IOException: NameNode is not formatted.`; init log: `BootstrapStandby: Access denied for user hadoop. Superuser privilege is required`.
- **Root cause**: `hdfs namenode -bootstrapStandby` calls Active NN's `isUpgradeFinalized` RPC, which requires **superuser**. Active NN `fsOwner=root` (main container runs as root via `securityContext.runAsUser:0`), but init wrapped hdfs commands in `su hadoop -c "..."`, downgrading to non-superuser `hadoop` → RPC rejected → bootstrap aborts → name.dir never formatted → `NameNode is not formatted`.
- **Fix**: Run ALL hdfs commands in init container as root (drop `su hadoop -c`); chown data dir to running user (`id -u`/`:id -g`, =0:0). Use `hdfs namenode -bootstrapStandby -force -nonInteractive` (idempotent, -force overrides "data appears to exist / Not formatting" in -nonInteractive).
- **Secondary**: `|| echo` masking on bootstrap previously hid the failure → never mask bootstrap errors.
- **Note**: format/formatZK/initializeSharedEdits work as hadoop (local ops / ZK-JN only, no remote superuser RPC), but bootstrap is the one that requires root. Running everything as root is simplest + consistent.
