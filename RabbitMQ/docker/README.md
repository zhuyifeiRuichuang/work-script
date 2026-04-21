# 说明

容器化部署，推荐最新稳定版本。

参考`https://hub.docker.com/_/rabbitmq/`



`rabbitmq.conf`来自`https://github.com/rabbitmq/rabbitmq-server/blob/main/deps/rabbit/docs/rabbitmq.conf.example`



docker的配置文件`/etc/docker/daemon.json`需追加配置

```bash
"default-ulimits": {
    "nofile": {
      "Name": "nofile",
      "Hard": 64000,
      "Soft": 64000
    }
  }
```



# 访问

浏览器访问`IP:15672`