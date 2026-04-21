# 说明
部署MySQL9最新子版本单实例最佳实践。

# 参考资料
文档`https://dev.mysql.com/doc/refman/9.6/en/`

# 配置说明
对`deploy.sh`说明
```bash
docker run -d \
  --name mysql9-single \
  --restart=unless-stopped \
  # 当前最佳实践，确保数据库真实启动，而非仅MySQL进程启动。
  --health-cmd='mysql -h 127.0.0.1 -u root -p$MYSQL_ROOT_PASSWORD -e "SELECT 1"' \
  --health-interval=10s \
  --health-timeout=5s \
  --health-retries=3 \
  --health-start-period=30s \
  -p 3306:3306 \
  # my.cnf可按业务需求修改，此处默认存放在当前目录，建议填写绝对路径。
  -v $(pwd)/my.cnf:/etc/my.cnf:ro \
  # 不用数据卷时，可配置本机目录的绝对路径。
  -v mysql9-data:/var/lib/mysql \
  # 同上
  -v mysql9-log:/var/log/mysql \
  -e MYSQL_ROOT_PASSWORD=root123456 \
  -e MYSQL_ROOT_HOST=% \
  -e TZ=Asia/Shanghai \
  -e LANG=C.UTF-8 \
  mysql:9.6.0

```

# 部署
## 联网部署
容器镜像加速下载可参考`1ms.run`

```bash
bash deploy.sh
```
## 离线部署

下载容器镜像，打包容器镜像为文件，导入镜像到部署环境。执行部署脚本。