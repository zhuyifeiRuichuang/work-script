# 说明
在k8s环境部署hive。

# 版本说明

| 版本 | 说明 |
|----|----|
| v1 | 原初版，可用 |
| v2 | 基于v1，测试可用，依赖现有的Hadoop环境。 |

# 部署
应创建专用的namespace，并修改所有配置文件中所有关于`namespace`的配置，包括与Hadoop的对接配置。 
按顺序执行命令，
```bash
kubectl create namespace hive
kubectl apply -f mysql.yaml -n hive
kubectl apply -f hive-configmap.yaml -n hive
kubectl apply -f hadoop-configmap.yaml -n hive
kubectl apply -f metastore.yaml -n hive
kubectl apply -f server2.yaml -n hive
```
