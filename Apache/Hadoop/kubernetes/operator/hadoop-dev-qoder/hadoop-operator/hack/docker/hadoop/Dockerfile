# Copyright 2024 Apache Software Foundation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Apache Hadoop 集群组件基础镜像
# 支持 NameNode, DataNode, ResourceManager, NodeManager, JournalNode
# 基于 Hadoop 3.3.6 官方发行版

# ============================================
# 阶段 1: 下载和准备阶段
# ============================================
FROM alpine:3.19 AS downloader

# 构建参数
ARG HADOOP_VERSION=3.3.6
ARG HADOOP_SHA512_URL=https://downloads.apache.org/hadoop/common/hadoop-${HADOOP_VERSION}/hadoop-${HADOOP_VERSION}.tar.gz.sha512
ARG HADOOP_DOWNLOAD_URL=https://archive.apache.org/dist/hadoop/common/hadoop-${HADOOP_VERSION}/hadoop-${HADOOP_VERSION}.tar.gz

# 安装下载工具
RUN apk add --no-cache curl gnupg tar gzip

WORKDIR /tmp

# 下载 Hadoop 并验证校验和
RUN echo "Downloading Hadoop ${HADOOP_VERSION}..." && \
    curl -fsSL -o hadoop.tar.gz "${HADOOP_DOWNLOAD_URL}" && \
    curl -fsSL -o hadoop.tar.gz.sha512 "${HADOOP_SHA512_URL}" || true && \
    echo "Verifying checksum..." && \
    if [ -f hadoop.tar.gz.sha512 ]; then \
        sha512sum -c hadoop.tar.gz.sha512 || exit 1; \
    else \
        echo "Warning: SHA512 file not available, skipping verification"; \
    fi && \
    echo "Extracting Hadoop..." && \
    tar -xzf hadoop.tar.gz && \
    mv hadoop-${HADOOP_VERSION} hadoop && \
    rm -f hadoop.tar.gz hadoop.tar.gz.sha512

# 清理不必要的文件以减少镜像体积
RUN rm -rf /tmp/hadoop/share/doc \
    /tmp/hadoop/share/hadoop/common/jdiff \
    /tmp/hadoop/share/hadoop/common/sources \
    /tmp/hadoop/share/hadoop/hdfs/jdiff \
    /tmp/hadoop/share/hadoop/hdfs/sources \
    /tmp/hadoop/share/hadoop/mapreduce/jdiff \
    /tmp/hadoop/share/hadoop/mapreduce/sources \
    /tmp/hadoop/share/hadoop/tools/sources \
    /tmp/hadoop/share/hadoop/yarn/jdiff \
    /tmp/hadoop/share/hadoop/yarn/sources

# ============================================
# 阶段 2: 运行时基础镜像
# ============================================
FROM eclipse-temurin:11-jre-alpine

# 维护者信息
LABEL maintainer="Apache Hadoop Operator Team"
LABEL org.opencontainers.image.title="Apache Hadoop"
LABEL org.opencontainers.image.description="Apache Hadoop cluster components for Kubernetes"
LABEL org.opencontainers.image.source="https://github.com/apache/hadoop-operator"

# 构建参数
ARG HADOOP_VERSION=3.3.6
ARG HADOOP_USER=hadoop
ARG HADOOP_UID=1000
ARG HADOOP_GROUP=hadoop
ARG HADOOP_GID=1000

# 环境变量
ENV HADOOP_VERSION=${HADOOP_VERSION}
ENV HADOOP_HOME=/opt/hadoop
ENV HADOOP_CONF_DIR=/opt/hadoop/etc/hadoop
ENV HADOOP_LOG_DIR=/opt/hadoop/logs
ENV HADOOP_DATA_DIR=/opt/hadoop/data
ENV PATH=${PATH}:${HADOOP_HOME}/bin:${HADOOP_HOME}/sbin

# 创建 hadoop 用户和组
RUN addgroup -g ${HADOOP_GID} ${HADOOP_GROUP} && \
    adduser -u ${HADOOP_UID} -G ${HADOOP_GROUP} -h /home/${HADOOP_USER} -s /bin/sh -D ${HADOOP_USER}

# 安装必要的运行时依赖
RUN apk add --no-cache \
    bash \
    curl \
    netcat-openbsd \
    procps \
    su-exec \
    tzdata \
    && rm -rf /var/cache/apk/*

# 从下载阶段复制 Hadoop
COPY --from=downloader /tmp/hadoop ${HADOOP_HOME}

# 创建必要的目录并设置权限
RUN mkdir -p ${HADOOP_LOG_DIR} \
    ${HADOOP_DATA_DIR} \
    /tmp/hadoop-yarn \
    /tmp/hadoop-hdfs && \
    chown -R ${HADOOP_USER}:${HADOOP_GROUP} ${HADOOP_HOME} \
    /tmp/hadoop-yarn \
    /tmp/hadoop-hdfs

# 复制配置文件模板
COPY conf/ ${HADOOP_CONF_DIR}/

# 复制启动脚本
COPY scripts/entrypoint.sh /usr/local/bin/
COPY scripts/healthcheck.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/entrypoint.sh /usr/local/bin/healthcheck.sh

# 设置工作目录
WORKDIR ${HADOOP_HOME}

# 暴露常用端口
# HDFS NameNode
EXPOSE 8020 9000 9870 9871
# HDFS DataNode
EXPOSE 9864 9866 9867
# YARN ResourceManager
EXPOSE 8030 8031 8032 8033 8088
# YARN NodeManager
EXPOSE 8040 8041 8042
# MapReduce JobHistory
EXPOSE 10020 19888
# JournalNode
EXPOSE 8485 8480 8481

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD /usr/local/bin/healthcheck.sh

# 入口点
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

# 默认命令（可以被覆盖）
CMD ["hdfs", "namenode"]
