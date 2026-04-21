# 说明

定制容器镜像。

# 基于官方原版构建

适用于仅添加驱动文件的场景。例如将新增驱动文件放入本地目录driver，复制到容器中。

```bash
FROM apache/seatunnel:2.3.13

WORKDIR /opt/seatunnel
# add driver
COPY driver/* /opt/seatunnel/plugins/
COPY driver/* /opt/seatunnel/lib/
```



# Docker使用

适用于定制开发场景。

适用于Docker环境。

可添加自定义文件，例如将新增驱动文件放入本地目录driver，复制到容器中。

参考`https://seatunnel.apache.org/docs/2.3.13/getting-started/docker/`

```bash
FROM seatunnelhub/openjdk:8u342

ARG VERSION
# Build from Source Code And Copy it into image
COPY ./target/apache-seatunnel-${VERSION}-bin.tar.gz /opt/

# Download From Internet
# Please Note this file only include fake/console connector, You'll need to download the other connectors manually
# wget -P /opt https://dlcdn.apache.org/seatunnel/${VERSION}/apache-seatunnel-${VERSION}-bin.tar.gz

RUN cd /opt && \
    tar -zxvf apache-seatunnel-${VERSION}-bin.tar.gz && \
    mv apache-seatunnel-${VERSION} seatunnel && \
    rm apache-seatunnel-${VERSION}-bin.tar.gz && \
    sed -i 's/#rootLogger.appenderRef.consoleStdout.ref/rootLogger.appenderRef.consoleStdout.ref/' seatunnel/config/log4j2.properties && \
    sed -i 's/#rootLogger.appenderRef.consoleStderr.ref/rootLogger.appenderRef.consoleStderr.ref/' seatunnel/config/log4j2.properties && \
    sed -i 's/rootLogger.appenderRef.file.ref/#rootLogger.appenderRef.file.ref/' seatunnel/config/log4j2.properties && \    
    cp seatunnel/config/hazelcast-master.yaml seatunnel/config/hazelcast-worker.yaml

WORKDIR /opt/seatunnel

# add driver
COPY driver/* /opt/seatunnel/plugins/
COPY driver/* /opt/seatunnel/lib/
```



# k8s环境使用

适用于定制开发场景。

适用于k8s环境。

可添加自定义文件，例如将新增驱动文件放入本地目录driver，复制到容器中。

参考`https://seatunnel.apache.org/docs/2.3.13/getting-started/kubernetes/`

```bash
FROM seatunnelhub/openjdk:8u342

ENV SEATUNNEL_VERSION="2.3.13"
ENV SEATUNNEL_HOME="/opt/seatunnel"

RUN wget https://dlcdn.apache.org/seatunnel/${SEATUNNEL_VERSION}/apache-seatunnel-${SEATUNNEL_VERSION}-bin.tar.gz
RUN tar -xzvf apache-seatunnel-${SEATUNNEL_VERSION}-bin.tar.gz
RUN mv apache-seatunnel-${SEATUNNEL_VERSION} ${SEATUNNEL_HOME}
RUN mkdir -p $SEATUNNEL_HOME/logs
RUN cd ${SEATUNNEL_HOME} && sh bin/install-plugin.sh ${SEATUNNEL_VERSION}
WORKDIR /opt/seatunnel

# add driver
COPY driver/* /opt/seatunnel/plugins/
COPY driver/* /opt/seatunnel/lib/
```


