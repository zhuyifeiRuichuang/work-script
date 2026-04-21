#!/bin/bash

set -e

# Source hadoop environment
source ${HADOOP_HOME}/etc/hadoop/hadoop-env.sh

# Function to wait for service
wait_for_service() {
    local host=$1
    local port=$2
    local service=$3
    local max_attempts=${4:-30}
    local attempt=1

    echo "Waiting for $service at $host:$port..."
    while ! nc -z "$host" "$port"; do
        if [ $attempt -eq $max_attempts ]; then
            echo "ERROR: $service at $host:$port is not available after $max_attempts attempts"
            return 1
        fi
        echo "Attempt $attempt/$max_attempts: $service not ready, waiting..."
        sleep 5
        attempt=$((attempt + 1))
    done
    echo "$service at $host:$port is ready"
    return 0
}

# Function to format NameNode if needed
format_namenode() {
    local nn_dir=${HADOOP_HOME}/data/nn
    if [ ! -f "$nn_dir/current/VERSION" ]; then
        echo "NameNode not formatted, formatting now..."
        $HADOOP_HOME/bin/hdfs namenode -format -force -nonInteractive
        echo "NameNode formatted successfully"
    else
        echo "NameNode already formatted, skipping format"
    fi
}

# Determine the component to start
COMPONENT=${1:-"namenode"}

echo "Starting Hadoop component: $COMPONENT"
echo "HADOOP_HOME: $HADOOP_HOME"
echo "HADOOP_CONF_DIR: $HADOOP_CONF_DIR"

# Execute the requested component
case "$COMPONENT" in
    namenode)
        echo "Starting NameNode..."
        format_namenode
        exec ${HADOOP_HOME}/bin/hdfs namenode
        ;;
    datanode)
        echo "Starting DataNode..."
        exec ${HADOOP_HOME}/bin/hdfs datanode
        ;;
    resourcemanager)
        echo "Starting ResourceManager..."
        exec ${HADOOP_HOME}/bin/yarn resourcemanager
        ;;
    nodemanager)
        echo "Starting NodeManager..."
        exec ${HADOOP_HOME}/bin/yarn nodemanager
        ;;
    all)
        echo "Starting all Hadoop services..."
        echo "This mode is not recommended for Kubernetes, use individual components"
        exec ${HADOOP_HOME}/bin/start-all.sh
        ;;
    *)
        echo "Unknown component: $COMPONENT"
        echo "Usage: $0 {namenode|datanode|resourcemanager|nodemanager|all}"
        exit 1
        ;;
esac
