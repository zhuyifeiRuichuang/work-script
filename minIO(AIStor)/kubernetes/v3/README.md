# 说明

minio(AIStor)社区闭源版。单节点，每节点双驱动器。

# 申请license

访问`https://www.min.io/pricing`，获取免费的license 文件`minio.license`。

# 配置namesapce

```bash
kubectl create namespace aistor
```

# 配置license

```bash
kubectl -n aistor create secret generic aistor-minio-license --from-file=minio.license=./minio.license
```

# 部署

```bash
kubectl apply -f aistor.yaml
```

