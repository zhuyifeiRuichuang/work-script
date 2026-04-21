#!/bin/bash

# ==========================================================
# Ubuntu 多版本软件源自动替换脚本
# 支持 Ubuntu 20.04, 22.04, 24.04 LTS
# 包含DNS客户端配置更新功能
# ==========================================================

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# 函数：打印带颜色的消息
print_message() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# 函数：运行apt update并显示简洁输出
run_apt_update() {
    print_message "$CYAN" "⏳ 正在更新软件包列表..."
    
    # 创建临时文件来捕获输出
    local temp_file=$(mktemp)
    
    # 在后台运行apt update
    apt update > "$temp_file" 2>&1 &
    local pid=$!
    
    # 显示简单的等待提示
    echo -n "    "
    local dots=0
    while kill -0 "$pid" 2>/dev/null; do
        echo -n "."
        dots=$((dots + 1))
        if [ $dots -ge 10 ]; then
            echo ""
            echo -n "    "
            dots=0
        fi
        sleep 0.3
    done
    
    echo ""  # 换行
    
    # 等待进程完成
    wait "$pid"
    local exit_code=$?
    
    # 处理输出
    if [ $exit_code -eq 0 ]; then
        print_message "$GREEN" "✅ 软件包列表更新完成"
        
        # 统计成功更新的源数量（安全地处理可能为空的情况）
        local hits=$(grep -c "Get:" "$temp_file" 2>/dev/null)
        if [ -n "$hits" ] && [ "$hits" -gt 0 ] 2>/dev/null; then
            echo "    └─ 获取了 $hits 个源的软件包信息"
        fi
        
        # 检查可以升级的包数量
        local packages_line=$(grep "packages can be upgraded" "$temp_file" 2>/dev/null)
        if [ -n "$packages_line" ]; then
            # 安全地提取数字
            local upgrade_count=$(echo "$packages_line" | grep -o '[0-9]*' | head -1)
            if [ -n "$upgrade_count" ] && [ "$upgrade_count" -gt 0 ] 2>/dev/null; then
                print_message "$YELLOW" "    ⓘ  有 $upgrade_count 个软件包可以升级"
            fi
        fi
        
        # 检查是否有警告
        local warnings=$(grep -i "warning" "$temp_file" 2>/dev/null | head -2)
        if [ -n "$warnings" ]; then
            print_message "$YELLOW" "    ⚠️  发现警告信息:"
            echo "$warnings" | while read -r line; do
                # 清理警告信息，限制长度
                local clean_line=$(echo "$line" | sed 's/WARNING: //' | tr -d '\r\n')
                echo "       ${clean_line:0:60}..."
            done
        fi
        
    else
        print_message "$RED" "❌ 软件包列表更新失败"
        
        # 显示简化的错误信息
        echo "    └─ 错误摘要:"
        grep -i "error\|failed\|E:" "$temp_file" 2>/dev/null | head -3 | while read -r line; do
            local clean_line=$(echo "$line" | tr -d '\r\n')
            echo "       ${clean_line:0:80}"
        done
        
        # 提供有用的建议
        echo ""
        print_message "$YELLOW" "💡 建议检查:"
        echo "    1. 网络连接是否正常"
        echo "    2. 镜像源地址是否正确"
        echo "    3. 使用 '$0 --restore' 恢复备份"
    fi
    
    rm -f "$temp_file"
    return $exit_code
}

# 函数：检测 Ubuntu 版本
detect_ubuntu_version() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        if [ "$ID" = "ubuntu" ]; then
            # 使用 VERSION_CODENAME 变量，它通常包含简写代号
            if [ -n "$VERSION_CODENAME" ]; then
                echo "$VERSION_ID:$VERSION_CODENAME"
            else
                # 如果没有 VERSION_CODENAME，从 VERSION 中提取
                local codename=$(echo "$VERSION" | grep -oP '(?<=\()\w+(?=\))' | head -1)
                echo "$VERSION_ID:$codename"
            fi
        else
            print_message "$RED" "❌ 错误: 此脚本仅支持 Ubuntu 系统。"
            exit 1
        fi
    else
        print_message "$RED" "❌ 错误: 无法检测操作系统版本。"
        exit 1
    fi
}

