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

// reconcileDataNodeService creates or updates the DataNode service
func (r *HadoopClusterReconciler) ReconcileDataNodeService(ctx context.Context, cluster *hadoopv1.HadoopCluster) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	// Headless service for StatefulSet
	headlessSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-datanode",
			Namespace: cluster.Namespace,
			Labels:    r.labelsForDataNode(cluster),
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, headlessSvc, func() error {
		headlessSvc.Spec = corev1.ServiceSpec{
			ClusterIP: "None",
			Selector:  r.labelsForDataNode(cluster),
			Ports: []corev1.ServicePort{
				{
					Name:       "data",
					Port:       9866,
					TargetPort: intstr.FromInt(9866),
				},
				{
					Name:       "web",
					Port:       9864,
					TargetPort: intstr.FromInt(9864),
				},
			},
		}
		return controllerutil.SetControllerReference(cluster, headlessSvc, r.Scheme)
	})
	if err != nil {
		log.Error(err, "Failed to create DataNode headless service")
		return ctrl.Result{}, err
	}

	// External service (NodePort/LoadBalancer)
	svcType := cluster.Spec.HDFS.DataNode.Service.Type
	if svcType == "" {
		svcType = corev1.ServiceTypeNodePort
	}

	externalSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-datanode-external",
			Namespace: cluster.Namespace,
			Labels:    r.labelsForDataNode(cluster),
			Annotations: cluster.Spec.HDFS.DataNode.Service.Annotations,
		},
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, externalSvc, func() error {
		externalSvc.Spec = corev1.ServiceSpec{
			Type:     svcType,
			Selector: r.labelsForDataNode(cluster),
			Ports: []corev1.ServicePort{
				{
					Name:       "data",
					Port:       9866,
					TargetPort: intstr.FromInt(9866),
					NodePort:   cluster.Spec.HDFS.DataNode.Service.NodePorts["data"],
				},
				{
					Name:       "web",
					Port:       9864,
					TargetPort: intstr.FromInt(9864),
					NodePort:   cluster.Spec.HDFS.DataNode.Service.NodePorts["web"],
				},
			},
		}
		return controllerutil.SetControllerReference(cluster, externalSvc, r.Scheme)
	})
	if err != nil {
		log.Error(err, "Failed to create DataNode external service")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// reconcileDataNode creates or updates the DataNode StatefulSet
func (r *HadoopClusterReconciler) ReconcileDataNode(ctx context.Context, cluster *hadoopv1.HadoopCluster) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	replicas := cluster.Spec.HDFS.DataNode.Replicas
	if replicas == 0 {
		replicas = 3
	}

	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-datanode",
			Namespace: cluster.Namespace,
			Labels:    r.labelsForDataNode(cluster),
		},
	}

	image := r.getImage(cluster)
	resources := cluster.Spec.HDFS.DataNode.Resources
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

	storageSize := cluster.Spec.HDFS.DataNode.Storage.Size
	if storageSize == "" {
		storageSize = "100Gi"
	}

	storageClass := cluster.Spec.HDFS.DataNode.Storage.StorageClassName
	accessMode := cluster.Spec.HDFS.DataNode.Storage.AccessMode
	if accessMode == "" {
		accessMode = corev1.ReadWriteOnce
	}

	// Get NameNode service name
	nnService := cluster.Name + "-namenode-0." + cluster.Name + "-namenode." + cluster.Namespace + ".svc.cluster.local"

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, sts, func() error {
		sts.Spec = appsv1.StatefulSetSpec{
			ServiceName: cluster.Name + "-datanode",
			Replicas:    &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: r.labelsForDataNode(cluster),
			},
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      r.labelsForDataNode(cluster),
					Annotations: r.podTemplateAnnotations(ctx, cluster),
				},
				Spec: corev1.PodSpec{
					ImagePullSecrets: cluster.Spec.Image.PullSecrets,
					Affinity:         cluster.Spec.HDFS.DataNode.Affinity,
					Tolerations:      cluster.Spec.HDFS.DataNode.Tolerations,
					SecurityContext: &corev1.PodSecurityContext{
						FSGroup: int64Ptr(1000),
					},
					TerminationGracePeriodSeconds: int64Ptr(30),
					InitContainers: []corev1.Container{
						{
							Name:            "wait-for-namenode",
							Image:           image,
							ImagePullPolicy: cluster.Spec.Image.PullPolicy,
							Command:         []string{"/bin/sh", "-c"},
							Args: []string{fmt.Sprintf(`
echo "Waiting for NameNode..."
while ! nc -z %s 9000; do
  echo "Waiting for NameNode..."
  sleep 5
done
echo "NameNode is ready"
`, nnService)},
						},
						{
							Name:            "init-permissions",
							Image:           image,
							ImagePullPolicy: cluster.Spec.Image.PullPolicy,
							SecurityContext: &corev1.SecurityContext{
								RunAsUser: int64Ptr(0),
							},
							Command: []string{"/bin/sh", "-c"},
							Args: []string{`
echo "Preparing data directories..."
mkdir -p /opt/hadoop/data/dn
if id "hadoop" >/dev/null 2>&1; then
    chown -R hadoop:hadoop /opt/hadoop/data
else
    chown -R 1000:1000 /opt/hadoop/data
fi
chmod -R 775 /opt/hadoop/data
echo "Permissions set"
`},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "datanode-data",
									MountPath: "/opt/hadoop/data/dn",
									SubPath:   "dn",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "datanode",
							Image:           image,
							ImagePullPolicy: cluster.Spec.Image.PullPolicy,
							SecurityContext: &corev1.SecurityContext{
								RunAsUser: int64Ptr(0),
							},
							Command: []string{"/bin/bash", "-c"},
							Args:    []string{"export HADOOP_CONF_DIR=/opt/hadoop/etc/hadoop; exec hdfs datanode"},
							Env: []corev1.EnvVar{
								{
									Name:  "HADOOP_CONF_DIR",
									Value: "/opt/hadoop/etc/hadoop",
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
									Name:          "data",
									ContainerPort: 9866,
								},
								{
									Name:          "web",
									ContainerPort: 9864,
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
										Path: "/",
										Port: intstr.FromString("web"),
									},
								},
								InitialDelaySeconds: 60,
								PeriodSeconds:       20,
								TimeoutSeconds:      10,
								FailureThreshold:    5,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/",
										Port: intstr.FromString("web"),
									},
								},
								InitialDelaySeconds: 20,
								PeriodSeconds:       10,
								TimeoutSeconds:      5,
								FailureThreshold:    3,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "datanode-data",
									MountPath: "/opt/hadoop/data/dn",
									SubPath:   "dn",
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
						Name: "datanode-data",
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
		log.Error(err, "Failed to create DataNode StatefulSet")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *HadoopClusterReconciler) labelsForDataNode(cluster *hadoopv1.HadoopCluster) map[string]string {
	return map[string]string{
		"app":                          "hadoop-datanode",
		"hadoop.apache.org/cluster":    cluster.Name,
		"hadoop.apache.org/component":  "datanode",
		"hadoop.apache.org/managed-by": "hadoop-operator",
	}
}
