# 说明
基于Apache Hive3软件包，构建容器镜像，必须使用JDK8。适用于中国地区网络。

容器镜像加速配置参考网站`1ms.run`。



# 参考文档

> hive官方文档: https://github.com/apache/hive/tree/master/packaging/src/docker



# 资源说明
`entrypoint.sh`是hive官方提供的标准[内置文件](https://github.com/apache/hive/blob/master/packaging/src/docker/entrypoint.sh)，不要修改。

`software`存放软件包。

`conf`存放hive官方标准配置文件，默认内置配置。

`Dockerfile.ubuntu25` 基于Ubuntu25，JDK8的容器镜像构建配置。

`Dockerfile.ubuntu22`基于Ubuntu22，JDK8的容器镜像构建配置。

## 镜像构建思路

对于Apache Hive3的任意子版本，可参考以下思路打包定制容器镜像。

1. 选择hive版本。将所有所需软件存入`software`目录。  
2. 构建容器镜像。  
3. 测试容器镜像。  
4. 上传容器镜像。测试镜像确认可用，上传至镜像仓库正式使用。  

# 准备软件
详见目录`software`中说明文件。必须准备软件。

## 构建hive镜像
镜像名字推荐采用格式`hive:3.1.2-mysql`，区分hive可对接的数据库及hive版本。

```bash
docker build -t hive:dev1 -f Dockerfile.ubuntu22 .
```
若构建镜像次数过多且频繁测试异常，可使用以下命令，跳过缓存，全新构建，
```bash
docker build --no-cache -t hive:dev1 -f Dockerfile.ubuntu22 .
```

## 测试容器镜像

容器镜像构建后可通过docker快速测试。

启动metastore
```bash
docker run -d -p 9083:9083 --env SERVICE_NAME=metastore --name metastore-standalone hive:dev1
```
启动hive server2
```bash
docker run -d -p 10000:10000 -p 10002:10002 --env SERVICE_NAME=hiveserver2 --name hiveserver2 hive:dev1
```
查询容器状态和日志，状态需UP，日志需无erro。
```bash
docker ps -a
docker logs metastore-standalone | grep erro
docker logs hiveserver2 | grep erro
```
若出现类似以下报错，说明镜像构建失败。

```bash
2025-11-26T06:03:40,139  WARN [main] server.HiveServer2: Error starting HiveServer2 on attempt 1, will retry in 60000ms
java.lang.RuntimeException: Error applying authorization policy on hive configuration: class jdk.internal.loader.ClassLoaders$AppClassLoader cannot be cast to class java.net.URLClassLoader (jdk.internal.loader.ClassLoaders$AppClassLoader and java.net.URLClassLoader are in module java.base of loader 'bootstrap')
	at org.apache.hive.service.cli.CLIService.init(CLIService.java:118) ~[hive-service-3.1.2.jar:3.1.2]
	at org.apache.hive.service.CompositeService.init(CompositeService.java:59) ~[hive-service-3.1.2.jar:3.1.2]
	at org.apache.hive.service.server.HiveServer2.init(HiveServer2.java:230) ~[hive-service-3.1.2.jar:3.1.2]
	at org.apache.hive.service.server.HiveServer2.startHiveServer2(HiveServer2.java:1036) [hive-service-3.1.2.jar:3.1.2]
	at org.apache.hive.service.server.HiveServer2.access$1600(HiveServer2.java:140) [hive-service-3.1.2.jar:3.1.2]
	at org.apache.hive.service.server.HiveServer2$StartOptionExecutor.execute(HiveServer2.java:1305) [hive-service-3.1.2.jar:3.1.2]
	at org.apache.hive.service.server.HiveServer2.main(HiveServer2.java:1149) [hive-service-3.1.2.jar:3.1.2]
	at jdk.internal.reflect.DirectMethodHandleAccessor.invoke(Unknown Source) ~[?:?]
	at java.lang.reflect.Method.invoke(Unknown Source) ~[?:?]
	at org.apache.hadoop.util.RunJar.run(RunJar.java:318) [hadoop-common-3.1.1.jar:?]
	at org.apache.hadoop.util.RunJar.main(RunJar.java:232) [hadoop-common-3.1.1.jar:?]
Caused by: java.lang.ClassCastException: class jdk.internal.loader.ClassLoaders$AppClassLoader cannot be cast to class java.net.URLClassLoader (jdk.internal.loader.ClassLoaders$AppClassLoader and java.net.URLClassLoader are in module java.base of loader 'bootstrap')
	at org.apache.hadoop.hive.ql.session.SessionState.<init>(SessionState.java:413) ~[hive-exec-3.1.2.jar:3.1.2]
	at org.apache.hadoop.hive.ql.session.SessionState.<init>(SessionState.java:389) ~[hive-exec-3.1.2.jar:3.1.2]
	at org.apache.hive.service.cli.CLIService.applyAuthorizationConfigPolicy(CLIService.java:128) ~[hive-service-3.1.2.jar:3.1.2]
	at org.apache.hive.service.cli.CLIService.init(CLIService.java:115) ~[hive-service-3.1.2.jar:3.1.2]
	... 10 more

```



# k8s环境使用

可能会用到以下配置，
```bash
securityContext:
          runAsUser: 1000
```



