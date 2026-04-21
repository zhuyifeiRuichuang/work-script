# 说明
nacos在k8s环境部署

在k8s环境应优先部署集群模式。



参考文档`https://nacos.io/docs/latest/quickstart/quick-start-kubernetes/?spm=5238cd80.5687ac34.0.0.74f0215fmrId3r`, nacos server v3.1

# 创建MySQL数据库
单独创建一个MySQL数据库，准备好MySQL的主机域名或者ip，端口号，账户，密码，提前创建一个数据库名字是`nacos`，编码`utf8mb4`，倒入`mysql-schema.sql`文件

# 修改nacos的yaml文件
`nacos-k8s-v1.yaml`的`nacos-cm`部分，`mysql`配置改为上述配置的。  
`namespace`可修改。  
如果k8s集群资源太少，就修改副本数为1.`StatefulSet`的`replicas: 3` 改为1。常规3副本足够，若想配置更多副本，需修改`NACOS_SERVERS`部分，追加配置更多的同格式服务名字。  
配置文件里的`nacos.auth.token`使用命令生成，可以直接用默认值，
```bash
openssl rand -base64 48
```

# 部署nacos
```bash
kubectl apply -f nacos-k8s-v1.yaml -n nacos
```

# 查询
若仅部署1个副本，查询结果如下。
```bash
root@master1:/data/zhuyifei/nacos# kubectl get all -n nacos 
NAME          READY   STATUS    RESTARTS        AGE
pod/nacos-0   1/1     Running   5 (3m18s ago)   6m1s

NAME                     TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                                                                      AGE
service/nacos-headless   ClusterIP   None            <none>        8848/TCP,9848/TCP,9849/TCP,7848/TCP,8080/TCP                                 6m1s
service/nacos-nodeport   NodePort    10.233.53.242   <none>        8848:32006/TCP,9848:30110/TCP,9849:30975/TCP,7848:32056/TCP,8080:30498/TCP   6m1s
service/nacos-operator   ClusterIP   10.233.17.137   <none>        8080/TCP                                                                     36m

NAME                     READY   AGE
statefulset.apps/nacos   1/1     6m1s
```
