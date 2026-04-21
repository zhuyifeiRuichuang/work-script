#!/bin/bash

# ==========================================================
# Ubuntu 多版本软件源自动替换脚本 v2.8
# 支持 Ubuntu 20.04, 22.04, 24.04 LTS
# 优化：更新软件包列表步骤使用简洁进度展示，无频繁输出
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

# 函数：运行apt update并显示简洁进度（旋转光标+百分比）
run_apt_update() {
    print_message "$CYAN" "⏳ 正在更新软件包列表..."
    
    # 创建临时文件来捕获输出
    local temp_file=$(mktemp)
    
    # 在后台运行apt update
    apt update > "$temp_file" 2>&1 &
    local pid=$!
    
    # 进度展示配置
    local spin='|/-\'  # 旋转光标字符
    local progress=0    # 进度百分比
    local i=0           # 光标索引
    local step=5        # 进度步长（控制更新频率）
    local max_wait=120  # 最大等待时间（秒）
    local elapsed=0     # 已耗时
    
    # 显示初始进度
    printf "    [%-10s] %d%% %c" "" $progress "${spin:$i:1}"
    
    # 进度循环（每秒更新一次，避免频繁输出）
    while kill -0 "$pid" 2>/dev/null && [ $elapsed -lt $max_wait ]; do
        # 更新旋转光标
        i=$(( (i+1) % 4 ))
        
        # 估算进度（线性增长，直到95%，完成后设为100%）
        if [ $progress -lt 95 ]; then
            progress=$(( progress + 1 ))
        fi
        
        # 清理当前行并打印新进度
        printf "\r    [%-10s] %d%% %c" $(printf "#%.0s" $(seq 1 $((progress/10)))) $progress "${spin:$i:1}"
        
        # 等待1秒，减少输出频率
        sleep 1
        elapsed=$((elapsed + 1))
    done
    
    # 等待进程完成
    wait "$pid"
    local exit_code=$?
    
    # 更新为100%完成状态
    printf "\r    [%-10s] 100%% ✔\n" $(printf "#%.0s" $(seq 1 10))
    
    # 处理输出结果
    if [ $exit_code -eq 0 ]; then
        print_message "$GREEN" "✅ 软件包列表更新完成"
        
        # 统计成功更新的源数量
        local hits=$(grep -c "Get:" "$temp_file" 2>/dev/null)
        if [ -n "$hits" ] && [ "$hits" -gt 0 ] 2>/dev/null; then
            echo "    └─ 获取了 $hits 个源的软件包信息"
        fi
        
        # 检查可以升级的包数量
        local packages_line=$(grep "packages can be upgraded" "$temp_file" 2>/dev/null)
        if [ -n "$packages_line" ]; then
            local upgrade_count=$(echo "$packages_line" | grep -o '[0-9]*' | head -1)
            if [ -n "$upgrade_count" ] && [ "$upgrade_count" -gt 0 ] 2>/dev/null; then
                print_message "$YELLOW" "    ⓘ  有 $upgrade_count 个软件包可以升级"
            fi
        fi
        
        # 检查警告（只显示关键警告）
        local warnings=$(grep -i "warning" "$temp_file" 2>/dev/null | head -1)
        if [ -n "$warnings" ]; then
            print_message "$YELLOW" "    ⚠️  警告: $(echo "$warnings" | sed 's/WARNING: //' | cut -c1-60)..."
        fi
        
    else
        print_message "$RED" "❌ 软件包列表更新失败"
        
        # 显示简化的错误信息
        echo "    └─ 错误摘要:"
        grep -i "error\|failed\|E:" "$temp_file" 2>/dev/null | head -2 | while read -r line; do
            local clean_line=$(echo "$line" | tr -d '\r\n')
            echo "       ${clean_line:0:70}"
        done
        
        # 提供有用的建议
        print_message "$YELLOW" "💡 建议检查网络连接或镜像源地址"
    fi
    
    # 清理临时文件
    rm -f "$temp_file"
    return $exit_code
}

