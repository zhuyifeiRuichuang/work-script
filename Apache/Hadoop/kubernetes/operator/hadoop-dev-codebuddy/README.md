# Hadoop Kubernetes Operator

[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go%20version-1.21+-blue.svg)](https://golang.org/)
[![Kubernetes](https://img.shields.io/badge/kubernetes-1.24+-blue.svg)](https://kubernetes.io/)
[![Operator](https://img.shields.io/badge/operator-kubedoop-blue.svg)](https://kubedoop.dev/)

A Kubernetes operator for managing Apache Hadoop clusters on Kubernetes. This operator provides a declarative way to deploy and manage Hadoop clusters (HDFS, YARN, HBase) using Kubernetes custom resources.

## 📖 Table of Contents

- [Features](#-features)
- [Architecture](#-architecture)
- [Prerequisites](#-prerequisites)
- [Quick Start](#-quick-start)
- [Installation](#-installation)
- [Configuration](#-configuration)
- [CRD Reference](#-crd-reference)
- [Development](#-development)
- [Deployment](#-deployment)
- [Examples](#-examples)
- [Troubleshooting](#-troubleshooting)
- [Contributing](#-contributing)
- [License](#-license)

## ✨ Features

### Core Features
- **HDFS Management**: Deploy and manage NameNode, DataNode, and JournalNode
- **YARN Management**: Deploy and manage ResourceManager and NodeManager
- **HBase Integration**: Optional HBase cluster support (Master and RegionServers)
- **High Availability**: Native support for HDFS HA (with JournalNode QJM) and YARN HA (with ZooKeeper automatic failover)
- **HDFS Federation**: Support for multiple HDFS namespaces
- **Kerberos Security**: Optional Kerberos authentication support
- **HadoopApplication CRD**: Submit and manage Hadoop applications (MapReduce, Spark, Hive, etc.) as Kubernetes custom resources

### Advanced Features
- **Role Groups**: Flexible role group configuration for different workloads
- **Logging**: Per-component logging configuration (console + file, per-logger level control)
- **Resource Management**: CPU and memory resource limits per component
- **Storage**: Configurable persistent storage with PVCs (supports EmptyDir for testing)
- **Health Checks**: Built-in liveness and readiness probes
- **Rolling Updates**: Support for rolling upgrades
- **ConfigMap-based Configuration**: Auto-generated `core-site.xml`, `hdfs-site.xml`, `yarn-site.xml`, `mapred-site.xml`, and `entrypoint.sh`
- **Custom ConfigMap Merge**: Merge user-provided ConfigMap with operator-generated configs

### Operator Features
- **Webhook Validation**: CRD validation and defaulting (admission webhook)
- **Metrics**: Prometheus metrics endpoint
- **Leader Election**: High availability for the operator itself
- **Multi-namespace Support**: Watch single or all namespaces
- **Finalizer-based Cleanup**: Proper resource cleanup on CRD deletion

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Hadoop Operator                         │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              HadoopCluster Controller                │   │
│  │  ┌───────────┐ ┌───────────┐ ┌───────────┐          │   │
│  │  │ NameNode  │ │ DataNode  │ │JournalNode│          │   │
│  │  │ Builder   │ │ Builder   │ │ Builder   │          │   │
│  │  └───────────┘ └───────────┘ └───────────┘          │   │
│  │  ┌───────────┐ ┌───────────┐ ┌───────────┐          │   │
│  │  │ResourceMgr│ │NodeManager│ │ConfigMap  │          │   │
│  │  │ Builder   │ │ Builder   │ │ Builder   │          │   │
│  │  └───────────┘ └───────────┘ └───────────┘          │   │
│  │  ┌───────────┐ ┌───────────┐                         │   │
│  │  │HBaseMaster│ │HBaseRS    │  HadoopApplication      │   │
│  │  │ Builder   │ │ Builder   │  Controller             │   │
│  │  └───────────┘ └───────────┘                         │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   StatefulSet │  │   StatefulSet │  │   StatefulSet │     │
│  │   (NameNode)  │  │   (DataNode)  │  │ (JournalNode) │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │  Deployment   │  │   StatefulSet │  │   StatefulSet│     │
│  │(ResourceMgr)  │  │ (HBaseMaster) │  │(HBaseRegionS)│     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│  ┌──────────────┐  ┌──────────────┐                        │
│  │  DaemonSet   │  │   ConfigMap   │                        │
│  │ (NodeManager) │  │ (Hadoop Config)│                       │
│  └──────────────┘  └──────────────┘                        │
└─────────────────────────────────────────────────────────────┘
```

**API Group**: `hadoop.kubedoop.dev/v1`

**Custom Resources**:
- `HadoopCluster` (shortName: `hc`) — Hadoop cluster lifecycle management
- `HadoopApplication` (shortName: `ha`) — Hadoop application submission

## 📋 Prerequisites

- Kubernetes 1.24 or higher
- kubectl 1.24 or higher
- Go 1.21 or higher (for development)

### Optional Dependencies

For HDFS High Availability:
- ZooKeeper cluster (or use [zookeeper-operator](https://github.com/zncdatadev/zookeeper-operator))

## 🚀 Quick Start

### 1. Deploy the Operator

```bash
# Clone the repository
git clone https://github.com/hadoop-operator/hadoop-k8s-operator.git
cd hadoop-k8s-operator

# Install CRDs
kubectl apply -f operator/config/crd/hadoopcluster-crd.yaml

# Create namespace
kubectl create namespace hadoop-system

# Apply RBAC
kubectl apply -f operator/config/rbac/role.yaml

# Apply Webhook configuration
kubectl apply -f operator/config/webhook/webhook.yaml

# Deploy the operator
kubectl apply -f operator/config/deploy/operator-deployment.yaml
```

### 2. Deploy a Hadoop Cluster

```yaml
# hadoopcluster.yaml
apiVersion: hadoop.kubedoop.dev/v1
kind: HadoopCluster
metadata:
  name: my-hadoop-cluster
spec:
  image: apache/hadoop:3.4.1
  nameNodeSpec:
    replicas: 1
    storage:
      useEmptyDir: true
  dataNodeSpec:
    replicas: 3
    storage:
      useEmptyDir: true
  resourceManagerSpec:
    replicas: 1
  nodeManagerSpec:
    replicas: 3
```

```bash
kubectl apply -f hadoopcluster.yaml -n hadoop-system
```

### 3. Check the Cluster Status

```bash
# View the cluster
kubectl get hadoopcluster -n hadoop-system

# View all resources
kubectl get all -n hadoop-system -l hadoop.kubedoop.dev/cluster=my-hadoop-cluster

# View detailed status
kubectl describe hadoopcluster my-hadoop-cluster -n hadoop-system
```

## 📥 Installation

### Method 1: Manual (Recommended for development)

```bash
# Clone the repository
git clone https://github.com/hadoop-operator/hadoop-k8s-operator.git
cd hadoop-k8s-operator

# Install CRDs
kubectl apply -f operator/config/crd/hadoopcluster-crd.yaml

# Create namespace
kubectl create namespace hadoop-system

# Apply RBAC
kubectl apply -f operator/config/rbac/role.yaml

# Deploy the operator
kubectl apply -f operator/config/deploy/operator-deployment.yaml
```

### Method 2: Build and Deploy from Source

```bash
cd operator

# Install dependencies
go mod download

# Build the operator binary
go build -o bin/hadoop-operator ./cmd/main.go

# Build Docker image
docker build -t hadoop-operator:latest -f ../docker/hadoop/Dockerfile .
```

## ⚙️ Configuration

### Basic Configuration

```yaml
apiVersion: hadoop.kubedoop.dev/v1
kind: HadoopCluster
metadata:
  name: my-cluster
spec:
  image: apache/hadoop:3.4.1
  imagePullPolicy: IfNotPresent

  # Cluster-wide configuration
  clusterConfig:
    replicationFactor: 3
    blockSize: 134217728
    zooKeeperConfigMapName: my-zk-config

  # Hadoop XML and environment configuration
  configSpec:
    logDir: /opt/hadoop/logs
    dataDir: /data/hadoop
    hadoopEnv:
      HDFS_HEAPSIZE: "4096"
      YARN_HEAPSIZE: "2048"
    coreSite:
      fs.trash.interval: "360"
    hdfsSite:
      dfs.replication: "3"
      dfs.blocksize: "134217728"
    yarnSite:
      yarn.nodemanager.resource.memory-mb: "16384"
    mapredSite:
      mapreduce.framework.name: "yarn"

  # Component specifications
  nameNodeSpec:
    replicas: 1
    resources:
      limits:
        cpu: "2"
        memory: 4Gi
      requests:
        cpu: "1"
        memory: 2Gi
    storage:
      storageClassName: standard
      resources:
        requests:
          storage: 50Gi

  dataNodeSpec:
    replicas: 3
    volumesPerNode: 1

  resourceManagerSpec:
    replicas: 1

  nodeManagerSpec:
    replicas: 3
```

### High Availability Configuration

```yaml
apiVersion: hadoop.kubedoop.dev/v1
kind: HadoopCluster
metadata:
  name: my-ha-cluster
spec:
  image: apache/hadoop:3.4.1

  # ZooKeeper configuration
  clusterConfig:
    zooKeeperConfigMapName: my-zk-config

  # HA configuration
  ha:
    nameNodeHA:
      enabled: true
      nameServiceId: ns1
      journalClusterId: jc1
      replicas: 2
    resourceManagerHA:
      enabled: true
      clusterId: rm-cluster
      replicas: 2

  # Components
  nameNodeSpec:
    replicas: 2
  journalNodeSpec:
    replicas: 3
  dataNodeSpec:
    replicas: 3
  resourceManagerSpec:
    replicas: 2
  nodeManagerSpec:
    replicas: 3
```

### HBase Configuration

```yaml
apiVersion: hadoop.kubedoop.dev/v1
kind: HadoopCluster
metadata:
  name: my-cluster-with-hbase
spec:
  image: apache/hadoop:3.4.1

  # ... other components ...

  hbaseSpec:
    enabled: true
    masterSpec:
      replicas: 2
      resources:
        limits:
          cpu: "2"
          memory: 4Gi
    regionServerSpec:
      replicas: 3
      resources:
        limits:
          cpu: "2"
          memory: 4Gi
    config:
      hbaseSite:
        hbase.cluster.distributed: "true"
        hbase.rootdir: "hdfs://my-cluster-namenode:9000/hbase"
      hbaseEnv:
        HBASE_HEAPSIZE: "4096"
```

### Role Groups (Advanced)

```yaml
apiVersion: hadoop.kubedoop.dev/v1
kind: HadoopCluster
metadata:
  name: my-cluster
spec:
  dataNodeSpec:
    roleGroups:
      default:
        replicas: 3
        resources:
          limits:
            cpu: "4"
            memory: 8Gi
        config:
          hdfsSite:
            dfs.datanode.du.reserved: "10737418240"
      high-io:
        replicas: 2
        resources:
          limits:
            cpu: "8"
            memory: 16Gi
        nodeSelector:
          node-type: high-io
```

## 📚 CRD Reference

### HadoopCluster

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `spec.image` | string | No | Hadoop image (default: `apache/hadoop:3.4.1`) |
| `spec.imagePullPolicy` | PullPolicy | No | Image pull policy (default: `IfNotPresent`) |
| `spec.imagePullSecrets` | []LocalObjectReference | No | Image pull secrets |
| `spec.serviceAccountName` | string | No | ServiceAccount name (default: `hadoop-operator`) |
| `spec.clusterConfig` | ClusterConfigSpec | No | Cluster-wide configuration (replication factor, block size, ZooKeeper) |
| `spec.configSpec` | ConfigSpec | No | Hadoop XML and environment configuration |
| `spec.nameNodeSpec` | NameNodeSpec | No | NameNode configuration |
| `spec.dataNodeSpec` | DataNodeSpec | No | DataNode configuration |
| `spec.journalNodeSpec` | JournalNodeSpec | No | JournalNode configuration (required for HDFS HA) |
| `spec.resourceManagerSpec` | ResourceManagerSpec | No | ResourceManager configuration |
| `spec.nodeManagerSpec` | NodeManagerSpec | No | NodeManager configuration |
| `spec.hbaseSpec` | HBaseSpec | No | HBase configuration |
| `spec.ha` | HAConfig | No | High availability configuration |
| `spec.authentication` | AuthenticationSpec | No | Authentication configuration (TLS, Kerberos, OIDC) |
| `spec.federation` | FederationConfig | No | HDFS federation configuration |
| `spec.clusterOperation` | ClusterOperationSpec | No | Cluster operation settings (auto format, upgrade) |

### Component Spec Common Fields

| Field | Type | Description |
|-------|------|-------------|
| `replicas` | int32 | Number of replicas |
| `resources` | ResourceRequirements | CPU and memory limits/requests |
| `storage` | StorageSpec | Persistent storage configuration (PVC or EmptyDir) |
| `affinity` | Affinity | Pod affinity rules |
| `nodeSelector` | map[string]string | Node selector labels |
| `tolerations` | []Toleration | Pod tolerations |
| `image` | string | Override the default image |
| `imagePullPolicy` | PullPolicy | Image pull policy |
| `ports` | *Ports | Port configuration (component-specific) |
| `logging` | LoggingSpec | Logging configuration (console + file) |
| `roleGroups` | map[string]RoleGroupSpec | Role group configuration for heterogeneous workloads |
| `annotations` | map[string]string | Pod annotations |
| `labels` | map[string]string | Pod labels |

### Cluster Status

| Field | Type | Description |
|-------|------|-------------|
| `status.phase` | ClusterPhase | Cluster phase: Pending, Creating, Running, Upgrading, Deleting, Failed, Unknown |
| `status.nameNodeStatus` | ComponentStatus | NameNode component status |
| `status.dataNodeStatus` | ComponentStatus | DataNode component status |
| `status.journalNodeStatus` | ComponentStatus | JournalNode component status |
| `status.resourceManagerStatus` | ComponentStatus | ResourceManager component status |
| `status.nodeManagerStatus` | ComponentStatus | NodeManager component status |
| `status.hbaseStatus` | HBaseStatus | HBase Master and RegionServer status |
| `status.conditions` | []ClusterCondition | Cluster conditions (Ready, ConfigReady, ComponentReady, etc.) |

### HadoopApplication

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `spec.clusterRef` | ClusterRef | Yes | Reference to the target HadoopCluster |
| `spec.type` | ApplicationType | Yes | Application type: mapreduce, spark, hive, hbase, pig, sqoop |
| `spec.jarFile` | string | No | Application JAR file path |
| `spec.mainClass` | string | No | Main class |
| `spec.args` | []string | No | Command line arguments |
| `spec.env` | []EnvVar | No | Environment variables |
| `spec.resources` | ResourceRequirements | No | Resource requirements |
| `spec.config` | ApplicationConfig | No | Application-specific configuration |

## 🔧 Development

### Prerequisites

- Go 1.21+
- Docker or Podman
- kubectl
- kind (for local testing)

### Setup Development Environment

```bash
# Clone the repository
git clone https://github.com/hadoop-operator/hadoop-k8s-operator.git
cd hadoop-k8s-operator/operator

# Install dependencies
go mod download

# Build the operator
go build -o bin/hadoop-operator ./cmd/main.go

# Run tests
go test ./...
```

### Run the Operator Locally

```bash
# Create a kind cluster
kind create cluster --name=hadoop-operator-dev

# Install CRDs
kubectl apply -f config/crd/hadoopcluster-crd.yaml

# Apply RBAC
kubectl apply -f config/rbac/role.yaml

# Run the operator locally
go run ./cmd/main.go
```

### Build Docker Image

```bash
# Build the image
docker build -t <your-registry>/hadoop-operator:latest -f ../docker/hadoop/Dockerfile .

# Push the image
docker push <your-registry>/hadoop-operator:latest
```

### Project Structure

```
operator/
├── cmd/main.go                          # Entry point
├── pkg/apis/hadoop/v1/                  # CRD type definitions
│   ├── hadoopcluster_types.go           # HadoopCluster + HadoopApplication types
│   ├── hadoopcluster_webhook.go         # Admission webhook (validation/defaulting)
│   └── groupversion_info.go             # GroupVersion: hadoop.kubedoop.dev/v1
├── pkg/builder/                         # Component Builders
│   ├── builder.go                       # Builder interface + BuilderFactory + computePhase()
│   ├── configmap_builder.go             # ConfigMap + entrypoint.sh generation
│   ├── nn_dn_builder.go                 # NameNode/DataNode StatefulSet + Services
│   ├── journalnode_builder.go           # JournalNode StatefulSet + Services (HDFS HA)
│   ├── rm_nm_builder.go                 # ResourceManager Deployment + NodeManager DaemonSet
│   └── hbase_builder.go                 # HBase Master/RegionServer StatefulSet + Services
├── pkg/controller/
│   ├── hadoopcluster_controller.go      # HadoopCluster reconciliation loop
│   └── hadoopapplication_controller.go  # HadoopApplication lifecycle management
└── config/
    ├── crd/hadoopcluster-crd.yaml       # CRD definition
    ├── rbac/role.yaml                   # ClusterRole, ServiceAccount, Bindings
    ├── deploy/operator-deployment.yaml  # Operator Deployment
    ├── webhook/webhook.yaml             # Admission webhook config
    └── samples/                         # Example CRD manifests
```

## 🚢 Deployment

### Production Deployment

```bash
# Install CRDs
kubectl apply -f operator/config/crd/hadoopcluster-crd.yaml

# Create namespace
kubectl create namespace hadoop-system

# Apply RBAC
kubectl apply -f operator/config/rbac/role.yaml

# Deploy the operator
kubectl apply -f operator/config/deploy/operator-deployment.yaml -n hadoop-system

# Verify the installation
kubectl get pods -n hadoop-system
```

### Uninstalling

```bash
# Delete all HadoopClusters first
kubectl delete hadoopcluster --all --all-namespaces

# Delete the operator
kubectl delete -f operator/config/deploy/operator-deployment.yaml -n hadoop-system

# Delete CRDs (careful - this will delete all HadoopClusters)
kubectl delete crd hadoopclusters.hadoop.kubedoop.dev
```

## 📝 Examples

Examples are available in the [examples](examples/) directory:

- [Simple Cluster](examples/hadoopcluster-simple.yaml) - Basic cluster for development (EmptyDir storage)
- [HA Cluster](examples/hadoopcluster-ha.yaml) - High availability cluster with JournalNode and ZooKeeper failover
- [Cluster with HBase](examples/hadoopcluster-with-hbase.yaml) - Cluster with HBase enabled
- [Custom Configuration](examples/hadoopcluster-custom-config.yaml) - Advanced configuration with affinity, node selectors, and full XML overrides

Additional examples in [operator/config/samples/](operator/config/samples/):
- [Sample Cluster](operator/config/samples/hadoop_v1alpha1_hadoopcluster.yaml)
- [Sample HA Cluster](operator/config/samples/hadoop_v1alpha1_hadoopcluster-ha.yaml)
- [Sample with HBase](operator/config/samples/hadoop_v1alpha1_hadoopcluster-with-hbase.yaml)

## 🔍 Troubleshooting

### Common Issues

#### Pods not starting

```bash
# Check pod status
kubectl get pods -n <namespace>

# View pod events
kubectl describe pod <pod-name> -n <namespace>

# View pod logs
kubectl logs <pod-name> -n <namespace>
```

#### Storage issues

```bash
# Check PVC status
kubectl get pvc -n <namespace>

# Check storage class
kubectl get storageclass
```

#### Operator issues

```bash
# Check operator logs
kubectl logs -n hadoop-system deployment/hadoop-operator

# Check operator status
kubectl get deployment -n hadoop-system
```

#### Verify component status

```bash
# Check all components for a specific cluster
kubectl get pods -n <namespace> -l hadoop.kubedoop.dev/cluster=<cluster-name>

# Check specific component
kubectl get pods -n <namespace> -l hadoop.kubedoop.dev/component=namenode
kubectl get pods -n <namespace> -l hadoop.kubedoop.dev/component=datanode
kubectl get pods -n <namespace> -l hadoop.kubedoop.dev/component=journalnode
kubectl get pods -n <namespace> -l hadoop.kubedoop.dev/component=resourcemanager
```

## 🤝 Contributing

Contributions are welcome! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Kubedoop](https://kubedoop.dev/) - The data platform this operator is part of
- [zncdatadev/hdfs-operator](https://github.com/zncdatadev/hdfs-operator) - Reference for operator design
- [chriskery/hadoop-operator](https://github.com/chriskery/hadoop-operator) - Initial inspiration
- [Apache Hadoop](https://hadoop.apache.org/) - The distributed computing framework

## 📬 Contact

- GitHub Issues: [https://github.com/hadoop-operator/hadoop-k8s-operator/issues](https://github.com/hadoop-operator/hadoop-k8s-operator/issues)

## 🔗 Related Projects

- [zookeeper-operator](https://github.com/zncdatadev/zookeeper-operator) - ZooKeeper operator
- [commons-operator](https://github.com/zncdatadev/commons-operator) - Common operator utilities
- [listener-operator](https://github.com/zncdatadev/listener-operator) - Listener operator
- [secret-operator](https://github.com/zncdatadev/secret-operator) - Secret operator
