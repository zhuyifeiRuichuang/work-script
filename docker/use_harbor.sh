#!/bin/bash
# 在当前机器导入harbor信任证书，可docker pull harbor仓库镜像。

# 创建专用目录，仅在目录不存在时创建
CERTS_DIR="/etc/docker/certs.d/docker.ruichuang.com"
if [ ! -d "$CERTS_DIR" ]; then
    echo "创建证书目录: $CERTS_DIR"
    mkdir -p "$CERTS_DIR"
else
    echo "证书目录已存在: $CERTS_DIR"
fi

# 无论crt文件是否存在，都复制一次
echo "复制证书文件到目标目录"
cp ruichuang.com.crt "$CERTS_DIR/ca.crt"

# 配置hosts解析，仅在不存在时添加
HOST_ENTRY1="172.16.0.19	docker.ruichuang.com"
HOST_ENTRY2="172.16.0.19	ruichuang.com"

if ! grep -qxF "$HOST_ENTRY1" /etc/hosts; then
    echo "添加hosts配置: $HOST_ENTRY1"
    echo "$HOST_ENTRY1" >> /etc/hosts
else
    echo "hosts配置已存在: $HOST_ENTRY1"
fi

if ! grep -qxF "$HOST_ENTRY2" /etc/hosts; then
    echo "添加hosts配置: $HOST_ENTRY2"
    echo "$HOST_ENTRY2" >> /etc/hosts
else
    echo "hosts配置已存在: $HOST_ENTRY2"
fi

# 重启Docker并检查状态
echo "重启Docker服务..."
systemctl restart docker

# 检查Docker是否正常运行
if ! systemctl is-active --quiet docker; then
    echo "错误：Docker重启后异常，请检查Docker服务状态。"
    exit 1
fi

# 检查拉取镜像
echo "尝试拉取测试镜像..."
docker pull docker.ruichuang.com/library/test1:v1
