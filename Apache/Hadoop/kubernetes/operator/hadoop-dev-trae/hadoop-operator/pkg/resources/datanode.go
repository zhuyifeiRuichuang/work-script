package resources

import (
	"hadoop-operator/pkg/apis/hadoop/v1alpha1"

	"k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// NewDataNodeResources creates DataNode Service and StatefulSet
func NewDataNodeResources(cluster *v1alpha1.HadoopCluster) (*v1.Service, *v1.StatefulSet, error) {
	service := NewDataNodeService(cluster)
	statefulSet := NewDataNodeStatefulSet(cluster)

	return service, statefulSet, nil
}

// NewDataNodeService creates DataNode Service
func NewDataNodeService(cluster *v1alpha1.HadoopCluster) *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-datanode",
			Namespace: cluster.Namespace,
			Labels: map[string]string{
				"app":     "hadoop-datanode",
				"cluster": cluster.Name,
				"app.kubernetes.io/name":       "hadoop-datanode",
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
					Port:       9864,
					Name:       "web",
					TargetPort: intstr.FromInt(9864),
				},
				{
					Port:       9866,
					Name:       "data",
					TargetPort: intstr.FromInt(9866),
				},
			},
			ClusterIP: "None",
			Selector: map[string]string{
				"app":     "hadoop-datanode",
				"cluster": cluster.Name,
			},
		},
	}
}

// NewDataNodeStatefulSet creates DataNode StatefulSet
func NewDataNodeStatefulSet(cluster *v1alpha1.HadoopCluster) *v1.StatefulSet {
	// Get image from spec or use default
	image := cluster.Spec.Image
	if image == "" {
		image = "zhuyifeiruichuang/hadoop:3.1.1"
	}

	// Set default replicas if not specified
	replicas := cluster.Spec.DataNode.Replicas
	if replicas == 0 {
		replicas = 1
	}

	statefulSet := &v1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-datanode",
			Namespace: cluster.Namespace,
			Labels: map[string]string{
				"app":     "hadoop-datanode",
				"cluster": cluster.Name,
				"app.kubernetes.io/name":       "hadoop-datanode",
				"app.kubernetes.io/instance":   cluster.Name,
				"app.kubernetes.io/part-of":    "hadoop-operator",
			},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cluster, v1alpha1.SchemeGroupVersion.WithKind("HadoopCluster")),
			},
		},
		Spec: v1.StatefulSetSpec{
			ServiceName: cluster.Name + "-datanode",
			Replicas:    &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":     "hadoop-datanode",
					"cluster": cluster.Name,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":     "hadoop-datanode",
						"cluster": cluster.Name,
						"app.kubernetes.io/name":       "hadoop-datanode",
						"app.kubernetes.io/instance":   cluster.Name,
						"app.kubernetes.io/part-of":    "hadoop-operator",
					},
				},
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
						{
							Name:  "wait-for-namenode",
							Image: image,
							Command: []string{"/bin/sh", "-c"},
							Args: []string{
								`echo "⏳ 正在检测 NameNode RPC 端口 (9000)..."
								while ! nc -z ` + cluster.Name + `-namenode-0.` + cluster.Name + `-namenode.` + cluster.Namespace + `.svc.cluster.local 9000; do
								  echo "   等待 NameNode 响应中..."
								  sleep 5
								
done
								echo "✅ NameNode RPC 已就绪"`,
							},
						},
						{
							Name:  "init-permissions",
							Image: image,
							Command: []string{"/bin/sh", "-c"},
							Args: []string{
								`echo "🔧 正在准备数据目录并校准权限..."
								mkdir -p /opt/hadoop/data/dn
								# 动态检测镜像内是否存在 hadoop 用户并授权
								if id "hadoop" >/dev/null 2>&1; then
								  echo "检测到 hadoop 用户，执行 chown hadoop:hadoop"
								  chown -R hadoop:hadoop /opt/hadoop/data
								else
								  echo "未检测到 hadoop 用户，回退使用 UID 1000"
								  chown -R 1000:1000 /opt/hadoop/data
								fi
								# 授予 775 权限确保 DataNode 内部检查通过
								chmod -R 775 /opt/hadoop/data
								echo "✅ 权限校准完成"`,
							},
						},
					},
					Containers: []v1.Container{
						{
							Name:  "datanode",
							Image: image,
							Command: []string{"/bin/bash", "-c"},
							Args: []string{
								`export HADOOP_CONF_DIR=/opt/hadoop/etc/hadoop
								echo "🚀 启动 DataNode..."
								exec hdfs datanode`,
							},
							Env: []v1.EnvVar{
								{
									Name:  "HADOOP_CONF_DIR",
									Value: "/opt/hadoop/etc/hadoop",
								},
							},
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 9864,
									Name:          "web",
								},
								{
									ContainerPort: 9866,
									Name:          "data",
								},
							},
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									"cpu":    cluster.Spec.DataNode.Resources.Requests.CPU,
									"memory": cluster.Spec.DataNode.Resources.Requests.Memory,
								},
								Limits: v1.ResourceList{
									"cpu":    cluster.Spec.DataNode.Resources.Limits.CPU,
									"memory": cluster.Spec.DataNode.Resources.Limits.Memory,
								},
							},
							LivenessProbe: &v1.Probe{
								HTTPGet: &v1.HTTPGetAction{
									Path: "/",
									Port: intstr.FromInt(9864),
								},
								InitialDelaySeconds: 60,
								PeriodSeconds:       20,
							},
							ReadinessProbe: &v1.Probe{
								HTTPGet: &v1.HTTPGetAction{
									Path: "/",
									Port: intstr.FromInt(9864),
								},
								InitialDelaySeconds: 20,
								PeriodSeconds:       10,
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "hadoop-dn-data",
									MountPath: "/opt/hadoop/data/dn",
									SubPath:   "dn",
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
								{
									Name:      "hadoop-config-volume",
									MountPath: "/opt/hadoop/etc/hadoop/yarn-site.xml",
									SubPath:   "yarn-site.xml",
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
			VolumeClaimTemplates: []v1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "hadoop-dn-data",
					},
					Spec: v1.PersistentVolumeClaimSpec{
						AccessModes: []v1.PersistentVolumeAccessMode{
							v1.ReadWriteOnce,
						},
						StorageClassName: &cluster.Spec.DataNode.Storage.StorageClass,
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								"storage": v1.MustParse(cluster.Spec.DataNode.Storage.Size),
							},
						},
					},
				},
			},
		},
	}

	// Add affinity if specified
	if cluster.Spec.DataNode.Affinity != nil {
		statefulSet.Spec.Template.Spec.Affinity = &v1.Affinity{}
		if cluster.Spec.DataNode.Affinity.NodeAffinity != nil {
			statefulSet.Spec.Template.Spec.Affinity.NodeAffinity = &v1.NodeAffinity{}
			if cluster.Spec.DataNode.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
				nodeSelectorTerms := []v1.NodeSelectorTerm{}
				for _, term := range cluster.Spec.DataNode.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
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
			if len(cluster.Spec.DataNode.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
				preferredTerms := []v1.PreferredSchedulingTerm{}
				for _, term := range cluster.Spec.DataNode.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
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
