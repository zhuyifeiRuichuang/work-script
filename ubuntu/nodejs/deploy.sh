#!/bin/bash

# 智能 Node.js 离线安装脚本
# 自动检测当前目录下的 Node.js 压缩包并让用户选择安装
# 优化：支持在 /usr/bin 下创建 node 和 nodejs 两个命令

set -e  # 遇到错误时退出

print_header() {
    echo "==========================================="
    echo "     智能 Node.js 离线安装脚本 (增强版)"
    echo "  支持 /usr/bin/node 和 /usr/bin/nodejs 命令"
    echo "==========================================="
}

detect_node_packages() {
    # 查找当前目录下的所有 Node.js 压缩包
    NODE_PACKAGES=()
    while IFS= read -r -d '' file; do
        NODE_PACKAGES+=("$file")
    done < <(find . -maxdepth 1 -name "node-*.tar.xz" -print0)

    if [ ${#NODE_PACKAGES[@]} -eq 0 ]; then
        echo "错误: 当前目录下未找到任何 node-*.tar.xz 文件"
        echo "请将 Node.js 压缩包放在当前目录下"
        exit 1
    fi
}

extract_version_from_filename() {
    local filename="$1"
    # 提取版本号，例如从 node-v18.20.8-linux-x64.tar.xz 提取 18.20.8
    local version=$(basename "$filename" | sed -n 's/^node-v\([0-9]*\.[0-9]*\.[0-9]*\).*/\1/p')
    echo "$version"
}

show_available_versions() {
    echo "在当前目录找到以下 Node.js 版本:"
    echo
    for i in "${!NODE_PACKAGES[@]}"; do
        filename="${NODE_PACKAGES[$i]}"
        version=$(extract_version_from_filename "$filename")
        arch=$(basename "$filename" | sed -n 's/.*linux-\([^-]*\)\.tar\.xz/\1/p')
        
        printf "%2d. %s (架构: %s)\n" $((i+1)) "v$version" "$arch"
        printf "    文件: %s\n" "$(basename "$filename")"
        printf "    大小: %.2f MB\n" $(du -m "$filename" | cut -f1)
        echo
    done
}

select_package() {
    local choice
    while true; do
        read -p "请选择要安装的 Node.js 版本 (输入数字 1-${#NODE_PACKAGES[@]}): " choice
        
        # 验证输入是否为有效数字
        if [[ "$choice" =~ ^[0-9]+$ ]] && [ "$choice" -ge 1 ] && [ "$choice" -le ${#NODE_PACKAGES[@]} ]; then
            SELECTED_INDEX=$((choice-1))
            SELECTED_PACKAGE="${NODE_PACKAGES[$SELECTED_INDEX]}"
            SELECTED_VERSION=$(extract_version_from_filename "$SELECTED_PACKAGE")
            break
        else
            echo "无效选择，请输入 1 到 ${#NODE_PACKAGES[@]} 之间的数字"
        fi
    done
}

check_dependencies() {
    echo "检查系统依赖..."
    
    # 检查是否需要安装 libatomic
    if ! ldconfig -p | grep -q libatomic; then
        echo "警告: 缺少 libatomic 库，这可能导致 Node.js 运行时错误"
        echo "建议运行: apt-get update && apt-get install -y libatomic1"
        read -p "是否继续安装? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "安装已取消"
            exit 0
        fi
    fi
}

validate_and_extract() {
    echo "正在验证压缩包: $(basename "$SELECTED_PACKAGE")..."
    
    # 检查文件完整性
    if ! tar -tf "$SELECTED_PACKAGE" > /dev/null 2>&1; then
        echo "错误: 压缩包 $(basename "$SELECTED_PACKAGE") 损坏或不完整"
        exit 1
    fi
    
    echo "解压 Node.js 压缩包..."
    tar -xf "$SELECTED_PACKAGE"
    
    # 获取解压后的目录名
    EXTRACTED_DIR=$(basename "$SELECTED_PACKAGE" .tar.xz)
    
    if [ ! -d "$EXTRACTED_DIR" ]; then
        echo "错误: 解压失败，目录 $EXTRACTED_DIR 不存在"
        exit 1
    fi
}

install_nodejs() {
    # 获取架构信息
    ARCH=$(uname -m)
    if [ "$ARCH" = "x86_64" ]; then
        ARCH="x64"
    elif [ "$ARCH" = "aarch64" ]; then
        ARCH="arm64"
    else
        echo "警告: 检测到架构 $ARCH，尝试使用 x64"
        ARCH="x64"
    fi
    
    # 默认安装路径
    DEFAULT_INSTALL_PATH="/opt/nodejs-v${SELECTED_VERSION}-linux-${ARCH}"
    
    read -p "请输入安装路径 (默认: $DEFAULT_INSTALL_PATH): " INSTALL_PATH
    if [ -z "$INSTALL_PATH" ]; then
        INSTALL_PATH="$DEFAULT_INSTALL_PATH"
    fi
    
    # 检查是否以 root 权限运行
    if [ "$EUID" -ne 0 ]; then
        echo "警告: 脚本需要管理员权限才能安装到系统目录"
        read -p "是否继续并使用 sudo? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            USE_SUDO="sudo"
        else
            echo "取消安装"
            exit 1
        fi
    else
        USE_SUDO=""
    fi
    
    # 检查磁盘空间
    REQUIRED_SPACE=209715200  # 200MB in bytes
    AVAILABLE_SPACE=$(df . | awk 'NR==2 {print $4*1024}')
    if [ "$AVAILABLE_SPACE" -lt "$REQUIRED_SPACE" ]; then
        echo "错误: 磁盘空间不足。需要至少 200MB 可用空间"
        echo "当前可用空间: $(df -h . | awk 'NR==2 {print $4}')"
        exit 1
    fi
    
    # 检查是否有旧的安装并移除
    if [ -d "$INSTALL_PATH" ]; then
        echo "发现现有安装，正在移除旧版本..."
        $USE_SUDO rm -rf "$INSTALL_PATH"
    fi
    
    # 移动解压后的目录到安装路径
    echo "移动 Node.js 到 $INSTALL_PATH..."
    $USE_SUDO mkdir -p "$(dirname "$INSTALL_PATH")"
    $USE_SUDO mv "$EXTRACTED_DIR" "$INSTALL_PATH"
    
    # 移除可能存在的旧符号链接
    $USE_SUDO rm -f "/usr/local/bin/node" "/usr/local/bin/npm" "/usr/local/bin/npx"
    $USE_SUDO rm -f "/usr/bin/node" "/usr/bin/nodejs" "/usr/bin/npm" "/usr/bin/npx"
    
    # 创建新的符号链接 - 同时创建 /usr/local/bin 和 /usr/bin 的链接
    echo "创建符号链接到 /usr/local/bin..."
    $USE_SUDO ln -sf "$INSTALL_PATH/bin/node" "/usr/local/bin/node"
    $USE_SUDO ln -sf "$INSTALL_PATH/bin/npm" "/usr/local/bin/npm"
    $USE_SUDO ln -sf "$INSTALL_PATH/bin/npx" "/usr/local/bin/npx"
    
    # 在 /usr/bin 中创建 node 和 nodejs 符号链接
    echo "创建符号链接到 /usr/bin (支持 node 和 nodejs 命令)..."
    $USE_SUDO ln -sf "$INSTALL_PATH/bin/node" "/usr/bin/node"
    $USE_SUDO ln -sf "$INSTALL_PATH/bin/node" "/usr/bin/nodejs"  # 创建 nodejs 链接
    $USE_SUDO ln -sf "$INSTALL_PATH/bin/npm" "/usr/bin/npm"
    $USE_SUDO ln -sf "$INSTALL_PATH/bin/npx" "/usr/bin/npx"
    
    # 移除可能存在的旧环境变量配置
    for old_profile in /etc/profile.d/nodejs-*.sh; do
        if [ -f "$old_profile" ]; then
            $USE_SUDO rm -f "$old_profile"
        fi
    done
    
    # 添加到 PATH (通过 profile.d)
    PROFILE_FILE="/etc/profile.d/nodejs-${SELECTED_VERSION}.sh"
    echo "设置环境变量..."
    $USE_SUDO bash -c "cat > $PROFILE_FILE" << EOF
#!/bin/bash
export PATH="$INSTALL_PATH/bin:\$PATH"
export NODEJS_HOME="$INSTALL_PATH"
EOF

    $USE_SUDO chmod +x "$PROFILE_FILE"
    
    # 使环境变量立即生效
    export PATH="$INSTALL_PATH/bin:$PATH"
}

verify_installation() {
    echo "验证安装..."
    
    # 检查 node 命令是否可用
    if command -v node >/dev/null 2>&1; then
        # 尝试获取版本，处理可能的库依赖错误
        if INSTALLED_NODE_VERSION=$(node --version 2>/dev/null); then
            echo "✓ Node.js 版本: $INSTALLED_NODE_VERSION"
            
            # 检查 nodejs 命令是否也可用
            if command -v nodejs >/dev/null 2>&1; then
                if NODEJS_VERSION=$(nodejs --version 2>/dev/null); then
                    echo "✓ nodejs 命令版本: $NODEJS_VERSION"
                else
                    echo "! nodejs 命令存在但无法正常工作"
                fi
            else
                echo "! nodejs 命令不可用"
            fi
            
            # 检查 npm 是否也正常
            if INSTALLED_NPM_VERSION=$(npm --version 2>/dev/null); then
                echo "✓ NPM 版本: $INSTALLED_NPM_VERSION"
                
                if [[ "$INSTALLED_NODE_VERSION" == *"v$SELECTED_VERSION"* ]]; then
                    echo "✓ Node.js v$SELECTED_VERSION 安装成功!"
                else
                    echo "! 安装似乎有问题，版本不匹配"
                fi
            else
                echo "! NPM 无法正常工作，请检查安装"
            fi
        else
            # 检查是否有库依赖问题
            if node --version 2>&1 | grep -q "libatomic"; then
                echo "! Node.js 运行时错误: 缺少 libatomic 库"
                echo "! 请运行以下命令修复:"
                echo "! apt-get update && apt-get install -y libatomic1"
            else
                echo "! Node.js 无法正常运行，请检查安装"
                node --version 2>&1
            fi
            return 1
        fi
    else
        echo "✗ Node.js 安装失败，无法找到 node 命令"
        return 1
    fi
}

cleanup_old_symlinks() {
    echo "清理旧的符号链接..."
    # 清理旧的符号链接，确保没有冲突
    $USE_SUDO rm -f "/usr/bin/node" "/usr/bin/nodejs" "/usr/bin/npm" "/usr/bin/npx"
    $USE_SUDO rm -f "/usr/local/bin/node" "/usr/local/bin/nodejs" "/usr/local/bin/npm" "/usr/local/bin/npx"
}

main() {
    print_header
    detect_node_packages
    show_available_versions
    select_package
    
    echo
    echo "您选择了安装 Node.js v$SELECTED_VERSION"
    read -p "确认安装? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "安装已取消"
        exit 0
    fi
    
    check_dependencies
    validate_and_extract
    install_nodejs
    verify_installation
    
    echo
    echo "==========================================="
    echo "安装完成!"
    echo "  Node.js 版本: v$SELECTED_VERSION"
    echo "  安装路径: $INSTALL_PATH"
    echo "  node 命令位置: /usr/bin/node"
    echo "  nodejs 命令位置: /usr/bin/nodejs"
    echo "  npm 命令位置: /usr/bin/npm"
    echo "  npx 命令位置: /usr/bin/npx"
    echo "  卸载命令: $USE_SUDO rm -rf $INSTALL_PATH"
    echo "  要使更改永久生效，请运行: source ~/.bashrc"
    echo "  或重新打开终端窗口"
    echo "==========================================="
    
    # 如果遇到库依赖问题，提供解决方案提示
    if ! node --version >/dev/null 2>&1; then
        echo
        echo "注意: 如果 Node.js 无法正常运行，请尝试安装缺失的系统库:"
        echo "  apt-get update && apt-get install -y libatomic1"
    fi
}

# 运行主函数
main "$@"