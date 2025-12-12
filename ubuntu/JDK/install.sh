#!/bin/bash
# 脚本名称：jdk_offline_install.sh
# 功能：离线自动安装JDK，解压当前目录的tar.gz包并配置JAVA_HOME
# 用法：sudo bash jdk_offline_install.sh

# ====================== 基础配置 ======================
# JDK安装根目录（可根据需求修改）
JVM_DIR="/usr/lib/jvm"
# 环境变量配置文件（兼容Ubuntu/CentOS）
ENV_FILE="/etc/profile"

# ====================== 函数定义 ======================

# 1. 权限检查函数
check_permission() {
    if [ $EUID -ne 0 ]; then
        echo -e "\033[31m错误：请使用root权限运行此脚本（sudo bash $0）\033[0m"
        exit 1
    fi
}

# 2. 查找JDK安装包函数
find_jdk_package() {
    echo -e "\033[34m正在查找当前目录下的JDK tar.gz包...\033[0m"
    # 查找当前目录下的JDK压缩包（忽略大小写，匹配jdk开头/包含jdk的tar.gz）
    JDK_PACKAGES=($(ls ./ | grep -i 'jdk.*\.tar\.gz' | grep -v grep))
    
    # 检查包数量
    if [ ${#JDK_PACKAGES[@]} -eq 0 ]; then
        echo -e "\033[31m错误：当前目录未找到JDK tar.gz格式的安装包！\033[0m"
        exit 1
    elif [ ${#JDK_PACKAGES[@]} -gt 1 ]; then
        echo -e "\033[33m找到多个JDK安装包，请选择要安装的版本：\033[0m"
        for i in "${!JDK_PACKAGES[@]}"; do
            echo "  $((i+1)). ${JDK_PACKAGES[$i]}"
        done
        read -p "请输入序号（1-${#JDK_PACKAGES[@]}）：" SELECT_NUM
        # 验证输入合法性
        if ! [[ $SELECT_NUM =~ ^[0-9]+$ ]] || [ $SELECT_NUM -lt 1 ] || [ $SELECT_NUM -gt ${#JDK_PACKAGES[@]} ]; then
            echo -e "\033[31m错误：输入序号非法！\033[0m"
            exit 1
        fi
        SELECTED_PACKAGE=${JDK_PACKAGES[$((SELECT_NUM-1))]}
    else
        SELECTED_PACKAGE=${JDK_PACKAGES[0]}
    fi
    echo -e "\033[32m已选择安装包：$SELECTED_PACKAGE\033[0m"
}

# 3. 解压JDK包函数
extract_jdk() {
    echo -e "\033[34m正在创建JDK安装目录：$JVM_DIR\033[0m"
    mkdir -p $JVM_DIR || { echo -e "\033[31m错误：创建目录$JVM_DIR失败！\033[0m"; exit 1; }

    echo -e "\033[34m正在解压JDK包到$JVM_DIR...\033[0m"
    tar -zxf ./$SELECTED_PACKAGE -C $JVM_DIR/ || { echo -e "\033[31m错误：解压$SELECTED_PACKAGE失败！\033[0m"; exit 1; }

    # 获取解压后的JDK实际目录名（兼容jdk1.8.0_xxx、jdk-8uxxx-linux-x64等命名）
    JDK_UNPACK_DIR=$(ls $JVM_DIR | grep -i 'jdk' | grep -v '\.tar\.gz' | head -n1)
    if [ -z "$JDK_UNPACK_DIR" ]; then
        echo -e "\033[31m错误：解压后未找到JDK目录！\033[0m"
        exit 1
    fi
    JDK_HOME="$JVM_DIR/$JDK_UNPACK_DIR"
    echo -e "\033[32mJDK解压完成，实际安装路径：$JDK_HOME\033[0m"
}

# 4. 配置环境变量函数
config_env() {
    echo -e "\033[34m正在配置JAVA_HOME环境变量...\033[0m"
    
    # 先清理旧的JAVA_HOME配置（避免重复）
    sed -i '/^export JAVA_HOME=/d' $ENV_FILE
    sed -i '/^export JRE_HOME=/d' $ENV_FILE
    sed -i '/^export CLASSPATH=/d' $ENV_FILE
    sed -i '/^export PATH=.*JAVA_HOME.*bin/d' $ENV_FILE

    # 添加新的环境变量配置
    cat >> $ENV_FILE << EOF

# JDK Auto Install Config
export JAVA_HOME=$JDK_HOME
export JRE_HOME=\${JAVA_HOME}/jre
export CLASSPATH=.:\${JAVA_HOME}/lib:\${JRE_HOME}/lib
export PATH=\${JAVA_HOME}/bin:\$PATH
EOF

    # 生效环境变量（当前会话）
    source $ENV_FILE || echo -e "\033[33m提示：环境变量临时生效失败，重启终端后自动生效\033[0m"
    echo -e "\033[32m环境变量配置完成，已写入$ENV_FILE\033[0m"
}

# 5. 验证安装结果函数
verify_install() {
    echo -e "\033[34m正在验证JDK安装结果...\033[0m"
    # 重新加载环境变量
    source $ENV_FILE
    # 检查java版本
    if java -version 2>&1 | grep -q "1.8"; then
        echo -e "\033[32m====================== 安装成功 ======================\033[0m"
        echo "JAVA_HOME: $(echo $JAVA_HOME)"
        echo "Java版本："
        java -version
        echo -e "\033[32m======================================================\033[0m"
        echo -e "\033[33m提示：若新终端中java命令未生效，请执行 source $ENV_FILE 或重启终端\033[0m"
    else
        echo -e "\033[31m错误：JDK安装验证失败！\033[0m"
        exit 1
    fi
}

# ====================== 主执行流程 ======================
main() {
    clear
    echo -e "\033[34m===================== JDK离线安装脚本 ======================\033[0m"
    check_permission
    find_jdk_package
    extract_jdk
    config_env
    verify_install
    exit 0
}

# 执行主函数
main
