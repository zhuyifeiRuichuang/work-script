# Hadoop Docker Image for Kubernetes
# Base image: Eclipse Temurin OpenJDK JRE
FROM eclipse-temurin:21-jre-alpine AS base

# Install required packages
RUN apk add --no-cache \
    bash \
    curl \
    netcat-openbsd \
    coreutils \
    su-exec \
    tini

# Create Hadoop user and groups
RUN addgroup -g 1000 hadoop && \
    adduser -u 1000 -G hadoop -s /bin/bash -D hadoop

# Set working directory
WORKDIR /opt

# Download and install Hadoop
ENV HADOOP_VERSION=3.4.0
ENV HADOOP_URL=https://dlcdn.apache.org/hadoop/common/hadoop-${HADOOP_VERSION}/hadoop-${HADOOP_VERSION}.tar.gz
ENV HADOOP_HOME=/opt/hadoop
ENV HADOOP_CONF_DIR=${HADOOP_HOME}/etc/hadoop
ENV PATH=${HADOOP_HOME}/bin:${HADOOP_HOME}/sbin:${PATH}
ENV HADOOP_LOG_DIR=${HADOOP_HOME}/logs
ENV HADOOP_CACHE_DIR=/tmp/hadoop-cache

RUN apk add --no-cache wget tar gzip && \
    mkdir -p ${HADOOP_HOME} ${HADOOP_LOG_DIR} ${HADOOP_CACHE_DIR} && \
    wget -q -O /tmp/hadoop.tar.gz ${HADOOP_URL} && \
    tar -xzf /tmp/hadoop.tar.gz -C /opt && \
    mv /opt/hadoop-${HADOOP_VERSION}/* ${HADOOP_HOME}/ && \
    rm -rf /opt/hadoop-${HADOOP_VERSION} /tmp/hadoop.tar.gz && \
    chown -R hadoop:hadoop ${HADOOP_HOME} ${HADOOP_LOG_DIR} ${HADOOP_CACHE_DIR}

# Set JVM options
ENV HADOOP_HEAPSIZE_MAX=2048m
ENV HADOOP_HEAPSIZE_MIN=1024m
ENV JAVA_HOME=/opt/java/openjdk

# Create necessary directories with proper permissions
RUN mkdir -p ${HADOOP_HOME}/tmpdata ${HADOOP_HOME}/data/nn ${HADOOP_HOME}/data/dn && \
    chown -R hadoop:hadoop ${HADOOP_HOME}

# Copy configuration files (will be overridden by ConfigMap in Kubernetes)
# These are default configurations
COPY hadoop-env.sh ${HADOOP_CONF_DIR}/hadoop-env.sh
COPY core-site.xml.template ${HADOOP_CONF_DIR}/core-site.xml.template
COPY hdfs-site.xml.template ${HADOOP_CONF_DIR}/hdfs-site.xml.template
COPY yarn-site.xml.template ${HADOOP_CONF_DIR}/yarn-site.xml.template
COPY mapred-site.xml.template ${HADOOP_CONF_DIR}/mapred-site.xml.template

# Create entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Set ownership
RUN chown -R hadoop:hadoop ${HADOOP_HOME}

# Switch to hadoop user
USER hadoop

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD nc -z localhost 9870 || exit 1

# Expose ports
# HDFS: 9870 (NameNode Web UI), 9000 (RPC), 9864 (DataNode Web UI), 9866 (DataNode IPC), 9867 (DataNode)
# YARN: 8088 (ResourceManager Web UI), 8032 (ResourceManager RPC), 8030 (ResourceManager Admin), 8042 (NodeManager Web UI)
EXPOSE 9000 9870 9864 9866 9867 8088 8030 8032 8042

# Set entrypoint
ENTRYPOINT ["/entrypoint.sh"]
