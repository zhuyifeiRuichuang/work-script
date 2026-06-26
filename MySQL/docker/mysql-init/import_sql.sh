#!/bin/bash

# ==================== 脚本配置 ====================
SCRIPT_NAME="import_sql"
CONFIG_FILE="${CONFIG_FILE:-./import.conf}"
LOG_FILE="${LOG_FILE:-$LOG_DIR/${SCRIPT_NAME}.log}"

# 加载公共函数
source "$(dirname "$0")/common.sh"

# ==================== 配置解析函数 ====================
parse_import_config() {
    local section=""
    local current_db=""
    
    while IFS= read -r line || [ -n "$line" ]; do
        [[ -z "$line" ]] && continue
        [[ "$line" =~ ^[[:space:]]*# ]] && continue
        [[ "$line" =~ ^[[:space:]]*$ ]] && continue
        
        if [[ "$line" =~ ^\[import_order\]$ ]]; then
            section="import_order"
            continue
        elif [[ "$line" =~ ^\[sql_files:(.*)\]$ ]]; then
            current_db="${BASH_REMATCH[1]}"
            SQL_FILES["$current_db"]=""
            section="sql_files"
            continue
        elif [[ "$line" =~ ^\[.*\]$ ]]; then
            section=""
            current_db=""
            continue
        fi
        
        case "$section" in
            "import_order")
                if [[ "$line" =~ ^order[[:space:]]*=[[:space:]]*(.*)$ ]]; then
                    IFS=',' read -ra IMPORT_ORDER <<< "${BASH_REMATCH[1]}"
                    for i in "${!IMPORT_ORDER[@]}"; do
                        IMPORT_ORDER[$i]=$(echo "${IMPORT_ORDER[$i]}" | xargs)
                    done
                fi
                ;;
            "sql_files")
                if [[ -n "$current_db" && "$line" =~ ^files[[:space:]]*=[[:space:]]*(.*)$ ]]; then
                    SQL_FILES["$current_db"]="${BASH_REMATCH[1]}"
                fi
                ;;
        esac
    done < "$CONFIG_FILE"
}

# ==================== 导入SQL函数 ====================
import_sql_file() {
    local db_name=$1
    local sql_file=$2
    local sql_dir=$3
    
    if [[ ! -f "$sql_dir/$sql_file" ]]; then
        log_error "❌ SQL文件不存在: $sql_dir/$sql_file"
        return 1
    fi
    
    log_info "导入SQL文件: $sql_file -> $db_name"
    log_info "文件大小: $(du -h "$sql_dir/$sql_file" | cut -f1)"
    
    # 检查数据库是否存在
    local db_exists_cmd="SELECT EXISTS(SELECT 1 FROM information_schema.schemata WHERE schema_name = '$db_name') AS db_exists;"
    
    if ! docker exec "$CONTAINER_NAME" mysql -u"$DB_USER" -p"$DB_PASSWORD" -se "$db_exists_cmd" | grep -q "1"; then
        log_error "❌ 数据库 $db_name 不存在，无法导入SQL"
        return 1
    fi
    
    # 导入SQL文件
    if docker exec -i "$CONTAINER_NAME" mysql -u"$DB_USER" -p"$DB_PASSWORD" \
        --max_allowed_packet=1024M "$db_name" < "$sql_dir/$sql_file"; then
        log_success "✅ 文件 $sql_file 导入成功"
        return 0
    else
        log_error "❌ 文件 $sql_file 导入失败"
        return 1
    fi
}

