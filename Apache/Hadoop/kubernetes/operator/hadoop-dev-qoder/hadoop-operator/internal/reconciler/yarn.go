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

// reconcileResourceManagerService creates or updates the ResourceManager service
func (r *HadoopClusterReconciler) reconcileResourceManagerService(ctx context.Context, cluster *hadoopv1.HadoopCluster) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	// Headless service for StatefulSet
	headlessSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-resourcemanager",
			Namespace: cluster.Namespace,
			Labels:    r.labelsForResourceManager(cluster),
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, headlessSvc, func() error {
		headlessSvc.Spec = corev1.ServiceSpec{
			ClusterIP: "None",
			Selector:  r.labelsForResourceManager(cluster),
			Ports: []corev1.ServicePort{
				{
					Name:       "rpc",
					Port:       8032,
					TargetPort: intstr.FromInt(8032),
				},
				{
					Name:       "web",
					Port:       8088,
					TargetPort: intstr.FromInt(8088),
				},
			},
		}
		return controllerutil.SetControllerReference(cluster, headlessSvc, r.Scheme)
	})
	if err != nil {
		log.Error(err, "Failed to create ResourceManager headless service")
		return ctrl.Result{}, err
	}

	// External service (NodePort/LoadBalancer)
	svcType := cluster.Spec.YARN.ResourceManager.Service.Type
	if svcType == "" {
		svcType = corev1.ServiceTypeNodePort
	}

	externalSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-resourcemanager-external",
			Namespace: cluster.Namespace,
			Labels:    r.labelsForResourceManager(cluster),
			Annotations: cluster.Spec.YARN.ResourceManager.Service.Annotations,
		},
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, externalSvc, func() error {
		externalSvc.Spec = corev1.ServiceSpec{
			Type:     svcType,
			Selector: r.labelsForResourceManager(cluster),
			Ports: []corev1.ServicePort{
				{
					Name:       "rpc",
					Port:       8032,
					TargetPort: intstr.FromInt(8032),
					NodePort:   cluster.Spec.YARN.ResourceManager.Service.NodePorts["rpc"],
				},
				{
					Name:       "web",
					Port:       8088,
					TargetPort: intstr.FromInt(8088),
					NodePort:   cluster.Spec.YARN.ResourceManager.Service.NodePorts["web"],
				},
			},
		}
		return controllerutil.SetControllerReference(cluster, externalSvc, r.Scheme)
	})
	if err != nil {
		log.Error(err, "Failed to create ResourceManager external service")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// reconcileResourceManager creates or updates the ResourceManager StatefulSet
func (r *HadoopClusterReconciler) reconcileResourceManager(ctx context.Context, cluster *hadoopv1.HadoopCluster) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	replicas := cluster.Spec.YARN.ResourceManager.Replicas
	if replicas == 0 {
		replicas = 1
	}

	isHA := cluster.Spec.YARN.ResourceManager.HA != nil && cluster.Spec.YARN.ResourceManager.HA.Enabled
	if isHA && replicas < 2 {
		replicas = 2
	}

	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-resourcemanager",
			Namespace: cluster.Namespace,
			Labels:    r.labelsForResourceManager(cluster),
		},
	}

	image := r.getImage(cluster)
	resources := cluster.Spec.YARN.ResourceManager.Resources
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

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, sts, func() error {
		sts.Spec = appsv1.StatefulSetSpec{
			ServiceName: cluster.Name + "-resourcemanager",
			Replicas:    &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: r.labelsForResourceManager(cluster),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: r.labelsForResourceManager(cluster),
				},
				Spec: corev1.PodSpec{
					Affinity:    cluster.Spec.YARN.ResourceManager.Affinity,
					Tolerations: cluster.Spec.YARN.ResourceManager.Tolerations,
					Containers: []corev1.Container{
						{
							Name:            "resourcemanager",
							Image:           image,
							ImagePullPolicy: cluster.Spec.Image.PullPolicy,
							Command:         []string{"yarn", "resourcemanager"},
							Env: []corev1.EnvVar{
								{
									Name:  "HADOOP_CONF_DIR",
									Value: "/opt/hadoop/etc/hadoop",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "rpc",
									ContainerPort: 8032,
								},
								{
									Name:          "web",
									ContainerPort: 8088,
								},
							},
							Resources: resources,
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/ws/v1/cluster/info",
										Port: intstr.FromString("web"),
									},
								},
								InitialDelaySeconds: 60,
								PeriodSeconds:       30,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/ws/v1/cluster/info",
										Port: intstr.FromString("web"),
									},
								},
								InitialDelaySeconds: 30,
								PeriodSeconds:       10,
							},
							VolumeMounts: []corev1.VolumeMount{
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
		}
		return controllerutil.SetControllerReference(cluster, sts, r.Scheme)
	})
	if err != nil {
		log.Error(err, "Failed to create ResourceManager StatefulSet")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// reconcileNodeManagerService creates or updates the NodeManager service
func (r *HadoopClusterReconciler) reconcileNodeManagerService(ctx context.Context, cluster *hadoopv1.HadoopCluster) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	// Headless service for StatefulSet
	headlessSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-nodemanager",
			Namespace: cluster.Namespace,
			Labels:    r.labelsForNodeManager(cluster),
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, headlessSvc, func() error {
		headlessSvc.Spec = corev1.ServiceSpec{
			ClusterIP: "None",
			Selector:  r.labelsForNodeManager(cluster),
			Ports: []corev1.ServicePort{
				{
					Name:       "web",
					Port:       8042,
					TargetPort: intstr.FromInt(8042),
				},
			},
		}
		return controllerutil.SetControllerReference(cluster, headlessSvc, r.Scheme)
	})
	if err != nil {
		log.Error(err, "Failed to create NodeManager headless service")
		return ctrl.Result{}, err
	}

	// External service (NodePort/LoadBalancer)
	svcType := cluster.Spec.YARN.NodeManager.Service.Type
	if svcType == "" {
		svcType = corev1.ServiceTypeNodePort
	}

	externalSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-nodemanager-external",
			Namespace: cluster.Namespace,
			Labels:    r.labelsForNodeManager(cluster),
			Annotations: cluster.Spec.YARN.NodeManager.Service.Annotations,
		},
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, externalSvc, func() error {
		externalSvc.Spec = corev1.ServiceSpec{
			Type:     svcType,
			Selector: r.labelsForNodeManager(cluster),
			Ports: []corev1.ServicePort{
				{
					Name:       "web",
					Port:       8042,
					TargetPort: intstr.FromInt(8042),
					NodePort:   cluster.Spec.YARN.NodeManager.Service.NodePorts["web"],
				},
			},
		}
		return controllerutil.SetControllerReference(cluster, externalSvc, r.Scheme)
	})
	if err != nil {
		log.Error(err, "Failed to create NodeManager external service")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// reconcileNodeManager creates or updates the NodeManager StatefulSet
func (r *HadoopClusterReconciler) reconcileNodeManager(ctx context.Context, cluster *hadoopv1.HadoopCluster) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	replicas := cluster.Spec.YARN.NodeManager.Replicas
	if replicas == 0 {
		replicas = 2
	}

	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-nodemanager",
			Namespace: cluster.Namespace,
			Labels:    r.labelsForNodeManager(cluster),
		},
	}

	image := r.getImage(cluster)
	resources := cluster.Spec.YARN.NodeManager.Resources
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

	// Get ResourceManager service name
	rmService := cluster.Name + "-resourcemanager-0." + cluster.Name + "-resourcemanager." + cluster.Namespace + ".svc.cluster.local"

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, sts, func() error {
		sts.Spec = appsv1.StatefulSetSpec{
			ServiceName: cluster.Name + "-nodemanager",
			Replicas:    &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: r.labelsForNodeManager(cluster),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: r.labelsForNodeManager(cluster),
				},
				Spec: corev1.PodSpec{
					Affinity:    cluster.Spec.YARN.NodeManager.Affinity,
					Tolerations: cluster.Spec.YARN.NodeManager.Tolerations,
					InitContainers: []corev1.Container{
						{
							Name:            "wait-for-resourcemanager",
							Image:           image,
							ImagePullPolicy: cluster.Spec.Image.PullPolicy,
							Command:         []string{"/bin/sh", "-c"},
							Args: []string{fmt.Sprintf(`
echo "Waiting for ResourceManager..."
while ! nc -z %s 8032; do
  echo "Waiting for ResourceManager..."
  sleep 5
done
echo "ResourceManager is ready"
`, rmService)},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "nodemanager",
							Image:           image,
							ImagePullPolicy: cluster.Spec.Image.PullPolicy,
							Command:         []string{"yarn", "nodemanager"},
							Env: []corev1.EnvVar{
								{
									Name:  "HADOOP_CONF_DIR",
									Value: "/opt/hadoop/etc/hadoop",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "web",
									ContainerPort: 8042,
								},
							},
							Resources: resources,
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/node",
										Port: intstr.FromString("web"),
									},
								},
								InitialDelaySeconds: 60,
								PeriodSeconds:       30,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/node",
										Port: intstr.FromString("web"),
									},
								},
								InitialDelaySeconds: 30,
								PeriodSeconds:       10,
							},
							VolumeMounts: []corev1.VolumeMount{
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
		}
		return controllerutil.SetControllerReference(cluster, sts, r.Scheme)
	})
	if err != nil {
		log.Error(err, "Failed to create NodeManager StatefulSet")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *HadoopClusterReconciler) labelsForResourceManager(cluster *hadoopv1.HadoopCluster) map[string]string {
	return map[string]string{
		"app":                          "hadoop-resourcemanager",
		"hadoop.apache.org/cluster":    cluster.Name,
		"hadoop.apache.org/component":  "resourcemanager",
		"hadoop.apache.org/managed-by": "hadoop-operator",
	}
}

func (r *HadoopClusterReconciler) labelsForNodeManager(cluster *hadoopv1.HadoopCluster) map[string]string {
	return map[string]string{
		"app":                          "hadoop-nodemanager",
		"hadoop.apache.org/cluster":    cluster.Name,
		"hadoop.apache.org/component":  "nodemanager",
		"hadoop.apache.org/managed-by": "hadoop-operator",
	}
}
