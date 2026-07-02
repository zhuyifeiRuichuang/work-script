# 说明

部署Apache Seata 2.7.0，搭配Nacos3.2.2，MySQL8.4.10

特别提醒，不要将所有组件合并到一个compose里，务必拆分处理，方便替换组件。可使用默认配置直接部署。

# 上传资源

新建目录seata，上传所有配置文件。

# 部署MySQL

配置MySQL，部署之前可修改配置。在目录`mysql`中处理，

## 配置说明

目录`init-db`存放用于初始化环境的sql文件。

`01-init.sql`说明

```bash
# 创建Nacos组件专用的数据库
CREATE DATABASE IF NOT EXISTS `nacos`
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;
# 创建Seata专用的数据库
CREATE DATABASE IF NOT EXISTS `seata-server`
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

# 创建Nacos对接MySQL的专用账户
CREATE USER 'nacos'@'%' IDENTIFIED BY 'nacos';
GRANT ALL PRIVILEGES ON nacos.* TO 'nacos'@'%';
# 创建seata对接MySQL的专用账户
CREATE USER 'seata'@'%' IDENTIFIED BY 'seata';
GRANT ALL PRIVILEGES ON `seata-server`.* TO 'seata'@'%';

# 使上述配置生效
FLUSH PRIVILEGES;
```



`02-mysql-schema.sql`来自`https://github.com/alibaba/nacos/blob/master/distribution/conf/mysql-schema.sql`

，用于Nacos

```bash
# 开头指定数据库
USE nacos;
```



`03-mysql.sql` 来自`https://github.com/apache/incubator-seata/tree/develop/script/server/db`，用于seata

```bash
# 开头指定数据库
USE `seata-server`;
```

`.root_password`是MySQL数据库的root的密码。

`my.cnf`是MySQL的配置文件，可根据个人需求做调整。

`compose.yaml`是部署用的文件，内容说明如下

```bash
services:
  # 此处默认用mysql，不要改。若必须改，应当将nacos和seata中的所有连接MySQL的配置信息都修改。
  mysql:
    # 建议使用最新子版本，主版本务必是8
    image: mysql:8.4.10
    # 此处随意。
    container_name: mysql8
    restart: unless-stopped

    healthcheck:
      test:
        [
          "CMD-SHELL",
          "mysqladmin ping -h 127.0.0.1 -uroot -p\"$(cat /run/secrets/mysql_root_password)\" --silent 2>/dev/null || exit 1",
        ]
      interval: 15s
      timeout: 5s
      retries: 5
      start_period: 60s

# 特殊优化，Alibaba qwen3.7 max推荐
    stop_grace_period: 90s
    stop_signal: SIGTERM

# 资源限制
    deploy:
      resources:
        limits:
          cpus: "4"
          memory: 4g
        reservations:
          cpus: "2"
          memory: 2g

    ports:
      - "0.0.0.0:3306:3306"

    volumes:
    # 配置文件挂载到容器中
      - ./my.cnf:/etc/mysql/conf.d/my.cnf:ro
      # 自动初始化，该目录下sql文件的命名，必须按照开头0到9依次排列，sql文件的开头必须写明 USE db名字;
      - ./init-db:/docker-entrypoint-initdb.d:ro
      - mysql-data:/var/lib/mysql
      - mysql-log:/var/log/mysql

# alibaba qwen3.7 max推荐
    tmpfs:
      - /tmp:size=512M,noexec,nosuid,nodev

# 配置密码文件
    secrets:
      - mysql_root_password

    environment:
      # MySQL 官方镜像支持 *_FILE 后缀变量，从指定文件读取密码
      MYSQL_ROOT_PASSWORD_FILE: /run/secrets/mysql_root_password
      MYSQL_ROOT_HOST: "%"
      TZ: Asia/Shanghai
      LANG: C.UTF-8

# Alibaba qwen3.7 max推荐
    logging:
      driver: json-file
      options:
        max-size: "100m"
        max-file: "5"
        tag: "mysql8"

# Alibaba qwen3.7 max推荐
    security_opt:
      - no-new-privileges:true

# Alibaba qwen3.7 max推荐
    ulimits:
      nofile:
        soft: 65536
        hard: 65536
      nproc:
        soft: 65536
        hard: 65536
    networks:
      - seata-network

# ------------------------------------------------------------
# Secrets 定义 — 密码文件路径，仅宿主机本地文件，不进入镜像
# ------------------------------------------------------------
secrets:
  mysql_root_password:
    file: ./.root_password

volumes:
  mysql-data:
    name: mysql-data
  mysql-log:
    name: mysql-log
# 注意，此处尽量不要改，此处用于实现容器可直接访问service name，无需指定IP
networks:
  seata-network:
    name: seata-net
    driver: bridge
```