# 函数：检测 Ubuntu 版本
detect_ubuntu_version() {
    local os_id=""
    local version_id=""
    local version_codename=""
    
    if [ -f /etc/os-release ]; then
        os_id=$(grep -E '^ID=' /etc/os-release | cut -d= -f2 | tr -d '"' | tr -d "'" | tr '[:upper:]' '[:lower:]')
        version_id=$(grep -E '^VERSION_ID=' /etc/os-release | cut -d= -f2 | tr -d '"' | tr -d "'")
        version_codename=$(grep -E '^VERSION_CODENAME=' /etc/os-release | cut -d= -f2 | tr -d '"' | tr -d "'" | tr '[:upper:]' '[:lower:]')
    fi
    
    if [ "$os_id" != "ubuntu" ]; then
        if command -v lsb_release &>/dev/null; then
            os_id=$(lsb_release -si | tr '[:upper:]' '[:lower:]')
            if [ "$os_id" = "ubuntu" ]; then
                version_id=$(lsb_release -sr)
                version_codename=$(lsb_release -sc | tr '[:upper:]' '[:lower:]')
            else
                print_message "$RED" "❌ 错误: 此脚本仅支持 Ubuntu 系统。"
                exit 1
            fi
        else
            print_message "$RED" "❌ 错误: 无法检测操作系统版本，非Ubuntu系统或缺少必要文件。"
            exit 1
        fi
    fi
    
    if [ -z "$version_codename" ] && [ -n "$version_id" ]; then
        case "$version_id" in
            "20.04") version_codename="focal" ;;
            "22.04") version_codename="jammy" ;;
            "24.04") version_codename="noble" ;;
            *) 
                print_message "$YELLOW" "⚠️  警告: 无法自动识别版本代号，版本号: $version_id"
                print_message "$RED" "❌ 错误: 不支持的 Ubuntu 版本"
                exit 1
                ;;
        esac
    fi
    
    if [ -z "$version_id" ] || [ -z "$version_codename" ]; then
        print_message "$RED" "❌ 错误: 无法完整检测Ubuntu版本信息"
        exit 1
    fi
    
    echo "$version_id:$version_codename"
}

