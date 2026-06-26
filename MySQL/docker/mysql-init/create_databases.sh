#!/bin/bash

# ==================== 脚本配置 ====================
SCRIPT_NAME="create_databases"
CONFIG_FILE="${CONFIG_FILE:-./databases.conf}"
LOG_FILE="${LOG_FILE:-$LOG_DIR/${SCRIPT_NAME}.log}"

# 加载公共函数
source "$(dirname "$0")/common.sh"

# ==================== 配置解析函数 ====================
parse_database_config() {
    local section=""
    local current_db=""
    
    while IFS= read -r line || [ -n "$line" ]; do
        [[ -z "$line" ]] && continue
        [[ "$line" =~ ^[[:space:]]*# ]] && continue
        [[ "$line" =~ ^[[:space:]]*$ ]] && continue
        
        if [[ "$line" =~ ^\[[^\]]+\]$ ]]; then
            section="${line#[}"
            section="${section%]}"
            
            if [[ "$section" =~ ^database:(.*)$ ]]; then
                current_db="${BASH_REMATCH[1]}"
                DATABASES["$current_db"]="true"
                eval "DB_${current_db}_CHARSET=\"utf8mb4\""
                eval "DB_${current_db}_COLLATE=\"utf8mb4_unicode_ci\""
            else
                current_db=""
            fi
            continue
        fi
        
        [[ -z "$section" ]] && continue
        
        if [[ "$line" =~ ^[[:space:]]*([a-zA-Z0-9_]+)[[:space:]]*=[[:space:]]*(.*)$ ]]; then
            key="${BASH_REMATCH[1]}"
            value="${BASH_REMATCH[2]}"
            
            case "$section" in
                "mysql_connection")
                    case "$key" in
                        "container_name") CONTAINER_NAME="$value" ;;
                        "db_user") DB_USER="$value" ;;
                        "db_password") DB_PASSWORD="$value" ;;
                    esac
                    ;;
                *)
                    if [[ "$section" =~ ^database:(.*)$ ]]; then
                        db_name="${BASH_REMATCH[1]}"
                        case "$key" in
                            "charset") eval "DB_${db_name}_CHARSET=\"$value\"" ;;
                            "collate") eval "DB_${db_name}_COLLATE=\"$value\"" ;;
                        esac
                    fi
                    ;;
            esac
        fi
    done < "$CONFIG_FILE"
}

# ==================== 主函数 ====================
main() {
    > "$LOG_FILE"
    log_info "开始执行数据库创建脚本"
    
    # 验证配置文件
    if [[ ! -f "$CONFIG_FILE" ]]; then
        log_error "配置文件 $CONFIG_FILE 不存在"
        exit 1
    fi
    
    # 解析配置
    declare -A DATABASES
    parse_database_config
    
    # 设置默认值
    CONTAINER_NAME="${CONTAINER_NAME:-mysql8-guacamole}"
    DB_USER="${DB_USER:-root}"
    DB_PASSWORD="${DB_PASSWORD:-root123456}"
    
    log_info "MySQL 配置: 容器=$CONTAINER_NAME, 用户=$DB_USER"
    log_info "检测到 ${#DATABASES[@]} 个数据库需要创建"
    
    # 测试MySQL连接
    if ! test_mysql_connection "$CONTAINER_NAME" "$DB_USER" "$DB_PASSWORD"; then
        log_error "MySQL 连接失败，退出脚本"
        exit 1
    fi
    
    # 创建数据库
    SUCCESS_DBS=()
    FAILED_DBS=()
    
    for db_name in "${!DATABASES[@]}"; do
        eval "charset=\"\${DB_${db_name}_CHARSET}\""
        eval "collate=\"\${DB_${db_name}_COLLATE}\""
        
        log_info "创建数据库: $db_name (字符集: $charset, 排序规则: $collate)"
        
        sql_cmd="CREATE DATABASE IF NOT EXISTS \`$db_name\` CHARACTER SET $charset COLLATE $collate;"
        
        if execute_mysql_cmd "$CONTAINER_NAME" "$DB_USER" "$DB_PASSWORD" "$sql_cmd"; then
            log_success "✅ 数据库 $db_name 创建成功"
            SUCCESS_DBS+=("$db_name")
        else
            log_error "❌ 数据库 $db_name 创建失败"
            FAILED_DBS+=("$db_name")
        fi
    done
    
    # 结果汇总
    echo -e "\n${BLUE}==================== 执行结果 ====================${NC}" | tee -a "$LOG_FILE"
    [ ${#SUCCESS_DBS[@]} -gt 0 ]