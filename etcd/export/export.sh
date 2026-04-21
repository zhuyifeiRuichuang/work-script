#!/bin/bash

# 初始化变量
base_dir=""
log_dir=""
resource_file=""
config_file=""
ETCD_ENDPOINTS="http://127.0.0.1:2389"
ETCDCTL_PATH="etcdctl"
ETCDCTL_CMD=""
RESOURCES=()
TOTAL_RESOURCES=0
PROCESSED_RESOURCES=0
LOG_FILE=""

# 打印帮助信息
print_help() {
    echo "用法: $0 -c <配置文件> -r <资源文件> [选项]"
    echo "导出etcd中的资源到指定目录"
    echo
    echo "必选参数:"
    echo "  -c, --config        指定配置文件路径"
    echo "  -r, --resourcefile  指定包含资源类型列表的文件"
    echo
    echo "可选参数:"
    echo "  -h, --help          显示帮助信息"
    echo
    echo "配置文件格式示例:"
    echo "  base_file=\"/tmp\""
    echo "  log_dir=\"/tmp/logs\""
    echo "  ETCD_ENDPOINTS=\"http://127.0.0.1:2389\""
    echo "  ETCDCTL_PATH=\"etcdctl\""
    echo
    echo "资源文件格式示例:"
    echo "  # 注释行将被忽略"
    echo "  namespaces"
    echo "  devopsprojects"
    echo "  appprojects"
    echo
    echo "示例:"
    echo "  $0 -c export_etcd_resources.conf -r resources.list"
}

# 读取配置文件
read_config() {
    local file="$1"
    
    if [ ! -f "$file" ]; then
        echo "错误：配置文件 $file 不存在" >&2
        exit 1
    fi
    
    if [ ! -r "$file" ]; then
        echo "错误：没有权限读取配置文件 $file" >&2
        exit 1
    fi
    
    # 读取配置文件中的变量
    source "$file"
    
    # 检查必要的配置项
    if [ -z "$ETCD_ENDPOINTS" ] || [ -z "$ETCDCTL_PATH" ]; then
        echo "错误：配置文件中必须包含 ETCD_ENDPOINTS 和 ETCDCTL_PATH" >&2
        exit 1
    fi
    
    # 设置默认值（如果配置文件中未指定）
    if [ -z "$base_file" ]; then
        base_file="/tmp"
    fi
    
    if [ -z "$log_dir" ]; then
        log_dir="${base_file}/logs"
    fi
    
    # 导出目录使用base_file
    out_dir="$base_file"
}