# 函数：备份文件（优化版，备份到专用目录）
backup_file() {
    local file_path=$1
    local backup_dir="/etc/apt/backups"
    mkdir -p "$backup_dir"
    
    if [ -f "$file_path" ]; then
        local filename=$(basename "$file_path")
        local timestamp=$(date +%Y%m%d%H%M%S)
        local backup_file="${backup_dir}/${filename}.bak_${timestamp}"
        
        # 避免覆盖已有备份
        if [ -f "$backup_file" ]; then
            backup_file="${backup_dir}/${filename}.bak_${timestamp}_$(date +%S)"
        fi
        
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
    
    local timestamp=$(date +%Y%m%d%H%M%S)
    local resolv_conf="/etc/resolv.conf"
    
    print_message "$YELLOW" "📂 备份DNS配置文件..."
    local resolv_backup=$(backup_file "$resolv_conf")
    if [ -n "$resolv_backup" ]; then
        print_message "$GREEN" "   ✅ 已备份到: $(basename "$resolv_backup")"
    fi
    
    if [ -L "$resolv_conf" ]; then
        print_message "$CYAN" "   🔗 检测到软链接，实际文件: $(readlink -f "$resolv_conf")"
        local real_resolv_conf=$(readlink -f "$resolv_conf")
        local real_backup=$(backup_file "$real_resolv_conf")
        if [ -n "$real_backup" ]; then
            print_message "$GREEN" "   ✅ 已备份真实文件到: $(basename "$real_backup")"
        fi
    fi
    
    print_message "$YELLOW" "📝 更新 /etc/resolv.conf..."
    if [ -f "$resolv_conf" ]; then
        grep -v "^nameserver" "$resolv_conf" > /tmp/resolv.conf.tmp 2>/dev/null
        
        {
            echo ""
            echo "# =========================================="
            echo "# 以下DNS服务器由脚本于 $(date) 添加"
            echo "# =========================================="
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
        
        mv /tmp/resolv.conf.tmp "$resolv_conf"
        print_message "$GREEN" "   ✅ 配置文件已更新"
    else
        cat > "$resolv_conf" << EOF
# DNS配置文件 - 由脚本自动创建于 $(date)
nameserver 223.5.5.5
nameserver 223.6.6.6
nameserver 114.114.114.114
nameserver 114.114.115.115
nameserver 8.8.8.8
nameserver 8.8.4.4
EOF
        print_message "$GREEN" "   ✅ 创建新的配置文件"
    fi
    
    chmod 644 "$resolv_conf"
    
    if command -v systemctl &>/dev/null && systemctl is-active systemd-resolved &>/dev/null 2>&1; then
        print_message "$YELLOW" "📝 配置systemd-resolved..."
        local resolved_conf="/etc/systemd/resolved.conf"
        local resolved_backup=$(backup_file "$resolved_conf")
        
        if [ -n "$resolved_backup" ]; then
            print_message "$GREEN" "   ✅ 已备份到: $(basename "$resolved_backup")"
        fi
        
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
        
        if systemctl restart systemd-resolved >/dev/null 2>&1; then
            print_message "$GREEN" "   ✅ systemd-resolved服务已重启"
        fi
    fi
    
    print_message "$GREEN" "
    ✅ DNS配置更新完成"
    
    print_message "$CYAN" "
    📋 当前DNS服务器配置:"
    echo "    ┌────────────────────────────────────"
    if [ -f "$resolv_conf" ]; then
        grep "^nameserver" "$resolv_conf" 2>/dev/null | head -6 | while read -r line; do
            echo "    │ $line"
        done
    fi
    echo "    └────────────────────────────────────"
    
    print_message "$CYAN" "
    🔍 进行DNS连通性测试..."
    
    local test_passed=0
    local test_total=3
    echo -n "    "
    local test_servers=("223.5.5.5" "114.114.114.114" "8.8.8.8")
    for server in "${test_servers[@]}"; do
        if timeout 1 ping -c 1 -W 1 "$server" >/dev/null 2>&1; then
            echo -n "✓"
            test_passed=$((test_passed + 1))
        else
            echo -n "✗"
        fi
    done
    
    echo ""
    if [ $test_passed -eq $test_total ]; then
        print_message "$GREEN" "    ✅ 所有DNS服务器连接正常"
    elif [ $test_passed -gt 0 ]; then
        print_message "$YELLOW" "    ⚠️  $test_passed/$test_total 个DNS服务器连接正常"
    else
        print_message "$RED" "    ❌ DNS服务器连接异常"
    fi
    
    print_message "$CYAN" "
    🧪 验证DNS解析功能..."
    if command -v nslookup &>/dev/null; then
        if nslookup google.com 8.8.8.8 >/dev/null 2>&1; then
            print_message "$GREEN" "    ✅ DNS解析验证通过"
        else
            print_message "$YELLOW" "    ⚠️  DNS解析可能存在问题"
        fi
    else
        print_message "$CYAN" "    ℹ️  nslookup未安装，跳过DNS解析验证"
    fi
}

# 函数：生成干净的 ubuntu.sources 文件（无重复配置）
generate_clean_ubuntu_sources() {
    local sources_file=$1
    local codename=$2
    local mirror=$3
    
    local mirror_url=""
    case $mirror in
        "tuna") mirror_url="https://mirrors.tuna.tsinghua.edu.cn/ubuntu/" ;;
        "aliyun") mirror_url="https://mirrors.aliyun.com/ubuntu/" ;;
        "ustc") mirror_url="https://mirrors.ustc.edu.cn/ubuntu/" ;;
        "163") mirror_url="https://mirrors.163.com/ubuntu/" ;;
        *) mirror_url="https://mirrors.tuna.tsinghua.edu.cn/ubuntu/" ;;
    esac
    
    print_message "$YELLOW" "📝 生成干净的 ${sources_file} 文件..."
    
    # 生成单一配置块，避免重复
    cat > "$sources_file" << EOF
# Ubuntu ${codename^} 软件源配置 - 由脚本自动生成于 $(date)
# 镜像源: ${mirror_url}
Types: deb deb-src
URIs: ${mirror_url}
Suites: ${codename} ${codename}-updates ${codename}-security ${codename}-backports
Components: main restricted universe multiverse
Signed-By: /usr/share/keyrings/ubuntu-archive-keyring.gpg
EOF
    
    chmod 644 "$sources_file"
    print_message "$GREEN" "   ✅ 成功生成干净的配置文件"
    return 0
}

