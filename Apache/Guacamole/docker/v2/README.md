# 说明

在Docker环境部署guacamole搭配PostgreSQL

# 配置PostgreSQL

获取sql文件，注意，此处选择了1.6.0版本

```bash
docker run --rm guacamole/guacamole:1.6.0 /opt/guacamole/bin/initdb.sh --postgresql > initdb.sql
```

根据业务需求调整配置。可挂载定制的配置文件到容器内。

# 配置guacamole

根据业务需求调整配置。

# 部署

```bash
docker compose up -d
```



# 访问

物理机IP和容器的8080端口映射的物理机端口

默认账户密码`guacadmin`