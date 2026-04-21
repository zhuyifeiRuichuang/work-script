#!/bin/bash

# 检查jq工具是否安装
if ! command -v jq &> /dev/null; then
    echo "错误：未安装jq工具，请先安装："
    echo "Ubuntu/Debian: sudo apt-get install jq"
    echo "CentOS/RHEL: sudo yum install jq"
    exit 1
fi

# 明确指定repositories.json路径（以自己环境为准。默认：/var/lib/docker/image/overlay2/repositories.json）
REPO_JSON_PATH="image/overlay2/repositories.json"

# 验证文件存在性
if [ ! -f "$REPO_JSON_PATH" ]; then
    echo "错误：未找到文件 $REPO_JSON_PATH"
    echo "当前目录文件列表："
    ls -l image/overlay2/
    exit 1
fi

# 解析JSON并生成正确的镜像映射关系
echo "正在解析镜像信息..."
jq -r '
    .Repositories | to_entries[] | 
    .key as $repo | 
    .value | to_entries[] | 
    select(.key | contains("@") | not) |  # 排除摘要标签
    [.value, ($repo + ":" + .key)] | 
    @tsv
' "$REPO_JSON_PATH" > /tmp/valid_mappings.txt

# 检查解析结果
if [ ! -s /tmp/valid_mappings.txt ]; then
    echo "错误：未解析到有效镜像信息"
    echo "JSON结构检查："
    jq '.Repositories | keys' "$REPO_JSON_PATH"
    exit 1
fi

# 处理镜像标签
echo "开始修复镜像标签..."
while IFS=$'\t' read -r image_id full_tag; do
    # 修复重复仓库名的标签（如 "ubuntu:ubuntu:22.04" → "ubuntu:22.04"）
    fixed_tag=$(echo "$full_tag" | sed -E 's/^([^:]+):\1:/\1:/')
    
    # 检查镜像是否存在
    if docker images --no-trunc --format "{{.ID}}" | grep -q "^${image_id}$"; then
        # 检查标签是否已存在，不存在才添加
        if ! docker images --format "{{.Repository}}:{{.Tag}}" | grep -q "^${fixed_tag}$"; then
            echo "添加标签: $fixed_tag (ID: ${image_id:0:12})"
            docker tag "$image_id" "$fixed_tag"
        else
            echo "标签已存在: $fixed_tag (ID: ${image_id:0:12})"
        fi
    else
        echo "镜像不存在: $fixed_tag (ID: ${image_id:0:12})"
    fi
done < /tmp/valid_mappings.txt

# 清理临时文件
rm -f /tmp/valid_mappings.txt

echo "操作完成！请使用 'docker images' 查看修复结果。"
    