## 部署

在目录mysql下，大概等到20秒可以完成部署。

```bash
docker compose up -d
```

# 部署Nacos

在目录`nacos`中处理，

## 配置

`application.properties`是nacos的配置文件

`nacos.env`是nacos的环境变量，必须配置。无法用配置文件替代，nacos强制要求配置环境变量。内容说明如下

```bash
PREFER_HOST_MODE=hostname
# 单机模式
MODE=standalone
# 要连接MySQL
SPRING_DATASOURCE_PLATFORM=mysql
# MySQL的容器名或者service name
MYSQL_SERVICE_HOST=mysql
# MySQL里已经配置了这个数据库
MYSQL_SERVICE_DB_NAME=nacos
# MySQL的端口
MYSQL_SERVICE_PORT=3306
# MySQL里已配置这个账户密码
MYSQL_SERVICE_USER=nacos
MYSQL_SERVICE_PASSWORD=nacos
# 连接MySQL8及以上不报错
MYSQL_SERVICE_DB_PARAM=characterEncoding=utf8&connectTimeout=1000&socketTimeout=3000&autoReconnect=true&useUnicode=true&useSSL=false&serverTimezone=Asia/Shanghai&allowPublicKeyRetrieval=true
NACOS_AUTH_IDENTITY_KEY=NACOS_INNER_AUTH_KEY_8F2s7K
NACOS_AUTH_IDENTITY_VALUE=9xP3sK8f2Dg5zQ1mB7nV4cS6eR9tY2uI7oP5aS1dF0gH6jK
NACOS_AUTH_TOKEN=NmE5czJRNnZYUXE4QlZONW1EY0cyaEo1THFUNnZVOHdFMnlBNjhHRTk5S2pGcw==
```



`compose.yaml`是部署文件，内容说明如下，

```bash
services:
  nacos:
  # 务必使用这个镜像，nacos小版本之间也可能存在大差异。
    image: nacos/nacos-server:v3.2.2
    container_name: nacos-server
    env_file:
      - ./nacos.env
    ports:
      - "8080:8080"
      - "8848:8848"
      - "9848:9848"
    restart: always
    networks:
      - seata-network

networks:
  seata-network:
    name: seata-net
    external: true
```

`seataServer.properties`是给seata专用的配置内容存档。文件内容来自`https://github.com/apache/incubator-seata/blob/develop/script/config-center/config.txt`



## 部署

在目录 nacos下，大概10秒部署完成。

```bash
docker compose up -d
```

## 配置nacos中seata专用配置

web界面访问nacos`物理机IP:8080`，首次登录务必配置密码为`nacos`，因为seata连接nacos的部分配置了nacos的账户密码。登录后，左侧`平台管理`-`命名空间`-`新建命名空间`，名称和ID、描述务必填写`seata-server`，因为seata已配置。务必保存。

左侧`配置管理`-`配置列表`，上方`命名空间`切换为`seata-server`，右侧`新建配置`，dataID必须是`seataServer.properties`，group必须填写`SEATA_GROUP`，配置格式选择`Properties`，

内容填写`seataServer.properties`中的配置，如下所示，完全不用修改，若使用了自定义的MySQL，只需核对MySQL的配置信息。此处配置是给seata使用。务必点击发布保存。

