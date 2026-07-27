-- ====================================================================
-- 文件: app_schema_init.sql
-- 说明: app_db 业务系统核心表结构初始化
-- ====================================================================

-- 创建用户表
CREATE TABLE IF NOT EXISTS public.sys_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), -- 使用 UUID 作为主键
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100),
    status VARCHAR(20) DEFAULT 'ACTIVE',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 为常用查询字段创建索引
CREATE INDEX idx_sys_users_status ON public.sys_users(status);

-- 创建角色字典表
CREATE TABLE IF NOT EXISTS public.sys_roles (
    role_code VARCHAR(50) PRIMARY KEY,
    role_name VARCHAR(100) NOT NULL,
    description TEXT
);