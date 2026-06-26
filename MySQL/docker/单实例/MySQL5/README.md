# 说明
部署MySQL5.7.44单实例最佳实践。在生产环境不建议使用。

# 参考文档
`https://dev.mysql.com/doc/refman/5.7/en/`

# 配置说明

目录`init-db`存放数据库初始化需要执行的脚本。例如创建专用的数据库，创建专用的账户及访问权限，导入指定的SQL文件到指定的已创建的数据库。当有多个SQL文件时，文件名使用类似`01-aaa.sql`，`02-aaa.sql`的命名方式，控制执行SQL文件的顺序。

文件`.root_password`配置root用户密码。

`compose.yaml`和`my.cnf`可根据业务需求调整。

特殊情况说明：

当导入大量文件时，需将健康检查配置的更晚，更久一些，避免导入数据进行中导致的健康检查失败。

# 部署
```bash
docker compose up -d
```

# 访问测试

```bash
docker exec -it mysql5-single mysql -uroot -p"$(cat /opt/mysql/.root_password)"
```

