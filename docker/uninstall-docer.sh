#!/bin/bash
# 用于完整卸载Ubuntu环境中的docker和containerd，runc
set -euo pipefail

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()  { echo -e "${GREEN}[INFO]${NC} $1"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; }

info "========================================"
info "  Docker / Containerd / Runc 彻底卸载脚本"
info "  适用系统: Ubuntu 24.04 LTS"
info "========================================"

# ========================================
# 1. 停止并禁用服务
# ========================================
info "步骤 1/7: 停止并禁用相关服务..."
sudo systemctl stop docker docker.socket containerd 2>/dev/null || true
sudo systemctl disable docker docker.socket containerd 2>/dev/null || true
info "服务已停止并禁用"

# ========================================
# 2. 卸载 apt 安装的所有包
# ========================================
info "步骤 2/7: 卸载 apt 软件包..."
sudo apt remove --purge -y \
    docker-ce docker-ce-cli containerd.io \
    docker-buildx-plugin docker-compose-plugin \
    docker.io runc containerd 2>/dev/null || true
sudo apt autoremove -y
sudo apt autoclean
info "apt 包已卸载"

# ========================================
# 3. 删除所有可能的二进制文件（含手动安装）
# ========================================
info "步骤 3/7: 删除所有二进制残留..."

BINARIES=(
    /usr/bin/docker
    /usr/bin/dockerd
    /usr/bin/docker-init
    /usr/bin/docker-proxy
    /usr/bin/containerd
    /usr/bin/containerd-shim
    /usr/bin/containerd-shim-runc-v2
    /usr/bin/ctr
    /usr/local/bin/docker
    /usr/local/bin/dockerd
    /usr/local/bin/containerd
    /usr/local/bin/containerd-shim*
    /usr/local/bin/ctr
    /usr/local/sbin/runc
    /usr/sbin/runc
    /usr/bin/runc
    /opt/runc
    /snap/bin/docker
    /snap/bin/containerd
    /snap/bin/docker-compose
    /usr/bin/docker-compose
    /usr/local/bin/docker-compose
    /usr/local/bin/docker-compose-v1
    /usr/bin/docker-buildx
    /usr/bin/docker-buildx-plugin
)

for f in "${BINARIES[@]}"; do
    if ls $f 1>/dev/null 2>&1; then
        sudo rm -f $f
        info "  已删除: $f"
    fi
done

# ========================================
# 4. 删除所有数据、配置和状态目录
# ========================================
info "步骤 4/7: 删除数据与配置目录..."

# 删除 systemd 单元文件（系统级 + 用户级）
for unit_path in /lib/systemd/system /usr/lib/systemd/system /etc/systemd/system; do
    for unit in docker.service docker.socket containerd.service containerd.socket; do
        if [ -f "$unit_path/$unit" ]; then
            sudo rm -f "$unit_path/$unit"
            info "  已删除单元: $unit_path/$unit"
        fi
    done
done

DIRS=(
    /var/lib/docker
    /var/lib/containerd
    /run/docker.sock
    /run/docker
    /run/containerd
    /var/run/docker.sock
    /var/run/docker
    /var/run/containerd
    /etc/docker
    /etc/containerd
    /etc/default/docker
    /etc/systemd/system/docker.service.d
    /etc/systemd/system/docker.socket.d
    /etc/systemd/system/containerd.service.d
    /opt/cni
    /opt/containerd
    /etc/cni/net.d
    /root/.docker
    /snap/docker
)

for d in "${DIRS[@]}"; do
    if [ -e "$d" ]; then
        sudo rm -rf "$d"
        info "  已删除: $d"
    fi
done