```bash
store.mode=db
#-----db-----
store.db.datasource=druid
store.db.dbType=mysql
# 需要根据mysql的版本调整driverClassName
# mysql8及以上版本对应的driver：com.mysql.cj.jdbc.Driver
# mysql8以下版本的driver：com.mysql.jdbc.Driver
store.db.driverClassName=com.mysql.cj.jdbc.Driver
store.db.url=jdbc:mysql://mysql:3306/seata-server?useUnicode=true&characterEncoding=utf8&connectTimeout=1000&socketTimeout=3000&autoReconnect=true&useSSL=false
store.db.user=seata
store.db.password=seata
# 数据库初始连接数
store.db.minConn=1
# 数据库最大连接数
store.db.maxConn=20
# 获取连接时最大等待时间 默认5000，单位毫秒
store.db.maxWait=5000
# 全局事务表名 默认global_table
store.db.globalTable=global_table
# 分支事务表名 默认branch_table
store.db.branchTable=branch_table
# 全局锁表名 默认lock_table
store.db.lockTable=lock_table
# 查询全局事务一次的最大条数 默认100
store.db.queryLimit=100


# undo保留天数 默认7天,log_status=1（附录3）和未正常清理的undo
server.undo.logSaveDays=7
# undo清理线程间隔时间 默认86400000，单位毫秒
server.undo.logDeletePeriod=86400000
# 二阶段提交重试超时时长 单位ms,s,m,h,d,对应毫秒,秒,分,小时,天,默认毫秒。默认值-1表示无限重试
# 公式: timeout>=now-globalTransactionBeginTime,true表示超时则不再重试
# 注: 达到超时时间后将不会做任何重试,有数据不一致风险,除非业务自行可校准数据,否者慎用
server.maxCommitRetryTimeout=-1
# 二阶段回滚重试超时时长
server.maxRollbackRetryTimeout=-1
# 二阶段提交未完成状态全局事务重试提交线程间隔时间 默认1000，单位毫秒
server.recovery.committingRetryPeriod=1000
# 二阶段异步提交状态重试提交线程间隔时间 默认1000，单位毫秒
server.recovery.asynCommittingRetryPeriod=1000
# 二阶段回滚状态重试回滚线程间隔时间  默认1000，单位毫秒
server.recovery.rollbackingRetryPeriod=1000
# 超时状态检测重试线程间隔时间 默认1000，单位毫秒，检测出超时将全局事务置入回滚会话管理器
server.recovery.timeoutRetryPeriod=1000
```



# 部署Seata

在目录seata中，

## 配置

目录`config`是seata存放配置文件专用。`seata-server`和`seata-naming-server`有各自专用配置文件。

文件`config/seata-server/application.yml`来自`https://github.com/apache/incubator-seata/blob/develop/server/src/main/resources/application.example.yml`

内容说明如下

```bash

server:
  port: 8091
spring:
  application:
    name: seata-server
  main:
    web-application-type: none
logging:
  config: classpath:logback-spring.xml
  file:
    path: ${log.home:${user.home}/logs/seata}
  extend:
    logstash-appender:
      # off by default
      enabled: false
      destination: 127.0.0.1:4560
    kafka-appender:
      # off by default
      enabled: false
      bootstrap-servers: 127.0.0.1:9092
      topic: logback_to_logstash
      producer:
        acks: 0
        linger-ms: 1000
        max-block-ms: 0
    metric-appender:
      # off by default
      enabled: false

seata:
  config:
    # support: nacos, consul, apollo, zk, etcd3
    type: nacos
    nacos:
    # 配置nacos的容器名字、service name或者宿主机IP，端口必须是nacos的8848或者映射的端口
      server-addr: nacos:8848
      # 此处对应nacos中配置的命名空间
      namespace: seata-server
      # 此处对应nacos中配置文件写的group
      group: SEATA_GROUP
      # 此处写nacos web界面登录的账户密码
      username: nacos
      password: nacos
      # 此处写nacos里配置文件的data id
      data-id: seataServer.properties
  registry:
    # support: nacos, eureka, redis, zk, consul, etcd3, sofa, seata
    type: nacos
    nacos:
    # 同上
      application: seata-server
      server-addr: nacos:8848
      group: SEATA_GROUP
      namespace: seata-server
      # tc集群名称
      cluster: default
      username: nacos
      password: nacos
  store:
    # support: file 、 db 、 redis 、 raft
    mode: db
    db:
      datasource: druid
      db-type: mysql
      driver-class-name: com.mysql.cj.jdbc.Driver
      # 填写MySQL的配置信息，在MySQL部分已经自动配置。
      url: jdbc:mysql://mysql:3306/seata-server?useUnicode=true&characterEncoding=utf8&useSSL=false&serverTimezone=GMT%2B8
      user: seata
      password: seata
      min-conn: 5
      max-conn: 30
      global-table: global_table
      branch-table: branch_table
      lock-table: lock_table
      query-limit: 100
      max-wait: 5000
  security:
    secretKey: SeataSecretKey0c382ef121d778043159209298fd40bf3850a017
    tokenValidityInMilliseconds: 1800000
    ignore:
      urls: /,/**/*.css,/**/*.js,/**/*.html,/**/*.map,/**/*.svg,/**/*.png,/**/*.ico,/console-fe/public/**,/api/v1/auth/login
```

