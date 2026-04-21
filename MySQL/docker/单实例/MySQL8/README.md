# 说明
部署MySQL8最新子版本单实例最佳实践，采用挂载`my.cnf`方法，快捷调整MySQL配置，配置数据卷实现数据存储持久化。  

# 参考资料
参考文档`https://dev.mysql.com/doc/refman/8.4/en/`

# 配置说明
可根据业务需求修改本地的`my.cnf`，重启容器使配置生效。

`deploy.sh`说明如下，
```bash
# deploy
docker run -d \
  --name mysql8-single \
  --restart=unless-stopped \
  --health-cmd="mysqladmin ping -uroot -p\$MYSQL_ROOT_PASSWORD --silent" \
  --health-interval=10s \
  --health-timeout=5s \
  --health-retries=3 \
  --health-start-period=30s \
  -p 3306:3306 \
  # $(pwd)/my.cnf是my.cnf在本机存放的绝对路径。
  -v $(pwd)/my.cnf:/etc/my.cnf \
  # 若使用本机目录存放，mysql8-data改为目录绝对路径。
  -v mysql8-data:/var/lib/mysql \
  # 同上
  -v mysql8-log:/var/log/mysql \
  # MySQL初始化数据库的root默认密码
  -e MYSQL_ROOT_PASSWORD=root123456 \
  -e MYSQL_ROOT_HOST=% \
  -e TZ=Asia/Shanghai \
  -e LANG=C.UTF-8 \
  mysql:8.4.8

```

# 部署
## 联网部署
容器镜像下载加速参考`1ms.run`
```bash
bash deploy.sh
```
## 离线部署

下载容器镜像，打包容器镜像为文件，导入镜像到部署环境。执行部署脚本。