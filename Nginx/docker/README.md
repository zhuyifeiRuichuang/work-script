# 说明

基于docker的nginx部署管理。用于生产环境标准配置。

# 配置

`config`存放配置文件。

- `nginx.conf`是主配置文件。

- `conf.d`存放具体的站点的专用配置文件。

`html`存放前端代码打包的dist目录下的所有内容。

- `50x.html`是自定义的错误状态反馈的页面。

`compose.yaml`是部署文件，若有端口冲突，可调整物理机映射端口。



# 部署

```bash
docker compose up -d
```



# 访问测试

浏览器访问系统IP

# 更新前端资源

后续更新前端代码时，将dist目录中所有资源存入上述目录的`html`目录下，重启容器可更新前端内容。

```bash
docker compose restart
```

# 替换容器镜像

若使用定制的nginx容器镜像，需先停止容器，修改`compose.yaml`，再启动容器。

```bash
docker compose down
docker compose up -d
```

