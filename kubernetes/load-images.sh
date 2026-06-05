#!/bin/bash
# 将当前目录下所有 tar 文件加载为容器镜像到 k8s 集群节点

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TAR_FILES=("$SCRIPT_DIR"/*.tar)

# 检查是否存在 tar 文件
if [ ! -e "${TAR_FILES[0]}" ]; then
    echo "当前目录下没有找到 .tar 文件"
    exit 1
fi

echo "找到 ${#TAR_FILES[@]} 个 tar 文件，开始加载镜像..."
echo "============================================"

# 检测容器运行时
detect_runtime() {
    if command -v ctr &>/dev/null && ctr version &>/dev/null; then
        echo "containerd"
    elif command -v docker &>/dev/null && docker info &>/dev/null; then
        echo "docker"
    elif command -v crictl &>/dev/null; then
        echo "crictl"
    else
        echo "unknown"
    fi
}

RUNTIME=$(detect_runtime)
echo "检测到容器运行时: $RUNTIME"
echo "============================================"

SUCCESS=0
FAIL=0

for tar_file in "${TAR_FILES[@]}"; do
    filename=$(basename "$tar_file")
    echo ""
    echo ">>> 正在加载: $filename"

    case "$RUNTIME" in
        containerd)
            if ctr -n k8s.io images import "$tar_file"; then
                echo "    [成功] $filename"
                ((SUCCESS++))
            else
                echo "    [失败] $filename"
                ((FAIL++))
            fi
            ;;
        docker)
            if docker load -i "$tar_file"; then
                echo "    [成功] $filename"
                ((SUCCESS++))
            else
                echo "    [失败] $filename"
                ((FAIL++))
            fi
            ;;
        crictl)
            if crictl pull "tar://$tar_file"; then
                echo "    [成功] $filename"
                ((SUCCESS++))
            else
                echo "    [失败] $filename"
                ((FAIL++))
            fi
            ;;
        *)
            echo "    [错误] 未检测到可用的容器运行时 (containerd/docker/crictl)"
            echo "    请确认当前节点已安装容器运行时工具"
            exit 1
            ;;
    esac
done

echo ""
echo "============================================"
echo "加载完成! 成功: $SUCCESS, 失败: $FAIL"

# 如果使用 containerd，列出已加载的镜像
if [ "$RUNTIME" = "containerd" ]; then
    echo ""
    echo "当前 k8s.io 命名空间下的镜像列表:"
    ctr -n k8s.io images list
fi
