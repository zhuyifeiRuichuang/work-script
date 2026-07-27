# 说明
在k8s环境部署postgresql数据库。

想实现高可用和多副本，需要引入operator。

# yaml文件说明
| 文件 | 说明 |
|----|----|
| v1 | pg v18及以上，支持nodePort访问 |
| v2 | pg v17及以下，支持nodePort访问 |
| v3 | 直接用户对接apache hive，基于v1 |
| cnpg | 使用cnpg工具部署pg数据库集群，无nodePort访问 |

# 配置



# 部署

```bash
# 创建专用namespace
kubectl create namespace pg
# 创建密码
kubectl create secret generic pg-secret --from-literal=password='Ruichuang@123' -n pg
# 部署资源
kubectl apply -f v1.yaml -n pg
```



# 其他

除了CNPF实现pg集群，还有其他工具。

`https://github.com/crunchydata/postgres-operator`

`https://github.com/CrunchyData/postgres-operator-examples`

`https://github.com/zalando/postgres-operator`
