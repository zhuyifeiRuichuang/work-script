package hadoop

import (
	"context"
	"fmt"
	"os"
	"time"

	"hadoop-operator/pkg/apis/hadoop/v1alpha1"
	"hadoop-operator/pkg/resources"

	"k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Add creates a new HadoopCluster Controller and adds it to the Manager.
// The Manager will set fields on the Controller and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileHadoopCluster{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("hadoopcluster-controller", mgr, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource HadoopCluster
	err = c.Watch(&source.Kind{Type: &v1alpha1.HadoopCluster{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resources
	watchTypes := []runtime.Object{
		&v1.StatefulSet{},
		&v1.Service{},
		&v1.ConfigMap{},
		&v1.PersistentVolumeClaim{},
	}

	for _, t := range watchTypes {
		err = c.Watch(&source.Kind{Type: t}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &v1alpha1.HadoopCluster{},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// ReconcileHadoopCluster reconciles a HadoopCluster object
type ReconcileHadoopCluster struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a HadoopCluster object and makes changes based on the state read
// and what is in the HadoopCluster.Spec
func (r *ReconcileHadoopCluster) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.Background()

	// Fetch the HadoopCluster instance
	hadoopCluster := &v1alpha1.HadoopCluster{}
	err := r.client.Get(ctx, request.NamespacedName, hadoopCluster)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.  Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Initialize status if not set
	if hadoopCluster.Status.Phase == "" {
		hadoopCluster.Status.Phase = "Creating"
		hadoopCluster.Status.Components = v1alpha1.ComponentsStatus{
			NameNode:       v1alpha1.ComponentStatus{Status: "Pending"},
			DataNode:       v1alpha1.ComponentStatus{Status: "Pending"},
			ResourceManager: v1alpha1.ComponentStatus{Status: "Pending"},
			NodeManager:    v1alpha1.ComponentStatus{Status: "Pending"},
		}
		err = r.client.Status().Update(ctx, hadoopCluster)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	// Check if webhook is enabled
	enableWebhook := os.Getenv("ENABLE_WEBHOOK") == "true"
	if enableWebhook {
		// Webhook logic would go here
	}

	// Create or update ConfigMap
	configMap, err := resources.NewConfigMap(hadoopCluster)
	if err != nil {
		return reconcile.Result{}, err
	}

	existingConfigMap := &v1.ConfigMap{}
	err = r.client.Get(ctx, types.NamespacedName{Name: configMap.Name, Namespace: configMap.Namespace}, existingConfigMap)
	if err != nil {
		if errors.IsNotFound(err) {
			// Create ConfigMap
			err = r.client.Create(ctx, configMap)
			if err != nil {
				return reconcile.Result{}, err
			}
		} else {
			return reconcile.Result{}, err
		}
	} else {
		// Update ConfigMap
		existingConfigMap.Data = configMap.Data
		err = r.client.Update(ctx, existingConfigMap)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	// Create or update NameNode
	namenodeService, namenodeStatefulSet, err := resources.NewNameNodeResources(hadoopCluster)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Create or update NameNode Service
	existingNameNodeService := &v1.Service{}
	err = r.client.Get(ctx, types.NamespacedName{Name: namenodeService.Name, Namespace: namenodeService.Namespace}, existingNameNodeService)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.client.Create(ctx, namenodeService)
			if err != nil {
				return reconcile.Result{}, err
			}
		} else {
			return reconcile.Result{}, err
		}
	}

	// Create or update NameNode StatefulSet
	existingNameNodeStatefulSet := &v1.StatefulSet{}
	err = r.client.Get(ctx, types.NamespacedName{Name: namenodeStatefulSet.Name, Namespace: namenodeStatefulSet.Namespace}, existingNameNodeStatefulSet)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.client.Create(ctx, namenodeStatefulSet)
			if err != nil {
				return reconcile.Result{}, err
			}
		} else {
			return reconcile.Result{}, err
		}
	} else {
		// Update StatefulSet
		existingNameNodeStatefulSet.Spec = namenodeStatefulSet.Spec
		err = r.client.Update(ctx, existingNameNodeStatefulSet)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	// Create or update DataNode
	datanodeService, datanodeStatefulSet, err := resources.NewDataNodeResources(hadoopCluster)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Create or update DataNode Service
	existingDataNodeService := &v1.Service{}
	err = r.client.Get(ctx, types.NamespacedName{Name: datanodeService.Name, Namespace: datanodeService.Namespace}, existingDataNodeService)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.client.Create(ctx, datanodeService)
			if err != nil {
				return reconcile.Result{}, err
			}
		} else {
			return reconcile.Result{}, err
		}
	}

	// Create or update DataNode StatefulSet
	existingDataNodeStatefulSet := &v1.StatefulSet{}
	err = r.client.Get(ctx, types.NamespacedName{Name: datanodeStatefulSet.Name, Namespace: datanodeStatefulSet.Namespace}, existingDataNodeStatefulSet)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.client.Create(ctx, datanodeStatefulSet)
			if err != nil {
				return reconcile.Result{}, err
			}
		} else {
			return reconcile.Result{}, err
		}
	} else {
		existingDataNodeStatefulSet.Spec = datanodeStatefulSet.Spec
		err = r.client.Update(ctx, existingDataNodeStatefulSet)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	// Create or update ResourceManager
	resourceManagerService, resourceManagerStatefulSet, err := resources.NewResourceManagerResources(hadoopCluster)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Create or update ResourceManager Service
	existingResourceManagerService := &v1.Service{}
	err = r.client.Get(ctx, types.NamespacedName{Name: resourceManagerService.Name, Namespace: resourceManagerService.Namespace}, existingResourceManagerService)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.client.Create(ctx, resourceManagerService)
			if err != nil {
				return reconcile.Result{}, err
			}
		} else {
			return reconcile.Result{}, err
		}
	}

	// Create or update ResourceManager StatefulSet
	existingResourceManagerStatefulSet := &v1.StatefulSet{}
	err = r.client.Get(ctx, types.NamespacedName{Name: resourceManagerStatefulSet.Name, Namespace: resourceManagerStatefulSet.Namespace}, existingResourceManagerStatefulSet)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.client.Create(ctx, resourceManagerStatefulSet)
			if err != nil {
				return reconcile.Result{}, err
			}
		} else {
			return reconcile.Result{}, err
		}
	} else {
		existingResourceManagerStatefulSet.Spec = resourceManagerStatefulSet.Spec
		err = r.client.Update(ctx, existingResourceManagerStatefulSet)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	// Create or update NodeManager
	nodeManagerService, nodeManagerStatefulSet, err := resources.NewNodeManagerResources(hadoopCluster)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Create or update NodeManager Service
	existingNodeManagerService := &v1.Service{}
	err = r.client.Get(ctx, types.NamespacedName{Name: nodeManagerService.Name, Namespace: nodeManagerService.Namespace}, existingNodeManagerService)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.client.Create(ctx, nodeManagerService)
			if err != nil {
				return reconcile.Result{}, err
			}
		} else {
			return reconcile.Result{}, err
		}
	}

	// Create or update NodeManager StatefulSet
	existingNodeManagerStatefulSet := &v1.StatefulSet{}
	err = r.client.Get(ctx, types.NamespacedName{Name: nodeManagerStatefulSet.Name, Namespace: nodeManagerStatefulSet.Namespace}, existingNodeManagerStatefulSet)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.client.Create(ctx, nodeManagerStatefulSet)
			if err != nil {
				return reconcile.Result{}, err
			}
		} else {
			return reconcile.Result{}, err
		}
	} else {
		existingNodeManagerStatefulSet.Spec = nodeManagerStatefulSet.Spec
		err = r.client.Update(ctx, existingNodeManagerStatefulSet)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	// Update status with actual ready replicas
	hadoopCluster.Status.Phase = "Running"
	
	// Check NameNode status
	existingNameNodeStatefulSet := &v1.StatefulSet{}
	err = r.client.Get(ctx, types.NamespacedName{Name: hadoopCluster.Name + "-namenode", Namespace: hadoopCluster.Namespace}, existingNameNodeStatefulSet)
	if err == nil {
		hadoopCluster.Status.Components.NameNode.Status = "Running"
		hadoopCluster.Status.Components.NameNode.ReadyReplicas = int32(existingNameNodeStatefulSet.Status.ReadyReplicas)
	} else {
		hadoopCluster.Status.Components.NameNode.Status = "Error"
		hadoopCluster.Status.Components.NameNode.Message = err.Error()
	}

	// Check DataNode status
	existingDataNodeStatefulSet := &v1.StatefulSet{}
	err = r.client.Get(ctx, types.NamespacedName{Name: hadoopCluster.Name + "-datanode", Namespace: hadoopCluster.Namespace}, existingDataNodeStatefulSet)
	if err == nil {
		hadoopCluster.Status.Components.DataNode.Status = "Running"
		hadoopCluster.Status.Components.DataNode.ReadyReplicas = int32(existingDataNodeStatefulSet.Status.ReadyReplicas)
	} else {
		hadoopCluster.Status.Components.DataNode.Status = "Error"
		hadoopCluster.Status.Components.DataNode.Message = err.Error()
	}

	// Check ResourceManager status
	existingResourceManagerStatefulSet := &v1.StatefulSet{}
	err = r.client.Get(ctx, types.NamespacedName{Name: hadoopCluster.Name + "-resourcemanager", Namespace: hadoopCluster.Namespace}, existingResourceManagerStatefulSet)
	if err == nil {
		hadoopCluster.Status.Components.ResourceManager.Status = "Running"
		hadoopCluster.Status.Components.ResourceManager.ReadyReplicas = int32(existingResourceManagerStatefulSet.Status.ReadyReplicas)
	} else {
		hadoopCluster.Status.Components.ResourceManager.Status = "Error"
		hadoopCluster.Status.Components.ResourceManager.Message = err.Error()
	}

	// Check NodeManager status
	existingNodeManagerStatefulSet := &v1.StatefulSet{}
	err = r.client.Get(ctx, types.NamespacedName{Name: hadoopCluster.Name + "-nodemanager", Namespace: hadoopCluster.Namespace}, existingNodeManagerStatefulSet)
	if err == nil {
		hadoopCluster.Status.Components.NodeManager.Status = "Running"
		hadoopCluster.Status.Components.NodeManager.ReadyReplicas = int32(existingNodeManagerStatefulSet.Status.ReadyReplicas)
	} else {
		hadoopCluster.Status.Components.NodeManager.Status = "Error"
		hadoopCluster.Status.Components.NodeManager.Message = err.Error()
	}

	// Check if all components are ready
	allReady := true
	if hadoopCluster.Status.Components.NameNode.ReadyReplicas < hadoopCluster.Spec.NameNode.Replicas {
		allReady = false
	}
	if hadoopCluster.Status.Components.DataNode.ReadyReplicas < hadoopCluster.Spec.DataNode.Replicas {
		allReady = false
	}
	if hadoopCluster.Status.Components.ResourceManager.ReadyReplicas < hadoopCluster.Spec.ResourceManager.Replicas {
		allReady = false
	}
	if hadoopCluster.Status.Components.NodeManager.ReadyReplicas < hadoopCluster.Spec.NodeManager.Replicas {
		allReady = false
	}

	if allReady {
		hadoopCluster.Status.Phase = "Ready"
	} else {
		hadoopCluster.Status.Phase = "Running"
	}

	err = r.client.Status().Update(ctx, hadoopCluster)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Requeue after 5 minutes to check status
	return reconcile.Result{RequeueAfter: 5 * time.Minute}, nil
}
