#!/bin/bash
# run custom domino frontend server

# 容器配置
CONTAINER_NAME="domino_frontend_dev1"
API_URL="http://10.12.2.95:31"
IMAGE="zhuyifeiruichuang/domino-frontend:dev1"
HOST_PORT="3001"

# 打印分隔线，增强可读性
echo "=============================================="

# 停止并删除旧容器（抑制命令本身输出）
docker stop $CONTAINER_NAME > /dev/null 2>&1
docker rm $CONTAINER_NAME > /dev/null 2>&1

# 统一提示旧容器已删除
echo "✅ 旧容器 [$CONTAINER_NAME] 已删除"

# 启动新容器
docker run -d \
  --name $CONTAINER_NAME \
  --restart unless-stopped \
  -p $HOST_PORT:80 \
  -e DOMINO_DEPLOY_MODE=local-compose \
  -e API_URL=$API_URL \
  --pull always $IMAGE > /dev/null 2>&1

# 提示新容器已启动
echo "🚀 新容器 [$CONTAINER_NAME] 已启动"
echo "----------------------------------------------"
echo "容器状态信息："
docker ps -a | grep $CONTAINER_NAME
echo "=============================================="
