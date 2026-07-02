# 说明
部署MySQL8单实例最佳实践。  

# 参考资料
参考文档`https://dev.mysql.com/doc/refman/8.4/en/`

# 配置说明
文件`.root_password`配置root用户的密码。

可根据业务需求自定义配置`my.cnf`和`compose.yaml`

目录`init-db`存放数据库初始化需要执行的脚本。例如创建专用的数据库，创建专用的账户及访问权限，导入指定的SQL文件到指定的已创建的数据库。当有多个SQL文件时，文件名使用类似`01-aaa.sql`，`02-aaa.sql`的命名方式，控制执行SQL文件的顺序。

# 部署
```bash
docker compose up -d
```



# 访问测试

```bash
docker exec -it mysql8-single mysql -uroot -p"$(cat /opt/mysql/.root_password)"
```

