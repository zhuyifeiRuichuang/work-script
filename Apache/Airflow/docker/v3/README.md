# 说明

docker环境部署Apache Airflow

参考`https://airflow.apache.org/docs/apache-airflow/stable/howto/docker-compose/index.html`

仅适用于主版本3及更高版本。



# 基础了解

airflow的基础组件

- `airflow-scheduler`-[调度器](https://airflow.apache.org/docs/apache-airflow/stable/administration-and-deployment/scheduler.html)监视所有任务和 DAG，然后在它们的依赖关系完成后触发任务实例。
- `airflow-dag-processor`- Dag 处理器解析 Dag 文件。
- `airflow-api-server`- API 服务器地址为`http://localhost:8080`：
- `airflow-worker`- 执行调度器分配的任务的工作进程。
- `airflow-triggerer`- 触发器运行可延迟任务的事件循环。
- `airflow-init`- 初始化服务。
- `postgres`数据库。
- `redis`- [Redis 代理](https://redis.io/)，负责将消息从调度器转发到工作器。
- `flower`-[一款用于监测环境的花卉应用程序](https://flower.readthedocs.io/en/latest/)。可在以下网址获取`http://localhost:5555` 文档：`https://flower.readthedocs.io/en/latest/`



容器中的重要目录

- `./dags`您可以将您的 Dag 文件放在这里。
- `./logs`- 包含任务执行和调度程序的日志。
- `./config`- 您可以添加自定义日志解析器或添加`airflow_local_settings.py`到集群策略配置中。
- `./plugins`您可以在这里放置您的[自定义插件](https://airflow.apache.org/docs/apache-airflow/stable/administration-and-deployment/plugins.html)。



标准容器镜像：`https://hub.docker.com/r/apache/airflow`，支持AMD64和ARM64的CPU架构。

支持对接以下数据库

- PostgreSQL
- MySQL
- MSSQL

# 关于配置项

配置文件详细说明：`https://airflow.apache.org/docs/apache-airflow/stable/configurations-ref.html`

`https://airflow.apache.org/docs/apache-airflow/stable/howto/set-config.html`



# 部署

获取compose文件

`https://airflow.apache.org/docs/apache-airflow/3.2.0/docker-compose.yaml`

3.2.0是airflow的版本号。



## 初始化环境

```bash
mkdir -p ./dags ./logs ./plugins ./config
echo -e "AIRFLOW_UID=$(id -u)" > .env
```



## 可配置的环境变量

参考。`https://airflow.apache.org/docs/apache-airflow/stable/howto/docker-compose/index.html#docker-compose-env-variables`

想用插件就在打包镜像时安装，包括依赖软件包。在容器运行时安装软件会使环境启动很慢。

| 变量                         | 作用                             | 示例值                                |
| ---------------------------- | -------------------------------- | ------------------------------------- |
| AIRFLOW_IMAGE_NAME           | 容器镜像版本                     | apache/airflow:3.2.0                  |
| AIRFLOW_UID                  | 容器中运行进程的账户UID          | 50000                                 |
| _AIRFLOW_WWW_USER_USERNAME   | web界面访问airflow时的账户       | airflow                               |
| _AIRFLOW_WWW_USER_PASSWORD   | web界面访问airflow时的密码       | airflow                               |
| _PIP_ADDITIONAL_REQUIREMENTS | 需在容器运行时自动安装的扩展插件 | lxml==4.6.3 charset-normalizer==1.4.1 |



## 初始化airflow.cfg

将在config目录中获取到airflow.cfg

```bash
docker compose run airflow-cli airflow config list
```

对比新旧文件，修改新文件，将旧文件的配置同步到新配置文件中。



## 初始化数据库

```bash
docker compose up airflow-init
```

看到`airflow-init-1 exited with code 0` 和账户密码，表示初始化完成。



## 运行环境

```bash
docker compose up -d
# 单独部署flower组件
docker compose up flower -d
```



# 访问

Ariflow管理界面：`IP:8080`默认账户密码airflow或admin，以env文件中配置为准。

Flower工具管理界面：`IP:5555` 



# 卸载和清除

```bash
docker compose down --volumes --rmi all
docker compose donw flower -v
```

# gitsync

这是airflow的扩展功能，需单独部署一个容器使用

可将指定代码仓库同步到本地，同步给airflow使用。



参考配置如下

如果容器启动异常，就`cmod -R 777 /opt/domino/airflow/dags/git-data`

```bash
root@domino:/opt/domino# cat gitsync-compose.yaml

services:
  git-sync:
    image: zhuyifeiruichuang/git-sync:v4.3.0
    container_name: airflow-gitsync
    restart: always
    user: "50000:0"
    volumes:
      - /opt/domino/airflow/git-data:/sync
    environment:
      GITSYNC_REPO: https://github.com/代码仓库.git
      GITSYNC_BRANCH: main
      GITSYNC_PERIOD: 30s
      GITSYNC_DEPTH: 1
      GITSYNC_ROOT: /sync
      GITSYNC_DEST: repo
      GITSYNC_LINK: current
      GITSYNC_USERNAME: GitHub账户
      GITSYNC_PASSWORD: 认证令牌
      GITSYNC_KNOWN_HOSTS: "false"
      GITSYNC_ONE_TIME: "false"
root@domino:/opt/domino# cat gitsync-ssh-compose.yaml
services:
  git-sync:
    image: zhuyifeiruichuang/git-sync:v4.3.0
    container_name: airflow-git-sync
    user: "0:0"
    restart: always
    environment:
      - GITSYNC_REPO=git@github.com:代码仓库.git
      - GITSYNC_BRANCH=main
      - GITSYNC_PERIOD=60s
      - GITSYNC_SSH=true
      - GITSYNC_ROOT=/git
      - GITSYNC_DEST=repo
      - GITSYNC_LINK=dags
      - GITSYNC_SSH_KNOWN_HOSTS=false
      - GITSYNC_ADD_USER=true
    volumes:
      - /opt/domino/airflow/git-data:/git
      # privkey 文件保存ssh私钥
      - ./privkey:/etc/git-secret/ssh:ro
```

