#!/bin/bash

# 定义镜像名称和打包文件名
IMAGE_NAME="jnpf-scheduletask:v6.0"
TAR_FILE="jnpf-scheduletask.tar"

echo "===== 开始检查是否存在同名镜像 ====="
# 检查是否存在目标镜像（通过镜像ID是否存在判断，更精准）
if [ -n "$(docker images -q $IMAGE_NAME)" ]; then
    echo "⚠️ 已存在同名镜像：$IMAGE_NAME"
    echo "当前镜像信息："
    docker images | grep "$IMAGE_NAME"
    echo "===== 脚本终止 ====="
    exit 0
fi

echo "===== 不存在同名镜像，开始构建Docker镜像 ====="
# 构建镜像
docker build -t $IMAGE_NAME .

# 检查构建是否成功
if [ $? -ne 0 ]; then
    echo "❌ 镜像构建失败，终止脚本"
    exit 1
fi

echo "===== 镜像构建成功，开始打包 ====="
# 打包镜像为tar文件
docker save -o $TAR_FILE $IMAGE_NAME

# 检查打包是否成功
if [ $? -ne 0 ]; then
    echo "❌ 镜像打包失败，终止脚本"
    exit 1
fi

echo "===== 镜像打包成功，开始删除镜像 ====="
# 删除本地镜像
docker rmi $IMAGE_NAME

# 检查删除是否成功
if [ $? -ne 0 ]; then
    echo "⚠️ 镜像删除失败，可能存在依赖的容器，请手动清理"
    echo "当前镜像列表中该镜像状态："
    docker images | grep $IMAGE_NAME
else
    echo "✅ 操作完成：已成功构建、打包镜像，并删除本地镜像"
    echo "打包文件位置：$(pwd)/$TAR_FILE"
fi
