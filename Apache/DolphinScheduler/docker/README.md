# 说明



# 快捷部署体验

仅用于快速体验，缺失插件，很多功能无法使用。

访问`https://dolphinscheduler.apache.org/en-us/download` ,下载指定版本，推荐用最新版本。例如`https://downloads.apache.org/dolphinscheduler/3.4.1/apache-dolphinscheduler-3.4.1-src.tar.gz`

```bash
# 解压文件
tar -zxf apache-dolphinscheduler-3.4.1-src.tar.gz
cd apache-dolphinscheduler-3.4.1-src/deploy/docker

# 可选。若下载的文件存在故障，可克隆最新代码
git clone https://github.com/apache/dolphinscheduler.git
cd dolphinscheduler-dev/

# 初始化数据库
docker compose --profile schema up -d
# 启动所有服务
docker compose --profile all up -d
```



需要改造，将初始化阶段加到compose的前置部分。

去掉profile参数，部署时一键部署。给其他容器加启动依赖关系，等前置容器健康再启动。

指定容器名字，网络名字，数据卷名字。



# 访问

浏览器访问` http://IP:12345/dolphinscheduler/ui `

默认账户密码是`admin` 和 `dolphinscheduler123`

