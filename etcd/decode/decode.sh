#!/bin/bash

# 检查是否提供了目录参数
if [ $# -ne 1 ]; then
    echo "用法: $0 <目标目录>"
    echo "示例: $0 ./configs"
    exit 1
fi

TARGET_DIR="$1"

# 检查目录是否存在
if [ ! -d "$TARGET_DIR" ]; then
    echo "错误: 目录 '$TARGET_DIR' 不存在或不是一个有效的目录"
    exit 1
fi

# 检查auger命令是否可用
if ! command -v auger &> /dev/null; then
    echo "错误: 未找到auger命令，请确保它已安装并在PATH中"
    exit 1
fi

# 递归查找所有yaml文件并处理
find "$TARGET_DIR" -type f -name "*.yaml" | while read -r yaml_file; do
    # 生成输出文件名：在原文件名后添加.decode
    output_file="${yaml_file%.yaml}.decode.yaml"
    
    echo "处理文件: $yaml_file"
    echo "输出到: $output_file"
    
    # 执行转换命令
    if cat "$yaml_file" | auger decode > "$output_file"; then
        echo "转换成功"
    else
        echo "错误: 转换文件 '$yaml_file' 失败"
        # 移除可能生成的空文件
        if [ -f "$output_file" ]; then
            rm "$output_file"
        fi
    fi
    
    echo "-------------------------"
done

echo "所有文件处理完成"
