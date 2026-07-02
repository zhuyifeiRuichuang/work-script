# 说明

有商业版和社区版，都需要申请license文件。闭源。推荐直接用最新版本。

# 参考资料

`https://docs.min.io/enterprise/aistor-object-store/installation/`

# 准备基础环境

创建专用目录

```bash
mkdir -p /data/minio/data /data/minio/certs
```



# 申请license

访问`https://www.min.io/pricing`，获取免费的license 文件`minio.license`，上传至`/data/minio`

# 配置TLS(可选)

参考`https://docs.min.io/enterprise/aistor-object-store/installation/container/network-encryption/`



# 部署

部署前检查脚本是否满足业务需求。

```bash
docker compose up -d
```

# 访问

浏览器访问`IP:9001` ，账户密码均是`minioadmin`
