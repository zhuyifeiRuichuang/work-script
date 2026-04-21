package resources

import (
	"hadoop-operator/pkg/apis/hadoop/v1alpha1"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewConfigMap creates a ConfigMap for Hadoop configuration
func NewConfigMap(cluster *v1alpha1.HadoopCluster) (*v1.ConfigMap, error) {
	// Generate core-site.xml
	coreSiteXML := generateCoreSiteXML(cluster)
	
	// Generate hdfs-site.xml
	hdfsSiteXML := generateHdfSiteXML(cluster)
	
	// Generate yarn-site.xml
	yarnSiteXML := generateYarnSiteXML(cluster)
	
	// Generate mapred-site.xml
	mapredSiteXML := generateMapredSiteXML(cluster)
	
	// Generate capacity-scheduler.xml
	capacitySchedulerXML := generateCapacitySchedulerXML()

	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hadoop-config-" + cluster.Name,
			Namespace: cluster.Namespace,
			Labels: map[string]string{
				"app": "hadoop",
				"cluster": cluster.Name,
			},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cluster, v1alpha1.SchemeGroupVersion.WithKind("HadoopCluster")),
			},
		},
		Data: map[string]string{
			"core-site.xml":        coreSiteXML,
			"hdfs-site.xml":        hdfsSiteXML,
			"yarn-site.xml":        yarnSiteXML,
			"mapred-site.xml":      mapredSiteXML,
			"capacity-scheduler.xml": capacitySchedulerXML,
		},
	}

	return configMap, nil
}

// generateCoreSiteXML generates core-site.xml content
func generateCoreSiteXML(cluster *v1alpha1.HadoopCluster) string {
	return `<?xml version="1.0"?>
<configuration>
  <property>
    <name>hadoop.tmp.dir</name>
    <value>/opt/hadoop/tmpdata</value>
  </property>
  <property>
    <name>fs.defaultFS</name>
    <value>hdfs://` + cluster.Name + `-namenode-0.` + cluster.Name + `-namenode.` + cluster.Namespace + `.svc.cluster.local:9000</value>
  </property>
  <property>
    <name>hadoop.proxyuser.hadoop.hosts</name>
    <value>*</value>
  </property>
  <property>
    <name>hadoop.proxyuser.hadoop.groups</name>
    <value>*</value>
  </property>
  <property>
    <name>hadoop.proxyuser.hue.hosts</name>
    <value>*</value>
  </property>
  <property>
    <name>hadoop.proxyuser.hue.groups</name>
    <value>*</value>
  </property>
  <property>
    <name>hadoop.proxyuser.hive.hosts</name>
    <value>*</value>
  </property>
  <property>
    <name>hadoop.proxyuser.hive.groups</name>
    <value>*</value>
  </property>
  <property>
    <name>hadoop.proxyuser.root.hosts</name>
    <value>*</value>
  </property>
  <property>
    <name>hadoop.proxyuser.root.groups</name>
    <value>*</value>
  </property>
</configuration>`
}

// generateHdfSiteXML generates hdfs-site.xml content
func generateHdfSiteXML(cluster *v1alpha1.HadoopCluster) string {
	return `<?xml version="1.0"?>
<configuration>
  <property>
    <name>dfs.namenode.name.dir</name>
    <value>/opt/hadoop/data/nn</value>
  </property>
  <property>
    <name>dfs.datanode.data.dir</name>
    <value>/opt/hadoop/data/dn</value>
  </property>
  <property>
    <name>dfs.namenode.rpc-address</name>
    <value>` + cluster.Name + `-namenode-0.` + cluster.Name + `-namenode.` + cluster.Namespace + `.svc.cluster.local:9000</value>
  </property>
  <property>
    <name>dfs.namenode.rpc-bind-host</name>
    <value>0.0.0.0</value>
  </property>
  <property>
    <name>dfs.namenode.http-bind-host</name>
    <value>0.0.0.0</value>
  </property>
  <property>
    <name>dfs.replication</name>
    <value>1</value>
  </property>
  <property>
    <name>dfs.client.use.datanode.hostname</name>
    <value>true</value>
  </property>
  <property>
    <name>dfs.permissions.enabled</name>
    <value>false</value>
  </property>
  <property>
    <name>dfs.webhdfs.enabled</name>
    <value>true</value>
  </property>
  <property>
    <name>dfs.namenode.datanode.registration.ip-hostname-check</name>
    <value>false</value>
  </property>
</configuration>`
}

