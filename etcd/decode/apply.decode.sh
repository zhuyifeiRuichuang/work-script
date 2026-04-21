#!/bin/bash

# 检查是否提供了目录参数
if [ $# -ne 1 ]; then
    echo "用法: $0 <目标目录>"
    echo "示例: $0 ./workspaces"
    exit 1
fi

TARGET_DIR="$1"

# 检查目录是否存在
if [ ! -d "$TARGET_DIR" ]; then
    echo "错误: 目录 '$TARGET_DIR' 不存在或不是有效目录"
    exit 1
fi

# 检查kubectl是否可用
if ! command -v kubectl &> /dev/null; then
    echo "错误: 未找到kubectl，请确保已安装并在PATH中"
    exit 1
fi

# 检查yq是否可用（yq是处理YAML的工具，比sed更可靠）
if ! command -v yq &> /dev/null; then
    echo "错误: 未找到yq工具，请先安装yq"
    echo "安装方法（Ubuntu/Debian）: sudo apt install yq"
    echo "或通过二进制安装: https://github.com/mikefarah/yq#install"
    exit 1
fi

# 定义需要从metadata中删除的不可变字段（根据你的YAML文件补充）
FIELDS_TO_DELETE=(
    "uid"
    "creationTimestamp"
    "generation"
    "managedFields"
    "resourceVersion"
    "selfLink"
)

# 构建yq删除命令（例如：del(.metadata.uid) | del(.metadata.creationTimestamp) ...）
DELETE_COMMANDS=""
for field in "${FIELDS_TO_DELETE[@]}"; do
    DELETE_COMMANDS+="del(.metadata.$field) | "
done
# 移除最后一个多余的"|"
DELETE_COMMANDS="${DELETE_COMMANDS%| }"

# 处理所有包含decode的yaml文件
find "$TARGET_DIR" -type f -name "*decode*.yaml" | while read -r yaml_file; do
    echo "处理文件: $yaml_file"
    
    # 创建临时文件存储处理后的内容
    temp_file=$(mktemp)
    
    # 使用yq删除指定字段（精准处理YAML结构，不受缩进影响）
    if yq eval "$DELETE_COMMANDS" "$yaml_file" > "$temp_file"; then
        echo "已成功移除不可变字段"
    else
        echo "错误: 处理文件 '$yaml_file' 时出错"
        rm "$temp_file"
        continue
    fi
    
    # 应用处理后的文件
    echo "正在应用处理后的文件..."
    if kubectl apply -f "$temp_file"; then
        echo "文件 '$yaml_file' 应用成功"
        rm "$temp_file"  # 成功后删除临时文件
    else
        echo "错误: 文件 '$yaml_file' 应用失败"
        echo "临时文件保留以便调试: $temp_file"  # 失败时保留临时文件
    fi
    
    echo "-------------------------"
done

echo "所有文件处理完成"
