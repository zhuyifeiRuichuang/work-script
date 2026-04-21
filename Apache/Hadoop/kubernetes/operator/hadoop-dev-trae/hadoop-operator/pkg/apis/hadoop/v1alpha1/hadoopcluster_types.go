package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// HadoopClusterSpec defines the desired state of HadoopCluster
type HadoopClusterSpec struct {
	// Hadoop image to use
	Image string `json:"image,omitempty"`
	// NameNode configuration
	NameNode NameNodeSpec `json:"nameNode"`
	// DataNode configuration
	DataNode DataNodeSpec `json:"dataNode"`
	// ResourceManager configuration
	ResourceManager ResourceManagerSpec `json:"resourceManager"`
	// NodeManager configuration
	NodeManager NodeManagerSpec `json:"nodeManager"`
	// Hadoop configuration
	HadoopConfig HadoopConfigSpec `json:"hadoopConfig"`
}

// NameNodeSpec defines the configuration for NameNode
type NameNodeSpec struct {
	Replicas int32 `json:"replicas,omitempty"`
	Resources ResourceRequirements `json:"resources,omitempty"`
	PersistentVolumes []PersistentVolumeSpec `json:"persistentVolumes,omitempty"`
	Affinity *AffinitySpec `json:"affinity,omitempty"`
	SecurityContext *SecurityContext `json:"securityContext,omitempty"`
	ContainerSecurityContext *ContainerSecurityContext `json:"containerSecurityContext,omitempty"`
	EnvVars []EnvVar `json:"envVars,omitempty"`
	ReadinessProbePolicy *ProbePolicy `json:"readinessProbePolicy,omitempty"`
	LivenessProbePolicy *ProbePolicy `json:"livenessProbePolicy,omitempty"`
	PodLabels map[string]string `json:"podLabels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	ImagePullSecrets []LocalObjectReference `json:"imagePullSecrets,omitempty"`
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	Tolerations []Toleration `json:"tolerations,omitempty"`
	ConfigMaps []ConfigMapSpec `json:"configMaps,omitempty"`
	HostAliases []HostAlias `json:"hostAliases,omitempty"`
}

// DataNodeSpec defines the configuration for DataNode
type DataNodeSpec struct {
	Replicas int32 `json:"replicas,omitempty"`
	Resources ResourceRequirements `json:"resources,omitempty"`
	PersistentVolumes []PersistentVolumeSpec `json:"persistentVolumes,omitempty"`
	Affinity *AffinitySpec `json:"affinity,omitempty"`
	SecurityContext *SecurityContext `json:"securityContext,omitempty"`
	ContainerSecurityContext *ContainerSecurityContext `json:"containerSecurityContext,omitempty"`
	EnvVars []EnvVar `json:"envVars,omitempty"`
	ReadinessProbePolicy *ProbePolicy `json:"readinessProbePolicy,omitempty"`
	LivenessProbePolicy *ProbePolicy `json:"livenessProbePolicy,omitempty"`
	PodLabels map[string]string `json:"podLabels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	ImagePullSecrets []LocalObjectReference `json:"imagePullSecrets,omitempty"`
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	Tolerations []Toleration `json:"tolerations,omitempty"`
	ConfigMaps []ConfigMapSpec `json:"configMaps,omitempty"`
	HostAliases []HostAlias `json:"hostAliases,omitempty"`
}

// ResourceManagerSpec defines the configuration for ResourceManager
type ResourceManagerSpec struct {
	Replicas int32 `json:"replicas,omitempty"`
	Resources ResourceRequirements `json:"resources,omitempty"`
	Affinity *AffinitySpec `json:"affinity,omitempty"`
	SecurityContext *SecurityContext `json:"securityContext,omitempty"`
	ContainerSecurityContext *ContainerSecurityContext `json:"containerSecurityContext,omitempty"`
	EnvVars []EnvVar `json:"envVars,omitempty"`
	ReadinessProbePolicy *ProbePolicy `json:"readinessProbePolicy,omitempty"`
	LivenessProbePolicy *ProbePolicy `json:"livenessProbePolicy,omitempty"`
	PodLabels map[string]string `json:"podLabels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	ImagePullSecrets []LocalObjectReference `json:"imagePullSecrets,omitempty"`
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	Tolerations []Toleration `json:"tolerations,omitempty"`
	ConfigMaps []ConfigMapSpec `json:"configMaps,omitempty"`
	HostAliases []HostAlias `json:"hostAliases,omitempty"`
}

// NodeManagerSpec defines the configuration for NodeManager
type NodeManagerSpec struct {
	Replicas int32 `json:"replicas,omitempty"`
	Resources ResourceRequirements `json:"resources,omitempty"`
	Affinity *AffinitySpec `json:"affinity,omitempty"`
	SecurityContext *SecurityContext `json:"securityContext,omitempty"`
	ContainerSecurityContext *ContainerSecurityContext `json:"containerSecurityContext,omitempty"`
	EnvVars []EnvVar `json:"envVars,omitempty"`
	ReadinessProbePolicy *ProbePolicy `json:"readinessProbePolicy,omitempty"`
	LivenessProbePolicy *ProbePolicy `json:"livenessProbePolicy,omitempty"`
	PodLabels map[string]string `json:"podLabels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	ImagePullSecrets []LocalObjectReference `json:"imagePullSecrets,omitempty"`
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	Tolerations []Toleration `json:"tolerations,omitempty"`
	ConfigMaps []ConfigMapSpec `json:"configMaps,omitempty"`
	HostAliases []HostAlias `json:"hostAliases,omitempty"`
}

