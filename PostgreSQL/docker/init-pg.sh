#!/bin/bash
# 由豆包创建
# PostgreSQL 初始化脚本
# 功能：固定容器init-pg | 自动从规则提取数据库 | 支持仅建库(数据库名:) | 按序导入SQL | 错误追踪
# 支持格式：
# 1. 库名:文件1,文件2 → 建库+导入SQL
# 2. 库名:           → 仅建库，不导入任何SQL

# 严格模式（不强制退出，仅捕获错误）
set -uo pipefail

##############################################################################
# 1. 【核心配置】仅需修改这里
##############################################################################
# PostgreSQL 连接信息
PG_HOST="postgresql"
PG_PORT="5432"
PG_USER="hive"
PG_PWD="hive"

# SQL文件目录
SQL_DIR="./sql"

# 导入规则（支持 仅创建库 / 建库+导入SQL）
IMPORT_MAP=(
  "metastore_db1:test_schema.sql,test_data.sql"  # 建库+导入双文件
  "metastore_db2:"                                # 仅创建库，无SQL
  "test_db:"                                      # 仅创建库，无SQL
)

# Docker配置
PG_IMAGE="postgres:18.3"
CONTAINER_NAME="init-pg"
DOCKER_NETWORK="pg-net"

##############################################################################
# 2. 全局变量
##############################################################################
FAILED_ITEMS=()
SUCCESS_COUNT=0

##############################################################################
# 3. 工具函数
##############################################################################
# 检查Docker是否安装
check_docker() {
  if ! command -v docker &> /dev/null; then
    echo "❌ 错误：未安装Docker"
    exit 1
  fi
}

# 检查SQL目录
check_sql_dir() {
  if [ ! -d "${SQL_DIR}" ]; then
    echo "⚠️  警告：SQL目录 ${SQL_DIR} 不存在，仅执行创建数据库操作"
  fi
}

# 自动从IMPORT_MAP提取所有数据库名
get_databases() {
  local dbs=()
  for rule in "${IMPORT_MAP[@]}"; do
    db=$(echo "${rule}" | cut -d':' -f1)
    dbs+=("${db}")
  done
  echo "${dbs[@]}"
}

# 创建数据库（不存在则创建，存在跳过）
create_database() {
  local db="$1"
  echo -e "\n→ 检查数据库：${db}"

  # 检测数据库是否存在
  docker run --rm --name "${CONTAINER_NAME}" \
    --network "${DOCKER_NETWORK}" \
    -e PGPASSWORD="${PG_PWD}" \
    "${PG_IMAGE}" \
    psql -h "${PG_HOST}" -p "${PG_PORT}" -U "${PG_USER}" -d postgres -t \
    -c "SELECT 1 FROM pg_database WHERE datname='${db}';" | grep -q 1

  if [ $? -eq 0 ]; then
    echo "✅ 数据库 ${db} 已存在，跳过创建"
    return
  fi

  # 创建数据库
  echo "🔨 创建数据库：${db}"
  docker run --rm --name "${CONTAINER_NAME}" \
    --network "${DOCKER_NETWORK}" \
    -e PGPASSWORD="${PG_PWD}" \
    "${PG_IMAGE}" \
    psql -h "${PG_HOST}" -p "${PG_PORT}" -U "${PG_USER}" -d postgres \
    -c "CREATE DATABASE ${db};"

  if [ $? -eq 0 ]; then
    echo "✅ 数据库 ${db} 创建成功"
  else
    FAILED_ITEMS+=("创建数据库失败：${db}")
  fi
}

# 导入SQL（空文件则跳过）
import_sql() {
  local db="$1"
  local files="$2"

  # 如果文件为空 → 仅创建库，跳过导入
  if [ -z "${files}" ] || [ "${files}" = " " ]; then
    echo -e "\nℹ️  数据库 ${db} 无SQL文件，跳过导入"
    return
  fi

  # 分割文件列表
  IFS=',' read -ra FILE_LIST <<< "${files}"
  for file in "${FILE_LIST[@]}"; do
    local sql_path="${SQL_DIR}/${file}"
    echo -e "\n→ 导入 ${db} <- ${file}"

    # 检查文件是否存在
    if [ ! -f "${sql_path}" ]; then
      FAILED_ITEMS+=("文件不存在：${sql_path}")
      continue
    fi

    # 执行SQL导入
    docker run --rm --name "${CONTAINER_NAME}" \
      --network "${DOCKER_NETWORK}" \
      -e PGPASSWORD="${PG_PWD}" \
      -v "$(realpath ${SQL_DIR}):/sql" \
      "${PG_IMAGE}" \
      psql -h "${PG_HOST}" -p "${PG_PORT}" -U "${PG_USER}" -d "${db}" -f "/sql/${file}"

    if [ $? -eq 0 ]; then
      echo "✅ 导入成功：${file}"
      SUCCESS_COUNT=$((SUCCESS_COUNT+1))
    else
      FAILED_ITEMS+=("导入失败：${db} -> ${file}")
    fi
  done
}

##############################################################################
# 4. 主执行逻辑
##############################################################################
echo "============================================="
echo "      PostgreSQL 初始化脚本 (容器：init-pg)"
echo "============================================="

# 前置检查
check_docker
check_sql_dir

# 自动创建所有数据库
echo -e "\n========== 开始创建数据库 =========="
DATABASES=$(get_databases)
for db in ${DATABASES}; do
  create_database "${db}"
done

# 导入SQL文件
echo -e "\n========== 开始执行SQL导入 =========="
for rule in "${IMPORT_MAP[@]}"; do
  # 分割库名和文件
  IFS=':' read -ra PARTS <<< "${rule}"
  import_sql "${PARTS[0]}" "${PARTS[1]:-}"
done

##############################################################################
# 5. 最终结果输出
##############################################################################
echo -e "\n============================================="
echo "                  执行结果                  "
echo "============================================="

if [ ${#FAILED_ITEMS[@]} -eq 0 ]; then
  echo -e "\n🎉 所有操作执行成功！"
else
  echo "❌ 执行失败，共 ${#FAILED_ITEMS[@]} 项错误："
  for item in "${FAILED_ITEMS[@]}"; do
    echo "   - ${item}"
  done
  exit 1
fi