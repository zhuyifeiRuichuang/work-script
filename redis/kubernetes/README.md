# 更新

需测试最新版本。



# 说明

在k8s环境部署redis，用于生产环境。

# 资源说明



| 资源名             | 说明         | 状态   |
| ------------------ | ------------ | ------ |
| redis-single.yaml  | 单节点部署   | 可用   |
| redis-cluster.yaml | 集群模式部署 | 未测试 |
|                    |              |        |



# 配置

配置修改说明



# 部署

```bash
# 创建namespace或使用已有的namespace
kubectl create namespace redis
# 部署期望使用的配置
kubectl apply -f redis-single.yaml -n redis
```



# 访问测试



