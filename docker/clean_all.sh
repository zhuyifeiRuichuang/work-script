#!/bin/bash

# 用于测试环境快速清理环境。
# 注意：清理后不可恢复。高危命令！

# 清空容器
docker rm -f $(docker ps -aq)

# 清空数据卷
docker volume rm $(docker volume ls -q) -f

# 清空自配置的网络
docker network prune -f

# 清空容器镜像
docker rmi -f $(docker images -q)
