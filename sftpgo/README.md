# 注意

未测试数据持久化。

未测试安全权限隔离。

不可用于生产。

# 说明

sftpgo是开源的ftp和sftp server项目。有商业版本。

代码仓库`https://github.com/drakkan/sftpgo`

开源版项目文档`https://docs.sftpgo.com/latest/docker/`

企业版项目文档`https://docs.sftpgo.com/enterprise/`

开源版容器镜像仓库`https://hub.docker.com/r/drakkan/sftpgo`

配置文件`sftpgo.json`完整版`https://github.com/drakkan/sftpgo/blob/main/sftpgo.json`

# 部署

```bash
docker compose up -d
```

sftp默认使用端口2022 ，ftp默认使用端口2023



# 访问

浏览器访问`http://主机IP:8080/web/admin`， 首次访问需创建管理员账户密码。

需创建`user`，并配置密码，用于远程访问实现sftp和ftp登录。



# 配置

可通过修改`sftpgo.json`调整配置。

