# 说明

Alist是一个管理存储的工具。

官网`https://alistgo.com/zh/`

代码`https://github.com/AlistGo/alist`

容器镜像`https://github.com/AlistGo/alist/pkgs/container/alist`，必须选择明确的版本号的tag。

# 在docker环境部署

```bash
docker compose up -d
```



# 在k8s环境部署

```bash
kubectl create namespace alist
kubectl apply -f alist.yaml -n alist
```

获取登录密码

```bash
# 确认pod名字
kubectl get all -n alist
# 例如pod名字是pods/alist-deployment-6559978bc8-p9scp
kubectl exec -it -n alist pods/alist-deployment-6559978bc8-p9scp -- ./alist admin random
```



# 访问

访问端口5244即可。

默认账户admin



# 对接对象存储

参考`https://alistgo.com/zh/guide/drivers/s3.html`
