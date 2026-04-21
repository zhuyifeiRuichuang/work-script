# 说明
在kubernetes环境部署和测试MySQL任意版本(必须是docker hub 有镜像的)。  
yaml配置已将容器内`/var/lib/mysql`配置数据持久化到k8s集群的pvc。  
单实例可以使用deployment类型部署，集群和主从需使用statefulset类型部署。  
集群架构推荐使用MySQL的operator部署，例如`https://dev.mysql.com/doc/mysql-operator/en/`
| 特性               | 主从架构 (Master-Slave)                              | InnoDB Cluster (MGR / HA)                          |
|--------------------|------------------------------------------------------|---------------------------------------------------|
| 原理               | 异步复制。主库写，异步传给从库。                     | 组复制 (MGR)。基于 Paxos 协议，强一致性/最终一致性。 |
| 读写分离           | 适合。主库写，从库读。                               | 支持。通常通过 MySQL Router 自动路由。             |
| 数据一致性         | 可能丢数据。如果主库突然宕机，从库可能还没收到最新数据。 | 不丢数据。多节点确认写入才算成功。                 |
| 故障切换           | 手动/复杂。主库挂了，需要人工把从库提升为主，或者用 Orchestrator。 | 自动。主节点挂了，集群自动选主，秒级恢复。         |
| k8s 部署难度       | 中等 (可以用 StatefulSet + 脚本实现)。               | 极高 (必须用 Operator，手写 YAML 极难维护)。       |
| 适用场景           | 读多写少，对数据丢失有一点容忍度（如日志、报表）。   | 核心交易系统，金融级要求，不能丢数据。             |

# 参考

`https://dev.mysql.com/doc/mysql-operator/en/mysql-operator-innodbcluster-simple-helm.html`

## yaml版本说明

### deployment  
```
v1,单实例。mysql:5.7.44  
```
### statefulset  
```
v2,单实例。mysql:8.0  
v3,主从架构，1主2从，读写分离，mysql:9.5.0
```
# 快速部署
在k8s集群内快速部署MySQL
```bash
kubectl create namespace mysql
kubectl apply -f mysql.yaml -n mysql
```
# 访问测试
测试在k8s环境，集群内访问MySQL。
```bash
bash checkInternal.sh
```
## 数据持久化测试
仅供参考
```bash
# 查询pod名字
kubectl get pod -n mysql

# 访问pod
kubectl exec -it -n mysql mysql-55cf49d874-z6sr2 -- bash

# 登录数据库，默认密码root123456
mysql -u root -p 
# 配置测试数据
mysql> CREATE DATABASE IF NOT EXISTS test_persistence;
mysql> USE test_persistence;
mysql> CREATE TABLE IF NOT EXISTS user (id INT PRIMARY KEY, name VARCHAR(20));
mysql> INSERT INTO user (id, name) VALUES (1, "persistence_test");
mysql> SELECT * FROM user;
mysql> exit
bash-4.2# exit

# 删除pod，模拟MySQL被破坏
kubectl delete pod -n mysql mysql-55cf49d874-z6sr2 

# 查询pod，确认pod被deployment自动恢复
kubectl get pod -n mysql

# 访问pod
kubectl exec -it -n mysql mysql-55cf49d874-td2c7 -- bash
# 登录数据库，默认密码root123456
bash-4.2# mysql -u root -p
# 查询数据，若可见数据，则数据持久化测试成功
mysql> USE test_persistence;
mysql> SELECT * FROM user;
mysql> exit
Bye
bash-4.2# exit
exit
```
