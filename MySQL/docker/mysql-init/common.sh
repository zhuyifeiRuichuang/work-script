#!/bin/bash

# ==================== 公共配置 ====================
LOG_DIR="${LOG_DIR:-./logs}"
mkdir -p "$LOG_DIR"

# 日志颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 格式化输出函数
log_info()    { echo -e "${BLUE}$(date '+%Y-%m-%d %H:%M:%S') [INFO]${NC} $1" | tee -a "$LOG_FILE"; }
log_success() { echo -e "${GREEN}$(date '+%Y-%m-%d %H:%M:%S') [SUCCESS]${NC} $1" | tee -a "$LOG_FILE"; }
log_warn()    { echo -e "${YELLOW}$(date '+%Y-%m-%d %H:%M:%S') [WARN]${NC} $1" | tee -a "$LOG_FILE"; }
log_error()   { echo -e "${RED}$(date '+%Y-%m-%d %H:%M:%S') [ERROR]${NC} $1" | tee -a "$LOG_FILE"; }

# MySQL连接测试
test_mysql_connection() {
    local container_name=$1
    local db_user=$2
    local db_password=$3
    
    log_info "测试 MySQL 连接: $container_name"
    
    if ! docker ps --format '{{.Names}}' | grep -q "^${container_name}$"; then
        log_error "容器 $container_name 未运行"
        return 1
    fi
    
    local max_attempts=15
    local attempt=0
    
    until docker exec "$container_name" mysqladmin ping -u"$db_user" -p"$db_password" --silent > /dev/null 2>&1; do
        attempt=$((attempt + 1))
        if [[ $attempt -ge $max_attempts ]]; then
            log_error "MySQL 连接超时，已达最大尝试次数"
            return 1
        fi
        log_info "等待 MySQL 就绪... (尝试 $attempt/$max_attempts)"
        sleep 2
    done
    
    log_success "MySQL 连接成功"
    return 0
}

# 执行MySQL命令
execute_mysql_cmd() {
    local container_name=$1
    local db_user=$2
    local db_password=$3
    local sql_cmd=$4
    
    if ! docker exec -i "$container_name" mysql -u"$db_user" -p"$db_password" -e "$sql_cmd"; then
        log_error "MySQL 命令执行失败: $sql_cmd"
        return 1
    fi
    return 0
}