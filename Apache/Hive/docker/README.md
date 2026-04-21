# 更新计划

后续改造为配置文件都放conf目录，挂载到容器内。



# 说明

容器环境部署Apache Hive及配套数据库。以docker为例。



# 参考

>https://hub.docker.com/r/apache/hive
https://hive.apache.org/development/quickstart/

# 部署文件说明

Apache Hive的metastore通过对接数据库实现数据持久化存储，支持PostgreSQL，Oracle，MySQL，MsSQL。

跟进业务需求选择需对接的数据库，可自定义数据库配置，默认采用常见生产环境单实例数据库标准配置。

| 文件                    | 说明                             |
| ----------------------- | -------------------------------- |
| compose-PostgreSQL.yaml | 对接PostgreSQL数据库18及更高版本 |
| MySQL                   | 对接MySQL数据库9及更高版本       |



# 配置文件管理

hive组件的配置文件可单独管理。将配置文件存放在本地目录`/opt/hive/conf`，将所有配置文件都存入该目录，挂载至容器内，并将启动项指向该目录，例如在`docker run`命令追加

```bash
-v /opt/hive/conf:/hive_custom_conf \
--env HIVE_CUSTOM_CONF_DIR=/hive_custom_conf
```

或在`compose.yaml`中解除配置文件的注释。

# 部署

复制模板文件，检查配置，删除注释，确认可在部署环境使用。

```bash
cp compose-$(数据库名字).yaml compose.yaml
```

部署所有组件

```bash
docker compose up -d
```



# 卸载

不删数据

```bash
docker compose down
```

删除所有

```bash
docker compose down -v
```



# 对接Hadoop

对接Hadoop环境。



# 访问测试

确认各组件是否正常。

容器状态健康。

查看容器日志最新内容无异常信息。

访问使用正常。

## 数据库

案例使用的postgreSQL，容器状态健康即正常。

## meta store

容器状态健康即正常。

## hive server2 组件

容器健康检查正常。

浏览器访问`http://IP:10002/`

对容器内增强测试，

```bash
docker exec -it hiveserver2 beeline -u 'jdbc:hive2://hiveserver2:10000/'
# 注意：若出现异常，可切换引擎。 
set hive.execution.engine=tez;

show tables;
create table hive_example(a string, b int) partitioned by(c int);
alter table hive_example add partition(c=1);
insert into hive_example partition(c=1) values('a', 1), ('a', 2),('b',3);
select count(distinct a) from hive_example;
select sum(b) from hive_example;
```



