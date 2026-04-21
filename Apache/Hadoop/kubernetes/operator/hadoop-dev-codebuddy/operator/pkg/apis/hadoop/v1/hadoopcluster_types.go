package v1

import (
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HadoopClusterSpec defines the desired state of HadoopCluster
type HadoopClusterSpec struct {
	// Hadoop image repository for all components
	// +optional
	// +kubebuilder:default="apache/hadoop:3.4.1"
	Image string `json:"image,omitempty"`
	// Hadoop image pull policy
	// +optional
	// +kubebuilder:default="IfNotPresent"
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// Image pull secrets for private registry
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// Service account for Hadoop components
	// +optional
	// +kubebuilder:default="hadoop-operator"
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// Cluster operation specification
	// +optional
	ClusterOperation *ClusterOperationSpec `json:"clusterOperation,omitempty"`

	// Cluster configuration
	// +optional
	ClusterConfig *ClusterConfigSpec `json:"clusterConfig,omitempty"`

	// NameNode specification
	// +optional
	NameNodeSpec *NameNodeSpec `json:"nameNodeSpec,omitempty"`
	// DataNode specification
	// +optional
	DataNodeSpec *DataNodeSpec `json:"dataNodeSpec,omitempty"`
	// JournalNode specification (required for HDFS HA)
	// +optional
	JournalNodeSpec *JournalNodeSpec `json:"journalNodeSpec,omitempty"`
	// ResourceManager specification
	// +optional
	ResourceManagerSpec *ResourceManagerSpec `json:"resourceManagerSpec,omitempty"`
	// NodeManager specification
	// +optional
	NodeManagerSpec *NodeManagerSpec `json:"nodeManagerSpec,omitempty"`

	// HBase configuration
	// +optional
	HBaseSpec *HBaseSpec `json:"hbaseSpec,omitempty"`

	// Configuration specification
	// +optional
	ConfigSpec *ConfigSpec `json:"configSpec,omitempty"`
	// Hadoop authentication configuration
	// +optional
	Authentication *AuthenticationSpec `json:"authentication,omitempty"`
	// High Availability configuration
	// +optional
	HA *HAConfig `json:"ha,omitempty"`
	// Federation configuration for HDFS federation
	// +optional
	Federation *FederationConfig `json:"federation,omitempty"`

	// Pod disruption budget for the cluster
	// +optional
	PodDisruptionBudget *policyv1.PodDisruptionBudgetSpec `json:"podDisruptionBudget,omitempty"`
}

// ClusterOperationSpec defines the cluster operation specification
type ClusterOperationSpec struct {
	// Cluster operation provisioner
	// +optional
	Provisioner string `json:"provisioner,omitempty"`
	// Cluster operation settings
	// +optional
	Settings *ClusterOperationSettings `json:"settings,omitempty"`
}

// ClusterOperationSettings defines cluster operation settings
type ClusterOperationSettings struct {
	// Whether to auto format the namenode on format failure
	// +optional
	AutoFormat bool `json:"autoFormat,omitempty"`
	// Upgrade settings
	// +optional
	Upgrade *UpgradeSettings `json:"upgrade,omitempty"`
}

// UpgradeSettings defines upgrade settings
type UpgradeSettings struct {
	// Type of upgrade (rolling, full)
	// +optional
	Type string `json:"type,omitempty"`
}

// ClusterConfigSpec defines cluster-level configuration
type ClusterConfigSpec struct {
	// ZooKeeper config map name (required for HA)
	// +optional
	ZooKeeperConfigMapName string `json:"zooKeeperConfigMapName,omitempty"`
	// Cluster domain
	// +optional
	ClusterDomain string `json:"clusterDomain,omitempty"`
	// Service configuration
	// +optional
	ServiceSpec *ServiceSpec `json:"service,omitempty"`
	// HDFS replication factor
	// +optional
	// +kubebuilder:default=3
	ReplicationFactor int32 `json:"replicationFactor,omitempty"`
	// HDFS block size
	// +optional
	// +kubebuilder:default=134217728
	BlockSize int64 `json:"blockSize,omitempty"`
	// Vector aggregator config map name
	// +optional
	VectorAggregatorConfigMapName string `json:"vectorAggregatorConfigMapName,omitempty"`
}

// ServiceSpec defines service configuration
type ServiceSpec struct {
	// Service type
	// +optional
	Type corev1.ServiceType `json:"type,omitempty"`
	// Annotations for the service
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

// AuthenticationSpec defines authentication configuration
type AuthenticationSpec struct {
	// TLS configuration
	// +optional
	TLS *TLSSpec `json:"tls,omitempty"`
	// Kerberos configuration
	// +optional
	Kerberos *KerberosSpec `json:"kerberos,omitempty"`
	// OIDC configuration
	// +optional
	OIDC *OIDCSpec `json:"oidc,omitempty"`
}

// TLSSpec defines TLS configuration
type TLSSpec struct {
	// Enable TLS
	// +optional
	Enabled bool `json:"enabled,omitempty"`
	// Keystore password secret reference
	// +optional
	KeystorePasswordSecretRef *SecretRef `json:"keystorePasswordSecretRef,omitempty"`
	// Truststore password secret reference
	// +optional
	TruststorePasswordSecretRef *SecretRef `json:"truststorePasswordSecretRef,omitempty"`
}

// SecretRef defines a reference to a secret
type SecretRef struct {
	// Name of the secret
	Name string `json:"name"`
	// Namespace of the secret
	// +optional
	Namespace string `json:"namespace,omitempty"`
	// Key in the secret
	// +optional
	Key string `json:"key,omitempty"`
}

// KerberosSpec defines Kerberos configuration
type KerberosSpec struct {
	// Enable Kerberos authentication
	// +optional
	Enabled bool `json:"enabled,omitempty"`
	// Kerberos realm
	// +optional
	Realm string `json:"realm,omitempty"`
	// KDC host
	// +optional
	KDCHost string `json:"kdcHost,omitempty"`
	// KDC port
	// +optional
	// +kubebuilder:default=88
	KDCPort int32 `json:"kdcPort,omitempty"`
	// Admin principal
	// +optional
	AdminPrincipal string `json:"adminPrincipal,omitempty"`
	// Keytab secret reference
	// +optional
	KeytabSecretRef *SecretRef `json:"keytabSecretRef,omitempty"`
	// krb5.conf ConfigMap name
	// +optional
	Krb5ConfConfigMapName string `json:"krb5ConfConfigMapName,omitempty"`
}

// OIDCSpec defines OIDC configuration
type OIDCSpec struct {
	// Enable OIDC
	// +optional
	Enabled bool `json:"enabled,omitempty"`
	// OIDC issuer URL
	// +optional
	IssuerURL string `json:"issuerURL,omitempty"`
	// OIDC client ID
	// +optional
	ClientID string `json:"clientID,omitempty"`
	// OIDC client secret
	// +optional
	ClientSecret string `json:"clientSecret,omitempty"`
	// Additional scopes
	// +optional
	Scopes []string `json:"scopes,omitempty"`
}

// LoggingSpec defines logging configuration
type LoggingSpec struct {
	// Log level configuration
	// +optional
	Loggers map[string]LogLevel `json:"loggers,omitempty"`
	// Console configuration
	// +optional
	Console *ConsoleLoggingConfig `json:"console,omitempty"`
	// File configuration
	// +optional
	File *FileLoggingConfig `json:"file,omitempty"`
}

// LogLevel defines log level
type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
)

// ConsoleLoggingConfig defines console logging configuration
type ConsoleLoggingConfig struct {
	// Enable console logging
	// +optional
	Enabled bool `json:"enabled,omitempty"`
	// Log level for console
	// +optional
	// +kubebuilder:default="WARN"
	Level LogLevel `json:"level,omitempty"`
	// Format for console (PLAIN, JSON)
	// +optional
	Format string `json:"format,omitempty"`
}

// FileLoggingConfig defines file logging configuration
type FileLoggingConfig struct {
	// Enable file logging
	// +optional
	Enabled bool `json:"enabled,omitempty"`
	// Log level for file
	// +optional
	// +kubebuilder:default="ERROR"
	Level LogLevel `json:"level,omitempty"`
	// Directory for log files
	// +optional
	Dir string `json:"dir,omitempty"`
}

// RoleGroupSpec defines common role group specification
type RoleGroupSpec struct {
	// Role group name
	Name string `json:"name"`
	// Replicas for the role group
	// +optional
	Replicas int32 `json:"replicas,omitempty"`
	// Config for the role group
	// +optional
	Config *ConfigSpec `json:"config,omitempty"`
	// Resources for the role group
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// Affinity for the role group
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// Node selector
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Tolerations
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// Annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// Labels
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// Pod security context
	// +optional
	PodSecurityContext *corev1.PodSecurityContext `json:"podSecurityContext,omitempty"`
	// Security context
	// +optional
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`
}

// NameNodeSpec defines the specification for NameNode
type NameNodeSpec struct {
	// Role group configuration
	// +optional
	RoleGroups map[string]*NameNodeRoleGroupSpec `json:"roleGroups,omitempty"`

	// Standalone configuration (for backward compatibility)
	// +optional
	Replicas int32 `json:"replicas,omitempty"`
	// Image for NameNode (overrides global image)
	// +optional
	Image string `json:"image,omitempty"`
	// Image pull policy (overrides global policy)
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// Pod affinity configuration
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// Node selector labels
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Tolerations for the pod
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// Resource requirements
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// Storage configuration
	// +optional
	Storage *StorageSpec `json:"storage,omitempty"`
	// Environment variables
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`
	// Volume mounts for the pod
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
	// Volumes for the pod
	// +optional
	Volumes []corev1.Volume `json:"volumes,omitempty"`
	// Liveness probe configuration
	// +optional
	LivenessProbe *corev1.Probe `json:"livenessProbe,omitempty"`
	// Readiness probe configuration
	// +optional
	ReadinessProbe *corev1.Probe `json:"readinessProbe,omitempty"`
	// Annotations for the pod
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// Labels for the pod
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// Pod security context
	// +optional
	PodSecurityContext *corev1.PodSecurityContext `json:"podSecurityContext,omitempty"`
	// Security context
	// +optional
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`
	// Service type (ClusterIP, NodePort, LoadBalancer)
	// +optional
	// +kubebuilder:default="ClusterIP"
	ServiceType corev1.ServiceType `json:"serviceType,omitempty"`
	// Port configuration
	// +optional
	Ports *NameNodePorts `json:"ports,omitempty"`
	// Init containers
	// +optional
	InitContainers []corev1.Container `json:"initContainers,omitempty"`
	// Sidecar containers
	// +optional
	Sidecars []corev1.Container `json:"sidecars,omitempty"`
	// Whether to format HDFS on startup
	// +optional
	// +kubebuilder:default=false
	FormatOnStartup bool `json:"formatOnStartup,omitempty"`
	// Name directory for HDFS metadata
	// +optional
	// +kubebuilder:default="/tmp/hadoop/dfs/name"
	NameDir string `json:"nameDir,omitempty"`
	// Logging configuration
	// +optional
	Logging *LoggingSpec `json:"logging,omitempty"`
}

// NameNodeRoleGroupSpec defines the NameNode role group specification
type NameNodeRoleGroupSpec struct {
	// Replicas
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=1
	Replicas int32 `json:"replicas,omitempty"`
	// Config
	// +optional
	Config *ConfigSpec `json:"config,omitempty"`
	// Resources
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// Affinity
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// Node selector
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Tolerations
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// Annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// Labels
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// Pod security context
	// +optional
	PodSecurityContext *corev1.PodSecurityContext `json:"podSecurityContext,omitempty"`
	// Security context
	// +optional
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`
	// Storage
	// +optional
	Storage *StorageSpec `json:"storage,omitempty"`
	// Logging
	// +optional
	Logging *LoggingSpec `json:"logging,omitempty"`
}

// NameNodePorts defines ports for NameNode
type NameNodePorts struct {
	// RPC port for NameNode
	// +optional
	// +kubebuilder:default=9000
	RPCPort int32 `json:"rpcPort,omitempty"`
	// HTTP port for NameNode
	// +optional
	// +kubebuilder:default=9870
	HTTPPort int32 `json:"httpPort,omitempty"`
	// HTTPS port for NameNode
	// +optional
	// +kubebuilder:default=9871
	HTTPSPort int32 `json:"httpsPort,omitempty"`
}

// DataNodeSpec defines the specification for DataNode
type DataNodeSpec struct {
	// Role group configuration
	// +optional
	RoleGroups map[string]*DataNodeRoleGroupSpec `json:"roleGroups,omitempty"`

	// Standalone configuration (for backward compatibility)
	// +optional
	Replicas int32 `json:"replicas,omitempty"`
	// Image for DataNode (overrides global image)
	// +optional
	Image string `json:"image,omitempty"`
	// Image pull policy (overrides global policy)
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// Pod affinity configuration
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// Node selector labels
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Tolerations for the pod
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// Resource requirements
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// Storage configuration (per volume)
	// +optional
	Storage *StorageSpec `json:"storage,omitempty"`
	// Number of storage volumes per DataNode
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=1
	VolumesPerNode int32 `json:"volumesPerNode,omitempty"`
	// Environment variables
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`
	// Volume mounts for the pod
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
	// Volumes for the pod
	// +optional
	Volumes []corev1.Volume `json:"volumes,omitempty"`
	// Liveness probe configuration
	// +optional
	LivenessProbe *corev1.Probe `json:"livenessProbe,omitempty"`
	// Readiness probe configuration
	// +optional
	ReadinessProbe *corev1.Probe `json:"readinessProbe,omitempty"`
	// Annotations for the pod
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// Labels for the pod
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// Pod security context
	// +optional
	PodSecurityContext *corev1.PodSecurityContext `json:"podSecurityContext,omitempty"`
	// Security context
	// +optional
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`
	// Service type
	// +optional
	ServiceType corev1.ServiceType `json:"serviceType,omitempty"`
	// Port configuration
	// +optional
	Ports *DataNodePorts `json:"ports,omitempty"`
	// Init containers
	// +optional
	InitContainers []corev1.Container `json:"initContainers,omitempty"`
	// Sidecar containers
	// +optional
	Sidecars []corev1.Container `json:"sidecars,omitempty"`
	// Logging configuration
	// +optional
	Logging *LoggingSpec `json:"logging,omitempty"`
}

// DataNodeRoleGroupSpec defines the DataNode role group specification
type DataNodeRoleGroupSpec struct {
	// Replicas
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=3
	Replicas int32 `json:"replicas,omitempty"`
	// Config
	// +optional
	Config *ConfigSpec `json:"config,omitempty"`
	// Resources
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// Affinity
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// Node selector
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Tolerations
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// Annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// Labels
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// Pod security context
	// +optional
	PodSecurityContext *corev1.PodSecurityContext `json:"podSecurityContext,omitempty"`
	// Security context
	// +optional
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`
	// Storage
	// +optional
	Storage *StorageSpec `json:"storage,omitempty"`
	// Volumes per node
	// +optional
	VolumesPerNode int32 `json:"volumesPerNode,omitempty"`
	// Logging
	// +optional
	Logging *LoggingSpec `json:"logging,omitempty"`
}

// DataNodePorts defines ports for DataNode
type DataNodePorts struct {
	// Data transfer port
	// +optional
	// +kubebuilder:default=9866
	DataPort int32 `json:"dataPort,omitempty"`
	// IPC port for DataNode
	// +optional
	// +kubebuilder:default=9867
	IPCPort int32 `json:"ipcPort,omitempty"`
	// HTTP port for DataNode
	// +optional
	// +kubebuilder:default=9864
	HTTPPort int32 `json:"httpPort,omitempty"`
	// HTTPS port for DataNode
	// +optional
	// +kubebuilder:default=9865
	HTTPSPort int32 `json:"httpsPort,omitempty"`
}

// JournalNodeSpec defines the specification for JournalNode
type JournalNodeSpec struct {
	// Role group configuration
	// +optional
	RoleGroups map[string]*JournalNodeRoleGroupSpec `json:"roleGroups,omitempty"`

	// Standalone configuration (for backward compatibility)
	// +optional
	Replicas int32 `json:"replicas,omitempty"`
	// Image for JournalNode (overrides global image)
	// +optional
	Image string `json:"image,omitempty"`
	// Image pull policy (overrides global policy)
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// Pod affinity configuration
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// Node selector labels
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Tolerations for the pod
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// Resource requirements
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// Storage configuration
	// +optional
	Storage *StorageSpec `json:"storage,omitempty"`
	// Environment variables
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`
	// Volume mounts for the pod
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
	// Annotations for the pod
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// Labels for the pod
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// Pod security context
	// +optional
	PodSecurityContext *corev1.PodSecurityContext `json:"podSecurityContext,omitempty"`
	// Security context
	// +optional
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`
	// Service type
	// +optional
	ServiceType corev1.ServiceType `json:"serviceType,omitempty"`
	// Port configuration
	// +optional
	Ports *JournalNodePorts `json:"ports,omitempty"`
	// Init containers
	// +optional
	InitContainers []corev1.Container `json:"initContainers,omitempty"`
	// Sidecar containers
	// +optional
	Sidecars []corev1.Container `json:"sidecars,omitempty"`
	// Logging configuration
	// +optional
	Logging *LoggingSpec `json:"logging,omitempty"`
}

// JournalNodeRoleGroupSpec defines the JournalNode role group specification
type JournalNodeRoleGroupSpec struct {
	// Replicas
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=3
	Replicas int32 `json:"replicas,omitempty"`
	// Config
	// +optional
	Config *ConfigSpec `json:"config,omitempty"`
	// Resources
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// Affinity
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// Node selector
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Tolerations
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// Annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// Labels
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// Pod security context
	// +optional
	PodSecurityContext *corev1.PodSecurityContext `json:"podSecurityContext,omitempty"`
	// Security context
	// +optional
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`
	// Storage
	// +optional
	Storage *StorageSpec `json:"storage,omitempty"`
	// Logging
	// +optional
	Logging *LoggingSpec `json:"logging,omitempty"`
}

// JournalNodePorts defines ports for JournalNode
type JournalNodePorts struct {
	// RPC port for JournalNode
	// +optional
	// +kubebuilder:default=8485
	RPCPort int32 `json:"rpcPort,omitempty"`
	// HTTP port for JournalNode
	// +optional
	// +kubebuilder:default=8480
	HTTPPort int32 `json:"httpPort,omitempty"`
}

// ResourceManagerSpec defines the specification for ResourceManager
type ResourceManagerSpec struct {
	// Role group configuration
	// +optional
	RoleGroups map[string]*ResourceManagerRoleGroupSpec `json:"roleGroups,omitempty"`

	// Standalone configuration (for backward compatibility)
	// +optional
	Replicas int32 `json:"replicas,omitempty"`
	// Image for ResourceManager (overrides global image)
	// +optional
	Image string `json:"image,omitempty"`
	// Image pull policy (overrides global policy)
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// Pod affinity configuration
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// Node selector labels
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Tolerations for the pod
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// Resource requirements
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// Environment variables
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`
	// Volume mounts for the pod
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
	// Volumes for the pod
	// +optional
	Volumes []corev1.Volume `json:"volumes,omitempty"`
	// Liveness probe configuration
	// +optional
	LivenessProbe *corev1.Probe `json:"livenessProbe,omitempty"`
	// Readiness probe configuration
	// +optional
	ReadinessProbe *corev1.Probe `json:"readinessProbe,omitempty"`
	// Annotations for the pod
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// Labels for the pod
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// Pod security context
	// +optional
	PodSecurityContext *corev1.PodSecurityContext `json:"podSecurityContext,omitempty"`
	// Security context
	// +optional
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`
	// Service type
	// +optional
	// +kubebuilder:default="ClusterIP"
	ServiceType corev1.ServiceType `json:"serviceType,omitempty"`
	// Port configuration
	// +optional
	Ports *ResourceManagerPorts `json:"ports,omitempty"`
	// Init containers
	// +optional
	InitContainers []corev1.Container `json:"initContainers,omitempty"`
	// Sidecar containers
	// +optional
	Sidecars []corev1.Container `json:"sidecars,omitempty"`
	// Logging configuration
	// +optional
	Logging *LoggingSpec `json:"logging,omitempty"`
}

// ResourceManagerRoleGroupSpec defines the ResourceManager role group specification
type ResourceManagerRoleGroupSpec struct {
	// Replicas
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=1
	Replicas int32 `json:"replicas,omitempty"`
	// Config
	// +optional
	Config *ConfigSpec `json:"config,omitempty"`
	// Resources
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// Affinity
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// Node selector
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Tolerations
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// Annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// Labels
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// Pod security context
	// +optional
	PodSecurityContext *corev1.PodSecurityContext `json:"podSecurityContext,omitempty"`
	// Security context
	// +optional
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`
	// Logging
	// +optional
	Logging *LoggingSpec `json:"logging,omitempty"`
}

// ResourceManagerPorts defines ports for ResourceManager
type ResourceManagerPorts struct {
	// RPC port for ResourceManager
	// +optional
	// +kubebuilder:default=8032
	RPCPort int32 `json:"rpcPort,omitempty"`
	// HTTP port for ResourceManager
	// +optional
	// +kubebuilder:default=8088
	HTTPPort int32 `json:"httpPort,omitempty"`
	// HTTPS port for ResourceManager
	// +optional
	// +kubebuilder:default=8090
	HTTPSPort int32 `json:"httpsPort,omitempty"`
	// Scheduler port
	// +optional
	// +kubebuilder:default=8030
	SchedulerPort int32 `json:"schedulerPort,omitempty"`
}

// NodeManagerSpec defines the specification for NodeManager
type NodeManagerSpec struct {
	// Role group configuration
	// +optional
	RoleGroups map[string]*NodeManagerRoleGroupSpec `json:"roleGroups,omitempty"`

	// Standalone configuration (for backward compatibility)
	// +optional
	Replicas int32 `json:"replicas,omitempty"`
	// Image for NodeManager (overrides global image)
	// +optional
	Image string `json:"image,omitempty"`
	// Image pull policy (overrides global policy)
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// Pod affinity configuration
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// Node selector labels
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Tolerations for the pod
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// Resource requirements
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// Environment variables
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`
	// Volume mounts for the pod
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
	// Volumes for the pod
	// +optional
	Volumes []corev1.Volume `json:"volumes,omitempty"`
	// Liveness probe configuration
	// +optional
	LivenessProbe *corev1.Probe `json:"livenessProbe,omitempty"`
	// Readiness probe configuration
	// +optional
	ReadinessProbe *corev1.Probe `json:"readinessProbe,omitempty"`
	// Annotations for the pod
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// Labels for the pod
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// Pod security context
	// +optional
	PodSecurityContext *corev1.PodSecurityContext `json:"podSecurityContext,omitempty"`
	// Security context
	// +optional
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`
	// Service type
	// +optional
	ServiceType corev1.ServiceType `json:"serviceType,omitempty"`
	// Port configuration
	// +optional
	Ports *NodeManagerPorts `json:"ports,omitempty"`
	// Init containers
	// +optional
	InitContainers []corev1.Container `json:"initContainers,omitempty"`
	// Sidecar containers
	// +optional
	Sidecars []corev1.Container `json:"sidecars,omitempty"`
	// Logging configuration
	// +optional
	Logging *LoggingSpec `json:"logging,omitempty"`
}

// NodeManagerRoleGroupSpec defines the NodeManager role group specification
type NodeManagerRoleGroupSpec struct {
	// Replicas
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=3
	Replicas int32 `json:"replicas,omitempty"`
	// Config
	// +optional
	Config *ConfigSpec `json:"config,omitempty"`
	// Resources
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// Affinity
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// Node selector
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Tolerations
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// Annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// Labels
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// Pod security context
	// +optional
	PodSecurityContext *corev1.PodSecurityContext `json:"podSecurityContext,omitempty"`
	// Security context
	// +optional
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`
	// Logging
	// +optional
	Logging *LoggingSpec `json:"logging,omitempty"`
}

// NodeManagerPorts defines ports for NodeManager
type NodeManagerPorts struct {
	// HTTP port for NodeManager
	// +optional
	// +kubebuilder:default=8042
	HTTPPort int32 `json:"httpPort,omitempty"`
	// HTTPS port for NodeManager
	// +optional
	// +kubebuilder:default=8044
	HTTPSPort int32 `json:"httpsPort,omitempty"`
	// Localizer port
	// +optional
	// +kubebuilder:default=8040
	LocalizerPort int32 `json:"localizerPort,omitempty"`
	// Shuffle port for MapReduce
	// +optional
	// +kubebuilder:default=13562
	ShufflePort int32 `json:"shufflePort,omitempty"`
}

// HBaseSpec defines the specification for HBase
type HBaseSpec struct {
	// Enable HBase
	// +optional
	// +kubebuilder:default=false
	Enabled bool `json:"enabled,omitempty"`
	// HBase Master specification
	// +optional
	MasterSpec *HBaseMasterSpec `json:"masterSpec,omitempty"`
	// HBase RegionServer specification
	// +optional
	RegionServerSpec *HBaseRegionServerSpec `json:"regionServerSpec,omitempty"`
	// HBase configuration
	// +optional
	Config *HBaseConfig `json:"config,omitempty"`
}

// HBaseMasterSpec defines the specification for HBase Master
type HBaseMasterSpec struct {
	// Role group configuration
	// +optional
	RoleGroups map[string]*HBaseMasterRoleGroupSpec `json:"roleGroups,omitempty"`

	// Standalone configuration
	// +optional
	Replicas int32 `json:"replicas,omitempty"`
	// Image for HBase Master
	// +optional
	Image string `json:"image,omitempty"`
	// Image pull policy
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// Pod affinity configuration
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// Node selector labels
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Tolerations for the pod
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// Resource requirements
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// Storage configuration
	// +optional
	Storage *StorageSpec `json:"storage,omitempty"`
	// Environment variables
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`
	// Volume mounts
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
	// Annotations for the pod
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// Labels for the pod
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// Pod security context
	// +optional
	PodSecurityContext *corev1.PodSecurityContext `json:"podSecurityContext,omitempty"`
	// Security context
	// +optional
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`
	// Service type
	// +optional
	ServiceType corev1.ServiceType `json:"serviceType,omitempty"`
	// Port configuration
	// +optional
	Ports *HBaseMasterPorts `json:"ports,omitempty"`
	// Logging configuration
	// +optional
	Logging *LoggingSpec `json:"logging,omitempty"`
}

// HBaseMasterRoleGroupSpec defines the HBase Master role group specification
type HBaseMasterRoleGroupSpec struct {
	// Replicas
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=1
	Replicas int32 `json:"replicas,omitempty"`
	// Config
	// +optional
	Config *ConfigSpec `json:"config,omitempty"`
	// Resources
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// Affinity
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// Node selector
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Tolerations
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// Annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// Labels
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// Pod security context
	// +optional
	PodSecurityContext *corev1.PodSecurityContext `json:"podSecurityContext,omitempty"`
	// Security context
	// +optional
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`
	// Storage
	// +optional
	Storage *StorageSpec `json:"storage,omitempty"`
	// Logging
	// +optional
	Logging *LoggingSpec `json:"logging,omitempty"`
}

// HBaseMasterPorts defines ports for HBase Master
type HBaseMasterPorts struct {
	// Master web UI port
	// +optional
	// +kubebuilder:default=16010
	WebUIPort int32 `json:"webUIPort,omitempty"`
	// Master port
	// +optional
	// +kubebuilder:default=16000
	Port int32 `json:"port,omitempty"`
}

// HBaseRegionServerSpec defines the specification for HBase RegionServer
type HBaseRegionServerSpec struct {
	// Role group configuration
	// +optional
	RoleGroups map[string]*HBaseRegionServerRoleGroupSpec `json:"roleGroups,omitempty"`

	// Standalone configuration
	// +optional
	Replicas int32 `json:"replicas,omitempty"`
	// Image for HBase RegionServer
	// +optional
	Image string `json:"image,omitempty"`
	// Image pull policy
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// Pod affinity configuration
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// Node selector labels
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Tolerations for the pod
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// Resource requirements
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// Storage configuration
	// +optional
	Storage *StorageSpec `json:"storage,omitempty"`
	// Number of storage volumes per RegionServer
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=1
	VolumesPerNode int32 `json:"volumesPerNode,omitempty"`
	// Environment variables
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`
	// Volume mounts
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
	// Annotations for the pod
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// Labels for the pod
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// Pod security context
	// +optional
	PodSecurityContext *corev1.PodSecurityContext `json:"podSecurityContext,omitempty"`
	// Security context
	// +optional
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`
	// Service type
	// +optional
	ServiceType corev1.ServiceType `json:"serviceType,omitempty"`
	// Port configuration
	// +optional
	Ports *HBaseRegionServerPorts `json:"ports,omitempty"`
	// Logging configuration
	// +optional
	Logging *LoggingSpec `json:"logging,omitempty"`
}

// HBaseRegionServerRoleGroupSpec defines the HBase RegionServer role group specification
type HBaseRegionServerRoleGroupSpec struct {
	// Replicas
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=3
	Replicas int32 `json:"replicas,omitempty"`
	// Config
	// +optional
	Config *ConfigSpec `json:"config,omitempty"`
	// Resources
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// Affinity
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// Node selector
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Tolerations
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// Annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// Labels
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// Pod security context
	// +optional
	PodSecurityContext *corev1.PodSecurityContext `json:"podSecurityContext,omitempty"`
	// Security context
	// +optional
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`
	// Storage
	// +optional
	Storage *StorageSpec `json:"storage,omitempty"`
	// Volumes per node
	// +optional
	VolumesPerNode int32 `json:"volumesPerNode,omitempty"`
	// Logging
	// +optional
	Logging *LoggingSpec `json:"logging,omitempty"`
}

// HBaseRegionServerPorts defines ports for HBase RegionServer
type HBaseRegionServerPorts struct {
	// RegionServer web UI port
	// +optional
	// +kubebuilder:default=16030
	WebUIPort int32 `json:"webUIPort,omitempty"`
	// RegionServer port
	// +optional
	// +kubebuilder:default=16020
	Port int32 `json:"port,omitempty"`
}

// HBaseConfig defines HBase configuration
type HBaseConfig struct {
	// HBase site configuration
	// +optional
	HBaseSite map[string]string `json:"hbaseSite,omitempty"`
	// HBase environment variables
	// +optional
	HBaseEnv map[string]string `json:"hbaseEnv,omitempty"`
}

// StorageSpec defines storage configuration
type StorageSpec struct {
	// Storage class name
	// +optional
	StorageClassName string `json:"storageClassName,omitempty"`
	// Access modes
	// +optional
	// +kubebuilder:default="ReadWriteOnce"
	AccessModes []corev1.PersistentVolumeAccessMode `json:"accessModes,omitempty"`
	// Resources for PVC
	// +optional
	Resources *corev1.PersistentVolumeClaimSpec `json:"resources,omitempty"`
	// Use empty dir instead of PVC
	// +optional
	UseEmptyDir bool `json:"useEmptyDir,omitempty"`
}

// ConfigSpec defines configuration specification
type ConfigSpec struct {
	// ConfigMap name for Hadoop configuration
	// +optional
	ConfigMapName string `json:"configMapName,omitempty"`
	// Core site configuration
	// +optional
	CoreSite map[string]string `json:"coreSite,omitempty"`
	// HDFS site configuration
	// +optional
	HDFSSite map[string]string `json:"hdfsSite,omitempty"`
	// YARN site configuration
	// +optional
	YARNSite map[string]string `json:"yarnSite,omitempty"`
	// MapReduce site configuration
	// +optional
	MapRedSite map[string]string `json:"mapredSite,omitempty"`
	// Hadoop environment variables
	// +optional
	HadoopEnv map[string]string `json:"hadoopEnv,omitempty"`
	// Log directory
	// +optional
	LogDir string `json:"logDir,omitempty"`
	// Data directory
	// +optional
	DataDir string `json:"dataDir,omitempty"`
	// Additional configs
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	Config map[string]string `json:"config,omitempty"`
}

// HAConfig defines High Availability configuration
type HAConfig struct {
	// Enable NameNode HA
	// +optional
	NameNodeHA *NameNodeHAConfig `json:"nameNodeHA,omitempty"`
	// Enable ResourceManager HA
	// +optional
	ResourceManagerHA *ResourceManagerHAConfig `json:"resourceManagerHA,omitempty"`
	// ZooKeeper configuration for HA
	// +optional
	ZooKeeper *ZooKeeperConfig `json:"zookeeper,omitempty"`
}

// ZooKeeperConfig defines ZooKeeper configuration for HA
type ZooKeeperConfig struct {
	// Enable embedded ZooKeeper
	// +optional
	Enabled bool `json:"enabled,omitempty"`
	// ZooKeeper ensemble replicas
	// +optional
	// +kubebuilder:default=3
	Replicas int32 `json:"replicas,omitempty"`
	// ZooKeeper port
	// +optional
	// +kubebuilder:default=2181
	Port int32 `json:"port,omitempty"`
	// ZooKeeper data directory
	// +optional
	DataDir string `json:"dataDir,omitempty"`
}

// NameNodeHAConfig defines NameNode HA configuration
type NameNodeHAConfig struct {
	// Enable NameNode HA
	// +optional
	Enabled bool `json:"enabled,omitempty"`
	// Nameservice ID
	// +optional
	NameServiceID string `json:"nameServiceId,omitempty"`
	// Journal cluster ID
	// +optional
	JournalClusterID string `json:"journalClusterId,omitempty"`
	// Number of NameNodes (must be 2)
	// +kubebuilder:validation:Minimum=2
	// +kubebuilder:default=2
	Replicas int32 `json:"replicas,omitempty"`
}

// ResourceManagerHAConfig defines ResourceManager HA configuration
type ResourceManagerHAConfig struct {
	// Enable ResourceManager HA
	// +optional
	Enabled bool `json:"enabled,omitempty"`
	// Cluster ID
	// +optional
	ClusterID string `json:"clusterId,omitempty"`
	// Number of ResourceManagers (must be 2)
	// +kubebuilder:validation:Minimum=2
	// +kubebuilder:default=2
	Replicas int32 `json:"replicas,omitempty"`
}

// FederationConfig defines HDFS federation configuration
type FederationConfig struct {
	// Enable HDFS federation
	// +optional
	Enabled bool `json:"enabled,omitempty"`
	// Nameservice ID for this cluster
	// +optional
	NameServiceID string `json:"nameServiceId,omitempty"`
	// Block pool ID
	// +optional
	BlockPoolID string `json:"blockPoolId,omitempty"`
}

// HadoopClusterStatus defines the observed state of HadoopCluster
type HadoopClusterStatus struct {
	// Phase represents the current state of the cluster
	// +optional
	Phase ClusterPhase `json:"phase,omitempty"`
	// NameNode status
	// +optional
	NameNodeStatus *ComponentStatus `json:"nameNodeStatus,omitempty"`
	// DataNode status
	// +optional
	DataNodeStatus *ComponentStatus `json:"dataNodeStatus,omitempty"`
	// JournalNode status
	// +optional
	JournalNodeStatus *ComponentStatus `json:"journalNodeStatus,omitempty"`
	// ResourceManager status
	// +optional
	ResourceManagerStatus *ComponentStatus `json:"resourceManagerStatus,omitempty"`
	// NodeManager status
	// +optional
	NodeManagerStatus *ComponentStatus `json:"nodeManagerStatus,omitempty"`
	// HBase status
	// +optional
	HBaseStatus *HBaseStatus `json:"hbaseStatus,omitempty"`
	// External service status
	// +optional
	ExternalService *ExternalServiceStatus `json:"externalService,omitempty"`
	// Conditions represents the current conditions of the cluster
	// +optional
	Conditions []ClusterCondition `json:"conditions,omitempty"`
	// ObservedGeneration represents the generation of the most recently observed cluster
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Message provides human-readable status information
	// +optional
	Message string `json:"message,omitempty"`
	// StartTime is the time when the cluster was first observed by the controller
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`
}

// RoleGroupStatus defines the status of a role group
type RoleGroupStatus struct {
	// Current state of the role group
	// +optional
	Conditions []ClusterCondition `json:"conditions,omitempty"`
	// Role group statuses
	// +optional
	RoleGroups map[string]*ComponentStatus `json:"roleGroups,omitempty"`
}

// ClusterPhase represents the phase of the cluster
type ClusterPhase string

const (
	ClusterPhasePending   ClusterPhase = "Pending"
	ClusterPhaseCreating  ClusterPhase = "Creating"
	ClusterPhaseRunning   ClusterPhase = "Running"
	ClusterPhaseUpgrading ClusterPhase = "Upgrading"
	ClusterPhaseDeleting  ClusterPhase = "Deleting"
	ClusterPhaseFailed    ClusterPhase = "Failed"
	ClusterPhaseUnknown   ClusterPhase = "Unknown"
)

// ComponentStatus defines the status of a component
type ComponentStatus struct {
	// ReadyReplicas is the number of ready replicas
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`
	// AvailableReplicas is the number of available replicas
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`
	// Replicas is the total number of replicas
	Replicas int32 `json:"replicas,omitempty"`
	// Phase is the current phase of the component
	Phase string `json:"phase,omitempty"`
	// Selector is the label selector for the StatefulSet/Deployment
	Selector string `json:"selector,omitempty"`
	// Message provides additional information
	// +optional
	Message string `json:"message,omitempty"`
}

// HBaseStatus defines the status of HBase components
type HBaseStatus struct {
	// Master status
	// +optional
	MasterStatus *ComponentStatus `json:"masterStatus,omitempty"`
	// RegionServer status
	// +optional
	RegionServerStatus *ComponentStatus `json:"regionServerStatus,omitempty"`
}

// ExternalServiceStatus defines external service information
type ExternalServiceStatus struct {
	// NameNode external address
	NameNodeAddress string `json:"nameNodeAddress,omitempty"`
	// ResourceManager external address
	ResourceManagerAddress string `json:"resourceManagerAddress,omitempty"`
	// Web UI URLs
	NameNodeWebURL  string `json:"nameNodeWebURL,omitempty"`
	ResourceManagerWebURL string `json:"resourceManagerWebURL,omitempty"`
	// HBase Web UI URLs
	HBaseMasterWebURL string `json:"hbaseMasterWebURL,omitempty"`
}

// ClusterCondition defines a condition
type ClusterCondition struct {
	// Type of condition
	Type ClusterConditionType `json:"type"`
	// Status of the condition
	Status corev1.ConditionStatus `json:"status"`
	// Message provides human-readable status information
	// +optional
	Message string `json:"message,omitempty"`
	// Reason provides a machine-readable status information
	// +optional
	Reason string `json:"reason,omitempty"`
	// LastTransitionTime is the last time the condition transitioned
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`
	// LastUpdateTime is the last time the condition was updated
	// +optional
	LastUpdateTime *metav1.Time `json:"lastUpdateTime,omitempty"`
}

// ClusterConditionType represents a condition type
type ClusterConditionType string

const (
	ConditionInitialized              ClusterConditionType = "Initialized"
	ConditionReady                    ClusterConditionType = "Ready"
	ConditionConfigReady              ClusterConditionType = "ConfigReady"
	ConditionNameNodeReady            ClusterConditionType = "NameNodeReady"
	ConditionDataNodeReady            ClusterConditionType = "DataNodeReady"
	ConditionJournalNodeReady         ClusterConditionType = "JournalNodeReady"
	ConditionResourceManagerReady     ClusterConditionType = "ResourceManagerReady"
	ConditionNodeManagerReady         ClusterConditionType = "NodeManagerReady"
	ConditionHBaseReady               ClusterConditionType = "HBaseReady"
	ConditionFailure                  ClusterConditionType = "Failure"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion
//+kubebuilder:resource:shortName=hc
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
//+kubebuilder:printcolumn:name="NameNode",type="integer",JSONPath=".spec.nameNodeSpec.replicas"
//+kubebuilder:printcolumn:name="DataNode",type="integer",JSONPath=".spec.dataNodeSpec.replicas"
//+kubebuilder:printcolumn:name="JournalNode",type="integer",JSONPath=".spec.journalNodeSpec.replicas"
//+kubebuilder:printcolumn:name="ResourceManager",type="integer",JSONPath=".spec.resourceManagerSpec.replicas"
//+kubebuilder:printcolumn:name="NodeManager",type="integer",JSONPath=".spec.nodeManagerSpec.replicas"
//+kubebuilder:printcolumn:name="HBase",type="string",JSONPath=".spec.hbaseSpec.enabled"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// HadoopCluster is the Schema for the hadoopclusters API
type HadoopCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HadoopClusterSpec   `json:"spec,omitempty"`
	Status HadoopClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// HadoopClusterList contains a list of HadoopCluster
type HadoopClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HadoopCluster `json:"items"`
}

// HadoopApplicationSpec defines the desired state of HadoopApplication
type HadoopApplicationSpec struct {
	// Hadoop cluster reference
	ClusterRef *ClusterRef `json:"clusterRef,omitempty"`
	// Application type: mapreduce, spark, hbase, etc.
	Type ApplicationType `json:"type"`
	// Application JAR file
	JarFile string `json:"jarFile,omitempty"`
	// Main class
	MainClass string `json:"mainClass,omitempty"`
	// Command line arguments
	Args []string `json:"args,omitempty"`
	// Environment variables
	Env []corev1.EnvVar `json:"env,omitempty"`
	// Resource requirements for the application master
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// Configuration for the application
	// +optional
	Config *ApplicationConfig `json:"config,omitempty"`
	// parallelism for the application
	// +optional
	Parallelism int32 `json:"parallelism,omitempty"`
}

// ClusterRef references to a HadoopCluster
type ClusterRef struct {
	// Name of the HadoopCluster
	Name string `json:"name"`
	// Namespace of the HadoopCluster
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// ApplicationType defines the type of Hadoop application
type ApplicationType string

const (
	ApplicationTypeMapReduce ApplicationType = "mapreduce"
	ApplicationTypeSpark    ApplicationType = "spark"
	ApplicationTypeHive      ApplicationType = "hive"
	ApplicationTypeHBase     ApplicationType = "hbase"
	ApplicationTypePig       ApplicationType = "pig"
	ApplicationTypeSqoop     ApplicationType = "sqoop"
)

// ApplicationConfig defines application configuration
type ApplicationConfig struct {
	// MapReduce configuration
	// +optional
	MapReduce map[string]string `json:"mapreduce,omitempty"`
	// Spark configuration
	// +optional
	Spark map[string]string `json:"spark,omitempty"`
}

// HadoopApplicationStatus defines the observed state of HadoopApplication
type HadoopApplicationStatus struct {
	// Application ID in YARN
	ApplicationID string `json:"applicationId,omitempty"`
	// Application state
	State ApplicationState `json:"state,omitempty"`
	// Application URL
	URL string `json:"url,omitempty"`
	// Final status (SUCCEEDED, FAILED, KILLED)
	FinalStatus FinalStatus `json:"finalStatus,omitempty"`
	// Message provides additional information
	Message string `json:"message,omitempty"`
	// Start time
	StartTime *metav1.Time `json:"startTime,omitempty"`
	// Completion time
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
	// Observed generation
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// ApplicationState defines the state of the application
type ApplicationState string

const (
	ApplicationStateNew         ApplicationState = "NEW"
	ApplicationStateNewSaving   ApplicationState = "NEW_SAVING"
	ApplicationStateSubmitted   ApplicationState = "SUBMITTED"
	ApplicationStateAccepted    ApplicationState = "ACCEPTED"
	ApplicationStateRunning     ApplicationState = "RUNNING"
	ApplicationStateFinished    ApplicationState = "FINISHED"
	ApplicationStateFailed      ApplicationState = "FAILED"
	ApplicationStateKilled      ApplicationState = "KILLED"
)

// FinalStatus defines the final status of the application
type FinalStatus string

const (
	FinalStatusUndetermined FinalStatus = "UNDEFINED"
	FinalStatusSucceeded    FinalStatus = "SUCCEEDED"
	FinalStatusFailed       FinalStatus = "FAILED"
	FinalStatusKilled       FinalStatus = "KILLED"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion
//+kubebuilder:resource:shortName=ha
//+kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".spec.clusterRef.name"
//+kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
//+kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
//+kubebuilder:printcolumn:name="ApplicationID",type="string",JSONPath=".status.applicationId"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// HadoopApplication is the Schema for the hadoopapplications API
type HadoopApplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HadoopApplicationSpec   `json:"spec,omitempty"`
	Status HadoopApplicationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// HadoopApplicationList contains a list of HadoopApplication
type HadoopApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HadoopApplication `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HadoopCluster{}, &HadoopClusterList{})
	SchemeBuilder.Register(&HadoopApplication{}, &HadoopApplicationList{})
}
