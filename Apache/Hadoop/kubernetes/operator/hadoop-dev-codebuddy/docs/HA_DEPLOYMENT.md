# Hadoop HA Cluster Deployment Guide

This document describes how to deploy a production-ready Hadoop cluster with High Availability (HA) support.

## Architecture Overview

```
                    ┌─────────────────────────────────────────────────────────┐
                    │                   ZooKeeper Cluster                     │
                    │         (External or Embedded - 3 nodes)                │
                    └─────────────────────────────────────────────────────────┘
                                          │
          ┌────────────────────────────────┼────────────────────────────────┐
          │                                │                                │
          ▼                                ▼                                ▼
┌─────────────────┐              ┌─────────────────┐              ┌─────────────────┐
│   NameNode 1    │◄────────────►│  JournalNode    │◄────────────►│   NameNode 2    │
│   (Active)      │   Edits QJM  │   (3 nodes)     │   Edits QJM  │   (Standby)     │
└─────────────────┘              └─────────────────┘              └─────────────────┘
          │                                                                │
          │                        HDFS Client                             │
          └────────────────────────────────────────────────────────────────┘

┌─────────────────┐              ┌─────────────────┐              ┌─────────────────┐
│ResourceManager 1│◄────────────►│     ZooKeeper   │◄────────────►│ResourceManager 2│
│   (Active)      │  ZK Failover │    (Leader)     │  ZK Failover │   (Standby)     │
└─────────────────┘              └─────────────────┘              └─────────────────┘
          │                                                                │
          │                        YARN Client                             │
          └────────────────────────────────────────────────────────────────┘
```

## Prerequisites

### 1. Kubernetes Cluster Requirements

- Kubernetes 1.24+
- At least 6 worker nodes (recommended)
- Minimum resources per node: 4 CPU cores, 8GB RAM
- StorageClass with dynamic provisioning

### 2. ZooKeeper Cluster (Required for HA)

Before deploying Hadoop HA, you must have a working ZooKeeper ensemble.

#### Option A: Deploy ZooKeeper using Helm (Recommended)

```bash
# Add Bitnami repository
helm repo add bitnami https://charts.bitnami.com/bitnami

# Install ZooKeeper
helm install zookeeper bitnami/zookeeper \
  --set replicaCount=3 \
  --set persistence.enabled=true \
  --set persistence.storageClass=standard \
  --set persistence.size=2Gi
```

#### Option B: Create ZooKeeper ConfigMap

After ZooKeeper is deployed, create a ConfigMap with ZooKeeper connection string:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: hadoopcluster-ha-zk
  namespace: default
data:
  ZOOKEEPER_SERVERS: "zookeeper-0.zookeeper-headless:2181,zookeeper-1.zookeeper-headless:2181,zookeeper-2.zookeeper-headless:2181"
```

## Deploying HA Hadoop Cluster

### Method 1: Using CRD

```bash
# Apply the HA cluster configuration
kubectl apply -f config/samples/hadoop_v1alpha1_hadoopcluster-ha.yaml
```

### Method 2: Using Helm Chart

```bash
# Install with HA values
helm install hadoop-ha deploy/helm/hadoop-operator \
  --values deploy/helm/hadoop-operator/examples/values-ha.yaml \
  --set ha.nameNodeHA.enabled=true \
  --set ha.resourceManagerHA.enabled=true \
  --set zookeeper.servers="zookeeper-0.zookeeper-headless:2181,zookeeper-1.zookeeper-headless:2181,zookeeper-2.zookeeper-headless:2181"
