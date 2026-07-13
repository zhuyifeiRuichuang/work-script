# 说明

在Docker环境部署guacamole+mysql，将此目录内容全部上传。

# 配置MySQL

获取最新的sql文件 ，注意此处以1.6.0为例。

```bash
docker run --rm guacamole/guacamole:1.6.0 /opt/guacamole/bin/initdb.sh --mysql > initdb.sql
```

根据业务需求调整my.cnf和compose.yaml中MySQL部分



# 配置guacamole

根据业务需求调整compose.yaml中guacamole部分。



# 部署

```bash
docker compose up -d
```



# 访问

物理机IP和容器的8080端口映射的物理机端口

默认账户密码`guacadmin`