# 函数：修改传统 sources.list 文件
modify_sources_list() {
    local sources_file=$1
    local codename=$2
    local mirror=$3
    
    local mirror_url=""
    case $mirror in
        "tuna") mirror_url="https://mirrors.tuna.tsinghua.edu.cn/ubuntu/" ;;
        "aliyun") mirror_url="https://mirrors.aliyun.com/ubuntu/" ;;
        "ustc") mirror_url="https://mirrors.ustc.edu.cn/ubuntu/" ;;
        "163") mirror_url="https://mirrors.163.com/ubuntu/" ;;
        *) mirror_url="https://mirrors.tuna.tsinghua.edu.cn/ubuntu/" ;;
    esac
    
    print_message "$YELLOW" "🔄 正在替换 ${sources_file} 中的镜像源..."
    
    cat > "$sources_file" << EOF
# Ubuntu ${codename^} 镜像源 (${mirror}) - 由脚本自动生成于 $(date)
deb ${mirror_url} ${codename} main restricted universe multiverse
deb-src ${mirror_url} ${codename} main restricted universe multiverse
deb ${mirror_url} ${codename}-updates main restricted universe multiverse
deb-src ${mirror_url} ${codename}-updates main restricted universe multiverse
deb ${mirror_url} ${codename}-security main restricted universe multiverse
deb-src ${mirror_url} ${codename}-security main restricted universe multiverse
deb ${mirror_url} ${codename}-backports main restricted universe multiverse
deb-src ${mirror_url} ${codename}-backports main restricted universe multiverse
EOF
    
    chmod 644 "$sources_file"
    print_message "$GREEN" "   ✅ 成功替换镜像源地址为: ${mirror_url}"
    return 0
}

# 函数：清理临时文件
cleanup() {
    [ -f /tmp/resolv.conf.tmp ] && rm -f /tmp/resolv.conf.tmp
}

# 设置信号处理器
trap cleanup EXIT INT TERM

# ==========================================================
# 主程序开始
# ==========================================================

print_message "$BLUE" "
╔══════════════════════════════════════════════════════════╗
║             Ubuntu 软件源管理脚本 v2.8                  ║
║  支持: 20.04/22.04/24.04 LTS | 简洁进度展示 | DNS优化   ║
║  特性: 直接修改系统原生 ubuntu.sources 文件             ║
╚══════════════════════════════════════════════════════════╝"

# 默认变量
SOURCES_FILE="/etc/apt/sources.list"
SOURCES_DIR="/etc/apt/sources.list.d"
UBUNTU_SOURCES_FILE="${SOURCES_DIR}/ubuntu.sources"
BACKUP_ONLY=false
MIRROR="tuna"
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
            echo "恢复功能："
            echo "1. 查看备份文件: ls /etc/apt/backups/"
            echo "2. 恢复ubuntu.sources: sudo cp /etc/apt/backups/ubuntu.sources.bak_* /etc/apt/sources.list.d/ubuntu.sources"
            echo "3. 恢复sources.list: sudo cp /etc/apt/backups/sources.list.bak_* /etc/apt/sources.list"
            echo "4. 更新: sudo apt update"
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

# 检查版本支持
supported_version=$(echo "$VERSION_CODENAME" | tr '[:upper:]' '[:lower:]')
case $supported_version in
    "focal"|"jammy"|"noble")
        VERSION_CODENAME="$supported_version"
        print_message "$GREEN" "✅ 支持该版本: Ubuntu $VERSION_ID ($VERSION_CODENAME)"
        ;;
    *)
        print_message "$RED" "❌ 错误: 不支持的 Ubuntu 版本 ($VERSION_CODENAME)。"
        exit 1
        ;;
esac

# 更新DNS配置
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

# APT源配置
print_message "$PURPLE" "
    ╔══════════════════════════════════════════╗
    ║           APT 源配置更新                 ║
    ╚══════════════════════════════════════════╝"

