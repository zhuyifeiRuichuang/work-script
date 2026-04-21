# 说明
docker环境部署redis

# redis容器镜像版本
访问`https://hub.docker.com/_/redis`查询具体版本，推荐使用最新版本规避软件漏洞。

# 常见部署模式
单实例。在云主机仅启动一个redis server。  
主从。一主多从，master负责读写，slave负责读。slave实时备份master完整数据。  
集群。
