-- ====================================================================
-- PostgreSQL 生产环境综合初始化脚本
-- 执行机制：仅在数据库容器初次启动且 PGDATA 目录为空时执行
-- ====================================================================

\echo '=================== [1/4] 开始创建业务用户 ==================='

-- 创建业务系统应用账号
CREATE USER app_user WITH ENCRYPTED PASSWORD 'AppSecret2026!#';
-- 创建只读/数据分析账号
CREATE USER report_user WITH ENCRYPTED PASSWORD 'ReportRead2026@!';

\echo '=================== [2/4] 开始创建数据库 ==================='

-- 创建业务数据库，并指定所有者为 app_user
CREATE DATABASE app_db OWNER app_user;
-- 创建报表/数仓数据库，指定所有者为 report_user
CREATE DATABASE report_db OWNER report_user;

\echo '=================== [3/4] 初始化 app_db 数据库及权限 ==================='

-- 使用 \connect (\c) 切换到刚创建的业务数据库
\connect app_db;

-- 安装生产必备扩展 (必须在具体数据库内启用才对该库生效)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 提升安全性：撤销默认 public schema 下开放给所有人的访问权限
REVOKE ALL ON SCHEMA public FROM PUBLIC;

-- 分配读写权限：赋予业务账号在 public schema 的全量权限
GRANT ALL ON SCHEMA public TO app_user;

-- 分配只读权限：给报表账号开放当前表的查询权限
GRANT USAGE ON SCHEMA public TO report_user;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO report_user;
-- 确保未来通过 app_user 新建的表，report_user 也自动拥有只读权限
ALTER DEFAULT PRIVILEGES FOR ROLE app_user IN SCHEMA public GRANT SELECT ON TABLES TO report_user;

-- 【核心需求】：在当前 app_db 库中，导入指定的 SQL 文件
-- 假设我们在 docker-compose.yml 中将这些文件挂载到了 imports/ 目录下
\echo '-> 正在向 app_db 导入外部结构和数据脚本...'
\i /docker-entrypoint-initdb.d/imports/app_schema_init.sql
\i /docker-entrypoint-initdb.d/imports/app_base_data.sql


\echo '=================== [4/4] 初始化 report_db 数据库及权限 ==================='

-- 切换到报表数据库
\connect report_db;

-- 回收默认权限
REVOKE ALL ON SCHEMA public FROM PUBLIC;
GRANT ALL ON SCHEMA public TO report_user;

-- 导入指定的 SQL 文件到 report_db
\echo '-> 正在向 report_db 导入初始化报表结构...'
\i /docker-entrypoint-initdb.d/imports/report_schema_init.sql

\echo '=================== 初始化脚本执行完毕 ==================='