# 函数：备份文件
backup_file() {
    local file_path=$1
    local backup_suffix=$2
    
    if [ -f "$file_path" ]; then
        local backup_file="${file_path}.bak_${backup_suffix}"
        cp -f "$file_path" "$backup_file"
        echo "$backup_file"
    else
        echo ""
    fi
}

# 函数：更新DNS客户端配置
update_dns_config() {
    print_message "$PURPLE" "
    ╔══════════════════════════════════════════╗
    ║           DNS 配置更新                   ║
    ╚══════════════════════════════════════════╝"
    
    # 备份时间戳
    local timestamp=$(date +%Y%m%d%H%M%S)
    
    # 主要的DNS配置文件
    local resolv_conf="/etc/resolv.conf"
    
    # 备份原始resolv.conf
    print_message "$YELLOW" "📂 备份DNS配置文件..."
    local resolv_backup=$(backup_file "$resolv_conf" "$timestamp")
    if [ -n "$resolv_backup" ]; then
        print_message "$GREEN" "   ✅ 已备份到: $(basename "$resolv_backup")"
    fi
    
    # 检查resolv.conf的类型（是否是软链接）
    if [ -L "$resolv_conf" ]; then
        print_message "$CYAN" "   🔗 检测到软链接，实际文件: $(readlink -f "$resolv_conf")"
        local real_resolv_conf=$(readlink -f "$resolv_conf")
        local real_backup=$(backup_file "$real_resolv_conf" "$timestamp")
        if [ -n "$real_backup" ]; then
            print_message "$GREEN" "   ✅ 已备份真实文件到: $(basename "$real_backup")"
        fi
    fi
    
    # 定义DNS服务器列表（精选最稳定的几个）
    local dns_servers=(
        "223.5.5.5"         # 阿里云 DNS（国内推荐）
        "223.6.6.6"         # 阿里云 DNS
        "114.114.114.114"   # 114 DNS（国内稳定）
        "114.114.115.115"   # 114 DNS
        "8.8.8.8"           # Google DNS（国际）
        "8.8.4.4"           # Google DNS
    )
    
    # 方法1: 追加DNS服务器到resolv.conf（保留原有配置）
    print_message "$YELLOW" "📝 更新 /etc/resolv.conf..."
    
    # 先保留原始文件内容（除了nameserver行）
    if [ -f "$resolv_conf" ]; then
        # 创建临时文件，保留原始注释和其他配置
        grep -v "^nameserver" "$resolv_conf" > /tmp/resolv.conf.tmp 2>/dev/null
        
        # 添加分隔线和说明
        {
            echo ""
            echo "# =========================================="
            echo "# 以下DNS服务器由脚本于 $(date) 添加"
            echo "# =========================================="
        } >> /tmp/resolv.conf.tmp
        
        # 添加DNS服务器
        {
            echo "# 阿里云 DNS (国内推荐)"
            echo "nameserver 223.5.5.5"
            echo "nameserver 223.6.6.6"
            echo ""
            echo "# 114 DNS (国内备用)"
            echo "nameserver 114.114.114.114"
            echo "nameserver 114.114.115.115"
            echo ""
            echo "# Google DNS (国际备用)"
            echo "nameserver 8.8.8.8"
            echo "nameserver 8.8.4.4"
        } >> /tmp/resolv.conf.tmp
        
        # 移动临时文件到目标位置
        mv /tmp/resolv.conf.tmp "$resolv_conf"
        print_message "$GREEN" "   ✅ 配置文件已更新"
    else
        # 如果文件不存在，创建新的
        cat > "$resolv_conf" << EOF
# DNS配置文件 - 由脚本自动创建于 $(date)

# ==========================================
# 主要DNS服务器配置
# ==========================================

# 阿里云 DNS (国内推荐，速度快)
nameserver 223.5.5.5
nameserver 223.6.6.6

# 114 DNS (国内稳定备用)
nameserver 114.114.114.114
nameserver 114.114.115.115

# Google DNS (国际访问)
nameserver 8.8.8.8
nameserver 8.8.4.4
EOF
        print_message "$GREEN" "   ✅ 创建新的配置文件"
    fi
    
    # 设置正确的权限
    chmod 644 "$resolv_conf"
    
    # 方法2: 配置systemd-resolved (如果可用)
    if command -v systemctl &> /dev/null && systemctl is-active systemd-resolved &> /dev/null 2>&1; then
        print_message "$YELLOW" "📝 配置systemd-resolved..."
        local resolved_conf="/etc/systemd/resolved.conf"
        local resolved_backup=$(backup_file "$resolved_conf" "$timestamp")
        
        if [ -n "$resolved_backup" ]; then
            print_message "$GREEN" "   ✅ 已备份到: $(basename "$resolved_backup")"
        fi
        
        # 更新或创建配置
        cat > "$resolved_conf" << EOF
[Resolve]
DNS=223.5.5.5 223.6.6.6 114.114.114.114
FallbackDNS=8.8.8.8 8.8.4.4
Domains=~.
DNSSEC=allow-downgrade
DNSOverTLS=opportunistic
Cache=yes
DNSStubListener=yes
ReadEtcHosts=yes
EOF
        
        # 重启systemd-resolved服务
        if systemctl restart systemd-resolved > /dev/null 2>&1; then
            print_message "$GREEN" "   ✅ systemd-resolved服务已重启"
        fi
    fi
    
    # 备份其他可能的DNS配置文件
    print_message "$YELLOW" "📂 备份其他DNS相关配置文件..."
    
    local dns_configs_to_backup=(
        "/etc/resolvconf/resolv.conf.d/base"
        "/etc/resolvconf/resolv.conf.d/head"
        "/etc/resolvconf/resolv.conf.d/original"
        "/etc/dhcp/dhclient.conf"
    )
    
    local backup_count=0
    for config_file in "${dns_configs_to_backup[@]}"; do
        if [ -f "$config_file" ]; then
            local backup_file_result=$(backup_file "$config_file" "$timestamp")
            if [ -n "$backup_file_result" ]; then
                backup_count=$((backup_count + 1))
            fi
        fi
    done
    
    if [ $backup_count -gt 0 ]; then
        print_message "$GREEN" "   ✅ 备份了 $backup_count 个相关配置文件"
    fi
    
    print_message "$GREEN" "
    ✅ DNS配置更新完成"
    
    # 显示更新后的DNS配置摘要
    print_message "$CYAN" "
    📋 当前DNS服务器配置:"
    echo "    ┌────────────────────────────────────"
    if [ -f "$resolv_conf" ]; then
        grep "^nameserver" "$resolv_conf" 2>/dev/null | head -6 | while read -r line; do
            echo "    │ $line"
        done
    fi
    echo "    └────────────────────────────────────"
    
    # 测试DNS解析
    print_message "$CYAN" "
    🔍 进行DNS连通性测试..."
    
    local test_passed=0
    local test_total=3
    
    echo -n "    "
    # 测试服务器连通性
    local test_servers=("223.5.5.5" "114.114.114.114" "8.8.8.8")
    for server in "${test_servers[@]}"; do
        if timeout 1 ping -c 1 -W 1 "$server" > /dev/null 2>&1; then
            echo -n "✓"
            test_passed=$((test_passed + 1))
        else
            echo -n "✗"
        fi
    done
    
    echo ""  # 换行
    
    if [ $test_passed -eq $test_total ]; then
        print_message "$GREEN" "    ✅ 所有DNS服务器连接正常"
    elif [ $test_passed -gt 0 ]; then
        print_message "$YELLOW" "    ⚠️  $test_passed/$test_total 个DNS服务器连接正常"
        echo "    └─ 部分DNS服务器不可达，但DNS解析可能仍然可用"
    else
        print_message "$RED" "    ❌ DNS服务器连接异常"
        echo "    └─ 请检查网络连接或防火墙设置"
    fi
    
    # 验证DNS解析功能
    print_message "$CYAN" "
    🧪 验证DNS解析功能..."
    if command -v nslookup &> /dev/null; then
        if nslookup google.com 8.8.8.8 > /dev/null 2>&1; then
            print_message "$GREEN" "    ✅ DNS解析验证通过"
        else
            print_message "$YELLOW" "    ⚠️  DNS解析可能存在问题"
        fi
    else
        print_message "$CYAN" "    ℹ️  nslookup未安装，跳过DNS解析验证"
    fi
}

