# 说明

生产环境，docker部署elasticsearch最佳实践。集群架构部署。

参考`https://www.elastic.co/docs/deploy-manage/deploy/self-managed/install-elasticsearch-with-docker`

# 文件说明

`.env`来自`https://github.com/elastic/elasticsearch/blob/main/docs/reference/setup/install/docker/.env`

```bash
# 密码必须的数字字母组合，最少6个
ELASTIC_PASSWORD=changeme

# 密码必须的数字字母组合，最少6个
KIBANA_PASSWORD=changeme

# 推荐写最新的版本，可在docker hub查询
STACK_VERSION=9.3.2

# Set the cluster name
CLUSTER_NAME=docker-cluster

# Set to 'basic' or 'trial' to automatically start the 30-day trial
LICENSE=basic
#LICENSE=trial

# Port to expose Elasticsearch HTTP API to the host
ES_PORT=9200
#ES_PORT=127.0.0.1:9200

# Port to expose Kibana to the host
KIBANA_PORT=5601
#KIBANA_PORT=80

# Increase or decrease based on the available host memory (in bytes)
MEM_LIMIT=1073741824

# Project namespace (defaults to the current folder name if not set)
#COMPOSE_PROJECT_NAME=myproject
```



`docker-compose.yml`来自`https://github.com/elastic/elasticsearch/blob/main/docs/reference/setup/install/docker/docker-compose.yml`

# 部署

```bash
docker compose up -d
```

# 访问

浏览器访问`IP:5601`，默认账户`elastic`，默认密码是`.env`的`KIBANA_PASSWORD`的值。

# 进入容器

部署后，可进入容器查看和调整配置

```bash
docker exec -it docker-es01-1 bash
```

配置文件在容器内的`/usr/share/elasticsearch/config`

存储数据在容器内的`/usr/share/elasticsearch/data`

# 采集数据

参考：`https://www.elastic.co/docs/reference/ingestion-tools`

目录结构`https://www.elastic.co/docs/reference/beats/filebeat/directory-layout`

