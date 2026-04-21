#!/bin/bash

# ==================== 配置区域 ====================
# 定义需连接的MySQL
CONTAINER_NAME="${CONTAINER_NAME:-mysql8-guacamole}"
DB_USER="${DB_USER:-root}"
DB_PASSWORD="${DB_PASSWORD:-root123456}"
SQL_DIR="${SQL_DIR:-./sql}"

# 定义数据库与SQL文件的映射关系
declare -A DB_SQL_MAP=(
    ["jnpf_init"]="jnpf_db_init.sql"
    ["xxl_job"]="jnpf_xxljob_init.sql"
    ["jnpf_flow"]="jnpf_flow_init.sql"
    ["jnpf_sso"]="jnpf_sso_3.5.5.sql jnpf_sso_3.5.5_data.sql"
    ["jnpf_tenant"]="jnpf_tenant_init.sql"
    ["jnpf_init_column"]="jnpf_db_init.sql"
)

# 日志颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 格式化输出函数
log_info()    { echo -e "${BLUE}$(date '+%Y-%m-%d %H:%M:%S') [INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}$(date '+%Y-%m-%d %H:%M:%S') [SUCCESS]${NC} $1"; }
log_warn()    { echo -e "${YELLOW}$(date '+%Y-%m-%d %H:%M:%S') [WARN]${NC} $1"; }
log_error()   { echo -e "${RED}$(date '+%Y-%m-%d %H:%M:%S') [ERROR]${NC} $1"; }

# 开启严格模式
set -e
set -u
set -o pipefail

# ==================== 预检查 ====================
if [ ! -d "$SQL_DIR" ]; then
    log_error "本地SQL目录 $SQL_DIR 不存在。"
    exit 1
fi

log_info "验证本地 SQL 文件完整性..."
for db in "${!DB_SQL_MAP[@]}"; do
    for sql_file in ${DB_SQL_MAP[$db]}; do
        if [ ! -f "$SQL_DIR/$sql_file" ]; then
            log_error "数据库 [$db] 所需文件缺失: $SQL_DIR/$sql_file"
            exit 1
        fi
    done
done

# ==================== 主流程 ====================

# 1. 检查容器及 MySQL 状态
if ! docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    log_error "容器 $CONTAINER_NAME 未运行，请先启动容器。"
    exit 1
fi

log_info "等待 MySQL 响应 (尝试连接)..."
until docker exec "$CONTAINER_NAME" mysqladmin ping -u"$DB_USER" -p"$DB_PASSWORD" --silent > /dev/null 2>&1; do
    echo -n "."
    sleep 2
done
echo ""
log_success "MySQL 已就绪。"

# 2. 核心函数定义
create_database() {
    local db_name=$1
    log_info "正在创建数据库: $db_name"
    # 直接输出执行结果到屏幕
    docker exec -i "$CONTAINER_NAME" mysql -u"$DB_USER" -p"$DB_PASSWORD" \
        -e "CREATE DATABASE IF NOT EXISTS \`$db_name\` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
}

import_sql() {
    local db_name=$1
    local sql_files=(${DB_SQL_MAP[$db_name]})
    
    for sql_file in "${sql_files[@]}"; do
        log_info "正在导入: $sql_file -> $db_name"
        # 使用管道导入，错误信息将直接打印在屏幕上
        # --max_allowed_packet 增加容错性
        if docker exec -i "$CONTAINER_NAME" mysql -u"$DB_USER" -p"$DB_PASSWORD" \
            --max_allowed_packet=1024M "$db_name" < "$SQL_DIR/$sql_file"; then
            log_success "文件 $sql_file 导入成功。"
        else
            log_error "文件 $sql_file 导入失败！"
            return 1
        fi
    done
}

# ==================== 开始执行 ====================
SUCCESS_DBS=()
FAILED_DBS=()

for db_name in "${!DB_SQL_MAP[@]}"; do
    echo -e "\n------------------------------------------------"
    log_info "处理数据库: $db_name"
    
    # 执行创建和导入，失败则记录到 FAILED_DBS 数组
    if create_database "$db_name" && import_sql "$db_name"; then
        SUCCESS_DBS+=("$db_name")
    else
        FAILED_DBS+=("$db_name")
    fi
done

# ==================== 结果汇总 ====================
echo -e "\n${BLUE}==================== 执行汇总 ====================${NC}"
[ ${#SUCCESS_DBS[@]} -ne 0 ] && log_success "成功数据库: ${SUCCESS_DBS[*]}"
if [ ${#FAILED_DBS[@]} -ne 0 ]; then
    log_error "失败数据库: ${FAILED_DBS[*]}"
    echo -e "${RED}请检查上方输出的 MySQL 错误详情。${NC}"
    exit 1
else
    log_success "所有初始化任务已完成！"
fi
echo -e "${BLUE}==================================================${NC}"