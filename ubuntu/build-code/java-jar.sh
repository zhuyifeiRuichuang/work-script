#!/bin/bash
set -euo pipefail  # 严格模式，确保命令失败时脚本终止并捕获错误

# ==============================================
# 可自定义变量
# ==============================================
BASE_IMAGE="ubuntu:24.04"         # 基础镜像
LOCAL_PROJECT_DIR="./"            # 本地项目目录
JDK_PACKAGE="openjdk-21-jdk"      # JDK包名
MAVEN_VERSION="3.9.11"            # Maven版本
CONTAINER_NAME="build-java-code"  # 临时容器名称
# ==============================================

# 容器内固定挂载目录（非根目录，彻底避免挂载到/的错误）
CONTAINER_PROJECT_DIR="/app"
# 临时日志文件（存储命令错误信息）
ERROR_LOG=$(mktemp)

# 清理函数：删除临时文件和容器（即使出错也执行）
cleanup() {
    rm -f "$ERROR_LOG"
    if docker inspect "$CONTAINER_NAME" &> /dev/null; then
        docker stop "$CONTAINER_NAME" &> /dev/null
    fi
}
trap cleanup EXIT  # 脚本退出时触发清理


# 步骤1：检查依赖和目录
echo "[1/6] 检查环境依赖..."
if ! command -v docker &> /dev/null; then
    echo "错误：未安装Docker，请先安装Docker"
    exit 1
fi
if [ ! -d "$LOCAL_PROJECT_DIR" ]; then
    echo "错误：本地项目目录 $LOCAL_PROJECT_DIR 不存在"
    exit 1
fi


# 步骤2：拉取基础镜像
echo "[2/6] 拉取基础镜像 $BASE_IMAGE..."
if ! docker pull "$BASE_IMAGE" &> "$ERROR_LOG"; then
    echo "错误：拉取镜像失败"
    echo "详细错误："
    cat "$ERROR_LOG"
    exit 1
fi


# 步骤3：启动容器并挂载目录（使用固定的/app目录，无根目录挂载风险）
echo "[3/6] 启动临时容器并挂载项目目录（本地:$LOCAL_PROJECT_DIR -> 容器:$CONTAINER_PROJECT_DIR）..."
if ! docker run -d --rm --name "$CONTAINER_NAME" \
    -v "$LOCAL_PROJECT_DIR:$CONTAINER_PROJECT_DIR" \
    "$BASE_IMAGE" sleep infinity &> "$ERROR_LOG"; then
    echo "错误：启动容器失败"
    echo "详细错误："
    cat "$ERROR_LOG"
    exit 1
fi


# 步骤4：在容器内安装依赖和编译（cd到固定的/app目录）
echo "[4/6] 安装依赖并编译项目..."
# 容器内命令：重定向输出到日志，仅保留错误码
if ! docker exec "$CONTAINER_NAME" bash -c "
    export DEBIAN_FRONTEND=noninteractive && \
    apt update -y &> /dev/null && \
    apt install -y --no-install-recommends wget tar $JDK_PACKAGE &> /dev/null && \
    wget -q --timeout=30 https://archive.apache.org/dist/maven/maven-3/${MAVEN_VERSION}/binaries/apache-maven-${MAVEN_VERSION}-bin.tar.gz -O /tmp/maven.tar.gz &> /dev/null && \
    tar -zxf /tmp/maven.tar.gz -C /usr/local &> /dev/null && \
    ln -s /usr/local/apache-maven-$MAVEN_VERSION /usr/local/maven &> /dev/null && \
    export PATH=\$PATH:/usr/local/maven/bin && \
    cd '$CONTAINER_PROJECT_DIR' && \
    mvn clean package -DskipTests
" &> "$ERROR_LOG"; then
    echo "错误：编译过程失败"
    echo "容器内详细日志："
    docker logs "$CONTAINER_NAME"  # 输出容器内完整操作日志
    exit 1
fi


# 步骤5：验证编译结果（优化部分）
echo "[5/6] 验证编译结果..."
# 递归查找项目目录下所有名为target的目录（排除权限错误）
TARGET_DIRS=$(find "$LOCAL_PROJECT_DIR" -type d -name "target" 2>/dev/null)

# 检查是否存在任何target目录
if [ -z "$TARGET_DIRS" ]; then
    echo "错误：项目目录下未找到任何target目录，编译可能未成功"
    exit 1
fi

# 检查是否存在非空的target目录
found_non_empty=0
while IFS= read -r dir; do
    # 检查目录是否非空（忽略隐藏文件的错误输出）
    if [ -n "$(ls -A "$dir" 2>/dev/null)" ]; then
        found_non_empty=1
        break  # 找到一个非空目录即可退出循环
    fi
done <<< "$TARGET_DIRS"

if [ "$found_non_empty" -eq 0 ]; then
    echo "错误：找到的所有target目录均为空，编译可能未成功"
    exit 1
fi

# 输出找到的非空target目录（可选，方便用户查看）
echo "找到有效编译结果目录："
while IFS= read -r dir; do
    if [ -n "$(ls -A "$dir" 2>/dev/null)" ]; then
        echo "  - $dir"
    fi
done <<< "$TARGET_DIRS"


# 步骤6：清理容器
echo "[6/6] 清理临时容器..."
# 容器会被trap中的cleanup自动删除，此处仅提示


echo "操作完成！编译结果已保存至上述target目录"