# ==================== 主函数 ====================
main() {
    > "$LOG_FILE"
    log_info "开始执行SQL文件导入脚本"
    
    # 验证配置文件
    if [[ ! -f "$CONFIG_FILE" ]]; then
        log_error "配置文件 $CONFIG_FILE 不存在"
        exit 1
    fi
    
    # 解析配置
    declare -A SQL_FILES
    parse_import_config
    
    # 读取MySQL连接配置
    if [[ -f "./databases.conf" ]]; then
        source <(grep -E '^\[(mysql_connection)\]' -A 3 ./databases.conf | grep -v '^--')
        CONTAINER_NAME="${CONTAINER_NAME:-mysql8-guacamole}"
        DB_USER="${DB_USER:-root}"
        DB_PASSWORD="${DB_PASSWORD:-root123456}"
    else
        CONTAINER_NAME="${CONTAINER_NAME:-mysql8-guacamole}"
        DB_USER="${DB_USER:-root}"
        DB_PASSWORD="${DB_PASSWORD:-root123456}"
    fi
    
    # 读取SQL目录配置
    if [[ -f "./databases.conf" ]]; then
        SQL_DIR=$(grep -E 'sql_dir[[:space:]]*=' ./databases.conf | cut -d'=' -f2 | xargs)
    fi
    SQL_DIR="${SQL_DIR:-./sql}"
    
    log_info "MySQL 配置: 容器=$CONTAINER_NAME, 用户=$DB_USER"
    log_info "SQL目录: $SQL_DIR"
    log_info "导入顺序: ${IMPORT_ORDER[*]}"
    
    # 验证SQL目录
    if [[ ! -d "$SQL_DIR" ]]; then
        log_error "SQL目录 $SQL_DIR 不存在"
        exit 1
    fi
    
    # 验证SQL文件
    log_info "验证SQL文件完整性..."
    for db_name in "${!SQL_FILES[@]}"; do
        IFS=',' read -ra files <<< "${SQL_FILES[$db_name]}"
        for sql_file in "${files[@]}"; do
            sql_file=$(echo "$sql_file" | xargs)
            if [[ ! -f "$SQL_DIR/$sql_file" ]]; then
                log_error "❌ 数据库 [$db_name] 所需文件缺失: $SQL_DIR/$sql_file"
                exit 1
            fi
            log_info "✅ 验证通过: $SQL_DIR/$sql_file"
        done
    done
    
    # 测试MySQL连接
    if ! test_mysql_connection "$CONTAINER_NAME" "$DB_USER" "$DB_PASSWORD"; then
        log_error "MySQL 连接失败，退出脚本"
        exit 1
    fi
    
    # 按顺序导入SQL
    SUCCESS_IMPORTS=()
    FAILED_IMPORTS=()
    
    for db_name in "${IMPORT_ORDER[@]}"; do
        if [[ -z "${SQL_FILES[$db_name]+_}" ]]; then
            log_warn "数据库 [$db_name] 未配置SQL文件，跳过"
            continue
        fi
        
        echo -e "\n${BLUE}--------------------------------------------------${NC}" | tee -a "$LOG_FILE"
        log_info "处理数据库: $db_name"
        
        IFS=',' read -ra files <<< "${SQL_FILES[$db_name]}"
        for sql_file in "${files[@]}"; do
            sql_file=$(echo "$sql_file" | xargs)
            if import_sql_file "$db_name" "$sql_file" "$SQL_DIR"; then
                SUCCESS_IMPORTS+=("$db_name/$sql_file")
            else
                FAILED_IMPORTS+=("$db_name/$sql_file")
            fi
        done
    done
    
    # 结果汇总
    echo -e "\n${BLUE}==================== 执行结果 ====================${NC}" | tee -a "$LOG_FILE"
    [ ${#SUCCESS_IMPORTS[@]} -gt 0 ] && log_success "成功导入: ${SUCCESS_IMPORTS[*]}"
    
    if [ ${#FAILED_IMPORTS[@]} -gt 0 ]; then
        log_error "导入失败: ${FAILED_IMPORTS[*]}"
        echo -e "${RED}请检查日志文件: $LOG_FILE${NC}" | tee -a "$LOG_FILE"
        exit 1
    else
        log_success "🎉 所有SQL文件导入成功！"
    fi
    echo -e "${BLUE}==================================================${NC}" | tee -a "$LOG_FILE"
}

# ==================== 执行主函数 ====================
main "$@"