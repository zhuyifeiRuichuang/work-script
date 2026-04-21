package resources

import (
	"hadoop-operator/pkg/apis/hadoop/v1alpha1"

	"k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// NewResourceManagerResources creates ResourceManager Service and StatefulSet
func NewResourceManagerResources(cluster *v1alpha1.HadoopCluster) (*v1.Service, *v1.StatefulSet, error) {
	service := NewResourceManagerService(cluster)
	statefulSet := NewResourceManagerStatefulSet(cluster)

	return service, statefulSet, nil
}

// NewResourceManagerService creates ResourceManager Service
func NewResourceManagerService(cluster *v1alpha1.HadoopCluster) *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-resourcemanager",
			Namespace: cluster.Namespace,
			Labels: map[string]string{
				"app":     "hadoop-resourcemanager",
				"cluster": cluster.Name,
				"app.kubernetes.io/name":       "hadoop-resourcemanager",
				"app.kubernetes.io/instance":   cluster.Name,
				"app.kubernetes.io/part-of":    "hadoop-operator",
			},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cluster, v1alpha1.SchemeGroupVersion.WithKind("HadoopCluster")),
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Port:       8088,
					Name:       "web",
					TargetPort: intstr.FromInt(8088),
				},
				{
					Port:       8032,
					Name:       "rpc",
					TargetPort: intstr.FromInt(8032),
				},
			},
			ClusterIP: "None",
			Selector: map[string]string{
				"app":     "hadoop-resourcemanager",
				"cluster": cluster.Name,
			},
		},
	}
}

// NewResourceManagerStatefulSet creates ResourceManager StatefulSet
func NewResourceManagerStatefulSet(cluster *v1alpha1.HadoopCluster) *v1.StatefulSet {
	// Get image from spec or use default
	image := cluster.Spec.Image
	if image == "" {
		image = "zhuyifeiruichuang/hadoop:3.1.1"
	}

	// Set default replicas if not specified
	replicas := cluster.Spec.ResourceManager.Replicas
	if replicas == 0 {
		replicas = 1
	}

	statefulSet := &v1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-resourcemanager",
			Namespace: cluster.Namespace,
			Labels: map[string]string{
				"app":     "hadoop-resourcemanager",
				"cluster": cluster.Name,
				"app.kubernetes.io/name":       "hadoop-resourcemanager",
				"app.kubernetes.io/instance":   cluster.Name,
				"app.kubernetes.io/part-of":    "hadoop-operator",
			},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cluster, v1alpha1.SchemeGroupVersion.WithKind("HadoopCluster")),
			},
		},
		Spec: v1.StatefulSetSpec{
			ServiceName: cluster.Name + "-resourcemanager",
			Replicas:    &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":     "hadoop-resourcemanager",
					"cluster": cluster.Name,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":     "hadoop-resourcemanager",
						"cluster": cluster.Name,
						"app.kubernetes.io/name":       "hadoop-resourcemanager",
						"app.kubernetes.io/instance":   cluster.Name,
						"app.kubernetes.io/part-of":    "hadoop-operator",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "resourcemanager",
							Image: image,
							Command: []string{"yarn", "resourcemanager"},
							Env: []v1.EnvVar{
								{
									Name:  "HADOOP_CONF_DIR",
									Value: "/opt/hadoop/etc/hadoop",
								},
							},
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 8088,
									Name:          "web",
								},
								{
									ContainerPort: 8032,
									Name:          "rpc",
								},
							},
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									"cpu":    cluster.Spec.ResourceManager.Resources.Requests.CPU,
									"memory": cluster.Spec.ResourceManager.Resources.Requests.Memory,
								},
								Limits: v1.ResourceList{
									"cpu":    cluster.Spec.ResourceManager.Resources.Limits.CPU,
									"memory": cluster.Spec.ResourceManager.Resources.Limits.Memory,
								},
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "hadoop-config-volume",
									MountPath: "/opt/hadoop/etc/hadoop/core-site.xml",
									SubPath:   "core-site.xml",
								},
								{
									Name:      "hadoop-config-volume",
									MountPath: "/opt/hadoop/etc/hadoop/hdfs-site.xml",
									SubPath:   "hdfs-site.xml",
								},
								{
									Name:      "hadoop-config-volume",
									MountPath: "/opt/hadoop/etc/hadoop/yarn-site.xml",
									SubPath:   "yarn-site.xml",
								},
								{
									Name:      "hadoop-config-volume",
									MountPath: "/opt/hadoop/etc/hadoop/mapred-site.xml",
									SubPath:   "mapred-site.xml",
								},
								{
									Name:      "hadoop-config-volume",
									MountPath: "/opt/hadoop/etc/hadoop/capacity-scheduler.xml",
									SubPath:   "capacity-scheduler.xml",
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "hadoop-config-volume",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "hadoop-config-" + cluster.Name,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Add affinity if specified
	if cluster.Spec.ResourceManager.Affinity != nil {
		statefulSet.Spec.Template.Spec.Affinity = &v1.Affinity{}
		if cluster.Spec.ResourceManager.Affinity.NodeAffinity != nil {
			statefulSet.Spec.Template.Spec.Affinity.NodeAffinity = &v1.NodeAffinity{}
			if cluster.Spec.ResourceManager.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
				nodeSelectorTerms := []v1.NodeSelectorTerm{}
				for _, term := range cluster.Spec.ResourceManager.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
					nodeSelectorTerm := v1.NodeSelectorTerm{}
					for _, expr := range term.MatchExpressions {
						nodeSelectorTerm.MatchExpressions = append(nodeSelectorTerm.MatchExpressions, v1.NodeSelectorRequirement{
							Key:      expr.Key,
							Operator: v1.NodeSelectorOperator(expr.Operator),
							Values:   expr.Values,
						})
					}
					nodeSelectorTerms = append(nodeSelectorTerms, nodeSelectorTerm)
				}
				statefulSet.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &v1.NodeSelector{
					NodeSelectorTerms: nodeSelectorTerms,
				}
			}
			if len(cluster.Spec.ResourceManager.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
				preferredTerms := []v1.PreferredSchedulingTerm{}
				for _, term := range cluster.Spec.ResourceManager.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
					nodeSelectorTerm := v1.NodeSelectorTerm{}
					for _, expr := range term.Preference.MatchExpressions {
						nodeSelectorTerm.MatchExpressions = append(nodeSelectorTerm.MatchExpressions, v1.NodeSelectorRequirement{
							Key:      expr.Key,
							Operator: v1.NodeSelectorOperator(expr.Operator),
							Values:   expr.Values,
						})
					}
					preferredTerms = append(preferredTerms, v1.PreferredSchedulingTerm{
						Weight:     term.Weight,
						Preference: nodeSelectorTerm,
					})
				}
				statefulSet.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = preferredTerms
			}
		}
	}

	return statefulSet
}
