# 注意

尚未测试验证

# 说明

脚本可用于初始化MySQL容器，注意，仅适用于容器。实现以下功能。

自动创建账户密码及权限。

自动创建数据库。

自动导入指定的SQL文件到指定的数据库。



# 使用方法

将所有用到的SQL文件存放到目录`sql`

配置`databases.conf`，`mysql_connection`填写数据库连接信息。



若需创建数据库

修改文件`mysql_connection`。

配置数据库基础信息，例如`database:jnpf_init`表示创建数据库`jnpf_init`

执行命令完成自动配置。

```bash
./create_databases.sh
```



若需创建用户

修改文件`users.conf`

`users` 用于配置通用账户

`database_grants:jnpf_init` 用于配置指定数据库的专属用户

```bash
./create_users.sh
```



若需导入SQL文件

修改文件`import.conf`

`import_order`必须是已经存在的数据库。

```bash
./import_sql.sh
```



若需全都配置

依次修改上述conf文件后，执行以下命令。

```bash
./run_all.sh
```

# 访问验证

访问数据库，查看是否有新增数据和配置。
