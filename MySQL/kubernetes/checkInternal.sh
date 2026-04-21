#!/bin/bash
set -euo pipefail

# 测试集群内访问有效性。
# ==============================================
# 可自定义变量（根据实际环境修改）
# ==============================================
MYSQL_NAMESPACE="mysql"          # MySQL所在的namespace
TEST_OTHER_NAMESPACE="default"      # 跨namespace测试用的目标namespace
TEST_IMAGE="mysql:5.7.44"           # 测试用容器镜像（需包含mysql客户端）
MYSQL_SERVICE_NAME="mysql-service"  # MySQL的Service名称
MYSQL_PORT="3306"                   # 集群内访问端口
MYSQL_USER="root"                   # MySQL用户名
MYSQL_PASSWORD="root123456"         # MySQL密码
# ==============================================

# 颜色定义（用于美化输出）
GREEN="\033[0;32m"
RED="\033[0;31m"
NC="\033[0m" # 无颜色

# 测试结果函数
print_result() {
    local scenario=$1
    local status=$2
    local message=$3
    echo -e "【测试场景】$scenario"
    if [ "$status" -eq 0 ]; then
        echo -e "【结果】${GREEN}成功${NC}"
    else
        echo -e "【结果】${RED}失败${NC}"
        echo -e "【原因】$message"
    fi
    echo "----------------------------------------"
}

# 1. 同namespace访问测试
echo "开始执行同namespace访问测试..."
same_ns_result=0
same_ns_msg=""
# 执行同namespace测试命令
same_ns_output=$(kubectl run -it --rm \
    --image="$TEST_IMAGE" \
    --namespace="$MYSQL_NAMESPACE" \
    --restart=Never \
    mysql-test-same-ns \
    -- sh -c "mysql -h $MYSQL_SERVICE_NAME -P $MYSQL_PORT -u $MYSQL_USER -p'$MYSQL_PASSWORD' -e 'SELECT 1;' 2>&1" || true)

# 判断结果
if echo "$same_ns_output" | grep -q "1"; then
    same_ns_result=0
else
    same_ns_result=1
    same_ns_msg=$(echo "$same_ns_output" | tail -n 3 | tr '\n' ' ') # 获取最后3行错误信息
fi
print_result "同namespace（$MYSQL_NAMESPACE）访问MySQL" $same_ns_result "$same_ns_msg"

# 2. 跨namespace访问测试
echo "开始执行跨namespace访问测试..."
cross_ns_result=0
cross_ns_msg=""
# 完整Service域名：service名称.命名空间.svc.cluster.local
full_service_name="${MYSQL_SERVICE_NAME}.${MYSQL_NAMESPACE}.svc.cluster.local"

# 执行跨namespace测试命令
cross_ns_output=$(kubectl run -it --rm \
    --image="$TEST_IMAGE" \
    --namespace="$TEST_OTHER_NAMESPACE" \
    --restart=Never \
    mysql-test-cross-ns \
    -- sh -c "mysql -h $full_service_name -P $MYSQL_PORT -u $MYSQL_USER -p'$MYSQL_PASSWORD' -e 'SELECT 1;' 2>&1" || true)

# 判断结果
if echo "$cross_ns_output" | grep -q "1"; then
    cross_ns_result=0
else
    cross_ns_result=1
    cross_ns_msg=$(echo "$cross_ns_output" | tail -n 3 | tr '\n' ' ') # 获取最后3行错误信息
fi
print_result "跨namespace（$TEST_OTHER_NAMESPACE -> $MYSQL_NAMESPACE）访问MySQL" $cross_ns_result "$cross_ns_msg"

# 最终总结
echo -e "\n【测试总结】"
if [ $same_ns_result -eq 0 ] && [ $cross_ns_result -eq 0 ]; then
    echo -e "${GREEN}所有测试场景均通过，MySQL集群内访问正常${NC}"
else
    echo -e "${RED}部分测试场景失败，请根据错误信息排查${NC}"
fi
