# 说明

容器化部署管理Apache Kafka，数据持久化到数据卷，有可靠的健康检查。支持离线部署，可灵活改造为k8s环境部署yaml。

容器镜像`https://hub.docker.com/r/apache/kafka`

容器部署参考`https://github.com/apache/kafka/blob/trunk/docker/examples/README.md`

# 配置

环境变量生效优先级高于配置文件挂载到容器内。

# 单节点

`single-node`

# 多节点集群

`cluster`



# 特殊配置

跨主机访问kafka时，需修改`compose.yaml`中的`KAFKA_ADVERTISED_LISTENERS`的 localhost改为真实的主机IP。这个位置的值目前无法做到自动适配真实主机IP。
