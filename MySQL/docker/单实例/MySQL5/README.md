# 说明
部署MySQL5最新子版本单实例最佳实践。采用挂载`my.cnf`方法，快捷调整MySQL配置，配置数据卷实现数据存储持久化。  

# 参考文档
`https://dev.mysql.com/doc/refman/5.7/en/`

# 配置说明
可根据业务需求修改本地的`my.cnf`，重启容器使配置生效。  
对`deploy.sh`的说明如下，
```bash
#!/bin/bash

# 部署容器
docker run -d \
# 定义容器名字
  --name mysql5-single \
  --restart unless-stopped \
  # 物理端口:容器内端口
  -p 3306:3306 \
  -e TZ=Asia/Shanghai \
  # MySQL初始化时root用户的密码
  -e MYSQL_ROOT_PASSWORD=root123456 \
  -e MYSQL_ROOT_HOST=% \
  # 若使用本机目录，mysql5-data 改为目录绝对路径。
  -v mysql5-data:/var/lib/mysql \
  # my.cnf默认在当前目录，可以存放在指定目录，改$(pwd)/my.cnf 为真实绝对路径。
  -v $(pwd)/my.cnf:/etc/mysql/conf.d/my.cnf \
  --health-cmd='mysqladmin ping -h127.0.0.1 -P3306 -uroot -p"root123456" --silent' \
  --health-interval=10s \
  --health-timeout=5s \
  --health-retries=3 \
  --health-start-period=30s \
  mysql:5.7.44
```


# 部署
## 联网部署
容器镜像加速下载可参考`1ms.run`

```bash
bash deploy.sh
```
## 离线部署

下载容器镜像，打包容器镜像为文件，导入镜像到部署环境。执行部署脚本。