// generateYarnSiteXML generates yarn-site.xml content
func generateYarnSiteXML(cluster *v1alpha1.HadoopCluster) string {
	return `<?xml version="1.0"?>
<configuration>
  <property>
    <name>yarn.resourcemanager.hostname</name>
    <value>` + cluster.Name + `-resourcemanager-0.` + cluster.Name + `-resourcemanager.` + cluster.Namespace + `.svc.cluster.local</value>
  </property>
  <property>
    <name>yarn.resourcemanager.bind-host</name>
    <value>0.0.0.0</value>
  </property>
  <property>
    <name>yarn.resourcemanager.webapp.address</name>
    <value>0.0.0.0:8088</value>
  </property>
  <property>
    <name>yarn.nodemanager.bind-host</name>
    <value>0.0.0.0</value>
  </property>
  <property>
    <name>yarn.nodemanager.pmem-check-enabled</name>
    <value>false</value>
  </property>
  <property>
    <name>yarn.nodemanager.delete.debug-delay-sec</name>
    <value>600</value>
  </property>
  <property>
    <name>yarn.nodemanager.vmem-check-enabled</name>
    <value>false</value>
  </property>
  <property>
    <name>yarn.nodemanager.aux-services</name>
    <value>mapreduce_shuffle</value>
  </property>
  <property>
    <name>yarn.nodemanager.aux-services.mapreduce_shuffle.class</name>
    <value>org.apache.hadoop.mapred.ShuffleHandler</value>
  </property>
  <property>
    <name>yarn.acl.enable</name>
    <value>false</value>
  </property>
  <property>
    <name>yarn.nodemanager.env-whitelist</name>
    <value>JAVA_HOME,HADOOP_COMMON_HOME,HADOOP_HDFS_HOME,HADOOP_CONF_DIR,CLASSPATH_PREPEND_DISTCACHE,HADOOP_YARN_HOME,HADOOP_HOME,PATH,LANG,TZ,HADOOP_MAPRED_HOME</value>
  </property>
</configuration>`
}

// generateMapredSiteXML generates mapred-site.xml content
func generateMapredSiteXML(cluster *v1alpha1.HadoopCluster) string {
	return `<?xml version="1.0"?>
<configuration>
  <property>
    <name>mapreduce.framework.name</name>
    <value>yarn</value>
  </property>
  <property>
    <name>yarn.app.mapreduce.am.env</name>
    <value>HADOOP_MAPRED_HOME=/opt/hadoop</value>
  </property>
  <property>
    <name>mapreduce.map.env</name>
    <value>HADOOP_MAPRED_HOME=/opt/hadoop</value>
  </property>
  <property>
    <name>mapreduce.reduce.env</name>
    <value>HADOOP_MAPRED_HOME=/opt/hadoop</value>
  </property>
</configuration>`
}

// generateCapacitySchedulerXML generates capacity-scheduler.xml content
func generateCapacitySchedulerXML() string {
	return `<?xml version="1.0"?>
<configuration>
  <property>
    <name>yarn.scheduler.capacity.maximum-applications</name>
    <value>10000</value>
  </property>
  <property>
    <name>yarn.scheduler.capacity.maximum-am-resource-percent</name>
    <value>0.1</value>
  </property>
  <property>
    <name>yarn.scheduler.capacity.resource-calculator</name>
    <value>org.apache.hadoop.yarn.util.resource.DefaultResourceCalculator</value>
  </property>
  <property>
    <name>yarn.scheduler.capacity.root.queues</name>
    <value>default</value>
  </property>
  <property>
    <name>yarn.scheduler.capacity.root.default.capacity</name>
    <value>100</value>
  </property>
  <property>
    <name>yarn.scheduler.capacity.root.default.user-limit-factor</name>
    <value>1</value>
  </property>
  <property>
    <name>yarn.scheduler.capacity.root.default.maximum-capacity</name>
    <value>100</value>
  </property>
  <property>
    <name>yarn.scheduler.capacity.root.default.state</name>
    <value>RUNNING</value>
  </property>
  <property>
    <name>yarn.scheduler.capacity.root.default.acl_submit_applications</name>
    <value>*</value>
  </property>
  <property>
    <name>yarn.scheduler.capacity.root.default.acl_administer_queue</name>
    <value>*</value>
  </property>
  <property>
    <name>yarn.scheduler.capacity.node-locality-delay</name>
    <value>40</value>
  </property>
  <property>
    <name>yarn.scheduler.capacity.root.default.acl_application_max_priority</name>
    <value>*</value>
  </property>
  <property>
    <name>yarn.scheduler.capacity.root.default.maximum-application-lifetime</name>
    <value>-1</value>
  </property>
  <property>
    <name>yarn.scheduler.capacity.rack-locality-additional-delay</name>
    <value>-1</value>
  </property>
</configuration>`
}
