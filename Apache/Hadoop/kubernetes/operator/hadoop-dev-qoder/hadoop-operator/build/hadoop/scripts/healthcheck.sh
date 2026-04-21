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

# Apache Hadoop Docker 镜像健康检查脚本
# 根据组件类型执行相应的健康检查

# 获取组件类型
COMPONENT_TYPE="${HADOOP_COMPONENT:-unknown}"

# 健康检查函数
check_namenode() {
    # 检查 NameNode Web UI
    if curl -sf http://localhost:9870/jmx > /dev/null 2>&1; then
        return 0
    fi
    # 检查 RPC 端口
    if nc -z localhost 9000 2>/dev/null; then
        return 0
    fi
    return 1
}

check_datanode() {
    # 检查 DataNode Web UI
    if curl -sf http://localhost:9864/ > /dev/null 2>&1; then
        return 0
    fi
    # 检查数据传输端口
    if nc -z localhost 9866 2>/dev/null; then
        return 0
    fi
    return 1
}

check_resourcemanager() {
    # 检查 ResourceManager Web UI
    if curl -sf http://localhost:8088/ws/v1/cluster/info > /dev/null 2>&1; then
        return 0
    fi
    # 检查 RPC 端口
    if nc -z localhost 8032 2>/dev/null; then
        return 0
    fi
    return 1
}

check_nodemanager() {
    # 检查 NodeManager Web UI
    if curl -sf http://localhost:8042/ws/v1/node/info > /dev/null 2>&1; then
        return 0
    fi
    return 1
}

check_journalnode() {
    # 检查 JournalNode RPC 端口
    if nc -z localhost 8485 2>/dev/null; then
        return 0
    fi
    return 1
}

check_historyserver() {
    # 检查 JobHistory Server
    if curl -sf http://localhost:19888/ws/v1/history/info > /dev/null 2>&1; then
        return 0
    fi
    return 1
}

# 主检查逻辑
case "$COMPONENT_TYPE" in
    "namenode")
        check_namenode
        ;;
    "datanode")
        check_datanode
        ;;
    "resourcemanager")
        check_resourcemanager
        ;;
    "nodemanager")
        check_nodemanager
        ;;
    "journalnode")
        check_journalnode
        ;;
    "historyserver")
        check_historyserver
        ;;
    *)
        # 未知组件，尝试通用检查
        # 检查 Java 进程是否存在
        if pgrep -f "org.apache.hadoop" > /dev/null 2>&1; then
            exit 0
        fi
        exit 1
        ;;
esac
