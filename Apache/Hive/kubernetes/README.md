# 说明
在k8s环境部署hive。

下一步计划增加资源限制，避免过度占用集群资源。默认配置最小资源占用。



# 版本说明

| 版本 | 说明 |
|----|----|
| v1 | 原初版，可用 |
| v2 | 基于v1，可用，依赖现有的Hadoop环境。 |

# 修改配置

待完善。

单独讲解各yaml的可配置部分。



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
