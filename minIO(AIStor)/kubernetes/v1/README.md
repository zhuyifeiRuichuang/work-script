# 说明

部署MinIO闭源版最新版本，现改名为AIStor。单副本，单节点。

# 获取license

访问`https://www.min.io/pricing`，获取免费的license 文件`minio.license`。

# 修改配置

`minio.yaml`是豆包提供的配置

`aistor.yaml`是Gemini提供的配置

可发给AI辅助你调整配置。

# 部署

若使用`minio.yaml`

```bash
kubectl create ns minio
kubectl create secret generic minio-license --from-file=minio.license=./minio.license -n minio
kubectl apply -f minio.yaml -n minio
```

若使用`aistor.yaml`

```bash
kubectl create namespace aistor
kubectl create secret generic aistor-license --from-file=minio.license=./minio.license -n aistor
kubectl apply -f aistor.yaml -n aistor
```

# 访问

账户密码`minioadmin`