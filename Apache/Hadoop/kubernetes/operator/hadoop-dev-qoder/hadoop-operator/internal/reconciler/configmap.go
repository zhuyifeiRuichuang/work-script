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
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	hadoopv1 "github.com/apache/hadoop-operator/api/v1"
)

// HadoopClusterReconciler holds the reconciler state
type HadoopClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// ComponentReconciler is a function type for component reconciliation
type ComponentReconciler func(ctx context.Context, cluster *hadoopv1.HadoopCluster) (ctrl.Result, error)

// reconcileConfigMap creates or updates the Hadoop ConfigMap
func (r *HadoopClusterReconciler) reconcileConfigMap(ctx context.Context, cluster *hadoopv1.HadoopCluster) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-config",
			Namespace: cluster.Namespace,
			Labels: map[string]string{
				"hadoop.apache.org/cluster":    cluster.Name,
				"hadoop.apache.org/managed-by": "hadoop-operator",
			},
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, cm, func() error {
		cm.Data = r.generateHadoopConfig(cluster)
		return controllerutil.SetControllerReference(cluster, cm, r.Scheme)
	})
	if err != nil {
		log.Error(err, "Failed to create ConfigMap")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *HadoopClusterReconciler) generateHadoopConfig(cluster *hadoopv1.HadoopCluster) map[string]string {
	config := make(map[string]string)

	// Get service names
	nnService := cluster.Name + "-namenode-0." + cluster.Name + "-namenode." + cluster.Namespace + ".svc.cluster.local"
	rmService := cluster.Name + "-resourcemanager-0." + cluster.Name + "-resourcemanager." + cluster.Namespace + ".svc.cluster.local"

	// Check if HA mode is enabled
	isNNHA := cluster.Spec.HDFS.NameNode.HA != nil && cluster.Spec.HDFS.NameNode.HA.Enabled
	isRMHA := cluster.Spec.YARN.ResourceManager.HA != nil && cluster.Spec.YARN.ResourceManager.HA.Enabled

	// core-site.xml
	coreSite := map[string]string{
		"hadoop.tmp.dir":                       "/opt/hadoop/tmpdata",
		"fs.defaultFS":                         fmt.Sprintf("hdfs://%s:9000", nnService),
		"hadoop.proxyuser.hadoop.hosts":        "*",
		"hadoop.proxyuser.hadoop.groups":       "*",
		"hadoop.proxyuser.hue.hosts":           "*",
		"hadoop.proxyuser.hue.groups":          "*",
		"hadoop.proxyuser.hive.hosts":          "*",
		"hadoop.proxyuser.hive.groups":         "*",
		"hadoop.proxyuser.root.hosts":          "*",
		"hadoop.proxyuser.root.groups":         "*",
	}

	if isNNHA {
		coreSite["fs.defaultFS"] = fmt.Sprintf("hdfs://%s", cluster.Name)
		coreSite["ha.zookeeper.quorum"] = r.getZooKeeperQuorum(cluster)
	}

	// Merge user overrides
	for k, v := range cluster.Spec.Config.CoreSite {
		coreSite[k] = v
	}
	config["core-site.xml"] = r.generateXMLConfig("configuration", coreSite)

	// hdfs-site.xml
	hdfsSite := map[string]string{
		"dfs.namenode.name.dir":                              "/opt/hadoop/data/nn",
		"dfs.datanode.data.dir":                              "/opt/hadoop/data/dn",
		"dfs.namenode.rpc-address":                           fmt.Sprintf("%s:9000", nnService),
		"dfs.namenode.rpc-bind-host":                         "0.0.0.0",
		"dfs.namenode.http-bind-host":                        "0.0.0.0",
		"dfs.replication":                                    "3",
		"dfs.client.use.datanode.hostname":                   "true",
		"dfs.permissions.enabled":                            "false",
		"dfs.webhdfs.enabled":                                "true",
		"dfs.namenode.datanode.registration.ip-hostname-check": "false",
	}

	if isNNHA {
		hdfsSite["dfs.nameservices"] = cluster.Name
		hdfsSite[fmt.Sprintf("dfs.ha.namenodes.%s", cluster.Name)] = "nn1,nn2"
		hdfsSite[fmt.Sprintf("dfs.namenode.rpc-address.%s.nn1", cluster.Name)] = fmt.Sprintf("%s-namenode-0.%s-namenode.%s.svc.cluster.local:9000", cluster.Name, cluster.Name, cluster.Namespace)
		hdfsSite[fmt.Sprintf("dfs.namenode.rpc-address.%s.nn2", cluster.Name)] = fmt.Sprintf("%s-namenode-1.%s-namenode.%s.svc.cluster.local:9000", cluster.Name, cluster.Name, cluster.Namespace)
		hdfsSite[fmt.Sprintf("dfs.namenode.http-address.%s.nn1", cluster.Name)] = fmt.Sprintf("%s-namenode-0.%s-namenode.%s.svc.cluster.local:9870", cluster.Name, cluster.Name, cluster.Namespace)
		hdfsSite[fmt.Sprintf("dfs.namenode.http-address.%s.nn2", cluster.Name)] = fmt.Sprintf("%s-namenode-1.%s-namenode.%s.svc.cluster.local:9870", cluster.Name, cluster.Name, cluster.Namespace)
		hdfsSite["dfs.client.failover.proxy.provider."+cluster.Name] = "org.apache.hadoop.hdfs.server.namenode.ha.ConfiguredFailoverProxyProvider"
		hdfsSite["dfs.ha.fencing.methods"] = "shell(/bin/true)"
		hdfsSite["dfs.ha.automatic-failover.enabled"] = "true"
	}

	// Merge user overrides
	for k, v := range cluster.Spec.Config.HDFSSite {
		hdfsSite[k] = v
	}
	config["hdfs-site.xml"] = r.generateXMLConfig("configuration", hdfsSite)

	// yarn-site.xml
	yarnSite := map[string]string{
		"yarn.resourcemanager.hostname":                           rmService,
		"yarn.resourcemanager.bind-host":                          "0.0.0.0",
		"yarn.resourcemanager.webapp.address":                     "0.0.0.0:8088",
		"yarn.nodemanager.bind-host":                              "0.0.0.0",
		"yarn.nodemanager.pmem-check-enabled":                     "false",
		"yarn.nodemanager.delete.debug-delay-sec":                 "600",
		"yarn.nodemanager.vmem-check-enabled":                     "false",
		"yarn.nodemanager.aux-services":                           "mapreduce_shuffle",
		"yarn.nodemanager.aux-services.mapreduce_shuffle.class":   "org.apache.hadoop.mapred.ShuffleHandler",
		"yarn.acl.enable":                                         "false",
		"yarn.nodemanager.env-whitelist":                          "JAVA_HOME,HADOOP_COMMON_HOME,HADOOP_HDFS_HOME,HADOOP_CONF_DIR,CLASSPATH_PREPEND_DISTCACHE,HADOOP_YARN_HOME,HADOOP_HOME,PATH,LANG,TZ,HADOOP_MAPRED_HOME",
	}

	if isRMHA {
		yarnSite["yarn.resourcemanager.ha.enabled"] = "true"
		yarnSite["yarn.resourcemanager.cluster-id"] = cluster.Name + "-rm"
		yarnSite["yarn.resourcemanager.ha.rm-ids"] = "rm1,rm2"
		yarnSite["yarn.resourcemanager.hostname.rm1"] = fmt.Sprintf("%s-resourcemanager-0.%s-resourcemanager.%s.svc.cluster.local", cluster.Name, cluster.Name, cluster.Namespace)
		yarnSite["yarn.resourcemanager.hostname.rm2"] = fmt.Sprintf("%s-resourcemanager-1.%s-resourcemanager.%s.svc.cluster.local", cluster.Name, cluster.Name, cluster.Namespace)
		yarnSite["yarn.resourcemanager.webapp.address.rm1"] = fmt.Sprintf("%s-resourcemanager-0.%s-resourcemanager.%s.svc.cluster.local:8088", cluster.Name, cluster.Name, cluster.Namespace)
		yarnSite["yarn.resourcemanager.webapp.address.rm2"] = fmt.Sprintf("%s-resourcemanager-1.%s-resourcemanager.%s.svc.cluster.local:8088", cluster.Name, cluster.Name, cluster.Namespace)
		yarnSite["yarn.resourcemanager.zk-address"] = r.getZooKeeperQuorum(cluster)
	}

	// Merge user overrides
	for k, v := range cluster.Spec.Config.YARNSite {
		yarnSite[k] = v
	}
	config["yarn-site.xml"] = r.generateXMLConfig("configuration", yarnSite)

	// mapred-site.xml
	mapredSite := map[string]string{
		"mapreduce.framework.name":       "yarn",
		"yarn.app.mapreduce.am.env":      "HADOOP_MAPRED_HOME=/opt/hadoop",
		"mapreduce.map.env":              "HADOOP_MAPRED_HOME=/opt/hadoop",
		"mapreduce.reduce.env":           "HADOOP_MAPRED_HOME=/opt/hadoop",
	}

	// Merge user overrides
	for k, v := range cluster.Spec.Config.MapredSite {
		mapredSite[k] = v
	}
	config["mapred-site.xml"] = r.generateXMLConfig("configuration", mapredSite)

	// capacity-scheduler.xml
	capacityScheduler := map[string]string{
		"yarn.scheduler.capacity.maximum-applications":             "10000",
		"yarn.scheduler.capacity.maximum-am-resource-percent":      "0.1",
		"yarn.scheduler.capacity.resource-calculator":              "org.apache.hadoop.yarn.util.resource.DefaultResourceCalculator",
		"yarn.scheduler.capacity.root.queues":                      "default",
		"yarn.scheduler.capacity.root.default.capacity":            "100",
		"yarn.scheduler.capacity.root.default.user-limit-factor":   "1",
		"yarn.scheduler.capacity.root.default.maximum-capacity":    "100",
		"yarn.scheduler.capacity.root.default.state":               "RUNNING",
		"yarn.scheduler.capacity.root.default.acl_submit_applications": "*",
		"yarn.scheduler.capacity.root.default.acl_administer_queue":    "*",
		"yarn.scheduler.capacity.node-locality-delay":              "40",
		"yarn.scheduler.capacity.root.default.acl_application_max_priority": "*",
		"yarn.scheduler.capacity.root.default.maximum-application-lifetime": "-1",
		"yarn.scheduler.capacity.rack-locality-additional-delay":   "-1",
	}

	// Merge user overrides
	for k, v := range cluster.Spec.Config.CapacityScheduler {
		capacityScheduler[k] = v
	}
	config["capacity-scheduler.xml"] = r.generateXMLConfig("configuration", capacityScheduler)

	return config
}

