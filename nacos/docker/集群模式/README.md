# 说明

集群模式提供生产环境可靠性保障，提供数据安全性保障。

选择k8s环境的限制

- 多主机
- 外接存储
- 有专职运维维护k8s平台和应用



选择docker环境的限制

- 资源少
- 可单主机部署伪集群
- 可多主机部署集群
- 无人维护k8s平台



参考文档`https://nacos.io/docs/latest/manual/admin/deployment/deployment-cluster/?spm=5238cd80.72a042d5.0.0.37cecd36jJ4LBJ`

综合说明`https://nacos.io/docs/latest/manual/admin/deployment/deployment-overview/?spm=5238cd80.72a042d5.0.0.37cecd36jJ4LBJ#1-Nacos%E9%83%A8%E7%BD%B2%E6%9E%B6%E6%9E%84`

参考配置文件`https://github.com/nacos-group/nacos-docker/tree/master/example?spm=5238cd80.382dab05.0.0.1fbd2909hnHO2g`
# docker环境
集群部署和单体部署的差异是，多了一个`cluster.conf`需要配置。可直接使用单体模式的配置文件，在三个不同的云主机部署，并修改`cluster.conf`文件，填写三个主机的IP地址或者域名，修改部署脚本，挂载`cluster.conf`到容器的`/home/nacos/conf/`目录。  
若三个云主机没有固定IP，没有固定域名，可以改云主机的的`hosts`文件，写一个固定域名和三个云主机的临时的IP，将其挂载到容器的`/etc/hosts`，后续只需要修改`hosts`里的IP为云主机的新IP即可，甚至可以写自动化脚本自动更新，每次更新后，重启nacos的容器。  
`cluster.conf` 配置例子，
```bash
# ip:port
200.8.9.16:8848
200.8.9.17:8848
200.8.9.18:8848
```

# k8s环境
参考

`https://nacos.io/docs/latest/manual/admin/deployment/deployment-cluster`

`https://github.com/nacos-group/nacos-k8s`

`https://github.com/nacos-group/nacos-k8s/blob/master/operator/README-CN.md`

`https://github.com/nacos-group/nacos-k8s/blob/master/README-CN.md`
