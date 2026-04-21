# 说明

minIO是开源对象存储项目，代码已停止更新。



# 参考资料

`https://www.minio.org.cn/docs/minio/container/operations/installation.html`

# 容器镜像

容器镜像`minio/minio:RELEASE.2025-04-22T22-12-26Z `最后一个自带控制台图形界面版本。备份镜像`zhuyifeiruichuang/minio:RELEASE.2025-04-22T22-12-26Z`

容器镜像`minio/minio:RELEASE.2025-09-07T16-13-09Z` 最后一次minIO官方更新，残缺的控制台图形界面，需`mc`工具管理。备份镜像`zhuyifeiruichuang/minio:RELEASE.2025-09-07T16-13-09Z`

# 代码

原项目地址`https://github.com/minio/minio`

代码备份`https://github.com/ruichuang-com/minio.git`

# 文件系统类型约束

minIO部署及数据所在目录的底层文件系统类型推荐使用xfs。可通过`df -hT`查看

# 部署

```bash
docker compose up -d
```

