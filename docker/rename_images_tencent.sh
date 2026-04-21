#!/bin/bash

# 启用错误处理，遇到错误时退出脚本
set -e

# 目标仓库基础路径
TARGET_BASE="ccr.ccs.tencentyun.com/kubespherev3.4.1"

# 获取所有带标签的Docker镜像（排除无标签的镜像）
images=$(docker images --format "{{.Repository}}:{{.Tag}}" | grep -v "<none>")

# 检查是否有需要处理的镜像
if [ -z "$images" ]; then
    echo "没有找到需要处理的镜像"
    exit 0
fi

# 存储需要删除的新镜像
new_images_to_remove=""

# 遍历处理每个镜像
while IFS= read -r image; do
    # 分离镜像仓库名和标签
    repo=$(echo "$image" | cut -d: -f1)
    tag=$(echo "$image" | cut -d: -f2)
    
    echo "处理镜像: $image"
    
    # 提取仓库名中的最后一部分
    # 对于a/b格式，取b；对于a/b/c格式，取c
    repo_suffix=$(echo "$repo" | awk -F '/' '{print $NF}')
    
    # 构建新的镜像名
    new_image="${TARGET_BASE}/${repo_suffix}:${tag}"
    
    echo "重命名为: $new_image"
    
    # 为镜像创建新标签
    docker tag "$image" "$new_image"
    
    # 推送镜像到目标仓库
    echo "推送镜像到仓库..."
    docker push "$new_image"
    
    # 检查推送是否成功
    if [ $? -eq 0 ]; then
        echo "镜像推送成功"
        new_images_to_remove="$new_images_to_remove $new_image"
    else
        echo "镜像推送失败，将保留新镜像"
    fi
    
    echo "----------------------------------------"
done <<< "$images"

# 删除所有推送成功的新镜像
if [ -n "$new_images_to_remove" ]; then
    # 去重处理
    unique_new_images=$(echo "$new_images_to_remove" | tr ' ' '\n' | sort -u | tr '\n' ' ')
    echo "开始删除以下推送成功的新镜像:"
    echo "$unique_new_images"
    
    # 删除镜像，忽略可能的错误
    docker rmi $unique_new_images || echo "部分镜像可能已被删除或无法删除"
fi

echo "所有镜像处理完成"
