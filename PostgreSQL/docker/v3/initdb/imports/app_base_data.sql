-- ====================================================================
-- 文件: app_base_data.sql
-- 说明: app_db 业务系统基础/字典数据写入
-- ====================================================================

-- 插入默认角色数据 (使用 ON CONFLICT 避免重复执行时报错)
INSERT INTO public.sys_roles (role_code, role_name, description) VALUES
('ROLE_ADMIN', '系统管理员', '拥有系统全部权限'),
('ROLE_USER', '普通用户', '基础业务操作权限')
ON CONFLICT (role_code) DO NOTHING;

-- 插入初始系统管理员账号
INSERT INTO public.sys_users (username, email, status) VALUES
('superadmin', 'admin@mycompany.com', 'ACTIVE'),
('system_bot', 'bot@mycompany.com', 'ACTIVE')
ON CONFLICT (username) DO NOTHING;