# 解析命令行参数
parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            -c|--config)
                config_file="$2"
                shift 2
                ;;
            -r|--resourcefile)
                resource_file="$2"
                shift 2
                ;;
            -h|--help)
                print_help
                exit 0
                ;;
            *)
                echo "错误：未知参数 $1" >&2
                print_help
                exit 1
                ;;
        esac
    done

    # 检查是否提供了必要参数
    if [ -z "$config_file" ]; then
        echo "错误：必须使用 -c 或 --config 指定配置文件" >&2
        print_help
        exit 1
    fi

    if [ -z "$resource_file" ]; then
        echo "错误：必须使用 -r 或 --resourcefile 指定资源文件" >&2
        print_help
        exit 1
    fi

    # 读取配置文件
    read_config "$config_file"

    # 从文件读取资源列表
    if [ ! -f "$resource_file" ]; then
        echo "错误：资源清单文件 $resource_file 不存在" >&2
        exit 1
    fi
    
    # 读取文件内容，忽略空行和注释
    while IFS= read -r line; do
        # 跳过空行和注释行
        [[ -z "$line" || "$line" =~ ^# ]] && continue
        RESOURCES+=("$line")
    done < "$resource_file"
    
    if [ ${#RESOURCES[@]} -eq 0 ]; then
        echo "错误：资源清单文件 $resource_file 中没有有效的资源类型" >&2
        exit 1
    fi
    
    TOTAL_RESOURCES=${#RESOURCES[@]}

    # 设置etcdctl命令
    ETCDCTL_CMD="ETCDCTL_API=3 ${ETCDCTL_PATH} --endpoints=${ETCD_ENDPOINTS}"
}

# 初始化日志
init_log() {
    # 确保日志目录存在
    mkdir -p "$log_dir" || {
        echo "错误：无法创建日志目录 $log_dir" >&2
        exit 1
    }

    LOG_FILE="${log_dir}/etcd_export_$(date +%Y%m%d_%H%M%S).log"
    
    # 获取日志文件绝对路径
    LOG_FILE=$(realpath "$LOG_FILE")
    
    # 重定向输出到日志文件和终端
    exec > >(tee -a "$LOG_FILE") 2>&1
}

# 显示总体进度
show_overall_progress() {
    local percent=$(( (PROCESSED_RESOURCES * 100) / TOTAL_RESOURCES ))
    local resource=$1
    local status=$2
    
    # 输出进度信息，使用回车而不是换行，实现覆盖效果
    echo -ne "总体进度: [${percent}%] 已处理: ${PROCESSED_RESOURCES}/${TOTAL_RESOURCES} | 最后处理: ${resource} [${status}] \r"
    
    # 最后一行输出后添加换行
    if [ $PROCESSED_RESOURCES -eq $TOTAL_RESOURCES ]; then
        echo -e "\n"
    fi
}

# 检查etcdctl是否可执行
check_etcdctl() {
    if [ ! -x "$ETCDCTL_PATH" ]; then
        echo "错误：etcdctl不存在或不可执行 - $ETCDCTL_PATH" >&2
        echo "请检查路径是否正确，并确保有执行权限" >&2
        exit 1
    fi
    
    # 测试连接
    echo "测试etcd连接..."
    if ! eval "${ETCDCTL_CMD} endpoint health > /dev/null 2>&1"; then
        echo "错误：无法连接到etcd服务 - $ETCD_ENDPOINTS" >&2
        echo "请检查endpoints是否正确，etcd服务是否运行" >&2
        exit 1
    fi
    echo "etcd连接测试成功"
}

# 导出资源的函数
export_resource() {
    local resource_type=$1
    local target_dir="${out_dir}/${resource_type}"
    local list_file="${target_dir}/${resource_type}.list"
    local status="失败"
    
    echo "开始处理 ${resource_type} 资源..."
    
    # 创建目标目录
    if ! mkdir -p "$target_dir"; then
        echo "错误：无法创建目录 $target_dir" >&2
        return 1
    fi
    
    # 构建查询命令
    local query_cmd="${ETCDCTL_CMD} get --prefix \"\" --keys-only"
    local filter_cmd="grep \"${resource_type}\""
    local full_cmd="${query_cmd} | ${filter_cmd} > \"${list_file}\""
    
    # 查询资源列表并保存到文件
    echo "查询 ${resource_type} 资源列表..."
    
    if ! eval "$full_cmd"; then
        echo "错误：查询 ${resource_type} 资源列表失败" >&2
        return 1
    fi
    
    # 检查是否有资源需要导出
    if [ ! -s "$list_file" ]; then
        echo "警告：未找到 ${resource_type} 资源"
        status="无资源"
        return 0
    fi
    
    # 显示找到的资源数量
    local count=$(wc -l < "$list_file")
    echo "找到 ${count} 个 ${resource_type} 资源，开始导出..."
    
    # 导出每个资源为yaml文件
    local success_count=0
    local total_count=$count
    
    while IFS= read -r resource_path; do
        # 跳过空行
        [ -z "$resource_path" ] && continue
        
        # 提取资源名称
        resource_name=$(basename "$resource_path")
        yaml_file="${target_dir}/${resource_name}.yaml"
        
        # 导出资源
        export_cmd="${ETCDCTL_CMD} get \"${resource_path}\" --print-value-only > \"${yaml_file}\""
        if eval "$export_cmd"; then
            ((success_count++))
        else
            echo "错误：导出 ${resource_path} 失败" >&2
        fi
        
        # 显示当前资源类型的导出进度
        local item_percent=$(( (success_count * 100) / total_count ))
        echo -ne "  ${resource_type} 导出进度: [${item_percent}%] (${success_count}/${total_count}) \r"
    done < "$list_file"
    
    # 完成当前资源类型的导出进度显示
    echo -e "\n${resource_type} 资源处理完成，成功导出 ${success_count}/${total_count} 个资源"
    echo "----------------------------------------"
    
    status="成功"
    return 0
}

# 主程序
parse_args "$@"
init_log
check_etcdctl

echo "使用配置文件: $config_file"
echo "使用资源文件: $resource_file"
echo "导出目录: $out_dir"
echo "日志目录: $log_dir"
echo "总资源类型数量: $TOTAL_RESOURCES"
echo "处理资源类型: ${RESOURCES[*]}"
echo
echo "开始处理资源..."
echo "----------------------------------------"

for res in "${RESOURCES[@]}"; do
    if export_resource "$res"; then
        status="成功"
    else
        status="失败"
    fi
    ((PROCESSED_RESOURCES++))
    show_overall_progress "$res" "$status"
done

echo "所有资源导出操作已完成"
echo "详细操作日志已保存至: $LOG_FILE"
echo "资源已导出至目录: $out_dir"
