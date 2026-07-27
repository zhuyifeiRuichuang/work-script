# 说明

适用于Docker环境部署pg18及以上版本。

目录`config`存放pg数据库配置文件。

目录`initdb`存放初始化环境的sql文件，生产环境务必替换示例文件为业务文件。

目录`secrets`存放pg管理员的账户密码。

`compose.yaml`，默认内置web界面数据库管理工具，生产环境可注释掉。