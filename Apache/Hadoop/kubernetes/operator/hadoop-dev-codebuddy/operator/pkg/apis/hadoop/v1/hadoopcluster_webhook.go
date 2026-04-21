package v1

import (
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var hadooplog = logf.Log.WithName("hadoop-resource")

// SetupWebhookWithManager sets up the webhook with the manager.
func (r *HadoopCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-hadoop-kubedoop-dev-v1-hadoopcluster,mutating=true,failurePolicy=fail,sideEffects=None,groups=hadoop.kubedoop.dev,resources=hadoopclusters,verbs=create;update,versions=v1,name=mhadoopcluster.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &HadoopCluster{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *HadoopCluster) Default() {
	hadooplog.Info("default", "name", r.Name)

	// Set default values for NameNode
	if r.Spec.NameNodeSpec != nil {
		if r.Spec.NameNodeSpec.Replicas == 0 {
			r.Spec.NameNodeSpec.Replicas = 1
		}
		if r.Spec.NameNodeSpec.Image == "" {
			r.Spec.NameNodeSpec.Image = r.Spec.Image
		}
		if r.Spec.NameNodeSpec.ImagePullPolicy == "" {
			r.Spec.NameNodeSpec.ImagePullPolicy = r.Spec.ImagePullPolicy
		}
		if r.Spec.NameNodeSpec.ServiceType == "" {
			r.Spec.NameNodeSpec.ServiceType = "ClusterIP"
		}
		if r.Spec.NameNodeSpec.NameDir == "" {
			r.Spec.NameNodeSpec.NameDir = "/tmp/hadoop/dfs/name"
		}
		// Set default ports
		if r.Spec.NameNodeSpec.Ports == nil {
			r.Spec.NameNodeSpec.Ports = &NameNodePorts{
				RPCPort:   9000,
				HTTPPort:  9870,
				HTTPSPort: 9871,
			}
		}
	}

	// Set default values for DataNode
	if r.Spec.DataNodeSpec != nil {
		if r.Spec.DataNodeSpec.Replicas == 0 {
			r.Spec.DataNodeSpec.Replicas = 3
		}
		if r.Spec.DataNodeSpec.Image == "" {
			r.Spec.DataNodeSpec.Image = r.Spec.Image
		}
		if r.Spec.DataNodeSpec.ImagePullPolicy == "" {
			r.Spec.DataNodeSpec.ImagePullPolicy = r.Spec.ImagePullPolicy
		}
		if r.Spec.DataNodeSpec.VolumesPerNode == 0 {
			r.Spec.DataNodeSpec.VolumesPerNode = 1
		}
		// Set default ports
		if r.Spec.DataNodeSpec.Ports == nil {
			r.Spec.DataNodeSpec.Ports = &DataNodePorts{
				DataPort:  9866,
				IPCPort:   9867,
				HTTPPort:  9864,
				HTTPSPort: 9865,
			}
		}
	}

	// Set default values for JournalNode
	if r.Spec.JournalNodeSpec != nil {
		if r.Spec.JournalNodeSpec.Replicas == 0 {
			r.Spec.JournalNodeSpec.Replicas = 3
		}
		if r.Spec.JournalNodeSpec.Image == "" {
			r.Spec.JournalNodeSpec.Image = r.Spec.Image
		}
		if r.Spec.JournalNodeSpec.ImagePullPolicy == "" {
			r.Spec.JournalNodeSpec.ImagePullPolicy = r.Spec.ImagePullPolicy
		}
		// Set default ports
		if r.Spec.JournalNodeSpec.Ports == nil {
			r.Spec.JournalNodeSpec.Ports = &JournalNodePorts{
				RPCPort:  8485,
				HTTPPort: 8480,
			}
		}
	}

	// Set default values for ResourceManager
	if r.Spec.ResourceManagerSpec != nil {
		if r.Spec.ResourceManagerSpec.Replicas == 0 {
			r.Spec.ResourceManagerSpec.Replicas = 1
		}
		if r.Spec.ResourceManagerSpec.Image == "" {
			r.Spec.ResourceManagerSpec.Image = r.Spec.Image
		}
		if r.Spec.ResourceManagerSpec.ImagePullPolicy == "" {
			r.Spec.ResourceManagerSpec.ImagePullPolicy = r.Spec.ImagePullPolicy
		}
		if r.Spec.ResourceManagerSpec.ServiceType == "" {
			r.Spec.ResourceManagerSpec.ServiceType = "ClusterIP"
		}
		// Set default ports
		if r.Spec.ResourceManagerSpec.Ports == nil {
			r.Spec.ResourceManagerSpec.Ports = &ResourceManagerPorts{
				RPCPort:       8032,
				HTTPPort:      8088,
				HTTPSPort:     8090,
				SchedulerPort: 8030,
			}
		}
	}

	// Set default values for NodeManager
	if r.Spec.NodeManagerSpec != nil {
		if r.Spec.NodeManagerSpec.Replicas == 0 {
			r.Spec.NodeManagerSpec.Replicas = 3
		}
		if r.Spec.NodeManagerSpec.Image == "" {
			r.Spec.NodeManagerSpec.Image = r.Spec.Image
		}
		if r.Spec.NodeManagerSpec.ImagePullPolicy == "" {
			r.Spec.NodeManagerSpec.ImagePullPolicy = r.Spec.ImagePullPolicy
		}
		// Set default ports
		if r.Spec.NodeManagerSpec.Ports == nil {
			r.Spec.NodeManagerSpec.Ports = &NodeManagerPorts{
				HTTPPort:      8042,
				HTTPSPort:     8044,
				LocalizerPort: 8040,
				ShufflePort:   13562,
			}
		}
	}

	// Set default values for HBase
	if r.Spec.HBaseSpec != nil && r.Spec.HBaseSpec.Enabled {
		if r.Spec.HBaseSpec.MasterSpec != nil {
			if r.Spec.HBaseSpec.MasterSpec.Replicas == 0 {
				r.Spec.HBaseSpec.MasterSpec.Replicas = 1
			}
			if r.Spec.HBaseSpec.MasterSpec.Image == "" {
				r.Spec.HBaseSpec.MasterSpec.Image = r.Spec.Image
			}
			// Set default ports
			if r.Spec.HBaseSpec.MasterSpec.Ports == nil {
				r.Spec.HBaseSpec.MasterSpec.Ports = &HBaseMasterPorts{
					WebUIPort: 16010,
					Port:      16000,
				}
			}
		}
		if r.Spec.HBaseSpec.RegionServerSpec != nil {
			if r.Spec.HBaseSpec.RegionServerSpec.Replicas == 0 {
				r.Spec.HBaseSpec.RegionServerSpec.Replicas = 3
			}
			if r.Spec.HBaseSpec.RegionServerSpec.Image == "" {
				r.Spec.HBaseSpec.RegionServerSpec.Image = r.Spec.Image
			}
			if r.Spec.HBaseSpec.RegionServerSpec.VolumesPerNode == 0 {
				r.Spec.HBaseSpec.RegionServerSpec.VolumesPerNode = 1
			}
			// Set default ports
			if r.Spec.HBaseSpec.RegionServerSpec.Ports == nil {
				r.Spec.HBaseSpec.RegionServerSpec.Ports = &HBaseRegionServerPorts{
					WebUIPort: 16030,
					Port:      16020,
				}
			}
		}
	}

	// Set global defaults
	if r.Spec.Image == "" {
		r.Spec.Image = "apache/hadoop:3.4.1"
	}
	if r.Spec.ImagePullPolicy == "" {
		r.Spec.ImagePullPolicy = "IfNotPresent"
	}
	if r.Spec.ServiceAccountName == "" {
		r.Spec.ServiceAccountName = "hadoop-operator"
	}

	// Set HA defaults if enabled
	if r.Spec.HA != nil {
		if r.Spec.HA.NameNodeHA != nil && r.Spec.HA.NameNodeHA.Enabled {
			if r.Spec.HA.NameNodeHA.Replicas == 0 {
				r.Spec.HA.NameNodeHA.Replicas = 2
			}
			// Enable JournalNode if not set
			if r.Spec.JournalNodeSpec == nil {
				r.Spec.JournalNodeSpec = &JournalNodeSpec{Replicas: 3}
			}
		}
		if r.Spec.HA.ResourceManagerHA != nil && r.Spec.HA.ResourceManagerHA.Enabled {
			if r.Spec.HA.ResourceManagerHA.Replicas == 0 {
				r.Spec.HA.ResourceManagerHA.Replicas = 2
			}
		}
	}

	// Set cluster config defaults
	if r.Spec.ClusterConfig != nil {
		if r.Spec.ClusterConfig.ReplicationFactor == 0 {
			r.Spec.ClusterConfig.ReplicationFactor = 3
		}
		if r.Spec.ClusterConfig.BlockSize == 0 {
			r.Spec.ClusterConfig.BlockSize = 134217728
		}
	}
}

//+kubebuilder:webhook:path=/validate-hadoop-kubedoop-dev-v1-hadoopcluster,mutating=false,failurePolicy=fail,sideEffects=None,groups=hadoop.kubedoop.dev,resources=hadoopclusters,verbs=create;update;delete,versions=v1,name=vhadoopcluster.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &HadoopCluster{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *HadoopCluster) ValidateCreate() (admission.Warnings, error) {
	hadooplog.Info("validate create", "name", r.Name)
	return nil, r.validateHadoopCluster()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *HadoopCluster) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	hadooplog.Info("validate update", "name", r.Name)

	// Check if the change is valid
	oldCluster, ok := old.(*HadoopCluster)
	if !ok {
		return nil, fmt.Errorf("expected old object to be a HadoopCluster")
	}

	// Validate immutability
	if !reflect.DeepEqual(r.Spec.Image, oldCluster.Spec.Image) {
		// Allow image changes during creation, but log warning
		hadooplog.Info("image change detected", "old", oldCluster.Spec.Image, "new", r.Spec.Image)
	}

	return nil, r.validateHadoopCluster()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *HadoopCluster) ValidateDelete() (admission.Warnings, error) {
	hadooplog.Info("validate delete", "name", r.Name)
	return nil, nil
}