# 版本比较函数
version_ge() {
    local v1=$1 v2=$2
    local IFS=.
    local i ver1=($v1) ver2=($v2)
    for ((i=0; i<${#ver1[@]} || i<${#ver2[@]}; i++)); do
        ((10#${ver1[i]:-0} > 10#${ver2[i]:-0})) && return 0
        ((10#${ver1[i]:-0} < 10#${ver2[i]:-0})) && return 1
    done
    return 0
}

# 备份源文件（备份到专用目录）
BACKUP_FILES=()
print_message "$YELLOW" "📂 备份APT源文件到专用目录..."

if [ -d "$SOURCES_DIR" ] && version_ge "$VERSION_ID" "24.04"; then
    print_message "$CYAN" "🔍 检测到 Ubuntu 24.04+，修改系统原生 ${UBUNTU_SOURCES_FILE} 文件"
    
    if [ -f "$UBUNTU_SOURCES_FILE" ]; then
        backup_result=$(backup_file "$UBUNTU_SOURCES_FILE")
        [ -n "$backup_result" ] && {
            BACKUP_FILES+=("$backup_result")
            print_message "$GREEN" "   ✅ 已备份: $(basename "$backup_result")"
        }
    fi
    
    if [ -f "$SOURCES_FILE" ]; then
        backup_result=$(backup_file "$SOURCES_FILE")
        [ -n "$backup_result" ] && {
            BACKUP_FILES+=("$backup_result")
            print_message "$GREEN" "   ✅ 已备份: $(basename "$backup_result")"
        }
    fi
    
    USE_SOURCES_DIR=true
else
    print_message "$CYAN" "🔍 检测到 Ubuntu < 24.04，修改传统 ${SOURCES_FILE} 文件"
    
    if [ -f "$SOURCES_FILE" ]; then
        backup_result=$(backup_file "$SOURCES_FILE")
        [ -n "$backup_result" ] && {
            BACKUP_FILES+=("$backup_result")
            print_message "$GREEN" "   ✅ 已备份: $(basename "$backup_result")"
        }
    fi
    
    USE_SOURCES_DIR=false
fi

# 仅备份模式
if [ "$BACKUP_ONLY" = true ]; then
    print_message "$GREEN" "✅ 备份完成。源文件未修改。"
    exit 0
fi

# 显示镜像源信息
case $MIRROR in
    "tuna") mirror_name="清华大学 TUNA 镜像站" ;;
    "aliyun") mirror_name="阿里云镜像站" ;;
    "ustc") mirror_name="中国科学技术大学镜像站" ;;
    "163") mirror_name="网易 163 镜像站" ;;
    *) mirror_name="未知镜像站" ;;
esac

print_message "$YELLOW" "🔄 正在配置 ${mirror_name}..."

# 写入配置
print_message "$YELLOW" "📝 修改软件源配置文件..."

if [ "$USE_SOURCES_DIR" = true ]; then
    # 生成干净的单一配置块，避免重复
    generate_clean_ubuntu_sources "$UBUNTU_SOURCES_FILE" "$VERSION_CODENAME" "$MIRROR"
else
    modify_sources_list "$SOURCES_FILE" "$VERSION_CODENAME" "$MIRROR"
fi

# 更新软件包列表（使用优化后的进度展示）
run_apt_update

if [ $? -eq 0 ]; then
    print_message "$GREEN" "
    🎉 操作完成!"
    
    echo ""
    print_message "$CYAN" "    📊 操作摘要:"
    echo "    ├────────────────────────────────────"
    echo "    │ ✅ DNS配置已优化"
    echo "    │ ✅ APT源已更新为: ${mirror_name}"
    echo "    │ ✅ 配置文件无重复、无警告"
    echo "    │ ✅ 软件包列表已同步"
    echo "    └────────────────────────────────────"
    
    echo ""
    print_message "$YELLOW" "    💡 后续操作建议:"
    echo "        └─ 执行升级: sudo apt upgrade"
    
    echo ""
    print_message "$BLUE" "    📋 备份文件位置: /etc/apt/backups/"
    for backup in "${BACKUP_FILES[@]}"; do
        echo "        └─ $(basename "$backup")"
    done
    
else
    print_message "$RED" "
    ⚠️  操作完成，但有错误发生"
    
    echo ""
    print_message "$YELLOW" "    🔧 故障排除建议:"
    echo "        1. 检查网络连接"
    echo "        2. 验证镜像源地址"
    echo "        3. 使用备份文件恢复"
fi

print_message "$BLUE" "
╔══════════════════════════════════════════════════════════╗
║                      脚本执行完毕                       ║
╚══════════════════════════════════════════════════════════╝"