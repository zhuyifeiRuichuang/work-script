# 说明

`alist-v1.yaml`是定制改造版。支持自定义配置文件并持久生效。

`alist.yaml`是原装普通版。不支持自定义配置文件，只能在web界面做配置。



部署

```bash
kubectl create namespace alist
kubectl apply -f alist.yaml -n alist
kubectl get all -n alist
# 查看admin的密码
kubectl logs -n alist pod/alist-pod真实名字
# 可以看到关键词password is:
```

