#!/usr/bin/env bash

# Hadoop Environment Configuration
# This file is sourced by all Hadoop shell scripts

# Set Java home
export JAVA_HOME=${JAVA_HOME:-"/opt/java/openjdk"}

# Heap settings
export HADOOP_HEAPSIZE_MAX=${HADOOP_HEAPSIZE_MAX:-"2048"}
export HADOOP_HEAPSIZE_MIN=${HADOOP_HEAPSIZE_MIN:-"1024"}

# Hadoop JVM options
export HADOOP_OPTS="${HADOOP_OPTS} -Djava.net.preferIPv4Stack=true"
export HADOOP_OPTS="${HADOOP_OPTS} -Djava.security.egd=file:/dev/./urandom"

# NameNode specific options
export HDFS_NAMENODE_OPTS="-Xmx1024m -Xms512m -verbose:gc -XX:+PrintGCDetails -XX:+PrintGCDateStamps"
export HDFS_DATANODE_OPTS="-Xmx1024m -Xms512m"

# YARN specific options
export YARN_RESOURCEMANAGER_OPTS="-Xmx1024m -Xms512m"
export YARN_NODEMANAGER_OPTS="-Xmx512m -Xms256m"

# Remote debugging (uncomment to enable)
# export HADOOP_OPTS="${HADOOP_OPTS} -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=5005"

# Native libraries
export HADOOP_COMMON_LIB_NATIVE_DIR=${HADOOP_HOME}/lib/native
export HADOOP_OPTS="${HADOOP_OPTS} -Djava.library.path=${HADOOP_HOME}/lib/native"

# Log settings
export HADOOP_LOG_DIR=${HADOOP_LOG_DIR:-${HADOOP_HOME}/logs}
export HADOOP_PID_DIR=${HADOOP_PID_DIR:-/var/run/hadoop}
export HADOOP_IDENT_STRING=${HADOOP_IDENT_STRING:-hadoop}

# Create log and pid directories
mkdir -p ${HADOOP_LOG_DIR} ${HADOOP_PID_DIR}
chown -R hadoop:hadoop ${HADOOP_LOG_DIR} ${HADOOP_PID_DIR}

# User for impersonation
export HADOOP_PROXY_USER=${HADOOP_PROXY_USER:-hadoop}

# Timezone
export TZ=Asia/Shanghai
