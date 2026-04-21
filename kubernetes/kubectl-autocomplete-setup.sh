#!/bin/bash

set -e

echo "🚀 开始配置 kubectl 自动补全..."

# 1. 检查 kubectl 是否安装
if ! command -v kubectl &> /dev/null; then
    echo "❌ 未检测到 kubectl。请先安装 kubectl 后再运行此脚本。"
    echo "   安装指南: https://kubernetes.io/docs/tasks/tools/"
    exit 1
fi

# 2. 安全获取 kubectl 客户端版本（兼容旧版本）
echo "✅ 已检测到 kubectl，客户端版本："
kubectl version --client 2>&1 | grep -o 'GitVersion.*' | head -n1 || kubectl version --client

# 3. 检测当前 Shell
SHELL_NAME=$(basename "$SHELL")
echo "🔍 当前 Shell: $SHELL_NAME"

if [[ "$SHELL_NAME" != "bash" && "$SHELL_NAME" != "zsh" ]]; then
    echo "⚠️  警告：当前 Shell ($SHELL_NAME) 不是 Bash 或 Zsh，跳过自动补全配置。"
    echo "   仅支持 Bash 和 Zsh。"
    exit 0
fi

# 4. 确定配置文件
if [[ "$SHELL_NAME" == "bash" ]]; then
    CONFIG_FILE="$HOME/.bashrc"
elif [[ "$SHELL_NAME" == "zsh" ]]; then
    CONFIG_FILE="$HOME/.zshrc"
fi

# 5. 构建配置内容
KUBECTL_COMPLETION_LINE="source <(kubectl completion $SHELL_NAME)"
ALIAS_LINE="alias k=kubectl"
ALIAS_COMPLETE_LINE=""

if [[ "$SHELL_NAME" == "bash" ]]; then
    ALIAS_COMPLETE_LINE="complete -F __start_kubectl k"
elif [[ "$SHELL_NAME" == "zsh" ]]; then
    # 确保 compinit 已加载（避免 zsh 补全失效）
    if ! grep -q "autoload -Uz compinit" "$CONFIG_FILE" 2>/dev/null; then
        echo "💡 Zsh: 添加 compinit 支持..."
        echo "autoload -Uz compinit && compinit" >> "$CONFIG_FILE"
    fi
fi

# 6. 检查是否已配置，避免重复
need_add=0

if ! grep -q "kubectl completion" "$CONFIG_FILE" 2>/dev/null; then
    need_add=1
else
    echo "ℹ️  kubectl 自动补全已配置在 $CONFIG_FILE 中，跳过添加。"
fi

if ! grep -q "alias k=kubectl" "$CONFIG_FILE" 2>/dev/null; then
    need_add=1
else
    echo "ℹ️  别名 'k' 已配置。"
fi

# 7. 写入配置（如需要）
if [[ $need_add -eq 1 ]]; then
    echo "📝 正在写入配置到 $CONFIG_FILE ..."
    {
        echo ""
        echo "# --- kubectl 自动补全 (由自动脚本添加) ---"
        echo "$KUBECTL_COMPLETION_LINE"
        echo "$ALIAS_LINE"
        if [[ -n "$ALIAS_COMPLETE_LINE" ]]; then
            echo "$ALIAS_COMPLETE_LINE"
        fi
        echo "# ----------------------------------------"
    } >> "$CONFIG_FILE"
    echo "✅ 配置已写入 $CONFIG_FILE"
else
    echo "✅ 配置已存在，无需重复操作。"
fi

# 8. 尝试立即生效
echo "🔄 尝试立即加载配置（仅限当前终端会话）..."
if [[ -f "$CONFIG_FILE" ]]; then
    source "$CONFIG_FILE"
fi

# 9. 提示
echo ""
echo "🎉 kubectl 自动补全配置完成！"
echo "   - 你可以使用 'kubectl get pod<TAB>' 进行补全"
echo "   - 也可以使用别名 'k get ns<TAB>'"
echo ""
echo "💡 注意：新打开的终端将自动生效。当前终端已尝试加载配置。"

exit 0