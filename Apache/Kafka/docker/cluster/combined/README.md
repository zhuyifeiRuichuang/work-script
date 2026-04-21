# 文件

推荐使用`compose.yml`

# 测试

在宿主机测试集群状态

```bash
# 查询集群状态，确认选举状态。
docker exec -it kafka-1 /opt/kafka/bin/kafka-metadata-quorum.sh --bootstrap-server kafka-1:19092 describe --status
# 查询集群在线Broker
docker exec -it kafka-1 /opt/kafka/bin/kafka-broker-api-versions.sh --bootstrap-server kafka-1:19092
```

