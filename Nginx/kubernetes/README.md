# 说明

在k8s集群部署nginx，使集群内和集群外可访问。

# 容器镜像

容器镜像应基于docker hub公版，将前端项目制作的dist目录内的资源存入容器镜像内指定目录，进行打包定制。

# 修改配置

对三个yaml文件进行自定义修改

`configmap.yaml`是nginx配置文件的内容，仅修改`server {}`中的内容。

`deployment.yaml`是应用的配置文件，可修改`image`，`replicas`，两个`name`

`service.yaml`是服务配置文件，使集群内外都可访问nginx，若不想集群外访问，可删除集群外整段配置。

# 部署

创建namespace，或使用已有的namespace

```bash
kubectl create namespace nginx
```

部署资源

```bash
kubectl apply -f configmap.yaml -n nginx
kubectl apply -f deployment.yaml -n nginx
kubectl apply -f service.yaml -n nginx
```

# 查询结果

```bash
kubectl get all -n nginx
```