# 删除所有用户目录下的 .docker
for user_home in /home/*; do
    if [ -d "$user_home/.docker" ]; then
        sudo rm -rf "$user_home/.docker"
        info "  已删除: $user_home/.docker"
    fi
done

# 删除 apt 源配置（可选，如需保留 Docker 源请注释掉下一行）
sudo rm -f /etc/apt/sources.list.d/docker.list /etc/apt/keyrings/docker.gpg /usr/share/keyrings/docker-archive-keyring.gpg 2>/dev/null || true

# ========================================
# 5. 清理网络残留
# ========================================
info "步骤 5/7: 清理网络残留..."

# 删除 docker0 网桥
if ip addr show docker0 >/dev/null 2>&1; then
    sudo ip link del docker0 2>/dev/null || true
    info "  已删除 docker0 网桥"
fi

# 清理 iptables
sudo iptables -F 2>/dev/null || true
sudo iptables -t nat -F 2>/dev/null || true
sudo iptables -t mangle -F 2>/dev/null || true
sudo iptables -X 2>/dev/null || true

# 清理 IPVS
sudo ipvsadm --clear 2>/dev/null || true

# 清理网络命名空间残留
ip netns list 2>/dev/null | while read -r ns; do
    ns_name=$(echo "$ns" | awk '{print $1}')
    if [ -n "$ns_name" ] && [ "$ns_name" != "rtnetlink" ]; then
        sudo ip netns delete "$ns_name" 2>/dev/null || true
    fi
done

info "网络残留已清理"

# ========================================
# 6. 刷新 systemd
# ========================================
info "步骤 6/7: 刷新 systemd..."
sudo systemctl daemon-reload
sudo systemctl reset-failed
sudo systemctl unmask docker 2>/dev/null || true
sudo systemctl unmask containerd 2>/dev/null || true
info "systemd 已刷新"

# ========================================
# 7. 清除 shell 命令缓存并验证
# ========================================
info "步骤 7/7: 清除命令缓存并执行验证..."
hash -r

echo ""
info "========== 验证结果 =========="

# 7.1 验证二进制残留
echo ""
echo "[1] 二进制文件检查:"
ALL_CLEAN=true
for cmd in docker dockerd containerd ctr runc docker-compose docker-buildx; do
    if command -v "$cmd" &>/dev/null; then
        error "  ❌ $cmd 残留: $(command -v $cmd)"
        ALL_CLEAN=false
    else
        echo -e "  ${GREEN}✅${NC} $cmd 已清理"
    fi
done

# 7.2 验证服务单元
echo ""
echo "[2] systemd 服务单元检查:"
if systemctl list-unit-files --type=service 2>/dev/null | grep -qE "docker|containerd"; then
    error "  ❌ 发现残留服务单元:"
    systemctl list-unit-files --type=service | grep -E "docker|containerd" || true
    ALL_CLEAN=false
else
    echo -e "  ${GREEN}✅${NC} 无残留服务单元"
fi

# 7.3 验证数据目录
echo ""
echo "[3] 数据目录检查:"
for dir in /var/lib/docker /var/lib/containerd /etc/docker /etc/containerd /opt/containerd /run/docker /run/containerd /var/run/docker; do
    if [ -e "$dir" ]; then
        error "  ❌ $dir 仍存在"
        ALL_CLEAN=false
    else
        echo -e "  ${GREEN}✅${NC} $dir 已删除"
    fi
done

# 7.4 验证网桥
echo ""
echo "[4] 网络网桥检查:"
if ip addr | grep -q "docker0"; then
    error "  ❌ docker0 网桥仍存在"
    ALL_CLEAN=false
else
    echo -e "  ${GREEN}✅${NC} docker0 已删除"
fi

# 7.5 验证端口监听
echo ""
echo "[5] 端口监听检查:"
if ss -tlnp 2>/dev/null | grep -qE "docker|containerd"; then
    warn "  ⚠️ 发现相关端口仍在监听:"
    ss -tlnp | grep -E "docker|containerd" || true
else
    echo -e "  ${GREEN}✅${NC} 无相关端口监听"
fi

# 7.6 验证 socket 文件
echo ""
echo "[6] Socket 文件检查:"
for sock in /run/docker.sock /var/run/docker.sock /run/containerd/containerd.sock; do
    if [ -e "$sock" ]; then
        error "  ❌ $sock 仍存在"
        ALL_CLEAN=false
    else
        echo -e "  ${GREEN}✅${NC} $sock 已删除"
    fi
done

echo ""
info "========== 验证结束 =========="
if [ "$ALL_CLEAN" = true ]; then
    info "${GREEN}所有组件已彻底清理完成！${NC}"
    info "建议执行 'reboot' 重启系统，确保内核模块完全释放。"
else
    warn "${YELLOW}发现残留项，请根据上方 ❌ 提示手动处理。${NC}"
fi