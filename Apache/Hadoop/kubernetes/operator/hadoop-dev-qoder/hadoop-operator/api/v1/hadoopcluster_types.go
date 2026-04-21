/*
Copyright 2024 Apache Software Foundation.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HadoopClusterSpec defines the desired state of HadoopCluster
type HadoopClusterSpec struct {
	// Image defines the Hadoop Docker image to use
	Image ImageSpec `json:"image,omitempty"`

	// HDFS configuration
	HDFS HDFSSpec `json:"hdfs,omitempty"`

	// YARN configuration
	YARN YARNSpec `json:"yarn,omitempty"`

	// Configuration overrides for Hadoop
	// +optional
	Config HadoopConfig `json:"config,omitempty"`

	// Security configuration
	// +optional
	Security SecuritySpec `json:"security,omitempty"`

	// Metrics and monitoring configuration
	// +optional
	Metrics MetricsSpec `json:"metrics,omitempty"`
}

// ImageSpec defines the container image configuration
type ImageSpec struct {
	// Repository is the Docker image repository
	Repository string `json:"repository,omitempty"`
	// Tag is the Docker image tag
	Tag string `json:"tag,omitempty"`
	// PullPolicy is the image pull policy
	PullPolicy corev1.PullPolicy `json:"pullPolicy,omitempty"`
	// PullSecrets are the secrets to use for pulling images
	// +optional
	PullSecrets []corev1.LocalObjectReference `json:"pullSecrets,omitempty"`
}

// HDFSSpec defines HDFS configuration
type HDFSSpec struct {
	// NameNode configuration
	NameNode NameNodeSpec `json:"nameNode,omitempty"`
	// DataNode configuration
	DataNode DataNodeSpec `json:"dataNode,omitempty"`
}

// NameNodeSpec defines NameNode configuration
type NameNodeSpec struct {
	// Replicas - for HA mode, use 2
	Replicas int32 `json:"replicas,omitempty"`
	// Resources defines CPU/Memory resource requests/limits
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// Storage configuration
	Storage StorageSpec `json:"storage,omitempty"`
	// Service configuration
	// +optional
	Service ServiceSpec `json:"service,omitempty"`
	// Enable high availability mode
	// +optional
	HA *HASpec `json:"ha,omitempty"`
	// Affinity configuration
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// Tolerations
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// HASpec defines high availability configuration
type HASpec struct {
	// Enabled enables HA mode
	Enabled bool `json:"enabled,omitempty"`
	// ZooKeeper ensemble for HA coordination
	// If not provided, internal ZooKeeper will be deployed
	// +optional
	ZooKeeper *ZooKeeperSpec `json:"zookeeper,omitempty"`
	// JournalNode configuration for HA
	JournalNode JournalNodeSpec `json:"journalNode,omitempty"`
}

// ZooKeeperSpec defines external ZooKeeper configuration
type ZooKeeperSpec struct {
	// Connection string for external ZooKeeper
	// e.g., "zk1:2181,zk2:2181,zk3:2181"
	ConnectionString string `json:"connectionString,omitempty"`
}

// JournalNodeSpec defines JournalNode configuration for HA
type JournalNodeSpec struct {
	// Replicas - should be at least 3 for quorum
	Replicas int32 `json:"replicas,omitempty"`
	// Resources
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// Storage
	Storage StorageSpec `json:"storage,omitempty"`
}

// DataNodeSpec defines DataNode configuration
type DataNodeSpec struct {
	// Replicas - number of DataNode instances
	Replicas int32 `json:"replicas,omitempty"`
	// Resources
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// Storage configuration
	Storage StorageSpec `json:"storage,omitempty"`
	// Service configuration
	// +optional
	Service ServiceSpec `json:"service,omitempty"`
	// Affinity configuration
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// Tolerations
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// YARNSpec defines YARN configuration
type YARNSpec struct {
	// ResourceManager configuration
	ResourceManager ResourceManagerSpec `json:"resourceManager,omitempty"`
	// NodeManager configuration
	NodeManager NodeManagerSpec `json:"nodeManager,omitempty"`
}

// ResourceManagerSpec defines ResourceManager configuration
type ResourceManagerSpec struct {
	// Replicas - for HA mode, use 2
	Replicas int32 `json:"replicas,omitempty"`
	// Resources
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// Service configuration
	// +optional
	Service ServiceSpec `json:"service,omitempty"`
	// Enable high availability mode
	// +optional
	HA *HASpec `json:"ha,omitempty"`
	// Affinity configuration
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// Tolerations
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// NodeManagerSpec defines NodeManager configuration
type NodeManagerSpec struct {
	// Replicas - number of NodeManager instances
	Replicas int32 `json:"replicas,omitempty"`
	// Resources
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// Service configuration
	// +optional
	Service ServiceSpec `json:"service,omitempty"`
	// Affinity configuration
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// Tolerations
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// StorageSpec defines storage configuration
type StorageSpec struct {
	// Size of the storage (e.g., "100Gi")
	Size string `json:"size,omitempty"`
	// StorageClassName for the PVC
	// +optional
	StorageClassName *string `json:"storageClassName,omitempty"`
	// AccessMode for the PVC
	// +optional
	AccessMode corev1.PersistentVolumeAccessMode `json:"accessMode,omitempty"`
}

// ServiceSpec defines service configuration
type ServiceSpec struct {
	// Type of the service (ClusterIP, NodePort, LoadBalancer)
	// +optional
	Type corev1.ServiceType `json:"type,omitempty"`
	// External port mappings for NodePort
	// +optional
	NodePorts map[string]int32 `json:"nodePorts,omitempty"`
	// Annotations to add to the service
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

// HadoopConfig defines Hadoop configuration overrides
type HadoopConfig struct {
	// Core-site.xml overrides
	// +optional
	CoreSite map[string]string `json:"coreSite,omitempty"`
	// HDFS-site.xml overrides
	// +optional
	HDFSSite map[string]string `json:"hdfsSite,omitempty"`
	// YARN-site.xml overrides
	// +optional
	YARNSite map[string]string `json:"yarnSite,omitempty"`
	// Mapred-site.xml overrides
	// +optional
	MapredSite map[string]string `json:"mapredSite,omitempty"`
	// Capacity-scheduler.xml overrides
	// +optional
	CapacityScheduler map[string]string `json:"capacityScheduler,omitempty"`
}

// SecuritySpec defines security configuration
type SecuritySpec struct {
	// Enable Kerberos authentication
	// +optional
	Kerberos *KerberosSpec `json:"kerberos,omitempty"`
	// Enable TLS for communication
	// +optional
	TLS *TLSSpec `json:"tls,omitempty"`
	// Enable Ranger integration
	// +optional
	Ranger *RangerSpec `json:"ranger,omitempty"`
}

// KerberosSpec defines Kerberos configuration
type KerberosSpec struct {
	Enabled bool `json:"enabled,omitempty"`
	// Kerberos realm
	Realm string `json:"realm,omitempty"`
	// KDC configuration
	KDC string `json:"kdc,omitempty"`
	// Admin server
	AdminServer string `json:"adminServer,omitempty"`
	// Keytab secret reference
	KeytabSecret string `json:"keytabSecret,omitempty"`
}

// TLSSpec defines TLS configuration
type TLSSpec struct {
	Enabled bool `json:"enabled,omitempty"`
	// Certificate secret reference
	CertificateSecret string `json:"certificateSecret,omitempty"`
}

// RangerSpec defines Apache Ranger integration
type RangerSpec struct {
	Enabled bool `json:"enabled,omitempty"`
	// Ranger admin URL
	AdminURL string `json:"adminURL,omitempty"`
}

// MetricsSpec defines monitoring configuration
type MetricsSpec struct {
	// Enable Prometheus metrics
	Enabled bool `json:"enabled,omitempty"`
	// Prometheus exporter image
	// +optional
	ExporterImage string `json:"exporterImage,omitempty"`
	// ServiceMonitor configuration
	// +optional
	ServiceMonitor *ServiceMonitorSpec `json:"serviceMonitor,omitempty"`
}

// ServiceMonitorSpec defines Prometheus ServiceMonitor configuration
type ServiceMonitorSpec struct {
	// Enable ServiceMonitor creation
	Enabled bool `json:"enabled,omitempty"`
	// Labels to match Prometheus instance
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// Scrape interval
	// +optional
	Interval string `json:"interval,omitempty"`
}

// HadoopClusterStatus defines the observed state of HadoopCluster
type HadoopClusterStatus struct {
	// Phase represents the current phase of the cluster
	Phase ClusterPhase `json:"phase,omitempty"`
	// Conditions represent the latest available observations of the cluster state
	Conditions []ClusterCondition `json:"conditions,omitempty"`
	// NameNode status
	// +optional
	NameNode NameNodeStatus `json:"nameNode,omitempty"`
	// DataNode status
	// +optional
	DataNode DataNodeStatus `json:"dataNode,omitempty"`
	// ResourceManager status
	// +optional
	ResourceManager ResourceManagerStatus `json:"resourceManager,omitempty"`
	// NodeManager status
	// +optional
	NodeManager NodeManagerStatus `json:"nodeManager,omitempty"`
}

// ClusterPhase represents the phase of the cluster
type ClusterPhase string

const (
	ClusterPhasePending     ClusterPhase = "Pending"
	ClusterPhaseCreating    ClusterPhase = "Creating"
	ClusterPhaseRunning     ClusterPhase = "Running"
	ClusterPhaseFailed      ClusterPhase = "Failed"
	ClusterPhaseDeleting    ClusterPhase = "Deleting"
	ClusterPhaseUpgrading   ClusterPhase = "Upgrading"
)

// ClusterCondition describes the state of a cluster at a certain point
type ClusterCondition struct {
	Type               ClusterConditionType   `json:"type"`
	Status             corev1.ConditionStatus `json:"status"`
	LastTransitionTime metav1.Time            `json:"lastTransitionTime,omitempty"`
	Reason             string                 `json:"reason,omitempty"`
	Message            string                 `json:"message,omitempty"`
}

// ClusterConditionType represents a cluster condition type
type ClusterConditionType string

const (
	ClusterConditionReady       ClusterConditionType = "Ready"
	ClusterConditionProgressing ClusterConditionType = "Progressing"
	ClusterConditionDegraded    ClusterConditionType = "Degraded"
)

// NameNodeStatus defines NameNode status
type NameNodeStatus struct {
	// Ready replicas
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`
	// Total replicas
	Replicas int32 `json:"replicas,omitempty"`
	// Active NameNode
	Active string `json:"active,omitempty"`
	// Standby NameNodes
	Standby []string `json:"standby,omitempty"`
}

// DataNodeStatus defines DataNode status
type DataNodeStatus struct {
	// Ready replicas
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`
	// Total replicas
	Replicas int32 `json:"replicas,omitempty"`
	// Live nodes count
	LiveNodes int32 `json:"liveNodes,omitempty"`
	// Dead nodes count
	DeadNodes int32 `json:"deadNodes,omitempty"`
}

// ResourceManagerStatus defines ResourceManager status
type ResourceManagerStatus struct {
	// Ready replicas
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`
	// Total replicas
	Replicas int32 `json:"replicas,omitempty"`
	// Active ResourceManager
	Active string `json:"active,omitempty"`
	// Standby ResourceManagers
	Standby []string `json:"standby,omitempty"`
}

// NodeManagerStatus defines NodeManager status
type NodeManagerStatus struct {
	// Ready replicas
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`
	// Total replicas
	Replicas int32 `json:"replicas,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=hc
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// HadoopCluster is the Schema for the hadoopclusters API
type HadoopCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HadoopClusterSpec   `json:"spec,omitempty"`
	Status HadoopClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HadoopClusterList contains a list of HadoopCluster
type HadoopClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HadoopCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HadoopCluster{}, &HadoopClusterList{})
}