# 函数：获取对应版本的源内容
get_sources_content() {
    local codename=$1
    local mirror=${2:-"tuna"}  # 默认使用清华源
    
    local mirror_url=""
    case $mirror in
        "tuna")
            mirror_url="https://mirrors.tuna.tsinghua.edu.cn/ubuntu/"
            ;;
        "aliyun")
            mirror_url="https://mirrors.aliyun.com/ubuntu/"
            ;;
        "ustc")
            mirror_url="https://mirrors.ustc.edu.cn/ubuntu/"
            ;;
        "163")
            mirror_url="https://mirrors.163.com/ubuntu/"
            ;;
        *)
            mirror_url="https://mirrors.tuna.tsinghua.edu.cn/ubuntu/"
            ;;
    esac
    
    cat << EOF
# Ubuntu ${codename^} 镜像源 (${mirror})
deb ${mirror_url} ${codename} main restricted universe multiverse
deb-src ${mirror_url} ${codename} main restricted universe multiverse

# 稳定版更新源
deb ${mirror_url} ${codename}-updates main restricted universe multiverse
deb-src ${mirror_url} ${codename}-updates main restricted universe multiverse

# 安全更新源
deb ${mirror_url} ${codename}-security main restricted universe multiverse
deb-src ${mirror_url} ${codename}-security main restricted universe multiverse

