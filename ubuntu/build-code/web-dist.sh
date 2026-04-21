#!/bin/bash
set -euo pipefail

# ========== 错误处理与清理 ==========
trap 'echo "❌ 脚本执行失败，清理临时容器..."; docker rm -f build-dist >/dev/null 2>&1 || true' ERR
trap 'docker rm -f build-dist >/dev/null 2>&1 || true' 0

# ========== 配置项 ==========
NODE_VERSION="20.15.0"
PNPM_VERSION="9.12.0"
SOURCE_DIR="$(pwd)"
CONTAINER_NAME="build-dist"
CONTAINER_WORKDIR="/app"
# 修复：移除末尾空格
NPM_REGISTRY="https://registry.npmmirror.com"
# 备用 registry（HTTP 协议）
FALLBACK_REGISTRY="http://registry.npmmirror.com"
APP_NAME=${APP_NAME:-"jnpf-web-apps-main"}

# ========== 排除目录配置清单 ==========
declare -a EXCLUDE_DIRS=(
    "node_modules"
    ".git"
    ".pnpm-store"
    ".cache"
    "backup-dist"
    "dist-backup"
    "*.backup"
    "scripts"
    "internal"
    "packages"
)

# ========== 前置检查 ==========
for cmd in docker; do
    if ! command -v $cmd &> /dev/null; then
        echo "❌ 错误：未安装 $cmd"
        exit 1
    fi
done

if ! docker info &> /dev/null; then
    echo "❌ 错误：Docker未运行"
    exit 1
fi

[ ! -f "${SOURCE_DIR}/package.json" ] && echo "❌ 错误：未找到package.json" && exit 1
docker rm -f "${CONTAINER_NAME}" >/dev/null 2>&1 || true

echo "✅ 前置检查通过"
echo "📌 构建配置：Node=${NODE_VERSION} PNPM=${PNPM_VERSION} 应用=${APP_NAME}"

# ========== 核心构建（优化部分）==========
docker run -d --name "${CONTAINER_NAME}" \
    -v "${SOURCE_DIR}:${CONTAINER_WORKDIR}" \
    -w "${CONTAINER_WORKDIR}" \
    --entrypoint /bin/bash \
    node:${NODE_VERSION} \
    -c "
        set -euo pipefail
        
        # 安装pnpm
        echo '📦 安装pnpm...@${PNPM_VERSION}'
        npm install -g pnpm@${PNPM_VERSION} --registry=${NPM_REGISTRY}
        pnpm config set registry ${NPM_REGISTRY}
        export PNPM_CATALOG_DISABLED=true
        
        # ========== 智能依赖安装（核心优化）==========
        echo '🔧 安装项目依赖...'
        
        # 检查锁文件是否存在
        if [ -f '${CONTAINER_WORKDIR}/pnpm-lock.yaml' ]; then
            echo '   ✅ 检测到 pnpm-lock.yaml，尝试严格模式安装...'
            
            # 严格模式安装，失败则回退
            pnpm install --frozen-lockfile || {
                echo '   ❌ 严格模式安装失败'
                echo '   🔄 切换到备用 registry 重试...'
                
                # 清除缓存并重试
                pnpm store prune
                pnpm install --registry=${FALLBACK_REGISTRY} || {
                    echo '   ❌ 备用 registry 也失败，退出构建'
                    exit 1
                }
            }
        else
            echo '   ⚠️  未检测到 pnpm-lock.yaml，生成锁文件并安装...'
            
            # 使用主 registry 安装
            pnpm install --registry=${NPM_REGISTRY} || {
                echo '   ❌ 主 registry 安装失败，尝试备用 registry...'
                
                # 使用备用 registry 重试
                pnpm install --registry=${FALLBACK_REGISTRY} || {
                    echo '   ❌ 安装失败，请检查网络或 registry 配置'
                    exit 1
                }
            }
        fi
        
        echo '   ✅ 依赖安装完成'
        
        # ========== 智能构建（优化部分）==========
        echo '🚀 执行前端构建...'
        
        # 首先尝试使用 pnpm 构建
        pnpm run build || {
            echo '   ❌ pnpm 构建失败，准备切换到 npm 备用方案...'
            
            # 清理 pnpm 相关文件以避免冲突
            echo '   🧹 清理 pnpm 相关文件...'
            rm -f pnpm-lock.yaml
            rm -rf node_modules
            
            # 使用 npm 重新安装依赖
            echo '   📦 使用 npm 安装依赖...'
            npm install -f --registry=${NPM_REGISTRY} || {
                echo '   ❌ npm 安装失败，尝试备用 registry...'
                npm install -f --registry=${FALLBACK_REGISTRY} || {
                    echo '   ❌ npm 备用 registry 也失败，退出构建'
                    exit 1
                }
            }
            
            # 使用 npm 执行构建
            echo '   🚀 使用 npm 执行构建...'
            npm run build || {
                echo '   ❌ npm 构建失败，请检查代码和配置'
                exit 1
            }
            
            echo '   ✅ npm 构建成功'
        }
        
        # 查找并显示所有dist目录（排除node_modules）
        echo '✅ 构建完成，容器内dist目录：'
        find ${CONTAINER_WORKDIR} -type d -name 'node_modules' -prune -o -type d -name 'dist' -print -exec ls -ld {} \\;
    "

