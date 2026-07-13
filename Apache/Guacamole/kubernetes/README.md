# 说明

在k8s环境部署完整guacamole

两种配置文件存在差异，

v1，对接MySQL数据库方法

v2，对接PostgreSQL数据库方法

# 配置

注意，sql文件内容来自Docker环境部署时获取的sql文件。

# 部署

```bash
kubectl create namespace guaca
kubectl apply -f v1.yaml -n guaca
kubectl get all -n guaca
```

# 访问

浏览器访问master node1 的IP和8080端口映射的物理机端口

默认账户密码是`guacadmin`