# 更新计划



更新为可灵活配置的配置文件+部署文件方法

区分单实例和集群部署。

配置数据初始化脚本

- 使用容器快速配置，不安装其他软件

- 可自动创建账户
- 自动创建数据库并导入指定的sql文件

# 说明

注意：pg17及以下版本的配置无法在18及以上版本使用。



权限配置文件`https://www.postgresql.org/docs/14/auth-pg-hba-conf.html`

可配置选项`https://www.postgresql.org/docs/18/runtime-config.html`



参考`https://hub.docker.com/_/postgres`