#!/usr/bin/env bash
set -e

# 获取脚本所在目录
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# 1. 等待逻辑
if [ -n "$SLEEP_SECONDS" ]; then
   echo "Sleeping for $SLEEP_SECONDS seconds"
   sleep "$SLEEP_SECONDS"
fi

# 2. 端口等待逻辑 (WAITFOR)
if [ -n "$WAITFOR" ]; then
  echo "Waiting for the service $WAITFOR"
  WAITFOR_HOST=$(printf "%s\n" "$WAITFOR"| cut -d : -f 1)
  WAITFOR_PORT=$(printf "%s\n" "$WAITFOR"| cut -d : -f 2)
  # 修正 seq 的使用方式，确保兼容性
  for i in $(seq ${WAITFOR_TIMEOUT:-300} -1 0) ; do
    set +e
    nc -z "$WAITFOR_HOST" "$WAITFOR_PORT" > /dev/null 2>&1
    result=$?
    set -e
    if [ $result -eq 0 ] ; then
      break
    fi
    sleep 1
  done
  if [ "$i" -eq 0 ]; then
      echo "Waiting for service $WAITFOR is timed out." >&2
      exit 1
  fi
fi

# 3. Kerberos 设置
if [ -n "$KERBEROS_ENABLED" ]; then
  echo "Setting up kerberos!!"
  KERBEROS_SERVER=${KERBEROS_SERVER:-krb5}
  ISSUER_SERVER=${ISSUER_SERVER:-$KERBEROS_SERVER:8081}
  echo "KDC ISSUER_SERVER => $ISSUER_SERVER"

  if [ -n "$SLEEP_SECONDS" ]; then
    # 修正了之前的 $(SLEEP_SECONDS) 语法错误
    echo "Sleeping for ${SLEEP_SECONDS} seconds"
    sleep "${SLEEP_SECONDS}"
  fi

  KEYTAB_DIR=${KEYTAB_DIR:-/etc/security/keytabs}

  while true; do
      set +e
      STATUS=$(curl -s -o /dev/null -w '%{http_code}' http://"$ISSUER_SERVER"/keytab/test/test)
      set -e
      if [ "$STATUS" -eq 200 ]; then
        echo "Got 200, KDC service ready!!"
        break
      else
        echo "Got $STATUS :( KDC service not ready yet..."
      fi
      sleep 5
  done

  HOST_NAME=$(hostname -f)
  export HOST_NAME
  for NAME in ${KERBEROS_KEYTABS}; do
    echo "Download $NAME/$HOST_NAME@EXAMPLE.COM keytab file to $KEYTAB_DIR/$NAME.keytab"
    wget -q "http://$ISSUER_SERVER/keytab/$HOST_NAME/$NAME" -O "$KEYTAB_DIR/$NAME.keytab"
    klist -kt "$KEYTAB_DIR/$NAME.keytab"
  done

  # 适配 Ubuntu 的配置文件路径
  sed "s/SERVER/$KERBEROS_SERVER/g" "$DIR"/krb5.conf | sudo tee /etc/krb5.conf > /dev/null
fi

# 4. 权限修复 (针对 Docker 挂载卷)
sudo chmod o+rwx /data

# 5. 调用 Python 3 转换配置 (关键修改)
python3 "$DIR"/envtoconf.py --destination "${HADOOP_CONF_DIR:-/opt/hadoop/etc/hadoop}"

# 6. Hadoop/Ozone 初始化逻辑 (保持原样，但确保路径正确)
if [ -n "$ENSURE_NAMENODE_DIR" ]; then
  CLUSTERID_OPTS=""
  if [ -n "$ENSURE_NAMENODE_CLUSTERID" ]; then
    CLUSTERID_OPTS="-clusterid $ENSURE_NAMENODE_CLUSTERID"
  fi
  if [ ! -d "$ENSURE_NAMENODE_DIR" ]; then
    /opt/hadoop/bin/hdfs namenode -format -force $CLUSTERID_OPTS
  fi
fi

if [ -n "$ENSURE_STANDBY_NAMENODE_DIR" ]; then
  if [ ! -d "$ENSURE_STANDBY_NAMENODE_DIR" ]; then
    /opt/hadoop/bin/hdfs namenode -bootstrapStandby
  fi
fi

# Ozone 相关的初始化
if [ -n "$ENSURE_SCM_INITIALIZED" ]; then
  if [ ! -f "$ENSURE_SCM_INITIALIZED" ]; then
    /opt/hadoop/bin/ozone scm --init || /opt/hadoop/bin/ozone scm -init
  fi
fi

if [ -n "$ENSURE_OM_INITIALIZED" ]; then
  if [ ! -f "$ENSURE_OM_INITIALIZED" ]; then
    /opt/hadoop/bin/ozone om --init || /opt/hadoop/bin/ozone om -createObjectStore
  fi
fi

# 7. Byteman 注入
if [ -n "$BYTEMAN_SCRIPT" ] || [ -n "$BYTEMAN_SCRIPT_URL" ]; then
  # 确保 BYTEMAN_DIR 已定义
  BYTEMAN_DIR=${BYTEMAN_DIR:-/opt/profiler} 
  export PATH=$PATH:$BYTEMAN_DIR/bin

  if [ -n "$BYTEMAN_SCRIPT_URL" ]; then
    sudo wget -q $BYTEMAN_SCRIPT_URL -O /tmp/byteman.btm
    export BYTEMAN_SCRIPT=/tmp/byteman.btm
  fi

  if [ ! -f "$BYTEMAN_SCRIPT" ]; then
    echo "ERROR: The defined $BYTEMAN_SCRIPT does not exist!!!"
    exit 1
  fi

  AGENT_STRING="-javaagent:/opt/byteman.jar=script:$BYTEMAN_SCRIPT"
  export HADOOP_OPTS="$AGENT_STRING $HADOOP_OPTS"
  echo "Process is instrumented with $AGENT_STRING"
fi

# 执行 CMD 传入的命令
exec "$@"