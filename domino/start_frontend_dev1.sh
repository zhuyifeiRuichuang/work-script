#!/bin/bash

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 容器配置
CONTAINER_NAME="domino_frontend_dev1"
HOST_PORT="3001"
IMAGE_NAME="zhuyifeiruichuang/domino-frontend:dev1"
API_URL="http://172.16.0.47:8000"

# 函数：显示进度
show_progress() {
    local task_name="$1"
    local status="$2"
    local color="$3"
    printf "%-25s -100%% -${color}%s${NC}\n" "$task_name" "$status"
}

# 函数：获取主机IP
get_host_ip() {
    # 获取非本地回环的IP地址
    local ip_address=$(ip route get 8.8.8.8 | awk '{print $7; exit}')
    if [ -z "$ip_address" ]; then
        # 如果上述方法失败，尝试其他方法
        ip_address=$(hostname -I | awk '{print $1}')
    fi
    if [ -z "$ip_address" ]; then
        ip_address="127.0.0.1"
    fi
    echo "$ip_address"
}

# 主脚本开始
clear

# 标题
cat << 'HEADER'
=== Domino前端容器部署脚本 ===
版本: 2.0
功能: 自动部署Domino前端容器
HEADER

# 检查Docker服务
if ! systemctl is-active --quiet docker; then
    echo -e "\n${RED}错误: Docker服务未运行，请先启动Docker服务${NC}"
    exit 1
fi

# 停止容器
show_progress "停止容器" "进行中" "$YELLOW"
docker stop "$CONTAINER_NAME" > /dev/null 2>&1
show_progress "停止容器" "完成" "$GREEN"

# 删除旧容器
show_progress "删除旧容器" "进行中" "$YELLOW"
docker rm "$CONTAINER_NAME" > /dev/null 2>&1
show_progress "删除旧容器" "完成" "$GREEN"

# 拉取最新镜像
show_progress "拉取最新镜像" "进行中" "$YELLOW"
docker pull "$IMAGE_NAME" > /dev/null 2>&1
show_progress "拉取最新镜像" "完成" "$GREEN"

# 创建新容器
show_progress "创建新容器" "进行中" "$YELLOW"
docker run -d \
  --name "$CONTAINER_NAME" \
  --restart unless-stopped \
  -p "$HOST_PORT:80" \
  -e DOMINO_DEPLOY_MODE=local-compose \
  -e API_URL="$API_URL" \
  --pull always "$IMAGE_NAME" > /dev/null 2>&1
show_progress "创建新容器" "完成" "$GREEN"

# 等待容器启动
show_progress "等待容器启动" "进行中" "$YELLOW"
sleep 3
show_progress "等待容器启动" "完成" "$GREEN"

# 获取容器信息
CONTAINER_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$CONTAINER_NAME" 2>/dev/null)
CONTAINER_STATUS=$(docker inspect -f '{{.State.Status}}' "$CONTAINER_NAME" 2>/dev/null)
HOST_IP=$(get_host_ip)

# 部署完成报告
cat << 'FOOTER'

=== 部署完成 ===
容器状态检查:
FOOTER

printf "%-15s: %s\n" "容器名称" "$CONTAINER_NAME"
printf "%-15s: %s\n" "运行状态" "$CONTAINER_STATUS"
printf "%-15s: %s\n" "本地端口" "$HOST_PORT"
printf "%-15s: %s\n" "容器IP" "$CONTAINER_IP"
printf "%-15s: %s\n" "主机IP" "$HOST_IP"
printf "%-15s: %s\n" "访问地址" "http://$HOST_IP:$HOST_PORT"

cat << 'END'

操作完成！
END

# 检查容器是否正常运行
if [ "$CONTAINER_STATUS" != "running" ]; then
    echo -e "\n${RED}警告: 容器状态异常，请检查日志${NC}"
    echo "查看日志: docker logs $CONTAINER_NAME"
fi
