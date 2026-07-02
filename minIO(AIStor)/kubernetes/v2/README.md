# 说明

部署MinIO开源版本，多副本，多POD。

# 修改配置

若需定制配置，请修改yaml文件，文件已定义namespace等资源。可投喂给AI处理。

关于容器镜像

```bash
# minio/minio:RELEASE.2025-09-07T16-13-09Z 是最后一个开源版本。
# minio/minio:RELEASE.2025-04-22T22-12-26Z 是最后一个带完整图形界面版本
# minio代码：https://github.com/minio/minio https://gitee.com/zhudev2026/minio
# console代码：https://github.com/georgmangold/console  https://github.com/Alevsk/console
```

console组件需要自己定制打包并构建容器镜像。也可以用默认的。

# 部署

```bash
kubectl apply -f minio.yaml
```

# 验证

```bash
kubectl exec -it minio-0 -n minio -- bash
mc alias set local http://localhost:9000 admin admin@123
mc admin info local
# 将查询结果发给AI，分析查询结果是否有异常。
# 进行数据测试
mc mb local/test-bucket
echo "test data" > test.txt
mc cp test.txt local/test-bucket/
# 验证每个POD的双副本生效。三个POD是6副本。
ls /data1/test-bucket/
ls /data2/test-bucket/
```

