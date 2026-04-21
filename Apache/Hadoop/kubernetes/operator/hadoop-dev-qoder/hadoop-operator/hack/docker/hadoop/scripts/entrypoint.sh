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

# Apache Hadoop Docker 镜像入口脚本
# 支持组件: namenode, datanode, resourcemanager, nodemanager, journalnode

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

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

# 环境变量默认值
: ${HADOOP_USER:=hadoop}
: ${HADOOP_HOME:=/opt/hadoop}
: ${HADOOP_CONF_DIR:=/opt/hadoop/etc/hadoop}
: ${HADOOP_LOG_DIR:=/opt/hadoop/logs}
: ${HADOOP_DATA_DIR:=/opt/hadoop/data}

# 获取当前组件类型（从命令参数或环境变量）
get_component_type() {
    local cmd="$1"
    case "$cmd" in
        "hdfs")
            case "$2" in
                "namenode") echo "namenode" ;;
                "datanode") echo "datanode" ;;
                "journalnode") echo "journalnode" ;;
                "zkfc") echo "zkfc" ;;
                *) echo "unknown" ;;
            esac
            ;;
        "yarn")
            case "$2" in
                "resourcemanager") echo "resourcemanager" ;;
                "nodemanager") echo "nodemanager" ;;
                "timelineserver") echo "timelineserver" ;;
                *) echo "unknown" ;;
            esac
            ;;
        "mapred")
            case "$2" in
                "historyserver") echo "historyserver" ;;
                *) echo "unknown" ;;
            esac
            ;;
        *)
            echo "unknown"
            ;;
    esac
}

# 等待服务就绪
wait_for_service() {
    local host="$1"
    local port="$2"
    local timeout="${3:-300}"
    local interval="${4:-5}"
    
    log_info "Waiting for $host:$port to be ready..."
    
    local elapsed=0
    while ! nc -z "$host" "$port" 2>/dev/null; do
        if [ $elapsed -ge $timeout ]; then
            log_error "Timeout waiting for $host:$port after ${timeout}s"
            return 1
        fi
        sleep $interval
        elapsed=$((elapsed + interval))
        echo -n "."
    done
    echo
    log_info "$host:$port is ready!"
}

# 初始化 NameNode
init_namenode() {
    log_info "Initializing NameNode..."
    
    # 确保数据目录存在
    mkdir -p "${HADOOP_DATA_DIR}/nn"
    
    # 检查是否已格式化
    if [ ! -f "${HADOOP_DATA_DIR}/nn/current/VERSION" ]; then
        log_info "NameNode not formatted. Formatting now..."
        if hdfs namenode -format -nonInteractive -force; then
            log_info "NameNode formatted successfully"
        else
            log_error "Failed to format NameNode"
            return 1
        fi
    else
        log_info "NameNode already formatted (VERSION file exists)"
    fi
}

# 初始化 DataNode
init_datanode() {
    log_info "Initializing DataNode..."
    
    # 确保数据目录存在
    mkdir -p "${HADOOP_DATA_DIR}/dn"
    
    # 等待 NameNode 就绪
    local namenode_host="${NAMENODE_HOST:-namenode}"
    local namenode_port="${NAMENODE_PORT:-9000}"
    
    wait_for_service "$namenode_host" "$namenode_port" 300 5
}

# 初始化 JournalNode
init_journalnode() {
    log_info "Initializing JournalNode..."
    
    # 确保数据目录存在
    mkdir -p "${HADOOP_DATA_DIR}/jn"
}

# 设置权限
setup_permissions() {
    log_info "Setting up permissions..."
    
    # 确保 hadoop 用户拥有数据目录
    if id "$HADOOP_USER" &>/dev/null; then
        chown -R "${HADOOP_USER}:${HADOOP_USER}" "${HADOOP_DATA_DIR}" 2>/dev/null || true
        chown -R "${HADOOP_USER}:${HADOOP_USER}" "${HADOOP_LOG_DIR}" 2>/dev/null || true
    fi
}

# 打印环境信息
print_env_info() {
    log_info "========================================"
    log_info "Apache Hadoop Docker Container"
    log_info "========================================"
    log_info "Hadoop Version: ${HADOOP_VERSION:-unknown}"
    log_info "Hadoop Home: ${HADOOP_HOME}"
    log_info "Hadoop Conf Dir: ${HADOOP_CONF_DIR}"
    log_info "Hadoop Data Dir: ${HADOOP_DATA_DIR}"
    log_info "Hadoop Log Dir: ${HADOOP_LOG_DIR}"
    log_info "Component: ${COMPONENT_TYPE}"
    log_info "========================================"
}

# 主函数
main() {
    # 确定组件类型
    COMPONENT_TYPE=$(get_component_type "$@")
    
    # 如果无法从参数确定，尝试从环境变量
    if [ "$COMPONENT_TYPE" = "unknown" ] && [ -n "$HADOOP_COMPONENT" ]; then
        COMPONENT_TYPE="$HADOOP_COMPONENT"
    fi
    
    print_env_info
    
    # 根据组件类型执行初始化
    case "$COMPONENT_TYPE" in
        "namenode")
            setup_permissions
            init_namenode
            ;;
        "datanode")
            setup_permissions
            init_datanode
            ;;
        "journalnode")
            setup_permissions
            init_journalnode
            ;;
        "resourcemanager"|"nodemanager"|"historyserver")
            setup_permissions
            log_info "Starting $COMPONENT_TYPE..."
            ;;
        *)
            log_warn "Unknown component type: $COMPONENT_TYPE"
            log_warn "Starting with provided command: $*"
            ;;
    esac
    
    # 使用 exec 启动进程，确保信号正确传递
    log_info "Starting Hadoop service: $*"
    exec "$@"
}

# 执行主函数
main "$@"