// validateHadoopCluster performs validation of the HadoopCluster spec
func (r *HadoopCluster) validateHadoopCluster() error {
	var allErrs []error

	// Validate NameNode replicas for HA mode
	if r.Spec.HA != nil && r.Spec.HA.NameNodeHA != nil && r.Spec.HA.NameNodeHA.Enabled {
		if r.Spec.NameNodeSpec != nil && r.Spec.NameNodeSpec.Replicas != 2 {
			allErrs = append(allErrs, fmt.Errorf("NameNode replicas must be 2 for HA mode, got %d", r.Spec.NameNodeSpec.Replicas))
		}
		// Check JournalNode is configured
		if r.Spec.JournalNodeSpec == nil || r.Spec.JournalNodeSpec.Replicas < 3 {
			allErrs = append(allErrs, fmt.Errorf("JournalNode replicas must be at least 3 for HDFS HA mode"))
		}
	}

	// Validate ResourceManager replicas for HA mode
	if r.Spec.HA != nil && r.Spec.HA.ResourceManagerHA != nil && r.Spec.HA.ResourceManagerHA.Enabled {
		if r.Spec.ResourceManagerSpec != nil && r.Spec.ResourceManagerSpec.Replicas != 2 {
			allErrs = append(allErrs, fmt.Errorf("ResourceManager replicas must be 2 for HA mode, got %d", r.Spec.ResourceManagerSpec.Replicas))
		}
	}

	// Validate DataNode replicas
	if r.Spec.DataNodeSpec != nil && r.Spec.DataNodeSpec.Replicas < 1 {
		allErrs = append(allErrs, fmt.Errorf("DataNode replicas must be at least 1"))
	}

	// Validate HBase configuration
	if r.Spec.HBaseSpec != nil && r.Spec.HBaseSpec.Enabled {
		if r.Spec.HBaseSpec.MasterSpec == nil {
			allErrs = append(allErrs, fmt.Errorf("HBase MasterSpec is required when HBase is enabled"))
		}
		if r.Spec.HBaseSpec.RegionServerSpec == nil {
			allErrs = append(allErrs, fmt.Errorf("HBase RegionServerSpec is required when HBase is enabled"))
		}
	}

	// Validate replication factor
	if r.Spec.ClusterConfig != nil {
		if r.Spec.ClusterConfig.ReplicationFactor < 1 {
			allErrs = append(allErrs, fmt.Errorf("replication factor must be at least 1"))
		}
		if r.Spec.DataNodeSpec != nil && r.Spec.ClusterConfig.ReplicationFactor > r.Spec.DataNodeSpec.Replicas {
			allErrs = append(allErrs, fmt.Errorf("replication factor (%d) cannot be greater than DataNode replicas (%d)",
				r.Spec.ClusterConfig.ReplicationFactor, r.Spec.DataNodeSpec.Replicas))
		}
	}

	if len(allErrs) > 0 {
		return fmt.Errorf("validation errors: %v", allErrs)
	}

	return nil
}
