# 说明

Docker环境部署Apache Hive及配套数据库。



# 参考

```bash
https://hub.docker.com/r/apache/hive
https://hive.apache.org/development/quickstart/
```



# 部署文件说明

Apache Hive的metastore通过对接数据库实现数据持久化存储，支持PostgreSQL，Oracle，MySQL，MsSQL。



| 文件 | 说明                                   |
| ---- | -------------------------------------- |
| v1   | 对接MySQL8.4.9。对接Hadoop3.1.1。      |
| v2   | 对接postgresql 18.4。对接Hadoop3.1.1。 |



# 配置文件管理

hive组件的配置文件可单独管理。将配置文件存放在本地目录`/opt/hive/conf`，将所有配置文件都存入该目录，挂载至容器内，并将启动项指向该目录，例如在`docker run`命令追加

```bash
-v /opt/hive/conf:/hive_custom_conf \
--env HIVE_CUSTOM_CONF_DIR=/hive_custom_conf
```

或在`compose.yaml`中解除配置文件的注释。

# 部署

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

默认已配置对接Hadoop，可在conf里做修改，对接其他Hadoop。

# 部署后检查

容器健康度检查。确认均显示healthy。

## 测试访问Hadoop

测试hive-metastore访问Hadoop

```bash
docker exec -it hive-metastore bash
hdfs dfs -ls /
hdfs dfs -mkdir -p /user/hive/warehouse
hdfs dfs -ls /
echo "hive test hadoop" > test.txt
hdfs dfs -put test.txt /tmp/
hdfs dfs -put test.txt /user/hive/warehouse
hdfs dfs -ls /tmp/test.txt
hdfs dfs -cat /tmp/test.txt
hdfs dfs -rm /tmp/test.txt
```



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



