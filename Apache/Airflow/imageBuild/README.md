# 说明

构建容器镜像，用于docker环境部署。

参考`https://airflow.apache.org/docs/docker-stack/index.html`



标准容器镜像：`https://hub.docker.com/r/apache/airflow`，支持AMD64和ARM64的CPU架构。

镜像tag带有特定python标记的，表示内置指定python版本。

带有slim标识的镜像是精简镜像，需自己额外安装插件。

参考Dockerfile`https://github.com/apache/airflow/blob/3.2.0/Dockerfile`



插件`https://airflow.apache.org/docs/apache-airflow/stable/extra-packages-ref.html`

构建指导`https://airflow.apache.org/docs/docker-stack/build.html#build-build-image`

自定义容器镜像`https://airflow.apache.org/docs/docker-stack/index.html`

标准容器镜像发布后将不再变动，不会升级依赖包，或修复安全漏洞，需要自己更新镜像内软件。

`/opt/airflow/`是默认工作目录。

DAG文件位于`/opt/airflow/dags`

日志文件在`/opt/airflow/logs`

不对接外部数据库时，使用SQLite，文件存放在`${AIRFLOW_HOME}/airflow.db`

启动命令`https://airflow.apache.org/docs/docker-stack/entrypoint.html#entrypoint-commands`

# 扩展镜像

基于原版镜像扩展插件和依赖包，

Dockerfile案例，

```bash
FROM apache/airflow:3.2.0
ADD requirements.txt .
RUN pip install apache-airflow==${AIRFLOW_VERSION} -r requirements.txt
```

requirements.txt 填写需要额外添加的依赖包。