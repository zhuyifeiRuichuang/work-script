# 说明
下载所需软件到当前目录。应明确要使用的hive版本，确认软件兼容性合格。

# 查询软件版本兼容性
Hive与Hadoop：[Hive官方文档](https://hive.apache.org/general/downloads/)

Hive与tez：无

Hadoop与tez：[tez官方文档](https://tez.apache.org/install.html)

Hive与iceberg：[Hive官网文档 ](https://iceberg.apache.org/docs/nightly/hive/#feature-support)  [seatunnel官网文档](https://seatunnel.apache.org/zh-CN/docs/2.3.12/connector-v2/sink/Iceberg/)

# 下载软件
应下载到当前目录。
## hive-必选
下载hive软件包存到当前目录，文件格式为`apache-hive-版本号-bin.tar.gz`。    
当前目录只能存放一个发行版软件包，若是定制开发，则必须打包文件格式为`apache-hive-版本号-bin.tar.gz`
Hive：[版本清单](https://archive.apache.org/dist/hive/)  

## Hadoop-必选
下载兼容性范围的Hadoop版本软件包到当前目录，文件格式为`hadoop-版本号.tar.gz`。  
当前目录只能存放一个发行版软件包，若是定制开发，则必须打包文件格式为`hadoop-版本号.tar.gz`  
Hadoop：[版本清单](https://archive.apache.org/dist/hadoop/common/)  

## tez-可选
下载到当前目录，文件格式为`apache-tez-版本号-bin.tar.gz`。
tez：[版本清单1](https://tez.apache.org/releases/index.html)，[版本清单2](https://dlcdn.apache.org/tez/)  

## iceberg-可选，推荐
下载到当前目录，文件格式为`iceberg-hive-runtime-版本号.jar`。

> 打包低于Hive4的版本时需要下载`iecberg-hive-runtime`。
>
> `libfb303 默认不下载，仅特殊低版本或特殊故障时需下载。

iceberg下载地址：[版本清单](https://repo1.maven.org/maven2/org/apache/iceberg/iceberg-hive-runtime/)  

libfb303下载地址：[版本清单](https://repo1.maven.org/maven2/org/apache/thrift/libfb303/) ，[备用下载地址](https://mvnrepository.com/artifact/org.apache.thrift/libfb303)  

## JDBC-可选，推荐
用于hive连接数据库。文件格式大多是`connector.jar` ，根据hive连接的数据库选择，推荐下载最新版本。
### MySQL
[版本清单](https://downloads.mysql.com/archives/c-j/)  
选择版本时，`Operating System`选择`Platform Independent`，`Product Version`选择期望使用的版本，推荐最新版本。将下载后的文件解压，仅取出名字格式为`mysql-connector-j-*.jar`的文件存放在本目录。
### Oracle
[版本清单](https://www.oracle.com/database/technologies/appdev/jdbc-downloads.html)  
下载到本目录。

### PostgreSQL
[版本清单](https://jdbc.postgresql.org/download/)  
下载到本目录。
