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

package controller

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	hadoopv1 "github.com/apache/hadoop-operator/api/v1"
	"github.com/apache/hadoop-operator/internal/reconciler"
)

// HadoopClusterReconciler reconciles a HadoopCluster object
type HadoopClusterReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=hadoop.apache.org,resources=hadoopclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=hadoop.apache.org,resources=hadoopclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=hadoop.apache.org,resources=hadoopclusters/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

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
		return r.reconcileDelete(ctx, cluster)
	}

	// Set phase to Creating if it's Pending
	if cluster.Status.Phase == hadoopv1.ClusterPhasePending {
		cluster.Status.Phase = hadoopv1.ClusterPhaseCreating
		if err := r.Status().Update(ctx, cluster); err != nil {
			return ctrl.Result{}, err
		}
		r.Recorder.Event(cluster, corev1.EventTypeNormal, "Creating", "Starting cluster creation")
	}

	// Reconcile components in order
	reconcilers := []reconciler.ComponentReconciler{
		r.reconcileConfigMap,
		r.reconcileNameNodeService,
		r.reconcileNameNode,
		r.reconcileDataNodeService,
		r.reconcileDataNode,
		r.reconcileResourceManagerService,
		r.reconcileResourceManager,
		r.reconcileNodeManagerService,
		r.reconcileNodeManager,
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
	if err := r.updateStatus(ctx, cluster); err != nil {
		return ctrl.Result{}, err
	}

	// Check if all components are ready
	if r.isClusterReady(cluster) {
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

const hadoopClusterFinalizer = "hadoop.apache.org/finalizer"

func (r *HadoopClusterReconciler) reconcileDelete(ctx context.Context, cluster *hadoopv1.HadoopCluster) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling HadoopCluster deletion")

	cluster.Status.Phase = hadoopv1.ClusterPhaseDeleting
	if err := r.Status().Update(ctx, cluster); err != nil {
		return ctrl.Result{}, err
	}

	// Cleanup logic here if needed
	// PVCs are retained by default unless explicitly deleted

	controllerutil.RemoveFinalizer(cluster, hadoopClusterFinalizer)
	if err := r.Update(ctx, cluster); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *HadoopClusterReconciler) isClusterReady(cluster *hadoopv1.HadoopCluster) bool {
	return cluster.Status.NameNode.ReadyReplicas > 0 &&
		cluster.Status.DataNode.ReadyReplicas > 0 &&
		cluster.Status.ResourceManager.ReadyReplicas > 0 &&
		cluster.Status.NodeManager.ReadyReplicas > 0
}

func (r *HadoopClusterReconciler) updateStatus(ctx context.Context, cluster *hadoopv1.HadoopCluster) error {
	// Update NameNode status
	nnSts := &appsv1.StatefulSet{}
	nnKey := types.NamespacedName{Name: cluster.Name + "-namenode", Namespace: cluster.Namespace}
	if err := r.Get(ctx, nnKey, nnSts); err == nil {
		cluster.Status.NameNode.Replicas = *nnSts.Spec.Replicas
		cluster.Status.NameNode.ReadyReplicas = nnSts.Status.ReadyReplicas
	}

	// Update DataNode status
	dnSts := &appsv1.StatefulSet{}
	dnKey := types.NamespacedName{Name: cluster.Name + "-datanode", Namespace: cluster.Namespace}
	if err := r.Get(ctx, dnKey, dnSts); err == nil {
		cluster.Status.DataNode.Replicas = *dnSts.Spec.Replicas
		cluster.Status.DataNode.ReadyReplicas = dnSts.Status.ReadyReplicas
	}

	// Update ResourceManager status
	rmSts := &appsv1.StatefulSet{}
	rmKey := types.NamespacedName{Name: cluster.Name + "-resourcemanager", Namespace: cluster.Namespace}
	if err := r.Get(ctx, rmKey, rmSts); err == nil {
		cluster.Status.ResourceManager.Replicas = *rmSts.Spec.Replicas
		cluster.Status.ResourceManager.ReadyReplicas = rmSts.Status.ReadyReplicas
	}

	// Update NodeManager status
	nmSts := &appsv1.StatefulSet{}
	nmKey := types.NamespacedName{Name: cluster.Name + "-nodemanager", Namespace: cluster.Namespace}
	if err := r.Get(ctx, nmKey, nmSts); err == nil {
		cluster.Status.NodeManager.Replicas = *nmSts.Spec.Replicas
		cluster.Status.NodeManager.ReadyReplicas = nmSts.Status.ReadyReplicas
	}

	// Update conditions
	readyCondition := hadoopv1.ClusterCondition{
		Type:               hadoopv1.ClusterConditionReady,
		Status:             corev1.ConditionFalse,
		LastTransitionTime: metav1.Now(),
	}
	if r.isClusterReady(cluster) {
		readyCondition.Status = corev1.ConditionTrue
		readyCondition.Reason = "AllComponentsReady"
		readyCondition.Message = "All cluster components are ready"
	} else {
		readyCondition.Reason = "ComponentsNotReady"
		readyCondition.Message = "Some cluster components are not ready"
	}

	meta.SetStatusCondition(&cluster.Status.Conditions, readyCondition)

	return r.Status().Update(ctx, cluster)
}

// SetupWithManager sets up the controller with the Manager.
func (r *HadoopClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hadoopv1.HadoopCluster{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Complete(r)
}
