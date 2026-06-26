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

package reconciler

import (
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	hadoopv1 "github.com/apache/hadoop-operator/api/v1"
)

const hadoopClusterFinalizer = "hadoop.apache.org/finalizer"

// +kubebuilder:rbac:groups=hadoop.apache.org,resources=hadoopclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hadoop.apache.org,resources=hadoopclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=hadoop.apache.org,resources=hadoopclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *HadoopClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the HadoopCluster instance
	cluster := &hadoopv1.HadoopCluster{}
	if err := r.Get(ctx, req.NamespacedName, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("HadoopCluster resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get HadoopCluster")
		return ctrl.Result{}, err
	}

	// Set initial status if not set
	if cluster.Status.Phase == "" {
		cluster.Status.Phase = hadoopv1.ClusterPhasePending
		if err := r.Status().Update(ctx, cluster); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Add finalizer for cleanup
	if !controllerutil.ContainsFinalizer(cluster, hadoopClusterFinalizer) {
		controllerutil.AddFinalizer(cluster, hadoopClusterFinalizer)
		if err := r.Update(ctx, cluster); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Handle deletion
	if !cluster.DeletionTimestamp.IsZero() {
		return r.ReconcileDelete(ctx, cluster)
	}

	// Set phase to Creating if it's Pending
	if cluster.Status.Phase == hadoopv1.ClusterPhasePending {
		cluster.Status.Phase = hadoopv1.ClusterPhaseCreating
		if err := r.Status().Update(ctx, cluster); err != nil {
			return ctrl.Result{}, err
		}
		r.Recorder.Event(cluster, corev1.EventTypeNormal, "Creating", "Starting cluster creation")
	}

	// Check if HA mode requires ZooKeeper/JournalNode
	isNNHA := cluster.Spec.HDFS.NameNode.HA != nil && cluster.Spec.HDFS.NameNode.HA.Enabled
	isRMHA := cluster.Spec.YARN.ResourceManager.HA != nil && cluster.Spec.YARN.ResourceManager.HA.Enabled

	// Build the ordered list of component reconcilers
	reconcilers := []ComponentReconciler{
		r.ReconcileConfigMap,
	}

	// HA components must be ready before NameNode/ResourceManager
	if isNNHA {
		reconcilers = append(reconcilers,
			r.ReconcileZooKeeper,
			r.ReconcileJournalNode,
		)
	}

	// Core HDFS and YARN components
	reconcilers = append(reconcilers,
		r.ReconcileNameNodeService,
		r.ReconcileNameNode,
		r.ReconcileDataNodeService,
		r.ReconcileDataNode,
		r.ReconcileResourceManagerService,
		r.ReconcileResourceManager,
		r.ReconcileNodeManagerService,
		r.ReconcileNodeManager,
	)

	// PodDisruptionBudgets for HA components
	if isNNHA || isRMHA {
		reconcilers = append(reconcilers, r.ReconcilePDBs)
	}

	for _, rec := range reconcilers {
		result, err := rec(ctx, cluster)
		if err != nil {
			r.Recorder.Event(cluster, corev1.EventTypeWarning, "ReconcileError", err.Error())
			return result, err
		}
		if result.Requeue || result.RequeueAfter > 0 {
			return result, nil
		}
	}

	// Update status
	if err := r.UpdateClusterStatus(ctx, cluster); err != nil {
		return ctrl.Result{}, err
	}

	// Check if all components are ready
	if r.IsClusterReady(cluster) {
		if cluster.Status.Phase != hadoopv1.ClusterPhaseRunning {
			cluster.Status.Phase = hadoopv1.ClusterPhaseRunning
			r.Recorder.Event(cluster, corev1.EventTypeNormal, "Running", "Cluster is now running")
			if err := r.Status().Update(ctx, cluster); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// ReconcileDelete handles cluster deletion with proper cleanup
func (r *HadoopClusterReconciler) ReconcileDelete(ctx context.Context, cluster *hadoopv1.HadoopCluster) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling HadoopCluster deletion")

	cluster.Status.Phase = hadoopv1.ClusterPhaseDeleting
	if err := r.Status().Update(ctx, cluster); err != nil {
		return ctrl.Result{}, err
	}

	r.Recorder.Event(cluster, corev1.EventTypeNormal, "Deleting", "Cluster deletion in progress")

	// PVCs are retained by default - Kubernetes garbage collection handles
	// StatefulSet-owned PVCs based on the retention policy

	controllerutil.RemoveFinalizer(cluster, hadoopClusterFinalizer)
	if err := r.Update(ctx, cluster); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// IsClusterReady checks if all required cluster components are running
func (r *HadoopClusterReconciler) IsClusterReady(cluster *hadoopv1.HadoopCluster) bool {
	return cluster.Status.NameNode.ReadyReplicas > 0 &&
		cluster.Status.DataNode.ReadyReplicas > 0 &&
		cluster.Status.ResourceManager.ReadyReplicas > 0 &&
		cluster.Status.NodeManager.ReadyReplicas > 0
}

// UpdateClusterStatus updates the status of all cluster components
func (r *HadoopClusterReconciler) UpdateClusterStatus(ctx context.Context, cluster *hadoopv1.HadoopCluster) error {
	// Update NameNode status
	nnSts := &appsv1.StatefulSet{}
	nnKey := types.NamespacedName{Name: cluster.Name + "-namenode", Namespace: cluster.Namespace}
	if err := r.Get(ctx, nnKey, nnSts); err == nil {
		cluster.Status.NameNode.Replicas = *nnSts.Spec.Replicas
		cluster.Status.NameNode.ReadyReplicas = nnSts.Status.ReadyReplicas
		if cluster.Spec.HDFS.NameNode.HA != nil && cluster.Spec.HDFS.NameNode.HA.Enabled {
			cluster.Status.NameNode.Active = cluster.Name + "-namenode-0"
			cluster.Status.NameNode.Standby = []string{cluster.Name + "-namenode-1"}
		}
	}

	// Update DataNode status
	dnSts := &appsv1.StatefulSet{}
	dnKey := types.NamespacedName{Name: cluster.Name + "-datanode", Namespace: cluster.Namespace}
	if err := r.Get(ctx, dnKey, dnSts); err == nil {
		cluster.Status.DataNode.Replicas = *dnSts.Spec.Replicas
		cluster.Status.DataNode.ReadyReplicas = dnSts.Status.ReadyReplicas
		cluster.Status.DataNode.LiveNodes = dnSts.Status.ReadyReplicas
		cluster.Status.DataNode.DeadNodes = dnSts.Status.Replicas - dnSts.Status.ReadyReplicas
	}

	// Update ResourceManager status
	rmSts := &appsv1.StatefulSet{}
	rmKey := types.NamespacedName{Name: cluster.Name + "-resourcemanager", Namespace: cluster.Namespace}
	if err := r.Get(ctx, rmKey, rmSts); err == nil {
		cluster.Status.ResourceManager.Replicas = *rmSts.Spec.Replicas
		cluster.Status.ResourceManager.ReadyReplicas = rmSts.Status.ReadyReplicas
		if cluster.Spec.YARN.ResourceManager.HA != nil && cluster.Spec.YARN.ResourceManager.HA.Enabled {
			cluster.Status.ResourceManager.Active = cluster.Name + "-resourcemanager-0"
			cluster.Status.ResourceManager.Standby = []string{cluster.Name + "-resourcemanager-1"}
		}
	}

	// Update NodeManager status
	nmSts := &appsv1.StatefulSet{}
	nmKey := types.NamespacedName{Name: cluster.Name + "-nodemanager", Namespace: cluster.Namespace}
	if err := r.Get(ctx, nmKey, nmSts); err == nil {
		cluster.Status.NodeManager.Replicas = *nmSts.Spec.Replicas
		cluster.Status.NodeManager.ReadyReplicas = nmSts.Status.ReadyReplicas
	}

	// Update conditions manually (ClusterCondition is our custom type, not metav1.Condition)
	readyCondition := hadoopv1.ClusterCondition{
		Type:               hadoopv1.ClusterConditionReady,
		Status:             corev1.ConditionFalse,
		LastTransitionTime: metav1.Now(),
	}
	if r.IsClusterReady(cluster) {
		readyCondition.Status = corev1.ConditionTrue
		readyCondition.Reason = "AllComponentsReady"
		readyCondition.Message = "All cluster components are ready"
	} else {
		readyCondition.Reason = "ComponentsNotReady"
		readyCondition.Message = "Some cluster components are not ready"
	}

	// Update or append the Ready condition
	found := false
	for i, c := range cluster.Status.Conditions {
		if c.Type == hadoopv1.ClusterConditionReady {
			cluster.Status.Conditions[i] = readyCondition
			found = true
			break
		}
	}
	if !found {
		cluster.Status.Conditions = append(cluster.Status.Conditions, readyCondition)
	}

	// Update observed generation
	cluster.Status.ObservedGeneration = cluster.Generation

	return r.Status().Update(ctx, cluster)
}

// ReconcilePDBs creates PodDisruptionBudgets for HA components
func (r *HadoopClusterReconciler) ReconcilePDBs(ctx context.Context, cluster *hadoopv1.HadoopCluster) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// PDB for NameNode (HA)
	if cluster.Spec.HDFS.NameNode.HA != nil && cluster.Spec.HDFS.NameNode.HA.Enabled {
		if err := r.ensurePDB(ctx, cluster, cluster.Name+"-namenode",
			map[string]string{"app": "hadoop-namenode", "hadoop.apache.org/cluster": cluster.Name},
			1); err != nil {
			log.Error(err, "Failed to create NameNode PDB")
			return ctrl.Result{}, err
		}
	}

	// PDB for ResourceManager (HA)
	if cluster.Spec.YARN.ResourceManager.HA != nil && cluster.Spec.YARN.ResourceManager.HA.Enabled {
		if err := r.ensurePDB(ctx, cluster, cluster.Name+"-resourcemanager",
			map[string]string{"app": "hadoop-resourcemanager", "hadoop.apache.org/cluster": cluster.Name},
			1); err != nil {
			log.Error(err, "Failed to create ResourceManager PDB")
			return ctrl.Result{}, err
		}
	}

	// PDB for DataNode (if replicas > 1)
	if cluster.Spec.HDFS.DataNode.Replicas > 1 {
		maxUnavailable := intstr.FromInt(1)
		if cluster.Spec.HDFS.DataNode.Replicas >= 3 {
			maxUnavailable = intstr.FromString("25%")
		}
		if err := r.ensurePDBWithMaxUnavailable(ctx, cluster, cluster.Name+"-datanode",
			map[string]string{"app": "hadoop-datanode", "hadoop.apache.org/cluster": cluster.Name},
			maxUnavailable); err != nil {
			log.Error(err, "Failed to create DataNode PDB")
			return ctrl.Result{}, err
		}
	}

	// PDB for NodeManager (if replicas > 1)
	if cluster.Spec.YARN.NodeManager.Replicas > 1 {
		maxUnavailable := intstr.FromInt(1)
		if cluster.Spec.YARN.NodeManager.Replicas >= 3 {
			maxUnavailable = intstr.FromString("25%")
		}
		if err := r.ensurePDBWithMaxUnavailable(ctx, cluster, cluster.Name+"-nodemanager",
			map[string]string{"app": "hadoop-nodemanager", "hadoop.apache.org/cluster": cluster.Name},
			maxUnavailable); err != nil {
			log.Error(err, "Failed to create NodeManager PDB")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *HadoopClusterReconciler) ensurePDB(ctx context.Context, cluster *hadoopv1.HadoopCluster, name string, matchLabels map[string]string, maxUnavailable int) error {
	return r.ensurePDBWithMaxUnavailable(ctx, cluster, name, matchLabels, intstr.FromInt(maxUnavailable))
}

func (r *HadoopClusterReconciler) ensurePDBWithMaxUnavailable(ctx context.Context, cluster *hadoopv1.HadoopCluster, name string, matchLabels map[string]string, maxUnavailable intstr.IntOrString) error {
	pdb := &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: cluster.Namespace,
			Labels: map[string]string{
				"hadoop.apache.org/cluster":    cluster.Name,
				"hadoop.apache.org/managed-by": "hadoop-operator",
			},
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, pdb, func() error {
		pdb.Spec = policyv1.PodDisruptionBudgetSpec{
			MaxUnavailable: &maxUnavailable,
			Selector: &metav1.LabelSelector{
				MatchLabels: matchLabels,
			},
		}
		return controllerutil.SetControllerReference(cluster, pdb, r.Scheme)
	})
	return err
}

// SetupWithManager sets up the controller with the Manager.
func (r *HadoopClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hadoopv1.HadoopCluster{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&policyv1.PodDisruptionBudget{}).
		Complete(r)
}
