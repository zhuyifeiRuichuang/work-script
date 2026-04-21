# 若自定义基础镜像，请替换
FROM apache/hadoop-runner
WORKDIR /opt

# 复制本地任意版本的hadoop tar.gz到容器内（需保证本地仅一个hadoop-*.tar.gz文件）
COPY hadoop-*.tar.gz /opt/hadoop.tar.gz

RUN sudo rm -rf /opt/hadoop \
    && tar zxf hadoop.tar.gz \
    && rm hadoop.tar.gz \
    && mv hadoop* hadoop \
    && rm -rf /opt/hadoop/share/doc

WORKDIR /opt/hadoop
ADD log4j.properties /opt/hadoop/etc/hadoop/log4j.properties
RUN sudo chown -R hadoop:users /opt/hadoop/etc/hadoop/*
ENV HADOOP_CONF_DIR=/opt/hadoop/etc/hadoop
