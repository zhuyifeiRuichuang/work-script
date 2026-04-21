#!/bin/bash
# Copyright 2024 Apache Software Foundation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Apache Hadoop Operator 镜像构建脚本
# 用于构建 Hadoop Operator 和 Hadoop 集群组件镜像

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 默认配置
REGISTRY="${REGISTRY:-}"
REPOSITORY="${REPOSITORY:-apache/hadoop-operator}"
HADOOP_REPOSITORY="${HADOOP_REPOSITORY:-apache/hadoop}"
VERSION="${VERSION:-latest}"
HADOOP_VERSION="${HADOOP_VERSION:-3.3.6}"
PLATFORMS="${PLATFORMS:-linux/amd64,linux/arm64}"
BUILDER="${BUILDER:-}"
PUSH="${PUSH:-false}"
SKIP_OPERATOR="${SKIP_OPERATOR:-false}"
SKIP_HADOOP="${SKIP_HADOOP:-false}"

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_debug() {
    echo -e "${BLUE}[DEBUG]${NC} $1"
}

# 显示帮助信息
show_help() {
    cat << EOF
Apache Hadoop Operator 镜像构建脚本

用法: $0 [选项]

选项:
    -h, --help              显示帮助信息
    -r, --registry          镜像仓库地址 (例如: registry.example.com)
    -o, --operator-repo     Operator 镜像名称 (默认: apache/hadoop-operator)
    -d, --hadoop-repo       Hadoop 镜像名称 (默认: apache/hadoop)
    -v, --version           镜像版本标签 (默认: latest)
    -H, --hadoop-version    Hadoop 版本 (默认: 3.3.6)
    -p, --platforms         构建平台 (默认: linux/amd64,linux/arm64)
    -b, --builder           指定 buildx builder 名称
    --push                  构建后推送镜像
    --skip-operator         跳过 Operator 镜像构建
    --skip-hadoop           跳过 Hadoop 镜像构建

示例:
    # 构建所有镜像（本地）
    $0

    # 构建并推送到私有仓库
    $0 --registry registry.example.com --push

    # 仅构建 Hadoop 镜像
    $0 --skip-operator

    # 构建指定版本并推送
    $0 -v 1.0.0 -H 3.3.6 --push

    # 多平台构建
    $0 -p linux/amd64,linux/arm64,linux/s390x --push

EOF
}

# 解析命令行参数
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -r|--registry)
                REGISTRY="$2"
                shift 2
                ;;
            -o|--operator-repo)
                REPOSITORY="$2"
                shift 2
                ;;
            -d|--hadoop-repo)
                HADOOP_REPOSITORY="$2"
                shift 2
                ;;
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            -H|--hadoop-version)
                HADOOP_VERSION="$2"
                shift 2
                ;;
            -p|--platforms)
                PLATFORMS="$2"
                shift 2
                ;;
            -b|--builder)
                BUILDER="$2"
                shift 2
                ;;
            --push)
                PUSH="true"
                shift
                ;;
            --skip-operator)
                SKIP_OPERATOR="true"
                shift
                ;;
            --skip-hadoop)
                SKIP_HADOOP="true"
                shift
                ;;
            *)
                log_error "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装"
        exit 1
    fi
    
    # 检查 buildx 是否可用
    if ! docker buildx version &> /dev/null; then
        log_error "Docker buildx 未启用"
        exit 1
    fi
    
    log_info "依赖检查通过"
}

# 设置 buildx builder
setup_builder() {
    if [ -n "$BUILDER" ]; then
        log_info "使用指定的 builder: $BUILDER"
        docker buildx use "$BUILDER" 2>/dev/null || {
            log_info "创建新的 builder: $BUILDER"
            docker buildx create --name "$BUILDER" --use
        }
    else
        # 使用默认 builder 或创建新的
        if ! docker buildx inspect default &> /dev/null; then
            BUILDER="hadoop-operator-builder"
            log_info "创建 builder: $BUILDER"
            docker buildx create --name "$BUILDER" --use
        fi
    fi
}

