#!/bin/bash
set -euo pipefail

# 定义x86-64各版本所需的核心指令集
declare -A X86_64_LEVELS=(
    [v1]="sse2"
    [v2]="sse3 ssse3 sse4_1 sse4_2 popcnt cmpxchg16b"
    [v3]="avx avx2 fma bmi1 bmi2 lzcnt"
    [v4]="avx512f avx512bw avx512cd avx512dq avx512vl"
)

# 颜色输出（增强可读性）
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly NC='\033[0m' # No Color

# 第一步：检查架构是否为x86_64
ARCH=$(uname -m)
if [ "$ARCH" != "x86_64" ]; then
    echo -e "${RED}Error: Current architecture is ${ARCH}, only x86_64 is supported!${NC}"
    exit 1
fi

# 第二步：提取CPU指令集标志（去重+清理首尾空格）
echo "=== Extracting CPU instruction set flags ==="
# 修复：去除开头空格，避免指令集列表前有多余空格
CPU_FLAGS=$(grep -m1 'flags' /proc/cpuinfo | awk -F: '{print $2}' | sed 's/^[ \t]*//' | tr ' ' '\n' | sort -u | tr '\n' ' ' | sed 's/ $//')
# 截取前100个字符并确保无末尾空格
TRUNCATED_FLAGS=$(echo "$CPU_FLAGS" | cut -c1-100 | sed 's/ $//')
echo -e "CPU supported instruction sets (partial): ${TRUNCATED_FLAGS}...\n"

# 第三步：逐个验证各版本
declare -A LEVEL_STATUS_RAW  # 存储原始状态（Supported/Not supported）
declare -A MISSING_FLAGS
MAX_SUPPORTED_LEVEL="none"

for LEVEL in v1 v2 v3 v4; do
    REQUIRED_FLAGS=${X86_64_LEVELS[$LEVEL]}
    MISSING=""
    echo "=== Verifying x86-64-${LEVEL} ==="
    echo "Required instruction sets: $REQUIRED_FLAGS"

    # 检查每个必需指令集是否存在
    for FLAG in $REQUIRED_FLAGS; do
        if [[ ! $CPU_FLAGS =~ (^| )$FLAG($| ) ]]; then  # 精确匹配指令集（避免子串匹配）
            MISSING+="$FLAG "
        fi
    done

    # 去除缺失指令集末尾的空格
    MISSING=$(echo "$MISSING" | sed 's/ $//')

    # 记录原始状态（不包含颜色码）
    if [ -z "$MISSING" ]; then
        LEVEL_STATUS_RAW[$LEVEL]="Supported"
        MAX_SUPPORTED_LEVEL=$LEVEL # 更新最高支持版本
        echo -e "Status: ${GREEN}${LEVEL_STATUS_RAW[$LEVEL]}${NC}\n"
    else
        LEVEL_STATUS_RAW[$LEVEL]="Not supported"
        MISSING_FLAGS[$LEVEL]=$MISSING
        echo -e "Status: ${RED}${LEVEL_STATUS_RAW[$LEVEL]}${NC}, missing instruction sets: $MISSING\n"
    fi
done

# 第四步：输出最终总结（动态添加颜色码，避免转义符暴露）
echo "==================== Detection Summary ===================="
echo "CPU Architecture: x86_64"
# 动态输出带颜色的状态，避免转义符存入变量
echo -e "x86-64-v1: $( [ "${LEVEL_STATUS_RAW[v1]}" = "Supported" ] && echo "${GREEN}Supported${NC}" || echo "${RED}Not supported${NC}" ) ${MISSING_FLAGS[v1]:+(Missing: ${MISSING_FLAGS[v1]})}"
echo -e "x86-64-v2: $( [ "${LEVEL_STATUS_RAW[v2]}" = "Supported" ] && echo "${GREEN}Supported${NC}" || echo "${RED}Not supported${NC}" ) ${MISSING_FLAGS[v2]:+(Missing: ${MISSING_FLAGS[v2]})}"
echo -e "x86-64-v3: $( [ "${LEVEL_STATUS_RAW[v3]}" = "Supported" ] && echo "${GREEN}Supported${NC}" || echo "${RED}Not supported${NC}" ) ${MISSING_FLAGS[v3]:+(Missing: ${MISSING_FLAGS[v3]})}"
echo -e "x86-64-v4: $( [ "${LEVEL_STATUS_RAW[v4]}" = "Supported" ] && echo "${GREEN}Supported${NC}" || echo "${RED}Not supported${NC}" ) ${MISSING_FLAGS[v4]:+(Missing: ${MISSING_FLAGS[v4]})}"
echo -e "=========================================================="
echo -e "${YELLOW}Final Conclusion: The highest x86-64 level supported by current CPU is x86-64-${MAX_SUPPORTED_LEVEL}${NC}"