func (r *HadoopClusterReconciler) generateXMLConfig(rootTag string, properties map[string]string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<?xml version=\"1.0\"?>\n<%s>\n", rootTag))
	for key, value := range properties {
		sb.WriteString("  <property>\n")
		sb.WriteString(fmt.Sprintf("    <name>%s</name>\n", key))
		sb.WriteString(fmt.Sprintf("    <value>%s</value>\n", value))
		sb.WriteString("  </property>\n")
	}
	sb.WriteString(fmt.Sprintf("</%s>\n", rootTag))
	return sb.String()
}

func (r *HadoopClusterReconciler) getZooKeeperQuorum(cluster *hadoopv1.HadoopCluster) string {
	// Check if external ZooKeeper is configured
	if cluster.Spec.HDFS.NameNode.HA != nil && cluster.Spec.HDFS.NameNode.HA.ZooKeeper != nil {
		if cluster.Spec.HDFS.NameNode.HA.ZooKeeper.ConnectionString != "" {
			return cluster.Spec.HDFS.NameNode.HA.ZooKeeper.ConnectionString
		}
	}
	// Default to internal ZooKeeper
	return fmt.Sprintf("%s-zookeeper-0.%s-zookeeper.%s.svc.cluster.local:2181", cluster.Name, cluster.Name, cluster.Namespace)
}

func (r *HadoopClusterReconciler) getImage(cluster *hadoopv1.HadoopCluster) string {
	repo := cluster.Spec.Image.Repository
	if repo == "" {
		repo = "apache/hadoop"
	}
	tag := cluster.Spec.Image.Tag
	if tag == "" {
		tag = "3.3.6"
	}
	return fmt.Sprintf("%s:%s", repo, tag)
}
