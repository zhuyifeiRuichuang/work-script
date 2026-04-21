"""Generate Hadoop XML configuration for a given cluster FQDN layout."""

from __future__ import annotations


def build_zookeeper_zoo_cfg(*, cluster: str, replicas: int) -> str:
    """Static zoo.cfg for official zookeeper image + start-foreground /conf/zoo.cfg."""
    svc = f"{cluster}-zookeeper"
    lines = [
        "tickTime=2000",
        "dataDir=/data",
        "clientPort=2181",
        "initLimit=10",
        "syncLimit=5",
        "4lw.commands.whitelist=srvr,ruok,mntr",
        "maxClientCnxns=256",
    ]
    for i in range(replicas):
        lines.append(
            f"server.{i + 1}={cluster}-zookeeper-{i}.{svc}:2888:3888:participant"
        )
    return "\n".join(lines) + "\n"


def build_core_site(fs_default: str) -> str:
    return f"""<?xml version="1.0"?>
<configuration>
  <property>
    <name>hadoop.tmp.dir</name>
    <value>/opt/hadoop/tmpdata</value>
  </property>
  <property>
    <name>fs.defaultFS</name>
    <value>{fs_default}</value>
  </property>
  <property>
    <name>hadoop.proxyuser.hadoop.hosts</name>
    <value>*</value>
  </property>
  <property>
    <name>hadoop.proxyuser.hadoop.groups</name>
    <value>*</value>
  </property>
</configuration>
"""


def build_core_site_ha(fs_default: str, ha_zookeeper_quorum: str) -> str:
    return f"""<?xml version="1.0"?>
<configuration>
  <property>
    <name>hadoop.tmp.dir</name>
    <value>/opt/hadoop/tmpdata</value>
  </property>
  <property>
    <name>fs.defaultFS</name>
    <value>{fs_default}</value>
  </property>
  <property>
    <name>ha.zookeeper.quorum</name>
    <value>{ha_zookeeper_quorum}</value>
  </property>
  <property>
    <name>hadoop.proxyuser.hadoop.hosts</name>
    <value>*</value>
  </property>
  <property>
    <name>hadoop.proxyuser.hadoop.groups</name>
    <value>*</value>
  </property>
</configuration>
"""


def build_hdfs_site(
    namenode_rpc: str,
    replication: int,
) -> str:
    return f"""<?xml version="1.0"?>
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
    <value>{namenode_rpc}</value>
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
    <value>{replication}</value>
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
</configuration>
"""


def build_hdfs_site_ha(
    *,
    nameservice: str,
    nn1_rpc: str,
    nn2_rpc: str,
    nn1_http: str,
    nn2_http: str,
    qjournal_uri: str,
    replication: int,
) -> str:
    """HDFS HA with QJM + automatic failover (ZKFC). Fencing uses shell(/bin/true) for Kubernetes."""
    return f"""<?xml version="1.0"?>
<configuration>
  <property>
    <name>dfs.nameservices</name>
    <value>{nameservice}</value>
  </property>
  <property>
    <name>dfs.ha.namenodes.{nameservice}</name>
    <value>nn1,nn2</value>
  </property>
  <property>
    <name>dfs.namenode.rpc-address.{nameservice}.nn1</name>
    <value>{nn1_rpc}</value>
  </property>
  <property>
    <name>dfs.namenode.rpc-address.{nameservice}.nn2</name>
    <value>{nn2_rpc}</value>
  </property>
  <property>
    <name>dfs.namenode.http-address.{nameservice}.nn1</name>
    <value>{nn1_http}</value>
  </property>
  <property>
    <name>dfs.namenode.http-address.{nameservice}.nn2</name>
    <value>{nn2_http}</value>
  </property>
  <property>
    <name>dfs.namenode.shared.edits.dir</name>
    <value>{qjournal_uri}</value>
  </property>
  <property>
    <name>dfs.journalnode.edits.dir</name>
    <value>/opt/hadoop/data/jn</value>
  </property>
  <property>
    <name>dfs.journalnode.rpc-address</name>
    <value>0.0.0.0:8485</value>
  </property>
  <property>
    <name>dfs.journalnode.http-address</name>
    <value>0.0.0.0:8480</value>
  </property>
  <property>
    <name>dfs.client.failover.proxy.provider.{nameservice}</name>
    <value>org.apache.hadoop.hdfs.server.namenode.ha.ConfiguredFailoverProxyProvider</value>
  </property>
  <property>
    <name>dfs.ha.fencing.methods</name>
    <value>shell(/bin/true)</value>
  </property>
  <property>
    <name>dfs.ha.automatic-failover.enabled</name>
    <value>true</value>
  </property>
  <property>
    <name>dfs.namenode.name.dir</name>
    <value>/opt/hadoop/data/nn</value>
  </property>
  <property>
    <name>dfs.datanode.data.dir</name>
    <value>/opt/hadoop/data/dn</value>
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
    <value>{replication}</value>
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
</configuration>
"""


def build_yarn_site(rm_hostname: str) -> str:
    return f"""<?xml version="1.0"?>
<configuration>
  <property>
    <name>yarn.resourcemanager.hostname</name>
    <value>{rm_hostname}</value>
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
</configuration>
"""


def build_mapred_site() -> str:
    return """<?xml version="1.0"?>
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
</configuration>
"""


def build_capacity_scheduler() -> str:
    return """<?xml version="1.0"?>
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
</configuration>
"""
