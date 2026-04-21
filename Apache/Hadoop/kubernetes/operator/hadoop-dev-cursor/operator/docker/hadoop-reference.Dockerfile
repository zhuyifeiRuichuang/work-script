# 参考用 Hadoop 镜像（与现用镜像无关时可自建）。多阶段 + 非 root 运行可选。
# 构建示例（在 operator/docker 目录）:
#   docker build -f hadoop-reference.Dockerfile -t your-registry/hadoop:3.3.6-ref .
#
# 说明：生产请固定 digest、在 CI 中校验 tarball 校验和，并只从可信源获取发行包。
ARG JAVA_IMAGE=eclipse-temurin:11-jdk-jammy
ARG HADOOP_VERSION=3.3.6
# 留空则跳过校验；生产请填写官方页公布的 sha512 全文。
ARG HADOOP_SHA512=

FROM ${JAVA_IMAGE} AS base
ARG HADOOP_VERSION
ARG HADOOP_SHA512
ENV DEBIAN_FRONTEND=noninteractive \
    HADOOP_VERSION=${HADOOP_VERSION} \
    HADOOP_HOME=/opt/hadoop \
    PATH=/opt/hadoop/bin:/opt/hadoop/sbin:${PATH}

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates curl bash tini netbase \
    && rm -rf /var/lib/apt/lists/*

RUN set -eux; \
    curl -fsSL "https://archive.apache.org/dist/hadoop/common/hadoop-${HADOOP_VERSION}/hadoop-${HADOOP_VERSION}.tar.gz" -o /tmp/hadoop.tgz; \
    if [ -n "${HADOOP_SHA512}" ]; then echo "${HADOOP_SHA512}  /tmp/hadoop.tgz" | sha512sum -c -; fi; \
    tar -xzf /tmp/hadoop.tgz -C /opt; \
    ln -s "/opt/hadoop-${HADOOP_VERSION}" /opt/hadoop; \
    rm /tmp/hadoop.tgz

RUN groupadd --system hadoop && useradd --system --gid hadoop --home-dir /opt/hadoop --shell /bin/bash hadoop \
    && mkdir -p /opt/hadoop/logs /opt/hadoop/data \
    && chown -R hadoop:hadoop /opt/hadoop

USER hadoop
WORKDIR /opt/hadoop
ENTRYPOINT ["/usr/bin/tini", "--"]
