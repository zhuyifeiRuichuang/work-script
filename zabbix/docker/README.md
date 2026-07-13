# 说明

Docker环境部署zabbix

参考

```bash
https://github.com/zabbix/zabbix-docker 
https://www.zabbix.com/container_images
https://www.zabbix.com/documentation/current/en/manual/installation/containers
```

# 配置

参考`https://www.zabbix.com/documentation/current/zh/manual/installation/containers`调整各环境变量和配置文件，例如

`https://github.com/zabbix/zabbix-docker/blob/7.4/env_vars/.env_web`可配置`PHP_TZ=Asia/Shanghai`，实现中国地区时间显示，默认是欧洲时间。

# 部署

```bash
git clone https://github.com/zabbix/zabbix-docker.git
cd zabbix-docker
git checkout 7.4
# 使用 MySQL 作为数据库：
docker compose -f ./compose.yaml up -d

# 使用 PostgreSQL 作为数据库：
docker compose -f ./compose_pgsql.yaml up -d
```



# 访问

首次访问web界面，IP是宿主机IP，账户`Admin`，密码`zabbix`