# 反向移植源
deb ${mirror_url} ${codename}-backports main restricted universe multiverse
deb-src ${mirror_url} ${codename}-backports main restricted universe multiverse
EOF
}

# 函数：清理临时文件
cleanup() {
    if [ -f /tmp/resolv.conf.tmp ]; then
        rm -f /tmp/resolv.conf.tmp
    fi
}

# 设置信号处理器
trap cleanup EXIT INT TERM

# ==========================================================
# 主程序开始
# ==========================================================

print_message "$BLUE" "
╔══════════════════════════════════════════════════════════╗
║             Ubuntu 软件源管理脚本 v2.1                  ║
║      支持: 20.04/22.04/24.04 | APT源更新 | DNS优化     ║
╚══════════════════════════════════════════════════════════╝"

# 默认变量
SOURCES_FILE="/etc/apt/sources.list"
BACKUP_ONLY=false
MIRROR="tuna"
RESTORE=false
UPDATE_DNS=true
DNS_ONLY=false

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            echo "用法: $0 [选项]"
            echo "选项:"
            echo "  -h, --help          显示帮助信息"
            echo "  -m, --mirror MIRROR 指定镜像源 (tuna, aliyun, ustc, 163)"
            echo "  -b, --backup-only   仅备份，不修改"
            echo "  --no-dns            跳过DNS配置"
            echo "  --dns-only          仅更新DNS"
            exit 0
            ;;
        -m|--mirror)
            MIRROR="$2"
            shift 2
            ;;
        -b|--backup-only)
            BACKUP_ONLY=true
            shift
            ;;
        -r|--restore)
            echo "恢复功能已简化，请手动操作:"
            echo "1. 查看备份文件: ls /etc/apt/sources.list.bak_*"
            echo "2. 恢复命令: cp /etc/apt/sources.list.bak_时间戳 /etc/apt/sources.list"
            echo "3. 更新: apt update"
            exit 0
            ;;
        --no-dns)
            UPDATE_DNS=false
            shift
            ;;
        --dns-only)
            DNS_ONLY=true
            UPDATE_DNS=true
            shift
            ;;
        *)
            print_message "$RED" "❌ 未知选项: $1"
            echo "使用 $0 -h 查看帮助"
            exit 1
            ;;
    esac
done

# 检查 root 权限
if [ "$EUID" -ne 0 ]; then
    print_message "$RED" "❌ 错误: 此脚本需要 root 权限才能执行。"
    echo "请使用 'sudo $0' 运行。"
    exit 1
fi

# 检测系统版本
print_message "$YELLOW" "🔍 正在检测系统版本..."
version_info=$(detect_ubuntu_version)
VERSION_ID=$(echo "$version_info" | cut -d: -f1)
VERSION_CODENAME=$(echo "$version_info" | cut -d: -f2)