# 构建 Operator 镜像
build_operator() {
    log_info "========================================"
    log_info "构建 Hadoop Operator 镜像"
    log_info "========================================"
    
    local image_tag="${REGISTRY:+${REGISTRY}/}${REPOSITORY}:${VERSION}"
    local dockerfile="Dockerfile"
    local build_context="."
    
    log_info "镜像标签: $image_tag"
    log_info "构建平台: $PLATFORMS"
    log_info "Dockerfile: $dockerfile"
    
    # 构建参数
    local build_args=(
        "--build-arg" "VERSION=$VERSION"
        "--build-arg" "BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')"
        "--build-arg" "GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
        "--build-arg" "GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo 'unknown')"
    )
    
    # 构建命令
    local build_cmd=(
        docker buildx build
        --platform "$PLATFORMS"
        --tag "$image_tag"
        --file "$dockerfile"
        "${build_args[@]}"
        --provenance=false
    )
    
    if [ "$PUSH" = "true" ]; then
        log_info "构建并推送镜像..."
        build_cmd+=(--push)
    else
        log_info "构建镜像（本地加载）..."
        build_cmd+=(--load)
    fi
    
    build_cmd+=("$build_context")
    
    log_debug "执行命令: ${build_cmd[*]}"
    "${build_cmd[@]}"
    
    log_info "Operator 镜像构建完成: $image_tag"
}

# 构建 Hadoop 镜像
build_hadoop() {
    log_info "========================================"
    log_info "构建 Hadoop 集群组件镜像"
    log_info "========================================"
    
    local image_tag="${REGISTRY:+${REGISTRY}/}${HADOOP_REPOSITORY}:${HADOOP_VERSION}"
    local dockerfile="hack/docker/hadoop/Dockerfile"
    local build_context="hack/docker/hadoop"
    
    log_info "镜像标签: $image_tag"
    log_info "Hadoop 版本: $HADOOP_VERSION"
    log_info "构建平台: $PLATFORMS"
    
    # 检查 Dockerfile 是否存在
    if [ ! -f "$dockerfile" ]; then
        log_error "Dockerfile 不存在: $dockerfile"
        exit 1
    fi
    
    # 构建参数
    local build_args=(
        "--build-arg" "HADOOP_VERSION=$HADOOP_VERSION"
    )
    
    # 构建命令
    local build_cmd=(
        docker buildx build
        --platform "$PLATFORMS"
        --tag "$image_tag"
        --file "$dockerfile"
        "${build_args[@]}"
        --provenance=false
    )
    
    if [ "$PUSH" = "true" ]; then
        log_info "构建并推送镜像..."
        build_cmd+=(--push)
    else
        log_info "构建镜像（本地加载）..."
        build_cmd+=(--load)
    fi
    
    build_cmd+=("$build_context")
    
    log_debug "执行命令: ${build_cmd[*]}"
    "${build_cmd[@]}"
    
    log_info "Hadoop 镜像构建完成: $image_tag"
}

# 打印构建摘要
print_summary() {
    log_info "========================================"
    log_info "构建摘要"
    log_info "========================================"
    
    if [ "$SKIP_OPERATOR" != "true" ]; then
        local operator_tag="${REGISTRY:+${REGISTRY}/}${REPOSITORY}:${VERSION}"
        log_info "Operator 镜像: $operator_tag"
    fi
    
    if [ "$SKIP_HADOOP" != "true" ]; then
        local hadoop_tag="${REGISTRY:+${REGISTRY}/}${HADOOP_REPOSITORY}:${HADOOP_VERSION}"
        log_info "Hadoop 镜像: $hadoop_tag"
    fi
    
    log_info "构建平台: $PLATFORMS"
    log_info "推送镜像: $PUSH"
    log_info "========================================"
}

# 主函数
main() {
    parse_args "$@"
    
    log_info "Apache Hadoop Operator 镜像构建脚本"
    log_info "版本: $VERSION, Hadoop: $HADOOP_VERSION"
    
    check_dependencies
    setup_builder
    
    # 构建 Operator 镜像
    if [ "$SKIP_OPERATOR" != "true" ]; then
        build_operator
    else
        log_warn "跳过 Operator 镜像构建"
    fi
    
    # 构建 Hadoop 镜像
    if [ "$SKIP_HADOOP" != "true" ]; then
        build_hadoop
    else
        log_warn "跳过 Hadoop 镜像构建"
    fi
    
    print_summary
    
    log_info "所有镜像构建完成!"
}

# 执行主函数
main "$@"
