# 说明

Domino项目是一个开源任务调度平台，结合Apache Airflow使用。

项目代码：`https://github.com/Tauffer-Consulting/domino`

通过docker全量部署完整版domino及依赖组件。



Domino组件

- 前端组件frontend
- 后端组件rest
- airflow
- domino database
- proxy，使用socat， 用于将每个pice创建专属容器。



新建的任务运行时，domino自动创建专用容器执行任务。当前版本不支持自动清理废弃容器，需使用专用脚本定期清理。



# 容器镜像

容器镜像下载困难时，可在`https://docker.aityp.com/`找替代品。或通过`1ms.run`加速下载。

# 配置GitHub

创建domino专用仓库，例如domino-dev1。

浏览器访问`https://github.com/settings/tokens`，创建token。授权访问者可对domino-dev1仓库有读写权限，可创建文件。

# 配置部署文件

将GitHub 获取的token填写到`.env`的`DOMINO_DEFAULT_PIECES_REPOSITORY_TOKEN`

修改`compose.yaml`中`domino_frontend`的`API_URL`为主机IP。

创建专用部署目录，例如目录`/opt/domino`。

# 部署

```bash
docker compose up -d
```

# 访问

等所有容器状态是`healthy`，

访问domino：浏览器访问`IP:3000` ，账户`admin@email.com`，密码`admin`

访问airflow：浏览器访问`IP:8080`，账户密码：`airlfow`

访问airflow flower：浏览器访问`IP:5555`

# 查看部署日志

```bash
docker compose logs -f
```

# 卸载

完整清空容器和数据。

```bash
docker compose down -v
```

