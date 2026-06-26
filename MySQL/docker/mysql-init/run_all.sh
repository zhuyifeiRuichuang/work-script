#!/bin/bash

set -e

echo "🚀 开始执行完整的MySQL初始化流程..."

# 1. 创建数据库
echo -e "\n📁 步骤1: 创建数据库"
./create_databases.sh
echo "✅ 数据库创建完成"

# 2. 创建用户和权限
echo -e "\n👤 步骤2: 创建用户和权限"
./create_users.sh
echo "✅ 用户和权限设置完成"

# 3. 导入SQL文件
echo -e "\n💾 步骤3: 导入SQL文件"
./import_sql.sh
echo "✅ SQL文件导入完成"

echo -e "\n🎉 完整的MySQL初始化流程成功完成！"