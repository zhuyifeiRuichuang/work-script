#!/bin/bash
# 脚本适用于联网环境。
set -euo pipefail

# 配置参数
# 版本自选，必填。
ETCD_VERSION=v3.5.9
# 旧数据存放的目录。必填。
ETCD_DATA_DIR=/data/etcd
ETCD_LISTEN_URL=http://127.0.0.1:2389
# etcd实例名字。可选。
ETCD_NAME=temp-etcd
TARBALL=etcd-${ETCD_VERSION}-linux-amd64.tar.gz
PRIMARY_URL="https://github.com/etcd-io/etcd/releases/download/${ETCD_VERSION}/${TARBALL}"
# 针对中国地区网络加速优化
BACKUP_URL="https://ghfast.top/https://github.com/etcd-io/etcd/releases/download/${ETCD_VERSION}/${TARBALL}"

# 检查是否已安装etcd
check_etcd_installed() {
    if command -v etcd &> /dev/null; then
        echo "检测到etcd已安装，版本: $(etcd --version | head -n1 | awk '{print $3}')"
        return 0
    else
        return 1
    fi
}

# 下载etcd
download_etcd() {
    echo "尝试从主地址下载etcd..."
    if wget -q --spider "$PRIMARY_URL"; then
        wget -q "$PRIMARY_URL" -O "$TARBALL"
        return 0
    else
        echo "主地址下载失败，尝试备用地址..."
        if wget -q "$BACKUP_URL" -O "$TARBALL"; then
            return 0
        else
            echo "错误: 无法从任何地址下载etcd"
            return 1
        fi
    fi
}

# 安装etcd
install_etcd() {
    echo "开始安装etcd ${ETCD_VERSION}..."
    
    # 下载
    if ! download_etcd; then
        exit 1
    fi
    
    # 解压
    echo "解压安装包..."
    tar zxf "$TARBALL"
    cd "etcd-${ETCD_VERSION}-linux-amd64"
    
    # 安装到系统目录
    echo "安装到系统目录..."
    sudo cp etcd etcdctl /usr/local/bin/
    
    # 清理
    cd ..
    rm -rf "etcd-${ETCD_VERSION}-linux-amd64" "$TARBALL"
    
    echo "etcd安装完成"
}

# 启动etcd
start_etcd() {
    # 确保数据目录存在
    if [ ! -d "$ETCD_DATA_DIR" ]; then
        echo "创建数据目录: $ETCD_DATA_DIR"
        sudo mkdir -p "$ETCD_DATA_DIR"
        sudo chown -R "$(whoami)" "$ETCD_DATA_DIR"
    fi
    
    # 检查是否已运行
    if pgrep -x "etcd" > /dev/null; then
        echo "etcd已在运行"
        return 0
    fi
    
    echo "启动etcd..."
    etcd \
        --data-dir="$ETCD_DATA_DIR" \
        --name="$ETCD_NAME" \
        --listen-client-urls="$ETCD_LISTEN_URL" \
        --advertise-client-urls="$ETCD_LISTEN_URL" &
    
    # 等待启动完成
    echo "等待etcd启动..."
    local max_attempts=30
    local attempt=1
    while [ $attempt -le $max_attempts ]; do
        if ETCDCTL_API=3 etcdctl --endpoints="$ETCD_LISTEN_URL" endpoint health > /dev/null 2>&1; then
            echo "etcd启动成功"
            return 0
        fi
        attempt=$((attempt + 1))
        sleep 1
    done
    
    echo "错误: etcd启动超时"
    return 1
}

# 主流程
main() {
    # 检查权限
    if [ "$(id -u)" -ne 0 ] && ! command -v sudo &> /dev/null; then
        echo "错误: 脚本需要root权限或sudo命令"
        exit 1
    fi
    
    # 检查是否已安装
    if ! check_etcd_installed; then
        install_etcd
    fi
    
    # 启动并验证
    if start_etcd; then
        echo "etcd操作完成，状态正常"
        exit 0
    else
        echo "etcd操作失败"
        exit 1
    fi
}

# 执行主流程
main
