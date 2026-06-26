#!/bin/bash

# ==================== 脚本配置 ====================
SCRIPT_NAME="create_users"
CONFIG_FILE="${CONFIG_FILE:-./users.conf}"
LOG_FILE="${LOG_FILE:-$LOG_DIR/${SCRIPT_NAME}.log}"

# 加载公共函数
source "$(dirname "$0")/common.sh"

# ==================== 配置解析函数 ====================
parse_users_config() {
    local section=""
    local in_users_section=false
    
    while IFS= read -r line || [ -n "$line" ]; do
        [[ -z "$line" ]] && continue
        [[ "$line" =~ ^[[:space:]]*# ]] && continue
        [[ "$line" =~ ^[[:space:]]*$ ]] && continue
        
        if [[ "$line" =~ ^\[users\]$ ]]; then
            in_users_section=true
            continue
        elif [[ "$line" =~ ^\[.*\]$ ]]; then
            in_users_section=false
            continue
        fi
        
        if $in_users_section && [[ "$line" =~ ^([a-zA-Z0-9_]+):([^:]+):(.+)$ ]]; then
            local username="${BASH_REMATCH[1]}"
            local password="${BASH_REMATCH[2]}"
            local host="${BASH_REMATCH[3]}"
            
            if [[ -z "${USERS[$username,$host]+_}" ]]; then
                USERS["$username,$host"]="$password"
                log_info "解析用户: $username@$host"
            else
                log_warn "用户 $username@$host 已存在，跳过重复定义"
            fi
        fi
    done < "$CONFIG_FILE"
    
    # 解析数据库权限
    while IFS= read -r line || [ -n "$line" ]; do
        [[ -z "$line" ]] && continue
        [[ "$line" =~ ^[[:space:]]*# ]] && continue
        [[ "$line" =~ ^[[:space:]]*$ ]] && continue
        
        if [[ "$line" =~ ^\[database_grants:(.*)\]$ ]]; then
            current_db="${BASH_REMATCH[1]}"
            continue
        fi
        
        if [[ -n "$current_db" && "$line" =~ ^([a-zA-Z0-9_]+):(.+)$ ]]; then
            username="${BASH_REMATCH[1]}"
            permissions="${BASH_REMATCH[2]}"
            GRANTS["$current_db,$username"]="$permissions"
            log_info "解析权限: $username -> $current_db ($permissions)"
        fi
    done < "$CONFIG_FILE"
}

# ==================== 创建用户函数 ====================
create_user() {
    local username=$1
    local password=$2
    local host=$3
    
    # 检查用户是否存在
    local user_exists_cmd="SELECT EXISTS(SELECT 1 FROM mysql.user WHERE user = '$username' AND host = '$host') AS user_exists;"
    
    if docker exec "$CONTAINER_NAME" mysql -u"$DB_USER" -p"$DB_PASSWORD" -se "$user_exists_cmd" | grep -q "1"; then
        log_info "用户 $username@$host 已存在，跳过创建"
        return 0
    fi
    
    local create_user_cmd="CREATE USER '$username'@'$host' IDENTIFIED BY '$password';"
    
    if execute_mysql_cmd "$CONTAINER_NAME" "$DB_USER" "$DB_PASSWORD" "$create_user_cmd"; then
        log_success "✅ 用户 $username@$host 创建成功"
        return 0
    else
        log_error "❌ 用户 $username@$host 创建失败"
        return 1
    fi
}

# ==================== 授予权限函数 ====================
grant_permissions() {
    local db_name=$1
    local username=$2
    local permissions=$3
    local host=$4
    
    # 检查用户是否存在
    local user_exists_cmd="SELECT EXISTS(SELECT 1 FROM mysql.user WHERE user = '$username' AND host = '$host') AS user_exists;"
    
    if ! docker exec "$CONTAINER_NAME" mysql -u"$DB_USER" -p"$DB_PASSWORD" -se "$user_exists_cmd" | grep -q "1"; then
        log_warn "用户 $username@$host 不存在，跳过权限授予"
        return 1
    fi
    
    local grant_cmd="GRANT $permissions ON \`$db_name\`.* TO '$username'@'$host';"
    
    if execute_mysql_cmd "$CONTAINER_NAME" "$DB_USER" "$DB_PASSWORD" "$grant_cmd"; then
        log_success "✅ 授予 $username@$host $permissions 权限 on $db_name.*"
        return 0
    else
        log_error "❌ 授予权限失败: $username@$host -> $db_name"
        return 1
    fi
}

# ==================== 主函数 ====================
main() {
    > "$LOG_FILE"
    log_info "开始执行用户创建和权限授予脚本"
    
    # 验证配置文件
    if [[ ! -f "$CONFIG_FILE" ]]; then
        log_error "配置文件 $CONFIG_FILE 不存在"
        exit 1
    fi
    
    # 解析配置
    declare -A USERS
    declare -A GRANTS
    
    parse_users_config
    
    # 读取MySQL连接配置（从数据库配置文件复用）
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
    
    log_info "MySQL 配置: 容器=$CONTAINER_NAME, 用户=$DB_USER"
    log_info "检测到 ${#USERS[@]} 个用户需要创建"
    log_info "检测到 ${#GRANTS[@]} 个权限授予需要执行"
    
    # 测试MySQL连接
    if ! test_mysql_connection "$CONTAINER_NAME" "$DB_USER" "$DB_PASSWORD"; then
        log_error "MySQL 连接失败，退出脚本"
        exit 1
    fi
    
    # 创建用户
    SUCCESS_USERS=()
    FAILED_USERS=()
    
    for user_host in "${!USERS[@]}"; do
        IFS=',' read -r username host <<< "$user_host"
        password="${USERS[$user_host]}"
        
        if create_user "$username" "$password" "$host"; then
            SUCCESS_USERS+=("$username@$host")
        else
            FAILED_USERS+=("$username@$host")
        fi
    done
    
    # 刷新权限
    execute_mysql_cmd "$CONTAINER_NAME" "$DB_USER" "$DB_PASSWORD" "FLUSH PRIVILEGES;"
    
    # 授予权限
    SUCCESS_GRANTS=()
    FAILED_GRANTS=()
    
    for db_user in "${!GRANTS[@]}"; do
        IFS=',' read -r db_name username <<< "$db_user"
        permissions="${GRANTS[$db_user]}"
        
        # 尝试常见的host配置
        grant_success=false
        for host in "%" "localhost"; do
            if [[ -n "${USERS[$username,$host]+_}" ]]; then
                if grant_permissions "$db_name" "$username" "$permissions" "$host"; then
                    SUCCESS_GRANTS+=("$username@$host -> $db_name")
                    grant_success=true
                    break
                fi
            fi
        done
        
        if ! $grant_success; then
            FAILED_GRANTS+=("$username -> $db_name")
        fi
    done
    
    # 再次刷新权限
    execute_mysql_cmd "$CONTAINER_NAME" "$DB_USER" "$DB_PASSWORD" "FLUSH PRIVILEGES;"
    
    # 结果汇总
    echo -e "\n${BLUE}==================== 执行结果 ====================${NC}" | tee -a "$LOG_FILE"
    [ ${#SUCCESS_USERS[@]} -gt 0 ] && log_success "成功创建用户: ${SUCCESS_USERS[*]}"
    [ ${#SUCCESS_GRANTS[@]} -gt 0 ] && log_success "成功授予权限: ${SUCCESS_GRANTS[*]}"
    
    if [ ${#FAILED_USERS[@]} -gt 0 ] || [ ${#FAILED_GRANTS[@]} -gt 0 ]; then
        [ ${#FAILED_USERS[@]} -gt 0 ] && log_error "创建失败的用户: ${FAILED_USERS[*]}"
        [ ${#FAILED_GRANTS[@]} -gt 0 ] && log_error "授予失败的权限: ${FAILED_GRANTS[*]}"
        echo -e "${RED}请检查日志文件: $LOG_FILE${NC}" | tee -a "$LOG_FILE"
        exit 1
    else
        log_success "🎉 所有用户创建和权限授予成功！"
    fi
    echo -e "${BLUE}==================================================${NC}" | tee -a "$LOG_FILE"
}

# ==================== 执行主函数 ====================
main "$@"