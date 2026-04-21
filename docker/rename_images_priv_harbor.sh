#!/bin/bash

# 确保脚本在遇到错误时退出
set -e

# 获取所有docker镜像，过滤掉带有goharbor的，提取镜像名和标签
images=$(docker images --format "{{.Repository}}:{{.Tag}}" | grep -v "goharbor" | grep -v "<none>")

# 检查是否有符合条件的镜像
if [ -z "$images" ]; then
    echo "没有找到符合条件的镜像（排除了带有goharbor的镜像）"
    exit 0
fi

# 存储需要删除的镜像
images_to_remove=""

# 遍历每个镜像
while IFS= read -r image; do
    # 分离镜像名和标签
    repo=$(echo "$image" | cut -d: -f1)
    tag=$(echo "$image" | cut -d: -f2)
    
    # 处理镜像名
    # 统计斜杠数量
    slash_count=$(tr -cd '/' <<< "$repo" | wc -c)
    
    if [ $slash_count -eq 1 ]; then
        # 格式为a/b，将a改为docker.ruichuang.com/kubesphere
        new_repo="docker.ruichuang.com/kubesphere/$(echo "$repo" | cut -d'/' -f2)"
    elif [ $slash_count -ge 2 ]; then
        # 格式为a/b/c或更多层级，将a/b改为docker.ruichuang.com/kubesphere
        # 获取最后一个部分
        last_part=$(echo "$repo" | awk -F '/' '{print $NF}')
        new_repo="docker.ruichuang.com/kubesphere/$last_part"
    else
        # 没有斜杠的情况，直接添加前缀
        new_repo="docker.ruichuang.com/kubesphere/$repo"
    fi
    
    new_image="$new_repo:$tag"
    
    echo "处理镜像: $image"
    echo "重命名为: $new_image"
    
    # 重命名镜像
    docker tag "$image" "$new_image"
    
    # 推送镜像
    echo "推送镜像: $new_image"
    docker push "$new_image"
    
    # 检查推送是否成功
    if [ $? -eq 0 ]; then
        echo "推送成功，标记镜像为待删除"
        # 添加到待删除列表
        images_to_remove="$images_to_remove $image $new_image"
    else
        echo "推送失败，保留镜像: $image 和 $new_image"
    fi
    
    echo "----------------------------------------"
done <<< "$images"

# 处理需要删除的镜像
if [ -n "$images_to_remove" ]; then
    # 去重
    unique_images=$(echo "$images_to_remove" | tr ' ' '\n' | sort -u | tr '\n' ' ')
    echo "删除以下镜像:"
    echo "$unique_images"
    docker rmi $unique_images || echo "部分镜像可能已被删除或无法删除"
fi

echo "所有镜像处理完成"
    
