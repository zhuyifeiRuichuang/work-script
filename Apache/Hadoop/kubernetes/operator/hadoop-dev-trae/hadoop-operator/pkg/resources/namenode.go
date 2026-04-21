package resources

import (
	"hadoop-operator/pkg/apis/hadoop/v1alpha1"

	"k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// NewNameNodeResources creates NameNode Service and StatefulSet
func NewNameNodeResources(cluster *v1alpha1.HadoopCluster) (*v1.Service, *v1.StatefulSet, error) {
	service := NewNameNodeService(cluster)
	statefulSet := NewNameNodeStatefulSet(cluster)

	return service, statefulSet, nil
}

// NewNameNodeService creates NameNode Service
func NewNameNodeService(cluster *v1alpha1.HadoopCluster) *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-namenode",
			Namespace: cluster.Namespace,
			Labels: map[string]string{
				"app":     "hadoop-namenode",
				"cluster": cluster.Name,
				"app.kubernetes.io/name":       "hadoop-namenode",
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
					Port:       9870,
					Name:       "web",
					TargetPort: intstr.FromInt(9870),
				},
				{
					Port:       9000,
					Name:       "rpc",
					TargetPort: intstr.FromInt(9000),
				},
			},
			ClusterIP: "None",
			Selector: map[string]string{
				"app":     "hadoop-namenode",
				"cluster": cluster.Name,
			},
		},
	}
}