```

## HA Configuration Details

### NameNode HA

| Property | Description | Default |
|----------|-------------|---------|
| `spec.ha.nameNodeHA.enabled` | Enable NameNode HA | `false` |
| `spec.ha.nameNodeHA.nameServiceId` | HDFS nameservice ID | Required |
| `spec.ha.nameNodeHA.journalClusterId` | Journal cluster ID | Optional |
| `spec.ha.nameNodeHA.replicas` | Number of NameNodes | `2` |

Generated Configuration (`hdfs-site.xml`):
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

| Property | Description | Default |
|----------|-------------|---------|
| `spec.ha.resourceManagerHA.enabled` | Enable ResourceManager HA | `false` |
| `spec.ha.resourceManagerHA.clusterId` | YARN cluster ID | `rm-cluster` |
| `spec.ha.resourceManagerHA.replicas` | Number of ResourceManagers | `2` |

Generated Configuration (`yarn-site.xml`):
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

## Verify HA Deployment

### 1. Check Cluster Status

```bash
kubectl get hadoopcluster
kubectl describe hadoopcluster hadoopcluster-ha
```

### 2. Verify NameNode HA

```bash
# Get NameNode pods
kubectl get pods -l hadoop.kubedoop.dev/component=namenode

# Check HDFS HA status (exec into active NameNode)
kubectl exec -it hadoopcluster-ha-namenode-0 -- hdfs haadmin -ns ns -getServiceState

# Expected output:
# namenode-0 active
# namenode-1 standby
```

### 3. Verify ResourceManager HA

```bash
# Get ResourceManager pods
kubectl get pods -l hadoop.kubedoop.dev/component=resourcemanager

# Check YARN HA status (exec into active ResourceManager)
kubectl exec -it hadoopcluster-ha-resourcemanager-rm1 -- yarn rmadmin -getServiceState rm1

# Expected output:
# rm1 active
# rm2 standby
```

### 4. Test Automatic Failover

#### NameNode Failover Test

```bash
# Simulate NameNode failure
kubectl exec -it hadoopcluster-ha-namenode-0 -- kill 1

# Wait for failover and check status
sleep 30
kubectl exec -it hadoopcluster-ha-namenode-1 -- hdfs haadmin -ns ns -getServiceState
```

#### ResourceManager Failover Test

```bash
# Simulate ResourceManager failure
kubectl delete pod hadoopcluster-ha-resourcemanager-rm1

# Wait for failover and check status
kubectl exec -it hadoopcluster-ha-resourcemanager-rm2 -- yarn rmadmin -getServiceState rm2
```

## Accessing HA Services

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

### Using HA Proxy (Recommended for Clients)

Create a service to proxy to the active ResourceManager:

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

## Troubleshooting

### Issue: ZooKeeper Connection Failed

**Symptom**: NameNode or ResourceManager fails to start with ZooKeeper connection error.

**Solution**:
1. Verify ZooKeeper is running: `kubectl get pods -l app.kubernetes.io/name=zookeeper`
2. Check ConfigMap has correct ZooKeeper addresses
3. Update `spec.clusterConfig.zooKeeperConfigMapName`

### Issue: NameNode Not Entering Safe Mode

**Symptom**: NameNode stuck in safe mode after HA setup.

**Solution**:
1. Ensure JournalNodes are running: `kubectl get pods -l hadoop.kubedoop.dev/component=journalnode`
2. Wait for JournalNodes to sync edits
3. Run: `kubectl exec -it <active-namenode> -- hdfs dfsadmin -safemode leave`

### Issue: ResourceManager Not Discovering Other RM

**Symptom**: ResourceManager not transitioning to active/standby state.

**Solution**:
1. Verify both ResourceManager pods are running
2. Check ZooKeeper access from pods
3. Verify Service is Headless (ClusterIP: None) for HA mode

## Component Requirements for HA

| Component | HA Required | Minimum Replicas | Notes |
|-----------|-------------|-----------------|-------|
| NameNode | Yes | 2 | Must be even number |
| JournalNode | Yes | 3 | Must be odd number |
| DataNode | No | 1+ | Replicas should be > replication factor |
| ResourceManager | Yes | 2 | Must be even number |
| NodeManager | No | 1+ | Depends on workload |

## Cleanup

```bash
# Delete HA cluster
kubectl delete -f config/samples/hadoop_v1alpha1_hadoopcluster-ha.yaml

# Delete ZooKeeper (if installed via Helm)
helm uninstall zookeeper
```
