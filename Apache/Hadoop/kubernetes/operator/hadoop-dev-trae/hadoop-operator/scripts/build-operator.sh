#!/bin/bash

set -e

# 解析参数
VERSION="v1.0.0"
REGISTRY=""
PUSH="false"
ARCH="amd64"

while getopts "v:r:pa:" opt; do
  case $opt in
    v)
      VERSION="$OPTARG"
      ;;
    r)
      REGISTRY="$OPTARG"
      ;;
    p)
      PUSH="true"
      ;;
    a)
      ARCH="$OPTARG"
      ;;
    *)
      echo "用法: $0 [-v version] [-r registry] [-p] [-a architecture]"
      echo "  -v: 指定版本号 (默认: v1.0.0)"
      echo "  -r: 指定镜像仓库"
      echo "  -p: 推送镜像到仓库"
      echo "  -a: 指定架构 (默认: amd64, 支持: amd64, arm64)"
      exit 1
      ;;
  esac
done

echo "=== 构建 Hadoop Operator ==="
echo "版本: $VERSION"
echo "架构: $ARCH"
echo "镜像仓库: $REGISTRY"
echo "推送镜像: $PUSH"

# 检查Go是否可用
if ! command -v go &> /dev/null; then
    echo "错误: go 命令不可用"
    exit 1
fi

# 检查Docker是否可用
if ! command -v docker &> /dev/null; then
    echo "错误: docker 命令不可用"
    exit 1
fi

# 设置Go环境变量
export GO111MODULE=on

# 安装依赖
echo "1. 安装依赖..."
go mod tidy

# 构建operator
echo "2. 构建 Operator 二进制文件..."
go build -ldflags "-w -s" -o build/hadoopoperator ./cmd/manager

# 构建Docker镜像
echo "3. 构建 Docker 镜像..."

# 构建单架构镜像
if [ "$ARCH" = "amd64" ]; then
    docker build -t hadoop-operator:$VERSION -f build/Dockerfile .
elif [ "$ARCH" = "arm64" ]; then
    docker build --platform linux/arm64 -t hadoop-operator:$VERSION-arm64 -f build/Dockerfile .
else
    echo "错误: 不支持的架构: $ARCH"
    exit 1
fi

# 构建多架构镜像
if [ "$REGISTRY" != "" ]; then
    echo "4. 构建多架构镜像..."
    docker buildx create --use || true
    docker buildx build --platform linux/amd64,linux/arm64 \
        -t $REGISTRY/hadoop-operator:$VERSION \
        -f build/Dockerfile \
        --push
else
    echo "4. 跳过多架构镜像构建 (未指定镜像仓库)"
fi

# 推送镜像
if [ "$PUSH" = "true" ] && [ "$REGISTRY" != "" ]; then
    echo "5. 推送镜像到仓库..."
    docker push $REGISTRY/hadoop-operator:$VERSION
fi

echo "=== 构建完成 ==="
if [ "$REGISTRY" != "" ]; then
    echo "镜像名称: $REGISTRY/hadoop-operator:$VERSION"
else
    echo "镜像名称: hadoop-operator:$VERSION"
fi
echo "使用示例:"
echo "  ./build-operator.sh -v v1.0.1 -r my-registry -p  # 构建并推送镜像"
echo "  ./build-operator.sh -a arm64  # 构建arm64架构镜像"

