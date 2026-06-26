#!/bin/bash

set -x

# =========================================================================
# DYNAMIC JAR LOADER (AWS/S3 Support)
# =========================================================================
STAGING_DIR="/tmp/ext-jars"

# Checks if /tmp/ext-jars is mounted (via Docker volume).
if [ -d "$STAGING_DIR" ]; then
  if ls "$STAGING_DIR"/*.jar 1> /dev/null 2>&1; then
    echo "--> Copying custom jars from volume to Hive..."
    cp -vf "$STAGING_DIR"/*.jar "${HIVE_HOME}/lib/"
  else
    echo "--> Volume mounted at $STAGING_DIR, but no jars found."
  fi
fi

# =========================================================================
# REPLACE ${VARS} in the template
# =========================================================================
: "${HIVE_WAREHOUSE_PATH:=/opt/hive/data/warehouse}"
export HIVE_WAREHOUSE_PATH

# 仅在环境变量未外部设置时，才使用默认值（兼容 compose/k8s 外部注入）
: "${HADOOP_CONF_DIR:=${HIVE_HOME}/conf}"
export HADOOP_CONF_DIR
: "${TEZ_CONF_DIR:=${HIVE_HOME}/conf}"
export TEZ_CONF_DIR

if [ -f "$HIVE_HOME/conf/core-site.xml.template" ]; then
  envsubst < $HIVE_HOME/conf/core-site.xml.template > $HIVE_HOME/conf/core-site.xml
fi
if [ -f "$HIVE_HOME/conf/hive-site.xml.template" ]; then
  envsubst < $HIVE_HOME/conf/hive-site.xml.template > $HIVE_HOME/conf/hive-site.xml
fi
# =========================================================================

# 修复1：给所有关键变量设置默认值，避免空值
: "${DB_DRIVER:=derby}"
: "${HIVE_VER:=3.1.3}"  # 设置默认Hive版本，适配无HIVE_VER的场景
: "${VERBOSE:=false}"
: "${IS_RESUME:=false}"
: "${SCHEMA_COMMAND:=initSchema}"
: "${SERVICE_NAME:=metastore}"  # 默认启动metastore
: "${SERVICE_OPTS:=}"
: "${TEZ_HOME:=/opt/tez}"  # 兼容无Tez的场景

SKIP_SCHEMA_INIT="${IS_RESUME}"
VERBOSE_MODE=""
[[ $VERBOSE = "true" ]] && VERBOSE_MODE="--verbose"

function initialize_hive {
  # 修复2：修正不合法的initOrUpgradeSchema参数，适配Hive版本
  local COMMAND=""
  HIVE_MAJOR_VER=$(echo "$HIVE_VER" | cut -d '.' -f1)
  
  # 处理空值/非数字的HIVE_MAJOR_VER，避免整数比较错误
  if ! [[ "$HIVE_MAJOR_VER" =~ ^[0-9]+$ ]]; then
    echo "WARN: 无法识别Hive主版本，默认使用-initSchema参数"
    COMMAND="-${SCHEMA_COMMAND}"
  elif [ "$HIVE_MAJOR_VER" -lt "4" ]; then
    COMMAND="-${SCHEMA_COMMAND}"
  else
    # Hive 4.x 仍使用-initSchema（无initOrUpgradeSchema参数）
    COMMAND="-${SCHEMA_COMMAND}"
  fi

  # 执行schematool，增加错误处理
  if [[ -n "$VERBOSE_MODE" ]]; then
    "$HIVE_HOME/bin/schematool" -dbType "$DB_DRIVER" "$COMMAND" "$VERBOSE_MODE"
  else
    "$HIVE_HOME/bin/schematool" -dbType "$DB_DRIVER" "$COMMAND"
  fi
  
  if [ $? -eq 0 ]; then
    echo "Initialized Hive Metastore Server schema successfully.."
  else
    echo "Hive Metastore Server schema initialization failed!"
    exit 1
  fi
}

export HIVE_CONF_DIR=$HIVE_HOME/conf
if [ -d "${HIVE_CUSTOM_CONF_DIR:-}" ]; then
  find "${HIVE_CUSTOM_CONF_DIR}" -type f -exec \
    ln -sfn {} "${HIVE_CONF_DIR}"/ \;
  export HADOOP_CONF_DIR=$HIVE_CONF_DIR
  export TEZ_CONF_DIR=$HIVE_CONF_DIR
fi

export HADOOP_CLIENT_OPTS="$HADOOP_CLIENT_OPTS -Xmx1G $SERVICE_OPTS"
if [[ "${SKIP_SCHEMA_INIT}" == "false" ]]; then
  # handles schema initialization
  initialize_hive
fi

# 修复3：兼容无Tez的场景 + 处理未知SERVICE_NAME
if [ "${SERVICE_NAME}" == "hiveserver2" ]; then
  # 仅当TEZ_HOME目录存在时，才添加Tez到CLASSPATH
  if [ -d "$TEZ_HOME" ]; then
    export HADOOP_CLASSPATH="$TEZ_HOME/*:$TEZ_HOME/lib/*:$HADOOP_CLASSPATH"
  fi
  exec "$HIVE_HOME/bin/hive" --skiphadoopversion --skiphbasecp --service "$SERVICE_NAME"
elif [ "${SERVICE_NAME}" == "metastore" ]; then
  export METASTORE_PORT=${METASTORE_PORT:-9083}
  if [[ -n "$VERBOSE_MODE" ]]; then
    exec "$HIVE_HOME/bin/hive" --skiphadoopversion --skiphbasecp "$VERBOSE_MODE" --service "$SERVICE_NAME"
  else
    exec "$HIVE_HOME/bin/hive" --skiphadoopversion --skiphbasecp --service "$SERVICE_NAME"
  fi
else
  # 修复4：处理未知SERVICE_NAME，给出提示并退出
  echo "ERROR: 不支持的SERVICE_NAME: ${SERVICE_NAME}，仅支持hiveserver2/metastore"
  exit 1
fi
