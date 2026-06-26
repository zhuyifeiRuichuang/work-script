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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	hadoopv1 "github.com/apache/hadoop-operator/api/v1"
)

// reconcileNameNodeService creates or updates the NameNode service
func (r *HadoopClusterReconciler) ReconcileNameNodeService(ctx context.Context, cluster *hadoopv1.HadoopCluster) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	// Headless service for StatefulSet
	headlessSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-namenode",
			Namespace: cluster.Namespace,
			Labels:    r.labelsForNameNode(cluster),
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, headlessSvc, func() error {
		headlessSvc.Spec = corev1.ServiceSpec{
			ClusterIP: "None",
			Selector:  r.labelsForNameNode(cluster),
			Ports: []corev1.ServicePort{
				{
					Name:       "rpc",
					Port:       9000,
					TargetPort: intstr.FromInt(9000),
				},
				{
					Name:       "web",
					Port:       9870,
					TargetPort: intstr.FromInt(9870),
				},
			},
		}
		return controllerutil.SetControllerReference(cluster, headlessSvc, r.Scheme)
	})
	if err != nil {
		log.Error(err, "Failed to create NameNode headless service")
		return ctrl.Result{}, err
	}

	// External service (NodePort/LoadBalancer)
	svcType := cluster.Spec.HDFS.NameNode.Service.Type
	if svcType == "" {
		svcType = corev1.ServiceTypeNodePort
	}

	externalSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-namenode-external",
			Namespace: cluster.Namespace,
			Labels:    r.labelsForNameNode(cluster),
			Annotations: cluster.Spec.HDFS.NameNode.Service.Annotations,
		},
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, externalSvc, func() error {
		externalSvc.Spec = corev1.ServiceSpec{
			Type:     svcType,
			Selector: r.labelsForNameNode(cluster),
			Ports: []corev1.ServicePort{
				{
					Name:       "rpc",
					Port:       9000,
					TargetPort: intstr.FromInt(9000),
					NodePort:   cluster.Spec.HDFS.NameNode.Service.NodePorts["rpc"],
				},
				{
					Name:       "web",
					Port:       9870,
					TargetPort: intstr.FromInt(9870),
					NodePort:   cluster.Spec.HDFS.NameNode.Service.NodePorts["web"],
				},
			},
		}
		return controllerutil.SetControllerReference(cluster, externalSvc, r.Scheme)
	})
	if err != nil {
		log.Error(err, "Failed to create NameNode external service")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// reconcileNameNode creates or updates the NameNode StatefulSet
func (r *HadoopClusterReconciler) ReconcileNameNode(ctx context.Context, cluster *hadoopv1.HadoopCluster) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	replicas := cluster.Spec.HDFS.NameNode.Replicas
	if replicas == 0 {
		replicas = 1
	}

	// Check if HA mode is enabled
	isHA := cluster.Spec.HDFS.NameNode.HA != nil && cluster.Spec.HDFS.NameNode.HA.Enabled
	if isHA && replicas < 2 {
		replicas = 2
	}

	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-namenode",
			Namespace: cluster.Namespace,
			Labels:    r.labelsForNameNode(cluster),
		},
	}

	image := r.getImage(cluster)
	resources := cluster.Spec.HDFS.NameNode.Resources
	if resources.Requests == nil {
		resources.Requests = corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("500m"),
			corev1.ResourceMemory: resource.MustParse("2Gi"),
		}
	}
	if resources.Limits == nil {
		resources.Limits = corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("1000m"),
			corev1.ResourceMemory: resource.MustParse("4Gi"),
		}
	}

	storageSize := cluster.Spec.HDFS.NameNode.Storage.Size
	if storageSize == "" {
		storageSize = "20Gi"
	}

	storageClass := cluster.Spec.HDFS.NameNode.Storage.StorageClassName
	accessMode := cluster.Spec.HDFS.NameNode.Storage.AccessMode
	if accessMode == "" {
		accessMode = corev1.ReadWriteOnce
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, sts, func() error {
		sts.Spec = appsv1.StatefulSetSpec{
			ServiceName: cluster.Name + "-namenode",
			Replicas:    &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: r.labelsForNameNode(cluster),
			},
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      r.labelsForNameNode(cluster),
					Annotations: r.podTemplateAnnotations(ctx, cluster),
				},
				Spec: corev1.PodSpec{
					ImagePullSecrets: cluster.Spec.Image.PullSecrets,
					Affinity:         cluster.Spec.HDFS.NameNode.Affinity,
					Tolerations:      cluster.Spec.HDFS.NameNode.Tolerations,
					SecurityContext: &corev1.PodSecurityContext{
						FSGroup: int64Ptr(1000),
					},
					TerminationGracePeriodSeconds: int64Ptr(60),
					InitContainers: []corev1.Container{
						{
							Name:            "init-namenode",
							Image:           image,
							ImagePullPolicy: cluster.Spec.Image.PullPolicy,
							SecurityContext: &corev1.SecurityContext{
								RunAsUser: int64Ptr(0),
							},
							Command: []string{"/bin/bash", "-c"},
							Args:    []string{r.getNameNodeInitScript(cluster, isHA)},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "namenode-data",
									MountPath: "/opt/hadoop/data/nn",
									SubPath:   "nn",
								},
								{
									Name:      "hadoop-config",
									MountPath: "/opt/hadoop/etc/hadoop",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "namenode",
							Image:           image,
							ImagePullPolicy: cluster.Spec.Image.PullPolicy,
							Command:         []string{"/bin/bash", "-c"},
							Args:            []string{r.getNameNodeStartScript(cluster, isHA)},
							Env: []corev1.EnvVar{
								{
									Name:  "HADOOP_CONF_DIR",
									Value: "/opt/hadoop/etc/hadoop",
								},
								{
									Name:  "HADOOP_LOG_DIR",
									Value: "/opt/hadoop/logs",
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
							Ports: []corev1.ContainerPort{
								{
									Name:          "rpc",
									ContainerPort: 9000,
								},
								{
									Name:          "web",
									ContainerPort: 9870,
								},
							},
							Resources: resources,
							Lifecycle: &corev1.Lifecycle{
								PreStop: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{
										Command: []string{"/bin/bash", "-c", "sleep 10"},
									},
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/jmx",
										Port: intstr.FromString("web"),
									},
								},
								InitialDelaySeconds: 120,
								PeriodSeconds:       30,
								TimeoutSeconds:      10,
								FailureThreshold:    5,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/jmx",
										Port: intstr.FromString("web"),
									},
								},
								InitialDelaySeconds: 40,
								PeriodSeconds:       10,
								TimeoutSeconds:      5,
								FailureThreshold:    3,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "namenode-data",
									MountPath: "/opt/hadoop/data/nn",
									SubPath:   "nn",
								},
								{
									Name:      "hadoop-config",
									MountPath: "/opt/hadoop/etc/hadoop",
								},
								{
									Name:      "logs",
									MountPath: "/opt/hadoop/logs",
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
						{
							Name: "logs",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "namenode-data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{accessMode},
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
		log.Error(err, "Failed to create NameNode StatefulSet")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *HadoopClusterReconciler) labelsForNameNode(cluster *hadoopv1.HadoopCluster) map[string]string {
	return map[string]string{
		"app":                          "hadoop-namenode",
		"hadoop.apache.org/cluster":    cluster.Name,
		"hadoop.apache.org/component":  "namenode",
		"hadoop.apache.org/managed-by": "hadoop-operator",
	}
}

func (r *HadoopClusterReconciler) getNameNodeInitScript(cluster *hadoopv1.HadoopCluster, isHA bool) string {
	if isHA {
		return `
set -e
echo "Initializing NameNode in HA mode..."
mkdir -p /opt/hadoop/data/nn
chown -R hadoop:hadoop /opt/hadoop/data || chown -R 1000:1000 /opt/hadoop/data
chmod -R 775 /opt/hadoop/data

# For HA mode, formatting is handled differently
if [ "${POD_NAME##*-}" = "0" ] && [ ! -f /opt/hadoop/data/nn/current/VERSION ]; then
    echo "Formatting NameNode on first instance..."
    su hadoop -c "hdfs namenode -format -nonInteractive -clusterId $(cat /proc/sys/kernel/random/uuid)" || true
fi
echo "Initialization complete"
`
	}
	return `
set -e
echo "Initializing NameNode..."
mkdir -p /opt/hadoop/data/nn
chown -R hadoop:hadoop /opt/hadoop/data || chown -R 1000:1000 /opt/hadoop/data
chmod -R 775 /opt/hadoop/data

if [ ! -f /opt/hadoop/data/nn/current/VERSION ]; then
    echo "Formatting NameNode..."
    su hadoop -c "hdfs namenode -format -nonInteractive"
fi
echo "Initialization complete"
`
}

func (r *HadoopClusterReconciler) getNameNodeStartScript(cluster *hadoopv1.HadoopCluster, isHA bool) string {
	return "exec hdfs namenode"
}

func int64Ptr(i int64) *int64 {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}