// NewNameNodeStatefulSet creates NameNode StatefulSet
func NewNameNodeStatefulSet(cluster *v1alpha1.HadoopCluster) *v1.StatefulSet {
	// Get image from spec or use default
	image := cluster.Spec.Image
	if image == "" {
		image = "zhuyifeiruichuang/hadoop:3.1.1"
	}

	// Set default replicas if not specified
	replicas := cluster.Spec.NameNode.Replicas
	if replicas == 0 {
		replicas = 1
	}

	// Prepare labels
	labels := map[string]string{
		"app":     "hadoop-namenode",
		"cluster": cluster.Name,
		"app.kubernetes.io/name":       "hadoop-namenode",
		"app.kubernetes.io/instance":   cluster.Name,
		"app.kubernetes.io/part-of":    "hadoop-operator",
	}

	// Add user-defined labels
	for k, v := range cluster.Spec.NameNode.PodLabels {
		labels[k] = v
	}

	statefulSet := &v1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-namenode",
			Namespace: cluster.Namespace,
			Labels:    labels,
			Annotations: cluster.Spec.NameNode.Annotations,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cluster, v1alpha1.SchemeGroupVersion.WithKind("HadoopCluster")),
			},
		},
		Spec: v1.StatefulSetSpec{
			ServiceName: cluster.Name + "-namenode",
			Replicas:    &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: cluster.Spec.NameNode.Annotations,
				},
				Spec: v1.PodSpec{
					ImagePullSecrets: []v1.LocalObjectReference{},
					NodeSelector:     cluster.Spec.NameNode.NodeSelector,
					Tolerations:      []v1.Toleration{},
					HostAliases:      []v1.HostAlias{},
					InitContainers: []v1.Container{
						{
							Name:  "init-namenode",
							Image: image,
							Command: []string{"/bin/bash", "-c"},
							Args: []string{
								`set -e
								echo "Step 1: Preparing directories and permissions..."
								mkdir -p /opt/hadoop/data/nn
								# 确保 hadoop 用户拥有数据目录权限
								chown -R hadoop:hadoop /opt/hadoop/data
								echo "Step 2: Checking if NameNode needs formatting..."
								if [ ! -f /opt/hadoop/data/nn/current/VERSION ]; then
								  echo "No VERSION file found. Formatting NameNode..."
								  # 使用非交互模式强制格式化
								  su hadoop -c "hdfs namenode -format -nonInteractive"
								  echo "Formatting completed successfully."
								else
								  echo "NameNode already formatted (VERSION file exists). Skipping."
								  grep -i "clusterID" /opt/hadoop/data/nn/current/VERSION
								fi`,
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "hadoop-nn-data",
									MountPath: "/opt/hadoop/data/nn",
									SubPath:   "nn",
								},
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
							},
						},
					},
					Containers: []v1.Container{
						{
							Name:  "namenode",
							Image: image,
							Command: []string{"/bin/bash", "-c"},
							Args:    []string{"exec hdfs namenode"},
							Env: []v1.EnvVar{
								{
									Name:  "HADOOP_CONF_DIR",
									Value: "/opt/hadoop/etc/hadoop",
								},
								{
									Name:  "HADOOP_LOG_DIR",
									Value: "/opt/hadoop/logs",
								},
								{
									Name:  "HADOOP_HEAPSIZE",
									Value: "1024",
								},
							},
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 9870,
									Name:          "web",
								},
								{
									ContainerPort: 9000,
									Name:          "rpc",
								},
							},
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									"cpu":    cluster.Spec.NameNode.Resources.Requests.CPU,
									"memory": cluster.Spec.NameNode.Resources.Requests.Memory,
								},
								Limits: v1.ResourceList{
									"cpu":    cluster.Spec.NameNode.Resources.Limits.CPU,
									"memory": cluster.Spec.NameNode.Resources.Limits.Memory,
								},
							},
							ReadinessProbe: &v1.Probe{
								HTTPGet: &v1.HTTPGetAction{
									Path: "/jmx",
									Port: intstr.FromInt(9870),
								},
								InitialDelaySeconds: 40,
								PeriodSeconds:       10,
							},
							LivenessProbe: &v1.Probe{
								HTTPGet: &v1.HTTPGetAction{
									Path: "/jmx",
									Port: intstr.FromInt(9870),
								},
								InitialDelaySeconds: 100,
								PeriodSeconds:       30,
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "hadoop-nn-data",
									MountPath: "/opt/hadoop/data/nn",
									SubPath:   "nn",
								},
								{
									Name:      "hadoop-logs",
									MountPath: "/opt/hadoop/logs",
								},
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
						{
							Name: "hadoop-logs",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []v1.PersistentVolumeClaim{},
		},
	}

	// Add persistent volumes
	for _, pv := range cluster.Spec.NameNode.PersistentVolumes {
		pvc := v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:        pv.Name,
				Annotations: pv.Annotations,
			},
			Spec: v1.PersistentVolumeClaimSpec{
				AccessModes:      []v1.PersistentVolumeAccessMode{},
				Resources:        v1.ResourceRequirements{},
				StorageClassName: &pv.PersistentVolumeClaimSpec.StorageClassName,
				VolumeName:       pv.PersistentVolumeClaimSpec.VolumeName,
			},
		}

		// Add access modes
		for _, mode := range pv.PersistentVolumeClaimSpec.AccessModes {
			pvc.Spec.AccessModes = append(pvc.Spec.AccessModes, v1.PersistentVolumeAccessMode(mode))
		}

		// Add resources
		if pv.PersistentVolumeClaimSpec.Resources.Requests.CPU != 0 {
			pvc.Spec.Resources.Requests = v1.ResourceList{
				"cpu":    pv.PersistentVolumeClaimSpec.Resources.Requests.CPU,
				"memory": pv.PersistentVolumeClaimSpec.Resources.Requests.Memory,
			}
		}

		// Add storage request
		if pv.PersistentVolumeClaimSpec.Resources.Requests.Storage() != nil {
			pvc.Spec.Resources.Requests["storage"] = *pv.PersistentVolumeClaimSpec.Resources.Requests.Storage()
		}

		statefulSet.Spec.VolumeClaimTemplates = append(statefulSet.Spec.VolumeClaimTemplates, pvc)
	}

	// Add security context
	if cluster.Spec.NameNode.SecurityContext != nil {
		statefulSet.Spec.Template.Spec.SecurityContext = &v1.PodSecurityContext{
			RunAsUser:    cluster.Spec.NameNode.SecurityContext.RunAsUser,
			RunAsGroup:   cluster.Spec.NameNode.SecurityContext.RunAsGroup,
			RunAsNonRoot: cluster.Spec.NameNode.SecurityContext.RunAsNonRoot,
			FSGroup:      cluster.Spec.NameNode.SecurityContext.FSGroup,
		}
	}

	// Add container security context
	if cluster.Spec.NameNode.ContainerSecurityContext != nil {
		statefulSet.Spec.Template.Spec.Containers[0].SecurityContext = &v1.SecurityContext{
			RunAsUser:              cluster.Spec.NameNode.ContainerSecurityContext.RunAsUser,
			RunAsGroup:             cluster.Spec.NameNode.ContainerSecurityContext.RunAsGroup,
			RunAsNonRoot:           cluster.Spec.NameNode.ContainerSecurityContext.RunAsNonRoot,
			ReadOnlyRootFilesystem: cluster.Spec.NameNode.ContainerSecurityContext.ReadOnlyRootFilesystem,
			Privileged:             cluster.Spec.NameNode.ContainerSecurityContext.Privileged,
		}
	}

	// Add environment variables
	for _, env := range cluster.Spec.NameNode.EnvVars {
		envVar := v1.EnvVar{
			Name:  env.Name,
			Value: env.Value,
		}

		if env.ValueFrom != nil {
			if env.ValueFrom.ConfigMapKeyRef != nil {
				envVar.ValueFrom = &v1.EnvVarSource{
					ConfigMapKeyRef: &v1.ConfigMapKeySelector{
						Name:      env.ValueFrom.ConfigMapKeyRef.Name,
						Key:       env.ValueFrom.ConfigMapKeyRef.Key,
						Optional:  env.ValueFrom.ConfigMapKeyRef.Optional,
					},
				}
			}

			if env.ValueFrom.SecretKeyRef != nil {
				envVar.ValueFrom = &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						Name:      env.ValueFrom.SecretKeyRef.Name,
						Key:       env.ValueFrom.SecretKeyRef.Key,
						Optional:  env.ValueFrom.SecretKeyRef.Optional,
					},
				}
			}
		}

		statefulSet.Spec.Template.Spec.Containers[0].Env = append(statefulSet.Spec.Template.Spec.Containers[0].Env, envVar)
	}

	// Add readiness probe policy
	if cluster.Spec.NameNode.ReadinessProbePolicy != nil {
		if cluster.Spec.NameNode.ReadinessProbePolicy.PeriodSeconds != nil {
			statefulSet.Spec.Template.Spec.Containers[0].ReadinessProbe.PeriodSeconds = *cluster.Spec.NameNode.ReadinessProbePolicy.PeriodSeconds
		}

		if cluster.Spec.NameNode.ReadinessProbePolicy.TimeoutSeconds != nil {
			statefulSet.Spec.Template.Spec.Containers[0].ReadinessProbe.TimeoutSeconds = *cluster.Spec.NameNode.ReadinessProbePolicy.TimeoutSeconds
		}

		if cluster.Spec.NameNode.ReadinessProbePolicy.FailureThreshold != nil {
			statefulSet.Spec.Template.Spec.Containers[0].ReadinessProbe.FailureThreshold = int32(*cluster.Spec.NameNode.ReadinessProbePolicy.FailureThreshold)
		}
	}

	// Add liveness probe policy
	if cluster.Spec.NameNode.LivenessProbePolicy != nil {
		if cluster.Spec.NameNode.LivenessProbePolicy.PeriodSeconds != nil {
			statefulSet.Spec.Template.Spec.Containers[0].LivenessProbe.PeriodSeconds = *cluster.Spec.NameNode.LivenessProbePolicy.PeriodSeconds
		}

		if cluster.Spec.NameNode.LivenessProbePolicy.TimeoutSeconds != nil {
			statefulSet.Spec.Template.Spec.Containers[0].LivenessProbe.TimeoutSeconds = *cluster.Spec.NameNode.LivenessProbePolicy.TimeoutSeconds
		}

		if cluster.Spec.NameNode.LivenessProbePolicy.FailureThreshold != nil {
			statefulSet.Spec.Template.Spec.Containers[0].LivenessProbe.FailureThreshold = int32(*cluster.Spec.NameNode.LivenessProbePolicy.FailureThreshold)
		}
	}

	// Add image pull secrets
	for _, secret := range cluster.Spec.NameNode.ImagePullSecrets {
		statefulSet.Spec.Template.Spec.ImagePullSecrets = append(statefulSet.Spec.Template.Spec.ImagePullSecrets, v1.LocalObjectReference{
			Name: secret.Name,
		})
	}

	// Add tolerations
	for _, toleration := range cluster.Spec.NameNode.Tolerations {
		statefulSet.Spec.Template.Spec.Tolerations = append(statefulSet.Spec.Template.Spec.Tolerations, v1.Toleration{
			Key:               toleration.Key,
			Operator:          v1.TolerationOperator(toleration.Operator),
			Value:             toleration.Value,
			Effect:            v1.TaintEffect(toleration.Effect),
			TolerationSeconds: toleration.TolerationSeconds,
		})
	}

	// Add config maps
	for _, configMap := range cluster.Spec.NameNode.ConfigMaps {
		// Add volume
		statefulSet.Spec.Template.Spec.Volumes = append(statefulSet.Spec.Template.Spec.Volumes, v1.Volume{
			Name: configMap.ConfigMapName,
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: configMap.ConfigMapName,
					},
				},
			},
		})

		// Add volume mount to containers
		for i := range statefulSet.Spec.Template.Spec.Containers {
			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = append(statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts, v1.VolumeMount{
				Name:      configMap.ConfigMapName,
				MountPath: configMap.MountPath,
			})
		}

		// Add volume mount to init containers
		for i := range statefulSet.Spec.Template.Spec.InitContainers {
			statefulSet.Spec.Template.Spec.InitContainers[i].VolumeMounts = append(statefulSet.Spec.Template.Spec.InitContainers[i].VolumeMounts, v1.VolumeMount{
				Name:      configMap.ConfigMapName,
				MountPath: configMap.MountPath,
			})
		}
	}

	// Add host aliases
	for _, hostAlias := range cluster.Spec.NameNode.HostAliases {
		statefulSet.Spec.Template.Spec.HostAliases = append(statefulSet.Spec.Template.Spec.HostAliases, v1.HostAlias{
			IP:        hostAlias.IP,
			Hostnames: hostAlias.Hostnames,
		})
	}

	// Add affinity if specified
	if cluster.Spec.NameNode.Affinity != nil {
		statefulSet.Spec.Template.Spec.Affinity = &v1.Affinity{}

		// Add node affinity
		if cluster.Spec.NameNode.Affinity.NodeAffinity != nil {
			statefulSet.Spec.Template.Spec.Affinity.NodeAffinity = &v1.NodeAffinity{}

			// Add required node affinity
			if cluster.Spec.NameNode.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
				nodeSelectorTerms := []v1.NodeSelectorTerm{}
				for _, term := range cluster.Spec.NameNode.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
					nodeSelectorTerm := v1.NodeSelectorTerm{}

					// Add match expressions
					for _, expr := range term.MatchExpressions {
						nodeSelectorTerm.MatchExpressions = append(nodeSelectorTerm.MatchExpressions, v1.NodeSelectorRequirement{
							Key:      expr.Key,
							Operator: v1.NodeSelectorOperator(expr.Operator),
							Values:   expr.Values,
						})
					}

					// Add match fields
					for _, field := range term.MatchFields {
						nodeSelectorTerm.MatchFields = append(nodeSelectorTerm.MatchFields, v1.NodeSelectorRequirement{
							Key:      field.Key,
							Operator: v1.NodeSelectorOperator(field.Operator),
							Values:   field.Values,
						})
					}

					nodeSelectorTerms = append(nodeSelectorTerms, nodeSelectorTerm)
				}

				statefulSet.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &v1.NodeSelector{
					NodeSelectorTerms: nodeSelectorTerms,
				}
			}

			// Add preferred node affinity
			if len(cluster.Spec.NameNode.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
				preferredTerms := []v1.PreferredSchedulingTerm{}
				for _, term := range cluster.Spec.NameNode.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
					nodeSelectorTerm := v1.NodeSelectorTerm{}

					// Add match expressions
					for _, expr := range term.Preference.MatchExpressions {
						nodeSelectorTerm.MatchExpressions = append(nodeSelectorTerm.MatchExpressions, v1.NodeSelectorRequirement{
							Key:      expr.Key,
							Operator: v1.NodeSelectorOperator(expr.Operator),
							Values:   expr.Values,
						})
					}

					// Add match fields
					for _, field := range term.Preference.MatchFields {
						nodeSelectorTerm.MatchFields = append(nodeSelectorTerm.MatchFields, v1.NodeSelectorRequirement{
							Key:      field.Key,
							Operator: v1.NodeSelectorOperator(field.Operator),
							Values:   field.Values,
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

		// Add pod affinity
		if cluster.Spec.NameNode.Affinity.PodAffinity != nil {
			statefulSet.Spec.Template.Spec.Affinity.PodAffinity = &v1.PodAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{},
				PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{},
			}

			// Add required pod affinity
			for _, term := range cluster.Spec.NameNode.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution {
				statefulSet.Spec.Template.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(
					statefulSet.Spec.Template.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution,
					v1.PodAffinityTerm{
						LabelSelector: term.LabelSelector,
						Namespaces:    term.Namespaces,
						TopologyKey:   term.TopologyKey,
					},
				)
			}

			// Add preferred pod affinity
			for _, term := range cluster.Spec.NameNode.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
				statefulSet.Spec.Template.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(
					statefulSet.Spec.Template.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
					v1.WeightedPodAffinityTerm{
						PodAffinityTerm: v1.PodAffinityTerm{
							LabelSelector: term.PodAffinityTerm.LabelSelector,
							Namespaces:    term.PodAffinityTerm.Namespaces,
							TopologyKey:   term.PodAffinityTerm.TopologyKey,
						},
						Weight:         term.Weight,
					},
				)
			}
		}

		// Add pod anti-affinity
		if cluster.Spec.NameNode.Affinity.PodAntiAffinity != nil {
			statefulSet.Spec.Template.Spec.Affinity.PodAntiAffinity = &v1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{},
				PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{},
			}

			// Add required pod anti-affinity
			for _, term := range cluster.Spec.NameNode.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution {
				statefulSet.Spec.Template.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(
					statefulSet.Spec.Template.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution,
					v1.PodAffinityTerm{
						LabelSelector: term.LabelSelector,
						Namespaces:    term.Namespaces,
						TopologyKey:   term.TopologyKey,
					},
				)
			}

			// Add preferred pod anti-affinity
			for _, term := range cluster.Spec.NameNode.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
				statefulSet.Spec.Template.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(
					statefulSet.Spec.Template.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
					v1.WeightedPodAffinityTerm{
						PodAffinityTerm: v1.PodAffinityTerm{
							LabelSelector: term.PodAffinityTerm.LabelSelector,
							Namespaces:    term.PodAffinityTerm.Namespaces,
							TopologyKey:   term.PodAffinityTerm.TopologyKey,
						},
						Weight:         term.Weight,
					},
				)
			}
		}
	}

	return statefulSet
}
