# YARN 配置

<cite>
**本文档引用的文件**
- [hadoopcluster_types.go](file://hadoop-operator/api/v1/hadoopcluster_types.go)
- [hadoopcluster_controller.go](file://hadoop-operator/internal/controller/hadoopcluster_controller.go)
- [yarn.go](file://hadoop-operator/internal/reconciler/yarn.go)
- [configmap.go](file://hadoop-operator/internal/reconciler/configmap.go)
- [hadoop_v1_hadoopcluster.yaml](file://hadoop-operator/config/samples/hadoop_v1_hadoopcluster.yaml)
- [hadoop_v1_hadoopcluster_ha.yaml](file://hadoop-operator/config/samples/hadoop_v1_hadoopcluster_ha.yaml)
- [offline-deployment.yaml](file://hadoop-operator/config/samples/offline-deployment.yaml)
- [resourcemanager.yaml](file://resourcemanager.yaml)
- [nodemanager.yaml](file://nodemanager.yaml)
- [configmap.yaml](file://configmap.yaml)
</cite>

## 目录
1. [简介](#简介)
2. [项目结构](#项目结构)
3. [核心组件](#核心组件)
4. [架构概览](#架构概览)
5. [详细组件分析](#详细组件分析)
6. [依赖关系分析](#依赖关系分析)
7. [性能考虑](#性能考虑)
8. [故障排除指南](#故障排除指南)
9. [结论](#结论)

## 简介

本文件为 YARN 组件配置的详细参考文档，基于 Apache Hadoop Operator 项目中的 YARN 实现。文档全面解释了 ResourceManagerSpec 和 NodeManagerSpec 的配置参数，包括副本数、资源限制、服务配置、高可用配置等。详细说明了 ResourceManager 的高可用模式配置和 NodeManager 的资源配置选项，并提供了完整的 YARN 配置示例，展示如何配置资源管理和节点管理。

## 项目结构

该项目采用 Kubernetes Operator 模式，通过自定义资源定义（CRD）来管理 Hadoop 集群。YARN 组件作为 HadoopCluster 资源的一部分进行配置和管理。

```mermaid
graph TB
subgraph "Hadoop Operator 项目结构"
A[HadoopCluster CRD] --> B[YARN 配置]
B --> C[ResourceManagerSpec]
B --> D[NodeManagerSpec]
C --> E[高可用配置]
C --> F[资源限制]
C --> G[服务配置]
D --> H[节点资源]
D --> I[网络配置]
J[配置映射] --> K[yarn-site.xml]
J --> L[capacity-scheduler.xml]
M[控制器] --> N[YARN 组件协调器]
N --> O[StatefulSet 管理]
N --> P[Service 管理]
end
```

**图表来源**
- [hadoopcluster_types.go:142-187](file://hadoop-operator/api/v1/hadoopcluster_types.go#L142-L187)
- [yarn.go:34-447](file://hadoop-operator/internal/reconciler/yarn.go#L34-L447)

**章节来源**
- [hadoopcluster_types.go:24-46](file://hadoop-operator/api/v1/hadoopcluster_types.go#L24-L46)
- [hadoopcluster_controller.go:104-127](file://hadoop-operator/internal/controller/hadoopcluster_controller.go#L104-L127)

## 核心组件

### YARNSpec 结构

YARNSpec 是 YARN 组件的顶层配置对象，包含 ResourceManager 和 NodeManager 的配置。

```mermaid
classDiagram
class YARNSpec {
+ResourceManagerSpec resourceManager
+NodeManagerSpec nodeManager
}
class ResourceManagerSpec {
+int32 replicas
+ResourceRequirements resources
+ServiceSpec service
+HASpec ha
+Affinity affinity
+Toleration[] tolerations
}
class NodeManagerSpec {
+int32 replicas
+ResourceRequirements resources
+ServiceSpec service
+Affinity affinity
+Toleration[] tolerations
}
YARNSpec --> ResourceManagerSpec
YARNSpec --> NodeManagerSpec
```

**图表来源**
- [hadoopcluster_types.go:142-187](file://hadoop-operator/api/v1/hadoopcluster_types.go#L142-L187)

### 高可用配置

高可用配置支持 NameNode 和 ResourceManager 的 HA 模式，通过 ZooKeeper 进行协调。

```mermaid
classDiagram
class HASpec {
+bool enabled
+ZooKeeperSpec zookeeper
+JournalNodeSpec journalNode
}
class ZooKeeperSpec {
+string connectionString
}
class JournalNodeSpec {
+int32 replicas
+ResourceRequirements resources
+StorageSpec storage
}
HASpec --> ZooKeeperSpec
HASpec --> JournalNodeSpec
```

**图表来源**
- [hadoopcluster_types.go:92-120](file://hadoop-operator/api/v1/hadoopcluster_types.go#L92-L120)

**章节来源**
- [hadoopcluster_types.go:142-187](file://hadoop-operator/api/v1/hadoopcluster_types.go#L142-L187)
- [hadoopcluster_types.go:92-120](file://hadoop-operator/api/v1/hadoopcluster_types.go#L92-L120)

## 架构概览

YARN 组件通过 Kubernetes StatefulSet 进行管理，每个组件都有对应的 Headless Service 和外部 Service。

```mermaid
graph TB
subgraph "YARN 组件架构"
A[ResourceManager StatefulSet] --> B[Headless Service]
A --> C[External Service]
D[NodeManager StatefulSet] --> E[Headless Service]
D --> F[External Service]
G[ConfigMap] --> H[yarn-site.xml]
G --> I[capacity-scheduler.xml]
J[Init Container] --> K[等待 ResourceManager 就绪]
L[探针配置] --> M[Liveness Probe]
L --> N[Readiness Probe]
end
```

**图表来源**
- [yarn.go:34-113](file://hadoop-operator/internal/reconciler/yarn.go#L34-L113)
- [yarn.go:242-310](file://hadoop-operator/internal/reconciler/yarn.go#L242-L310)

## 详细组件分析

### ResourceManager 配置

ResourceManager 是 YARN 集群的主控组件，负责集群资源的统一管理和调度。

#### 基本配置参数

| 参数 | 类型 | 描述 | 默认值 |
|------|------|------|--------|
| replicas | int32 | ResourceManager 副本数量 | 1 |
| resources.requests.cpu | string | CPU 请求量 | 500m |
| resources.requests.memory | string | 内存请求量 | 2Gi |
| resources.limits.cpu | string | CPU 限制量 | 1000m |
| resources.limits.memory | string | 内存限制量 | 4Gi |
| service.type | ServiceType | 服务类型 | NodePort |
| service.nodePorts.rpc | int32 | RPC 端口 | 自动分配 |
| service.nodePorts.web | int32 | Web UI 端口 | 自动分配 |

#### 高可用配置

ResourceManager 支持双活高可用模式，需要至少 2 个副本。

```mermaid
sequenceDiagram
participant Client as 客户端
participant RM1 as ResourceManager-0
participant RM2 as ResourceManager-1
participant ZK as ZooKeeper
Client->>RM1 : 请求资源
RM1->>ZK : 检查领导者状态
ZK-->>RM1 : 返回领导者信息
alt RM1 是领导者
RM1-->>Client : 处理请求
else RM1 不是领导者
RM1->>Client : 重定向到领导者
Client->>RM2 : 重新请求
RM2-->>Client : 处理请求
end
```

**图表来源**
- [yarn.go:115-240](file://hadoop-operator/internal/reconciler/yarn.go#L115-L240)
- [configmap.go:138-162](file://hadoop-operator/internal/reconciler/configmap.go#L138-L162)

#### 资源配置策略

```mermaid
flowchart TD
Start([配置验证]) --> CheckReplicas{检查副本数}
CheckReplicas --> |HA 模式且副本数<2| SetDefaultReplicas[设置默认副本数=2]
CheckReplicas --> |正常| CheckResources{检查资源配置}
SetDefaultReplicas --> CheckResources
CheckResources --> |未设置请求| SetDefaultRequests[设置默认请求]
CheckResources --> |未设置限制| SetDefaultLimits[设置默认限制]
SetDefaultRequests --> CreatePod[创建 Pod]
SetDefaultLimits --> CreatePod
CheckResources --> CreatePod
CreatePod --> End([完成])
```

**图表来源**
- [yarn.go:119-150](file://hadoop-operator/internal/reconciler/yarn.go#L119-L150)

**章节来源**
- [hadoopcluster_types.go:150-169](file://hadoop-operator/api/v1/hadoopcluster_types.go#L150-L169)
- [yarn.go:115-240](file://hadoop-operator/internal/reconciler/yarn.go#L115-L240)

### NodeManager 配置

NodeManager 是 YARN 集群的工作节点组件，负责单个节点上的资源管理和任务执行。

#### 基本配置参数

| 参数 | 类型 | 描述 | 默认值 |
|------|------|------|--------|
| replicas | int32 | NodeManager 副本数量 | 2 |
| resources.requests.cpu | string | CPU 请求量 | 500m |
| resources.requests.memory | string | 内存请求量 | 2Gi |
| resources.limits.cpu | string | CPU 限制量 | 1000m |
| resources.limits.memory | string | 内存限制量 | 4Gi |
| service.type | ServiceType | 服务类型 | NodePort |
| service.nodePorts.web | int32 | Web UI 端口 | 自动分配 |

#### 初始化容器配置

NodeManager 启动前需要等待 ResourceManager 就绪：

```mermaid
flowchart TD
Start([NodeManager 启动]) --> InitContainer[启动初始化容器]
InitContainer --> WaitRM[等待 ResourceManager 就绪]
WaitRM --> CheckPort{检查 8032 端口}
CheckPort --> |端口可达| StartNM[启动 NodeManager]
CheckPort --> |端口不可达| WaitRetry[等待后重试]
WaitRetry --> CheckPort
StartNM --> HealthCheck[健康检查]
HealthCheck --> End([完成])
```

**图表来源**
- [yarn.go:312-447](file://hadoop-operator/internal/reconciler/yarn.go#L312-L447)

**章节来源**
- [hadoopcluster_types.go:171-187](file://hadoop-operator/api/v1/hadoopcluster_types.go#L171-L187)
- [yarn.go:312-447](file://hadoop-operator/internal/reconciler/yarn.go#L312-L447)

### 服务发现配置

YARN 组件使用 Kubernetes DNS 进行服务发现，自动解析服务地址。

#### 服务地址解析

| 组件 | 服务名称 | 地址格式 | 端口 |
|------|----------|----------|------|
| ResourceManager | `{cluster}-resourcemanager` | `{pod}.{cluster}-resourcemanager.{namespace}.svc.cluster.local` | 8032(RPC), 8088(Web) |
| NodeManager | `{cluster}-nodemanager` | `{pod}.{cluster}-nodemanager.{namespace}.svc.cluster.local` | 8042(Web) |

**章节来源**
- [yarn.go:345-346](file://hadoop-operator/internal/reconciler/yarn.go#L345-L346)
- [configmap.go:74-75](file://hadoop-operator/internal/reconciler/configmap.go#L74-L75)

## 依赖关系分析

### 组件间依赖

```mermaid
graph TB
subgraph "配置依赖关系"
A[HadoopClusterSpec] --> B[YARNSpec]
B --> C[ResourceManagerSpec]
B --> D[NodeManagerSpec]
C --> E[HASpec]
C --> F[ServiceSpec]
C --> G[ResourceRequirements]
D --> H[ServiceSpec]
D --> I[ResourceRequirements]
J[ConfigMap] --> K[yarn-site.xml]
J --> L[capacity-scheduler.xml]
M[StatefulSet] --> N[PersistentVolumeClaim]
O[Service] --> P[Endpoints]
end
```

**图表来源**
- [hadoopcluster_types.go:24-46](file://hadoop-operator/api/v1/hadoopcluster_types.go#L24-L46)
- [yarn.go:34-447](file://hadoop-operator/internal/reconciler/yarn.go#L34-L447)

### 控制器协调流程

```mermaid
sequenceDiagram
participant Controller as HadoopClusterController
participant ConfigMap as ConfigMapReconciler
participant RM as ResourceManagerReconciler
participant NM as NodeManagerReconciler
Controller->>ConfigMap : 创建/更新配置映射
ConfigMap-->>Controller : 配置就绪
Controller->>RM : 创建/更新 ResourceManager
RM-->>Controller : RM 就绪
Controller->>NM : 创建/更新 NodeManager
NM-->>Controller : NM 就绪
Controller->>Controller : 更新集群状态
```

**图表来源**
- [hadoopcluster_controller.go:104-127](file://hadoop-operator/internal/controller/hadoopcluster_controller.go#L104-L127)

**章节来源**
- [hadoopcluster_controller.go:104-145](file://hadoop-operator/internal/controller/hadoopcluster_controller.go#L104-L145)

## 性能考虑

### 资源配置最佳实践

1. **CPU 配置**
   - 建议 ResourceManager 的 CPU 请求至少 500m，限制至少 1000m
   - NodeManager 的 CPU 请求至少 500m，限制至少 1000m
   - 根据集群规模调整副本数：生产环境建议至少 2 个 ResourceManager

2. **内存配置**
   - ResourceManager 内存请求建议 2Gi，限制 4Gi
   - NodeManager 内存请求建议 2Gi，限制 4Gi
   - 节点内存应根据工作负载类型进行调整

3. **存储配置**
   - 使用本地 SSD 存储以提高 I/O 性能
   - 配置适当的存储类和访问模式

### 调度器优化参数

| 参数 | 推荐值 | 说明 |
|------|--------|------|
| yarn.scheduler.maximum-allocation-mb | 8192 | 单任务最大内存 |
| yarn.scheduler.maximum-allocation-vcores | 4 | 单任务最大 CPU 核心数 |
| yarn.nodemanager.resource.memory-mb | 8192 | 节点总内存 |
| yarn.nodemanager.resource.cpu-vcores | 4 | 节点总 CPU 核心数 |

**章节来源**
- [hadoop_v1_hadoopcluster_ha.yaml:98-99](file://hadoop-operator/config/samples/hadoop_v1_hadoopcluster_ha.yaml#L98-L99)
- [offline-deployment.yaml:113-114](file://hadoop-operator/config/samples/offline-deployment.yaml#L113-L114)

## 故障排除指南

### 常见问题诊断

1. **ResourceManager 无法启动**
   - 检查 ZooKeeper 连接状态
   - 验证高可用配置是否正确
   - 查看 Pod 日志和事件

2. **NodeManager 注册失败**
   - 确认 ResourceManager 服务地址解析
   - 检查网络连通性和防火墙规则
   - 验证资源配额和限制

3. **服务发现问题**
   - 检查 DNS 解析是否正常
   - 验证 Service Selector 配置
   - 确认 Pod 标签匹配

### 配置验证

```mermaid
flowchart TD
Start([开始验证]) --> CheckConfig[检查配置语法]
CheckConfig --> ValidateSchema{验证模式}
ValidateSchema --> |通过| CheckServices[检查服务连通性]
ValidateSchema --> |失败| FixSchema[修复配置错误]
CheckServices --> TestRM[测试 ResourceManager]
CheckServices --> TestNM[测试 NodeManager]
TestRM --> RMReady{ResourceManager 就绪?}
TestNM --> NMReady{NodeManager 就绪?}
RMReady --> |否| DebugRM[调试 ResourceManager]
RMReady --> |是| NMReady
NMReady --> |否| DebugNM[调试 NodeManager]
NMReady --> |是| Complete[验证完成]
DebugRM --> FixRM[修复问题]
DebugNM --> FixNM[修复问题]
FixRM --> CheckConfig
FixNM --> CheckConfig
```

**图表来源**
- [hadoopcluster_controller.go:169-174](file://hadoop-operator/internal/controller/hadoopcluster_controller.go#L169-L174)

**章节来源**
- [hadoopcluster_controller.go:169-227](file://hadoop-operator/internal/controller/hadoopcluster_controller.go#L169-L227)

## 结论

本 YARN 配置参考文档详细介绍了基于 Kubernetes Operator 的 YARN 组件配置方法。通过 ResourceManagerSpec 和 NodeManagerSpec，用户可以灵活配置 YARN 集群的各项参数，包括副本数、资源限制、服务配置和高可用模式。文档提供了完整的配置示例和最佳实践建议，帮助用户构建稳定高效的 YARN 集群。

关键要点：
- ResourceManager 高可用模式需要至少 2 个副本
- NodeManager 启动前需要等待 ResourceManager 就绪
- 通过 ConfigMap 自动管理 Hadoop 配置文件
- 支持多种服务发现和服务暴露方式
- 提供完整的监控和故障排除机制