docker logs -f "${CONTAINER_NAME}"
CONTAINER_EXIT_CODE=$(docker inspect "${CONTAINER_NAME}" --format '{{.State.ExitCode}}')
[ "${CONTAINER_EXIT_CODE}" -ne 0 ] && exit 1

# ========== 智能查找：应用排除清单 ==========
echo -e "\n\n🎉 构建成功！"
echo "🔍 正在查找所有构建产物(dist目录)..."

# 构建find命令的排除参数
FIND_EXCLUDE_CMD="find \"$(pwd)\""
for exclude_dir in "${EXCLUDE_DIRS[@]}"; do
    FIND_EXCLUDE_CMD+=" -type d -name \"${exclude_dir}\" -prune -o"
done
FIND_EXCLUDE_CMD+=" -type d -name \"dist\" -print 2>/dev/null"

# 执行查找并排除空行
DIST_DIRS=$(eval "${FIND_EXCLUDE_CMD}" | grep -v '^$')

if [ -z "${DIST_DIRS}" ]; then
    echo "⚠️  警告：未找到任何dist目录"
    
    # 显示排除配置信息
    echo -e "\n🚫 已排除的目录清单："
    printf "   - %s\n" "${EXCLUDE_DIRS[@]}"
    
    echo -e "\n📂 当前目录结构（排除主要干扰目录）："
    # 显示目录结构时也应用排除
    tree -I "$(IFS='|'; echo "${EXCLUDE_DIRS[*]}")" -L 3 -d 2>/dev/null || \
        find . -type d \( -name "node_modules" -o -name ".git" \) -prune -o -type d -print
else
    echo -e "\n✅ 找到 $(echo "${DIST_DIRS}" | wc -l) 个dist目录（已应用排除清单）："
    echo "------------------------------------------------"
    
    # 显示排除配置
    echo "🚫 排除的目录：$(IFS=', '; echo "${EXCLUDE_DIRS[*]}")"
    echo "------------------------------------------------"
    
    # 遍历显示详细信息
    echo "${DIST_DIRS}" | while read -r dist_path; do
        if [ -n "${dist_path}" ]; then
            echo -e "\n📦 ${dist_path}"
            echo "   📊 大小: $(du -sh "${dist_path}" 2>/dev/null | cut -f1)"
            echo "   📄 文件数量: $(find "${dist_path}" -type f 2>/dev/null | wc -l)"
            echo "   🔍 前5项内容:"
            ls -lh "${dist_path}" 2>/dev/null | head -5 | sed 's/^/      /'
        fi
    done
    
    # 总统计
    TOTAL_SIZE=$(du -shc $(echo "${DIST_DIRS}") 2>/dev/null | tail -1 | cut -f1)
    echo -e "\n------------------------------------------------"
    echo "📈 总产物大小: ${TOTAL_SIZE}"
fi

# ========== 清理 ==========
echo -e "\n🧹 清理临时容器..."
docker rm -f "${CONTAINER_NAME}" >/dev/null 2>&1

echo -e "\n✅ 所有操作完成！"
trap - ERR 0