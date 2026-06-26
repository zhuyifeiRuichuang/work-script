# 说明
在k8s环境部署postgresql数据库。

任何改动调整都应复制为新文件处理。

# yaml文件说明
| 文件 | 说明 |
|----|----|
| v1 | pg v18及以上 |
| v2 | pg v17及以下 |
| v3 | 直接用户对接apache hive，基于v1 |



# CNPG

 **CloudNativePG**是主流的pg数据库在k8s环境的管理工具。

官网`https://cloudnative-pg.io/`

代码`https://github.com/cloudnative-pg/cloudnative-pg`

创建专用namespace

```bash
kubectl create namespace postgresql
```

部署operator

```bash
kubectl apply -f cnpg/cnpg-1.29.1.yaml
```

如果需要调整配置，需查询文档`https://cloudnative-pg.io/docs/1.29/scheduling`

部署pg集群

```bash
kubectl apply -f cnpg/cnpg-pg-cluster.yaml -n postgresql
```

部署后查询

```bash
root@master1:/data/postgresql# kubectl get all -n postgres-dev1 
NAME                    READY   STATUS    RESTARTS   AGE
pod/cluster-example-1   1/1     Running   0          5m11s
pod/cluster-example-2   1/1     Running   0          3m59s
pod/cluster-example-3   1/1     Running   0          3m37s

NAME                         TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/cluster-example-r    ClusterIP   10.233.23.99   <none>        5432/TCP   5m57s
service/cluster-example-ro   ClusterIP   10.233.2.107   <none>        5432/TCP   5m57s
service/cluster-example-rw   ClusterIP   10.233.38.79   <none>        5432/TCP   5m57s

```

