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
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	hadoopv1 "github.com/apache/hadoop-operator/api/v1"
)

// reconcileZooKeeper creates or updates ZooKeeper for HA coordination
func (r *HadoopClusterReconciler) ReconcileZooKeeper(ctx context.Context, cluster *hadoopv1.HadoopCluster) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	// Check if external ZooKeeper is configured
	if cluster.Spec.HDFS.NameNode.HA != nil && cluster.Spec.HDFS.NameNode.HA.ZooKeeper != nil {
		if cluster.Spec.HDFS.NameNode.HA.ZooKeeper.ConnectionString != "" {
			log.Info("Using external ZooKeeper, skipping internal deployment")
			return ctrl.Result{}, nil
		}
	}

	// Create ZooKeeper service
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-zookeeper",
			Namespace: cluster.Namespace,
			Labels:    r.labelsForZooKeeper(cluster),
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, svc, func() error {
		svc.Spec = corev1.ServiceSpec{
			ClusterIP: "None",
			Selector:  r.labelsForZooKeeper(cluster),
			Ports: []corev1.ServicePort{
				{
					Name:       "client",
					Port:       2181,
					TargetPort: intstr.FromInt(2181),
				},
				{
					Name:       "peer",
					Port:       2888,
					TargetPort: intstr.FromInt(2888),
				},
				{
					Name:       "election",
					Port:       3888,
					TargetPort: intstr.FromInt(3888),
				},
			},
		}
		return controllerutil.SetControllerReference(cluster, svc, r.Scheme)
	})
	if err != nil {
		log.Error(err, "Failed to create ZooKeeper service")
		return ctrl.Result{}, err
	}

	// Create ZooKeeper StatefulSet
	replicas := int32(3)
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-zookeeper",
			Namespace: cluster.Namespace,
			Labels:    r.labelsForZooKeeper(cluster),
		},
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, sts, func() error {
		sts.Spec = appsv1.StatefulSetSpec{
			ServiceName: cluster.Name + "-zookeeper",
			Replicas:    &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: r.labelsForZooKeeper(cluster),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: r.labelsForZooKeeper(cluster),
				},
				Spec: corev1.PodSpec{
					ImagePullSecrets:              cluster.Spec.Image.PullSecrets,
					TerminationGracePeriodSeconds: int64Ptr(30),
					Containers: []corev1.Container{
						{
							Name:            "zookeeper",
							Image:           "zookeeper:3.8",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Ports: []corev1.ContainerPort{
								{
									Name:          "client",
									ContainerPort: 2181,
								},
								{
									Name:          "peer",
									ContainerPort: 2888,
								},
								{
									Name:          "election",
									ContainerPort: 3888,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "ZOO_MY_ID",
									Value: "$(POD_NAME)" + "-zookeeper-",
								},
								{
									Name:  "ZOO_SERVERS",
									Value: r.getZooKeeperServers(cluster),
								},
								{
									Name: "POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/data",
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
						Resources: corev1.VolumeResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("10Gi"),
							},
						},
					},
				},
			},
		}
		return controllerutil.SetControllerReference(cluster, sts, r.Scheme)
	})
	if err != nil {
		log.Error(err, "Failed to create ZooKeeper StatefulSet")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// reconcileJournalNode creates or updates JournalNodes for HA
func (r *HadoopClusterReconciler) ReconcileJournalNode(ctx context.Context, cluster *hadoopv1.HadoopCluster) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	// Only create JournalNodes if HA is enabled for NameNode
	if cluster.Spec.HDFS.NameNode.HA == nil || !cluster.Spec.HDFS.NameNode.HA.Enabled {
		return ctrl.Result{}, nil
	}

	replicas := cluster.Spec.HDFS.NameNode.HA.JournalNode.Replicas
	if replicas < 3 {
		replicas = 3 // Minimum for quorum
	}

	// Create JournalNode service
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-journalnode",
			Namespace: cluster.Namespace,
			Labels:    r.labelsForJournalNode(cluster),
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, svc, func() error {
		svc.Spec = corev1.ServiceSpec{
			ClusterIP: "None",
			Selector:  r.labelsForJournalNode(cluster),
			Ports: []corev1.ServicePort{
				{
					Name:       "rpc",
					Port:       8485,
					TargetPort: intstr.FromInt(8485),
				},
				{
					Name:       "web",
					Port:       8480,
					TargetPort: intstr.FromInt(8480),
				},
			},
		}
		return controllerutil.SetControllerReference(cluster, svc, r.Scheme)
	})
	if err != nil {
		log.Error(err, "Failed to create JournalNode service")
		return ctrl.Result{}, err
	}

	// Create JournalNode StatefulSet
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-journalnode",
			Namespace: cluster.Namespace,
			Labels:    r.labelsForJournalNode(cluster),
		},
	}

	image := r.getImage(cluster)
	resources := cluster.Spec.HDFS.NameNode.HA.JournalNode.Resources
	if resources.Requests == nil {
		resources.Requests = corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("250m"),
			corev1.ResourceMemory: resource.MustParse("1Gi"),
		}
	}
	if resources.Limits == nil {
		resources.Limits = corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("500m"),
			corev1.ResourceMemory: resource.MustParse("2Gi"),
		}
	}

	storageSize := cluster.Spec.HDFS.NameNode.HA.JournalNode.Storage.Size
	if storageSize == "" {
		storageSize = "20Gi"
	}

	storageClass := cluster.Spec.HDFS.NameNode.HA.JournalNode.Storage.StorageClassName

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, sts, func() error {
		sts.Spec = appsv1.StatefulSetSpec{
			ServiceName: cluster.Name + "-journalnode",
			Replicas:    &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: r.labelsForJournalNode(cluster),
			},
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      r.labelsForJournalNode(cluster),
					Annotations: r.podTemplateAnnotations(ctx, cluster),
				},
				Spec: corev1.PodSpec{
					ImagePullSecrets:              cluster.Spec.Image.PullSecrets,
					TerminationGracePeriodSeconds: int64Ptr(30),
					SecurityContext: &corev1.PodSecurityContext{
						FSGroup: int64Ptr(1000),
					},
					Containers: []corev1.Container{
						{
							Name:            "journalnode",
							Image:           image,
							ImagePullPolicy: cluster.Spec.Image.PullPolicy,
							Command:         []string{"hdfs", "journalnode"},
							Env: []corev1.EnvVar{
								{
									Name:  "HADOOP_CONF_DIR",
									Value: "/opt/hadoop/etc/hadoop",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "rpc",
									ContainerPort: 8485,
								},
								{
									Name:          "web",
									ContainerPort: 8480,
								},
							},
							Resources: resources,
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/",
										Port: intstr.FromString("web"),
									},
								},
								InitialDelaySeconds: 60,
								PeriodSeconds:       30,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/",
										Port: intstr.FromString("web"),
									},
								},
								InitialDelaySeconds: 30,
								PeriodSeconds:       10,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "journalnode-data",
									MountPath: "/opt/hadoop/data/jn",
								},
								{
									Name:      "hadoop-config",
									MountPath: "/opt/hadoop/etc/hadoop",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "hadoop-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cluster.Name + "-config",
									},
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "journalnode-data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
						Resources: corev1.VolumeResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse(storageSize),
							},
						},
						StorageClassName: storageClass,
					},
				},
			},
		}
		return controllerutil.SetControllerReference(cluster, sts, r.Scheme)
	})
	if err != nil {
		log.Error(err, "Failed to create JournalNode StatefulSet")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *HadoopClusterReconciler) labelsForZooKeeper(cluster *hadoopv1.HadoopCluster) map[string]string {
	return map[string]string{
		"app":                          "zookeeper",
		"hadoop.apache.org/cluster":    cluster.Name,
		"hadoop.apache.org/component":  "zookeeper",
		"hadoop.apache.org/managed-by": "hadoop-operator",
	}
}

func (r *HadoopClusterReconciler) labelsForJournalNode(cluster *hadoopv1.HadoopCluster) map[string]string {
	return map[string]string{
		"app":                          "hadoop-journalnode",
		"hadoop.apache.org/cluster":    cluster.Name,
		"hadoop.apache.org/component":  "journalnode",
		"hadoop.apache.org/managed-by": "hadoop-operator",
	}
}

func (r *HadoopClusterReconciler) getZooKeeperServers(cluster *hadoopv1.HadoopCluster) string {
	servers := ""
	for i := 0; i < 3; i++ {
		if i > 0 {
			servers += " "
		}
		servers += fmt.Sprintf("server.%d=%s-zookeeper-%d.%s-zookeeper.%s.svc.cluster.local:2888:3888",
			i+1, cluster.Name, i, cluster.Name, cluster.Namespace)
	}
	return servers
}
