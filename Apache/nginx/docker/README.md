# 说明

基于docker的nginx部署管理。用于生产环境标准配置。

# 获取容器镜像

联网环境访问`https://hub.docker.com/_/nginx` 查看可用版本，推荐使用最新版本。容器镜像加速可参考网站`1ms.run`

```bash
docker pull nginx:1.29.5
```

离线环境部署时，先下载容器镜像，保存为文件，再导入至需部署容器的环境。

```bash
docker save -o nginx.tar nginx:1.29.5
docker load -i nginx.tar
```



# 部署

创建nginx专用目录，此方法有助于后续灵活调整nginx配置。可改为自定义目录，同时应修改部署脚本中对应目录。

`conf.d` 存放nginx配置文件

`html`存放前端代码打包生成的dist目录下所有资源。

```bash
mkdir -p /data/nginx/{conf.d,html}
```

若无自定义配置文件需求，可将当前目录下`default.conf` 复制到上述目录的`conf.d`下，后续修改此文件，重启容器，即可灵活调整容器配置。建议配置文件中的端口默认配置80，若希望使用其他端口，应修改容器映射的主机端口。

后续更新前端代码时，将dist目录中所有资源存入上述目录的`html`目录下，重启容器可更新前端内容。



修改部署脚本，说明如下

```bash
docker run -d \
  # 改为业务指定的名字，应有全局辨识度。
  --name nginx \
  --restart unless-stopped \
  # 物理机端口:容器内端口。物理机端口需全局唯一且与当前环境资源不冲突。容器端口默认80，建议不要修改。
  -p 80:80 \
  -e TZ=Asia/Shanghai \
  -v /data/nginx/conf.d:/etc/nginx/conf.d:ro \
  -v /data/nginx/html:/usr/share/nginx/html:rw \
  --health-cmd="curl -f http://localhost/ || exit 1" \
  --health-interval=30s \
  --health-timeout=10s \
  --health-retries=3 \
  --health-start-period=10s \
  # 容器镜像应选最新版本，规避安全漏洞。
  nginx:1.29.5
```

脚本可根据业务需求拓展配置，例如配置`--network`使容器部署在自定义网络，实现其他容器直接访问容器名字即可访问到资源。可配置CPU，内存的占用限制，也可配置nginx 的进程数限制，扩展配置需单独搜索。



执行部署脚本

```bash
bash deploy.sh
```

