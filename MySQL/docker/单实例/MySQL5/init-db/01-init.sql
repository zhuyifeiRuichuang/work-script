-- 创建数据参考
-- 创建业务数据库
-- CREATE DATABASE IF NOT EXISTS `my_database`
--   DEFAULT CHARACTER SET utf8mb4
--   DEFAULT COLLATE utf8mb4_unicode_ci;

-- 创建用户参考
-- 创建应用专属用户
-- CREATE USER 'app_user'@'%' IDENTIFIED BY 'YourAppPassword123';
-- GRANT ALL PRIVILEGES ON my_database.* TO 'app_user'@'%';
-- 创建只读用户
-- CREATE USER 'read_only'@'%' IDENTIFIED BY 'ReadOnlyPassword456';
-- GRANT SELECT ON my_database.* TO 'read_only'@'172.%';

-- 导入数据。应将SQL文件按照导入顺序，一次在文件名开头添加01，02，03等数字标识，并在SQL开头配置 USE my_database; 来指定数据库。

-- 使上述配置生效
-- FLUSH PRIVILEGES;