# 说明

官网`https://security.appspot.com/vsftpd.html`

配置文件说明

`https://security.appspot.com/vsftpd/vsftpd_conf.html`

在官网下载最新版本软件，仅选取一个发行版的软件包放在当前目录。

修改`Dockerfile`，配置自己期望的账户和密码和配置文件。也可以单独配置一个配置文件挂载到容器内。

如果有自定义配置，修改`compose.yaml`，并且参考`vsftpd.conf.example`修改配置文件。将配置文件挂载到容器中的`/etc/vsftpd.conf`

默认内置账户密码`ftp/Ftp@123`，默认端口20，21