文件`config/seata-naming-server/application.yml`来自`https://github.com/apache/incubator-seata/blob/2.x/namingserver/src/main/resources/application.yml`

内容说明如下，

```bash
server:
  port: 8081

spring:
  application:
    name: seata-namingserver

logging:
  config: classpath:logback-spring.xml
  file:
    path: ${log.home:${user.home}/logs/seata}

heartbeat:
  threshold: 90000  # Heartbeat timeout threshold (milliseconds)
  period: 60000     # Heartbeat period (milliseconds)

seata:
  security:
    # JWT secret key (must be Base64 encoded, generated with openssl rand -base64 32)
    secretKey: SeataSecretKey0c382ef121d778043159209298fd40bf3850a017
    # Token validity period (milliseconds)
    tokenValidityInMilliseconds: 1800000
    # CSRF ignored URLs
    csrf-ignore-urls: /naming/v1/**,/api/v1/naming/**
    ignore:
      urls: /,/**/*.css,/**/*.js,/**/*.html,/**/*.map,/**/*.svg,/**/*.png,/**/*.jpeg,/**/*.ico,/api/v1/auth/login,/version.json,/naming/v1/health,/error

# Console account configuration
console:
  user:
  # seata web界面的账户密码。
    username: admin      # Console login username, default: seata
    password: admin123   # Console login password, leave blank for auto-generated random password
```

文件`compose.yaml`是部署文件。内容说明如下

```bash
services:
  # Seata 事务协调器服务
  seata-server:
  # 务必选择可下载的稳定版，容器镜像最新版本是测试版。
    image: apache/seata-server:2.7.0
    container_name: seata-server
    ports:
      - "7091:7091"
      - "8091:8091"
    environment:
      - STORE_MODE=db
      # 此处必须填写宿主机IP
      - SEATA_IP=172.16.0.30
      - SEATA_PORT=8091
    volumes:
      - "/usr/share/zoneinfo/Asia/Shanghai:/etc/localtime"
      - "/usr/share/zoneinfo/Asia/Shanghai:/etc/timezone"
      - ./config/seata-server/application.yml:/seata-server/resources/application.yml
    networks:
      - seata-network
    restart: always

  # Seata 控制台服务（Namingserver）
  seata-namingserver:
  # 同上
    image: apache/seata-naming-server:2.7.0.jdk25
    container_name: seata-naming-server
    hostname: seata-namingserver
    ports:
      - "8081:8081"
      - "10055:10055"
    volumes:
      - ./config/seata-naming-server/application.yml:/seata-naming-server/resources/application.yml
    networks:
      - seata-network
    restart: always

networks:
  seata-network:
    name: seata-net
    external: true

```



## 部署

在目录seata中，

```bash
docker compose up -d
```



## 访问

部署后可访问web界面进行管理，`宿主机IP::8081`，账户`admin`，密码`admin123`

