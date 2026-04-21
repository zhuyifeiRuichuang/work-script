# 说明

docker环境部署并测试Seatunnel集群模式。

参考`https://seatunnel.apache.org/docs/2.3.13/getting-started/docker/#%E4%BD%BF%E7%94%A8docker-compose`



# 部署

```bash
docker compose up -d
```



# 测试

验证节点状态

```bash
docker exec -it seatunnel_master /opt/seatunnel/bin/seatunnel.sh cluster list
```



提交job

本地模式

```bash
docker run --name seatunnel_client \
    --network seatunnel-network \
    -e ST_DOCKER_MEMBER_LIST=seatunnel_master:5801 \
    --rm \
    apache/seatunnel:2.3.13 \
    ./bin/seatunnel.sh --config config/v2.batch.config.template -m local
```



集群模式

```bash
docker run --name seatunnel_client \
    --network seatunnel-network \
    -e ST_DOCKER_MEMBER_LIST=seatunnel_master:5801 \
    --rm \
    apache/seatunnel:2.3.13 \
    ./bin/seatunnel.sh --config config/v2.batch.config.template -m cluster
```



查询job

```bash
docker run --name seatunnel_client \
    --network seatunnel-network \
    -e ST_DOCKER_MEMBER_LIST=seatunnel_master:5801 \
    --rm \
    apache/seatunnel \
    ./bin/seatunnel.sh  -l
```

