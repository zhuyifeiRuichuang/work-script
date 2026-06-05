#!/bin/bash

#将当前目录下所有tar文件导入环境，实现容器镜像批量导入。

# 进入脚本所在目录（确保路径正确）
cd "$(dirname "$0")" || exit 1

echo "========================================"
echo " 开始导入当前目录下所有 tar 文件到 Docker"
echo "========================================"
echo

# 遍历当前目录下所有 .tar 文件
for tar_file in *.tar; do
    # 如果不存在匹配文件，直接跳过
    [ -e "$tar_file" ] || continue

    echo "👉 正在导入：$tar_file"
    docker load -i "$tar_file"

    # 检查导入是否成功
    if [ $? -eq 0 ]; then
        echo "✅ 成功导入：$tar_file"
    else
        echo "❌ 导入失败：$tar_file"
    fi
    echo
done

echo "========================================"
echo "           所有任务执行完成"
echo "========================================"