// ResourceRequirements defines the resource requirements for components
type ResourceRequirements struct {
	Requests ResourceList `json:"requests,omitempty"`
	Limits ResourceList `json:"limits,omitempty"`
}

// ResourceList defines the resource list for components
type ResourceList struct {
	CPU resource.Quantity `json:"cpu,omitempty"`
	Memory resource.Quantity `json:"memory,omitempty"`
}

// PersistentVolumeSpec defines the persistent volume configuration for components
type PersistentVolumeSpec struct {
	Name string `json:"name,omitempty"`
	MountPath string `json:"mountPath,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	PersistentVolumeClaimSpec PersistentVolumeClaimSpec `json:"persistentVolumeClaimSpec,omitempty"`
	Provisioner string `json:"provisioner,omitempty"`
}

// PersistentVolumeClaimSpec defines the persistent volume claim specification
type PersistentVolumeClaimSpec struct {
	AccessModes []string `json:"accessModes,omitempty"`
	Resources ResourceRequirements `json:"resources,omitempty"`
	StorageClassName string `json:"storageClassName,omitempty"`
	VolumeName string `json:"volumeName,omitempty"`
}

// SecurityContext defines the security context for pods
type SecurityContext struct {
	RunAsUser *int64 `json:"runAsUser,omitempty"`
	RunAsGroup *int64 `json:"runAsGroup,omitempty"`
	RunAsNonRoot *bool `json:"runAsNonRoot,omitempty"`
	FSGroup *int64 `json:"fsGroup,omitempty"`
}

// ContainerSecurityContext defines the security context for containers
type ContainerSecurityContext struct {
	RunAsUser *int64 `json:"runAsUser,omitempty"`
	RunAsGroup *int64 `json:"runAsGroup,omitempty"`
	RunAsNonRoot *bool `json:"runAsNonRoot,omitempty"`
	ReadOnlyRootFilesystem *bool `json:"readOnlyRootFilesystem,omitempty"`
	Privileged *bool `json:"privileged,omitempty"`
}

// EnvVar defines an environment variable for containers
type EnvVar struct {
	Name string `json:"name"`
	Value string `json:"value,omitempty"`
	ValueFrom *EnvVarSource `json:"valueFrom,omitempty"`
}

// EnvVarSource defines the source of an environment variable
type EnvVarSource struct {
	ConfigMapKeyRef *ConfigMapKeySelector `json:"configMapKeyRef,omitempty"`
	SecretKeyRef *SecretKeySelector `json:"secretKeyRef,omitempty"`
}

// ConfigMapKeySelector defines a selector for a ConfigMap key
type ConfigMapKeySelector struct {
	Name string `json:"name,omitempty"`
	Key string `json:"key"`
	Optional *bool `json:"optional,omitempty"`
}

// SecretKeySelector defines a selector for a Secret key
type SecretKeySelector struct {
	Name string `json:"name,omitempty"`
	Key string `json:"key"`
	Optional *bool `json:"optional,omitempty"`
}

// ProbePolicy defines the probe policy for containers
type ProbePolicy struct {
	PeriodSeconds *int32 `json:"periodSeconds,omitempty"`
	TimeoutSeconds *int32 `json:"timeoutSeconds,omitempty"`
	FailureThreshold *int32 `json:"failureThreshold,omitempty"`
}

// LocalObjectReference defines a reference to a local object
type LocalObjectReference struct {
	Name string `json:"name,omitempty"`
}

// Toleration defines a toleration for pods
type Toleration struct {
	Key string `json:"key,omitempty"`
	Operator string `json:"operator,omitempty"`
	Value string `json:"value,omitempty"`
	Effect string `json:"effect,omitempty"`
	TolerationSeconds *int64 `json:"tolerationSeconds,omitempty"`
}

// ConfigMapSpec defines a ConfigMap for pods
type ConfigMapSpec struct {
	ConfigMapName string `json:"configMapName"`
	MountPath string `json:"mountPath"`
}

// HostAlias defines a host alias for pods
type HostAlias struct {
	IP string `json:"ip"`
	Hostnames []string `json:"hostnames"`
}

// AffinitySpec defines the affinity rules for components
type AffinitySpec struct {
	NodeAffinity *NodeAffinitySpec `json:"nodeAffinity,omitempty"`
	PodAffinity *PodAffinitySpec `json:"podAffinity,omitempty"`
	PodAntiAffinity *PodAntiAffinitySpec `json:"podAntiAffinity,omitempty"`
}

// NodeAffinitySpec defines the node affinity rules for components
type NodeAffinitySpec struct {
	PreferredDuringSchedulingIgnoredDuringExecution []PreferredSchedulingTerm `json:"preferredDuringSchedulingIgnoredDuringExecution,omitempty"`
	RequiredDuringSchedulingIgnoredDuringExecution *RequiredSchedulingTerm `json:"requiredDuringSchedulingIgnoredDuringExecution,omitempty"`
}

// PreferredSchedulingTerm defines the preferred scheduling term for node affinity
type PreferredSchedulingTerm struct {
	Preference NodeSelectorTerm `json:"preference"`
	Weight int32 `json:"weight"`
}

// RequiredSchedulingTerm defines the required scheduling term for node affinity
type RequiredSchedulingTerm struct {
	NodeSelectorTerms []NodeSelectorTerm `json:"nodeSelectorTerms"`
}

// NodeSelectorTerm defines the node selector term for node affinity
type NodeSelectorTerm struct {
	MatchExpressions []NodeSelectorRequirement `json:"matchExpressions,omitempty"`
	MatchFields []NodeSelectorRequirement `json:"matchFields,omitempty"`
}

// NodeSelectorRequirement defines the node selector requirement for node affinity
type NodeSelectorRequirement struct {
	Key string `json:"key"`
	Operator string `json:"operator"`
	Values []string `json:"values,omitempty"`
}

// PodAffinitySpec defines the pod affinity rules for components
type PodAffinitySpec struct {
	RequiredDuringSchedulingIgnoredDuringExecution []PodAffinityTerm `json:"requiredDuringSchedulingIgnoredDuringExecution,omitempty"`
	PreferredDuringSchedulingIgnoredDuringExecution []WeightedPodAffinityTerm `json:"preferredDuringSchedulingIgnoredDuringExecution,omitempty"`
}

// PodAntiAffinitySpec defines the pod anti-affinity rules for components
type PodAntiAffinitySpec struct {
	RequiredDuringSchedulingIgnoredDuringExecution []PodAffinityTerm `json:"requiredDuringSchedulingIgnoredDuringExecution,omitempty"`
	PreferredDuringSchedulingIgnoredDuringExecution []WeightedPodAffinityTerm `json:"preferredDuringSchedulingIgnoredDuringExecution,omitempty"`
}

// PodAffinityTerm defines a pod affinity term
type PodAffinityTerm struct {
	LabelSelector *metav1.LabelSelector `json:"labelSelector,omitempty"`
	Namespaces []string `json:"namespaces,omitempty"`
	TopologyKey string `json:"topologyKey"`
}

// WeightedPodAffinityTerm defines a weighted pod affinity term
type WeightedPodAffinityTerm struct {
	PodAffinityTerm PodAffinityTerm `json:"podAffinityTerm"`
	Weight int32 `json:"weight"`
}

// HadoopConfigSpec defines the Hadoop configuration
type HadoopConfigSpec struct {
	CoreSite     map[string]string `json:"coreSite,omitempty"`
	HdfsSite     map[string]string `json:"hdfsSite,omitempty"`
	YarnSite     map[string]string `json:"yarnSite,omitempty"`
	MapredSite   map[string]string `json:"mapredSite,omitempty"`
}

// HadoopClusterStatus defines the observed state of HadoopCluster
type HadoopClusterStatus struct {
	// Phase indicates the current phase of the cluster
	Phase string `json:"phase,omitempty"`
	// Components status
	Components ComponentsStatus `json:"components,omitempty"`
	// Conditions
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// Version
	Version string `json:"version,omitempty"`
}

// ComponentsStatus defines the status of each component
type ComponentsStatus struct {
	NameNode       ComponentStatus `json:"nameNode,omitempty"`
	DataNode       ComponentStatus `json:"dataNode,omitempty"`
	ResourceManager ComponentStatus `json:"resourceManager,omitempty"`
	NodeManager    ComponentStatus `json:"nodeManager,omitempty"`
}

// ComponentStatus defines the status of a single component
type ComponentStatus struct {
	ReadyReplicas int32  `json:"readyReplicas,omitempty"`
	Status        string `json:"status,omitempty"`
	Message       string `json:"message,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HadoopCluster is the Schema for the hadoopclusters API
type HadoopCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HadoopClusterSpec   `json:"spec,omitempty"`
	Status HadoopClusterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HadoopClusterList contains a list of HadoopCluster
type HadoopClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HadoopCluster `json:"items"`
}