print_message "$GREEN" "✅ 检测到 Ubuntu $VERSION_ID ($VERSION_CODENAME)"

# 检查是否支持该版本（不区分大小写）
supported_version=$(echo "$VERSION_CODENAME" | tr '[:upper:]' '[:lower:]')
case $supported_version in
    "focal"|"jammy"|"noble")
        # 统一使用小写版本代号
        VERSION_CODENAME="$supported_version"
        print_message "$GREEN" "✅ 支持该版本: Ubuntu $VERSION_ID ($VERSION_CODENAME)"
        ;;
    *)
        print_message "$RED" "❌ 错误: 不支持的 Ubuntu 版本 ($VERSION_CODENAME)。"
        print_message "$YELLOW" "仅支持 Ubuntu 20.04 (focal), 22.04 (jammy), 24.04 (noble)"
        exit 1
        ;;
esac

# ==================== 第一步：更新DNS配置 ====================
if [ "$UPDATE_DNS" = true ]; then
    update_dns_config
    echo ""
else
    print_message "$YELLOW" "⏭️  跳过DNS配置更新"
fi

# 如果只需要更新DNS，则退出
if [ "$DNS_ONLY" = true ]; then
    print_message "$GREEN" "
    ✅ DNS配置更新完成，脚本结束。"
    exit 0
fi

# ==================== 第二步：APT源配置 ====================
print_message "$PURPLE" "
    ╔══════════════════════════════════════════╗
    ║           APT 源配置更新                 ║
    ╚══════════════════════════════════════════╝"

# 备份源文件
BACKUP_FILE="${SOURCES_FILE}.bak_$(date +%Y%m%d%H%M%S)"
print_message "$YELLOW" "📂 备份APT源文件..."
cp -f "${SOURCES_FILE}" "${BACKUP_FILE}"
print_message "$GREEN" "   ✅ 已备份到: $(basename "$BACKUP_FILE")"

if [ "$BACKUP_ONLY" = true ]; then
    print_message "$GREEN" "✅ 备份完成。源文件未修改。"
    exit 0
fi

# 显示镜像源信息
case $MIRROR in
    "tuna")
        mirror_name="清华大学 TUNA 镜像站"
        ;;
    "aliyun")
        mirror_name="阿里云镜像站"
        ;;
    "ustc")
        mirror_name="中国科学技术大学镜像站"
        ;;
    "163")
        mirror_name="网易 163 镜像站"
        ;;
    *)
        mirror_name="未知镜像站"
        ;;
esac

print_message "$YELLOW" "🔄 正在配置 ${mirror_name}..."

# 获取并写入新的源内容
print_message "$YELLOW" "📝 写入新的软件源配置..."
get_sources_content "$VERSION_CODENAME" "$MIRROR" > "$SOURCES_FILE"
print_message "$GREEN" "   ✅ 配置已写入"

# 更新软件包列表（使用简洁输出）
run_apt_update

if [ $? -eq 0 ]; then
    print_message "$GREEN" "
    🎉 操作完成!"
    
    echo ""
    print_message "$CYAN" "    📊 操作摘要:"
    echo "    ├────────────────────────────────────"
    echo "    │ ✅ DNS配置已优化"
    echo "    │ ✅ APT源已更新为: ${mirror_name}"
    echo "    │ ✅ 软件包列表已同步"
    echo "    └────────────────────────────────────"
    
    echo ""
    print_message "$YELLOW" "    💡 后续操作建议:"
    echo "        └─ 执行升级: sudo apt upgrade"
    
    echo ""
    print_message "$BLUE" "    📋 备份文件位置:"
    echo "        └─ APT源: $(basename "$BACKUP_FILE")"
    
else
    print_message "$RED" "
    ⚠️  操作完成，但有错误发生"
    
    echo ""
    print_message "$YELLOW" "    🔧 故障排除建议:"
    echo "        1. 检查网络连接"
    echo "        2. 验证镜像源地址"
    echo "        3. 手动恢复备份文件"
fi

print_message "$BLUE" "
╔══════════════════════════════════════════════════════════╗
║                      脚本执行完毕                       ║
╚══════════════════════════════════════════════════════════╝"
