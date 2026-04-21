#!/bin/bash

# nerdctl自动安装脚本

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

# 检查是否以root用户运行
check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_warn "建议使用root权限运行此脚本"
        read -p "是否继续? (y/n): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
}

# 步骤1: 扫描当前目录下的tar.gz文件
scan_tar_file() {
    local tar_files=()
    
    # 安全地查找tar.gz文件
    while IFS= read -r -d '' file; do
        tar_files+=("$file")
    done < <(find . -maxdepth 1 -name "*.tar.gz" -print0)
    
    if [[ ${#tar_files[@]} -eq 0 ]]; then
        log_error "当前目录下未找到tar.gz文件"
        exit 1
    fi
    
    # 查找nerdctl相关的tar.gz文件
    local nerdctl_files=()
    for file in "${tar_files[@]}"; do
        if [[ "$(basename "$file")" == *nerdctl* ]]; then
            nerdctl_files+=("$file")
        fi
    done
    
    if [[ ${#nerdctl_files[@]} -eq 0 ]]; then
        log_error "未找到包含'nerdctl'的tar.gz文件"
        exit 1
    fi
    
    # 如果有多个文件，让用户选择
    if [[ ${#nerdctl_files[@]} -gt 1 ]]; then
        echo "找到多个nerdctl压缩文件:"
        for i in "${!nerdctl_files[@]}"; do
            echo "$((i+1)). $(basename "${nerdctl_files[$i]}")"
        done
        
        read -p "请选择要安装的文件序号 (1-${#nerdctl_files[@]}): " selection
        if [[ ! $selection =~ ^[0-9]+$ ]] || (( selection < 1 || selection > ${#nerdctl_files[@]} )); then
            log_error "无效的选择"
            exit 1
        fi
        
        TAR_FILE="${nerdctl_files[$((selection-1))]}"
    else
        TAR_FILE="${nerdctl_files[0]}"
    fi
    
    # 移除开头的 "./"
    TAR_FILE="${TAR_FILE#./}"
    log_info "找到安装文件: $TAR_FILE"
    
    # 提取版本信息
    VERSION_NAME="$(basename "$TAR_FILE")"
    VERSION_NAME="${VERSION_NAME%.tar.gz}"
    VERSION_NAME="${VERSION_NAME#nerdctl-}"
    
    log_info "安装版本: $VERSION_NAME"
}

# 步骤2: 检测操作系统
detect_os() {
    if [[ "$(uname)" != "Linux" ]]; then
        log_error "仅支持Linux操作系统。"
        exit 1
    fi
    
    # 检测发行版
    if [[ -f /etc/os-release ]]; then
        # 使用source命令读取文件，避免变量污染
        if source /etc/os-release 2>/dev/null; then
            # 将OS信息保存到局部变量
            local os_name="$NAME"
            local os_version="$VERSION_ID"
            log_info "检测到操作系统: ${os_name} ${os_version}"
        else
            log_warn "无法读取/etc/os-release文件"
            OS="unknown"
        fi
    else
        log_warn "无法检测操作系统发行版，继续安装..."
    fi
}

# 步骤3: 安装nerdctl
install_nerdctl() {
    local install_dir="/usr/local/bin"
    
    log_info "开始安装nerdctl..."
    
    # 检查文件是否存在
    if [[ ! -f "$TAR_FILE" ]]; then
        log_error "文件 $TAR_FILE 不存在"
        exit 1
    fi
    
    # 显示文件大小
    file_size=$(du -h "$TAR_FILE" | cut -f1)
    log_info "文件大小: $file_size"
    
    # 解压文件
    log_info "解压文件到 $install_dir ..."
    
    # 先检查tar文件是否有效
    if ! tar -tzf "$TAR_FILE" >/dev/null 2>&1; then
        log_error "文件 $TAR_FILE 不是有效的tar.gz文件或已损坏"
        exit 1
    fi
    
    # 解压文件
    if tar -zxf "$TAR_FILE" -C "$install_dir"; then
        log_info "文件解压成功"
        
        # 确保文件有执行权限
        if [[ -f "$install_dir/nerdctl" ]]; then
            chmod +x "$install_dir/nerdctl"
            log_info "已设置执行权限"
        else
            # 检查解压后的目录结构
            log_warn "在 $install_dir 中未找到nerdctl二进制文件"
            log_warn "检查解压后的文件..."
            find "$install_dir" -name "nerdctl" -type f | head -5
        fi
        
        # 创建符号链接（可选）
        if [[ ! -f /usr/bin/nerdctl ]] && [[ -f "$install_dir/nerdctl" ]]; then
            ln -sf "$install_dir/nerdctl" /usr/bin/nerdctl 2>/dev/null && \
            log_info "已创建符号链接: /usr/bin/nerdctl"
        fi
    else
        log_error "文件解压失败"
        exit 1
    fi
}

# 步骤4: 配置命令自动补全
setup_completion() {
    log_info "配置命令自动补全..."
    
    # 创建补全目录
    mkdir -p ~/.bash_completion.d/
    
    # 检查nerdctl是否在PATH中
    if ! command -v nerdctl &> /dev/null; then
        # 尝试直接使用/usr/local/bin/nerdctl
        if [[ -f /usr/local/bin/nerdctl ]]; then
            /usr/local/bin/nerdctl completion bash > ~/.bash_completion.d/nerdctl 2>/dev/null
        else
            log_warn "nerdctl命令未找到，跳过补全配置"
            return 1
        fi
    else
        # 使用PATH中的nerdctl
        nerdctl completion bash > ~/.bash_completion.d/nerdctl 2>/dev/null
    fi
    
    if [[ $? -eq 0 ]] && [[ -s ~/.bash_completion.d/nerdctl ]]; then
        # 添加到bashrc
        if ! grep -q "bash_completion.d" ~/.bashrc; then
            echo 'for bcfile in ~/.bash_completion.d/* ; do . $bcfile; done' >> ~/.bashrc
            log_info "已将补全配置添加到 ~/.bashrc"
        fi
        
        # 立即生效
        if [[ -f ~/.bash_completion.d/nerdctl ]]; then
            source ~/.bash_completion.d/nerdctl 2>/dev/null
        fi
        
        log_info "命令补全配置完成"
    else
        log_warn "无法生成命令补全脚本，可能版本不支持"
        return 1
    fi
}

# 步骤5: 安装后检查
post_install_check() {
    log_info "安装后检查..."
    
    # 等待一下，确保系统更新PATH
    sleep 1
    
    echo "========================================"
    
    # 检查不同路径下的nerdctl
    local found=false
    
    if command -v nerdctl &> /dev/null; then
        found=true
        nerdctl_path="$(command -v nerdctl)"
        echo -e "${GREEN}✓ nerdctl 已安装到PATH中: $nerdctl_path${NC}"
    elif [[ -f /usr/local/bin/nerdctl ]]; then
        found=true
        echo -e "${GREEN}✓ nerdctl 已安装到: /usr/local/bin/nerdctl${NC}"
    elif [[ -f /usr/bin/nerdctl ]]; then
        found=true
        echo -e "${GREEN}✓ nerdctl 已安装到: /usr/bin/nerdctl${NC}"
    fi
    
    if $found; then
        echo "========================================"
        echo -e "${GREEN}nerdctl 安装成功!${NC}"
        echo "========================================"
        
        # 显示版本信息
        if [[ -x "$(command -v nerdctl || echo /usr/local/bin/nerdctl)" ]]; then
            "$(command -v nerdctl || echo /usr/local/bin/nerdctl)" --version
        fi
        
        echo "========================================"
        log_info "使用 'nerdctl --help' 查看帮助信息"
        
        # 检查是否在PATH中
        if ! command -v nerdctl &> /dev/null; then
            log_warn "nerdctl不在PATH中，可能需要重启终端或执行: source ~/.bashrc"
        fi
    else
        log_error "nerdctl 安装失败，请检查安装过程"
        
        # 检查可能的安装位置
        echo "检查可能的安装位置:"
        ls -la /usr/local/bin/nerdctl 2>/dev/null || echo "/usr/local/bin/nerdctl 不存在"
        ls -la /usr/bin/nerdctl 2>/dev/null || echo "/usr/bin/nerdctl 不存在"
        
        # 检查解压的文件
        echo "检查解压的文件:"
        find /usr/local/bin -name "*nerdctl*" -type f 2>/dev/null
        
        exit 1
    fi
}

# 显示帮助信息
show_help() {
    echo "用法: $0 [选项]"
    echo
    echo "选项:"
    echo "  -h, --help     显示此帮助信息"
    echo "  -v, --version  显示脚本版本"
    echo
    echo "说明:"
    echo "  此脚本会自动扫描当前目录下的nerdctl压缩包并安装"
    echo "  支持 tar.gz 格式的压缩包"
    echo
    echo "示例:"
    echo "  $0               # 自动安装"
    echo "  $0 --help        # 显示帮助"
}

# 清理函数（异常退出时调用）
cleanup() {
    if [[ $? -ne 0 ]]; then
        echo
        log_error "安装过程中出现错误"
        log_info "请检查:"
        log_info "1. 文件是否存在且完整"
        log_info "2. 是否有足够的权限"
        log_info "3. 磁盘空间是否充足"
    fi
}

# 主函数
main() {
    # 设置trap，在脚本退出时执行清理
    trap cleanup EXIT
    
    # 检查参数
    case "$1" in
        -h|--help)
            show_help
            exit 0
            ;;
        -v|--version)
            echo "nerdctl自动安装脚本 v1.1"
            exit 0
            ;;
    esac
    
    echo "========================================"
    echo "    nerdctl 自动安装脚本"
    echo "========================================"
    
    # 执行安装步骤
    check_root
    scan_tar_file
    detect_os
    install_nerdctl
    setup_completion
    post_install_check
    
    log_info "安装完成!请打开新的会话窗口!"
}

# 运行主函数
main "$@"