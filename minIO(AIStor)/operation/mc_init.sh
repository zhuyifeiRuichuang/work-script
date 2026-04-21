#!/bin/bash
# 自动配置mc client
# 获取脚本所在目录的绝对路径
SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
# 定义路径和域名
TARGET_DIR="$HOME/minio-binaries"
TARGET_PATH="$TARGET_DIR/mc"
MC_URL="https://dl.min.io/client/mc/release/linux-amd64/mc"
LOCAL_MC_PATH="$SCRIPT_DIR/mc"  # 脚本所在目录的mc文件
TARGET_DOMAIN="dl.min.io"       # 目标域名
PING_IP="8.8.8.8"               # 用于IP连通性检测的地址
is_connected=0                  # 网络连接状态标记(1=正常,0=异常)

# 1. 检测当前环境是否有mc命令，若有则结束
echo "1. 检测是否已安装mc客户端..."
if command -v mc &> /dev/null; then
    echo "检测到已安装mc客户端，版本信息："
    mc --version
    echo "无需进一步操作，流程结束"
    exit 0
fi
echo "未检测到mc客户端，继续执行流程..."

# 2. 检测当前环境联网状态（同时检测IP和域名）
echo -e "\n2. 检测网络连通性..."
echo "正在检测IP连接（$PING_IP）..."
if ping -c 2 -W 5 "$PING_IP" &> /dev/null; then
    echo "IP连接检测通过"
    echo "正在检测域名访问（$TARGET_DOMAIN）..."
    if ping -c 2 -W 5 "$TARGET_DOMAIN" &> /dev/null; then
        echo "域名访问检测通过"
        is_connected=1
    else
        echo "域名访问检测失败"
        is_connected=0
    fi
else
    echo "IP连接检测失败"
    is_connected=0
fi

# 根据网络状态执行不同流程
if [ $is_connected -eq 1 ]; then
    echo "网络状态：正常"
    # 3. 联网正常，自动下载mc文件
    echo -e "\n3. 开始下载mc客户端..."
    mkdir -p "$TARGET_DIR" || {
        echo "错误：无法创建目录 $TARGET_DIR"
        exit 1
    }
    
    if curl "$MC_URL" -o "$TARGET_PATH"; then
        echo "mc客户端下载成功"
    else
        echo "错误：mc客户端下载失败"
        exit 1
    fi
else
    echo "网络状态：异常"
    # 3. 联网异常，检测当前是否有mc文件
    echo -e "\n3. 检测本地mc文件..."
    if [ -f "$LOCAL_MC_PATH" ]; then
        echo "发现本地mc文件，将直接使用"
        # 确保目标目录存在
        mkdir -p "$TARGET_DIR" || {
            echo "错误：无法创建目录 $TARGET_DIR"
            exit 1
        }
        # 复制本地文件到目标路径
        cp "$LOCAL_MC_PATH" "$TARGET_PATH" || {
            echo "错误：复制本地mc文件失败"
            exit 1
        }
    else
        echo "未发现本地mc文件，检测是否有mc.7z压缩包..."
        # 检测是否有mc.7z文件
        if [ ! -f "./mc.7z" ]; then
            echo "错误：未找到mc.7z文件"
            echo "请手动从以下地址下载mc文件到当前目录："
            echo "$MC_URL"
            exit 1
        fi
        
        # 4. 存在mc.7z文件，检测7z解压工具
        echo -e "\n4. 发现mc.7z文件，检测解压工具..."
        if ! command -v 7z &> /dev/null; then
            echo "未检测到7z解压工具"
            # 联网异常时无法安装7z，提示手动下载
            echo "错误：网络异常，无法自动安装7z工具"
            echo "请手动从以下地址下载mc文件到当前目录："
            echo "$MC_URL"
            exit 1
        fi
        
        # 使用7z解压文件
        echo "正在解压mc.7z文件..."
        7z x ./mc.7z -o"$SCRIPT_DIR" > /dev/null 2>&1 || {
            echo "错误：解压mc.7z失败"
            exit 1
        }
        
        # 确认解压出mc文件
        if [ ! -f "$LOCAL_MC_PATH" ]; then
            echo "错误：解压后未找到mc文件"
            exit 1
        fi
        
        # 复制解压后的文件到目标路径
        mkdir -p "$TARGET_DIR" || {
            echo "错误：无法创建目录 $TARGET_DIR"
            exit 1
        }
        cp "$LOCAL_MC_PATH" "$TARGET_PATH" || {
            echo "错误：复制解压后的mc文件失败"
            exit 1
        }
    fi
fi

# 后续配置流程：添加执行权限
echo -e "\n5. 配置执行权限..."
chmod +x "$TARGET_PATH" || {
    echo "错误：无法设置执行权限"
    exit 1
}

# 配置环境变量（持久化）
echo -e "\n6. 配置环境变量..."
if ! grep -q "$TARGET_DIR" "$HOME/.bashrc"; then
    echo "export PATH=\$PATH:$TARGET_DIR" >> "$HOME/.bashrc"
    echo "环境变量已添加到 .bashrc"
else
    echo ".bashrc中已存在环境变量配置，无需重复添加"
fi

# 对于使用zsh的用户
if [ -f "$HOME/.zshrc" ] && ! grep -q "$TARGET_DIR" "$HOME/.zshrc"; then
    echo "export PATH=\$PATH:$TARGET_DIR" >> "$HOME/.zshrc"
    echo "环境变量已添加到 .zshrc"
else
    echo ".zshrc中已存在环境变量配置，无需重复添加"
fi

# 立即加载环境变量
export PATH="$PATH:$TARGET_DIR"

# 验证安装
echo -e "\n7. 验证mc安装..."
if command -v mc &> /dev/null; then
    echo "mc版本信息："
    mc --version
    echo "配置完成！请重启终端或执行 source ~/.bashrc (或 ~/.zshrc) 使配置生效"
else
    echo "警告：mc命令暂时无法识别，请手动执行 source ~/.bashrc 后重试"
    exit 1
fi
