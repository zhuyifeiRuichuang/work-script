# 说明
在k8s环境中部署Hadoop集群。

参考`https://github.com/stackabletech/hdfs-operator`

operator处于开发阶段，不可在生产环境使用。



## 容器镜像
推荐采用Hadoop官方镜像。
案例采用定制的Hadoop v3.1.1 
可参考[资料](https://github.com/zhuyifeiRuichuang/work-script/tree/main/hadoop/docker.build)打包定制容器镜像。

## 配置文件
| 版本 | 说明 |
|----|----|
| v1 | namenode,datanode支持数据持久化，支持集群内外访问。 |
| v2 | 基于v1，新增对hive的支持，对tez的支持。新增datanode延迟启动，规避datanode先于namenode启动产生的集群关系连接失败。|
| v3 | 基于v2, 解决了namenode初始化强制格式化的问题。解决了datanode因namenode反向DNS解析失败无法建立连接的问题。 |

# 遗留问题
## 问题1
重建namenode的POD时，Hadoop集群的`namespaceID`会变化，导致datanode无法连接namenode。
### 临时解决办法
方法一：将卷挂载给其他容器，修改卷里的VERSION文件的ID与namenode的一致。  
方法二：删掉datanode对应的数据卷。重建datanode  
方法三：使用v3配置文件，给namenode配置一个`init pod`，检测到存储卷有数据，就禁止格式化，直接使用。目前此方法效果最好。

## 问题2
v2版本在k8s环境部署时，出现namenode解析不到datanode，导致拒绝datanode注册。可使用v3配置文件，但该方法是Hadoop官方不推荐的配置。

# 后期计划
Hadoop可能需要一个类似apache Doris的 operator实现集群的部署管理。

# 修改配置文件
推荐使用namespace名称hadoop。除了超大规模环境，很少有一个k8s集群部署多套Hadoop的业务场景，除非基础资源超级多。
若需指定其他`namespace`部署Hadoop，需修改`configmap.yaml`中所有关于`namespace`的配置，例如`namenode-0.namenode.hadoop:9000`,`resourcemanager-0.resourcemanager.hadoop`，其中`hadoop`是namespace。需修改`datanode.yaml`中`namenode-0.namenode.hadoop`,其中`hadoop`是namespace。



多个datanode场景。若需要多个datanode实现数据多副本，应做以下配置

修改`configmap.yaml`的`hdfs-site.xml`部分，

```bash
<property>
        <name>dfs.replication</name>
        # 将副本数调整为3
        <value>3</value>
</property>
```

修改`datanode.yaml`的`StatefulSet`部分

```bash
spec:
  serviceName: datanode
  # 改为3副本
  replicas: 3
  selector:
    matchLabels:
      app: hadoop-datanode
  template:
    metadata:
      labels:
        app: hadoop-datanode
    spec:
    # 添加反亲和性，让三个datanode落在不同的node
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
              - key: app
                operator: In
                values:
                - hadoop-datanode
            topologyKey: "kubernetes.io/hostname"
```



# 部署
务必按照顺序部署。指定部署在namespace hadoop中。
```bash
kubectl create namespace hadoop
kubectl apply -f configmap.yaml -n hadoop
kubectl apply -f namenode.yaml -n hadoop
kubectl apply -f datanode.yaml -n hadoop
kubectl apply -f resourcemanager.yaml -n hadoop
kubectl apply -f nodemanager.yaml -n hadoop
```
# 部署后查询
```bash
root@master Mon Nov 24 [10:05:40] : /opt/bigdata2/hadoop/v3.1.1/v4
# kubectl get all -n hadoop 

NAME                    READY   STATUS    RESTARTS      AGE
pod/datanode-0          1/1     Running   0             9m20s
pod/namenode-0          1/1     Running   3 (14m ago)   15m
pod/nodemanager-0       1/1     Running   0             8s
pod/resourcemanager-0   1/1     Running   0             76s

NAME                      TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)                      AGE
service/datanode          ClusterIP   None         <none>        9864/TCP,9866/TCP            20m
service/namenode          ClusterIP   None         <none>        9870/TCP,9000/TCP,8020/TCP   27m
service/nodemanager       ClusterIP   None         <none>        8042/TCP                     8s
service/resourcemanager   ClusterIP   None         <none>        8088/TCP,8032/TCP            7m9s

NAME                               READY   AGE
statefulset.apps/datanode          1/1     20m
statefulset.apps/namenode          1/1     27m
statefulset.apps/nodemanager       1/1     8s
statefulset.apps/resourcemanager   1/1     76s
root@master Mon Nov 24 [10:05:42] : /opt/bigdata2/hadoop/v3.1.1/v4
# kubectl get pvc -n hadoop
NAME                        STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
hadoop-dn-data-datanode-0   Bound    pvc-1ce19f52-729d-4d2e-8b4a-6c064b5901fe   20Gi       RWO            local          26m
hadoop-nn-data-namenode-0   Bound    pvc-8d162dd9-c732-4337-acd2-01945a9ff929   10Gi       RWO            local          33m
```
# 测试
部署后请充分测试后再计划使用。
## 浏览器访问
`IP:9870`

## 数据持久化测试
测试验证数据持久化，数据持久依赖pv，pv在，数据就在。

### namenode
```bash
# 配置测试数据
root@master Mon Nov 24 [10:49:38] : /opt/bigdata2/hadoop/v3.1.1/v4
# kubectl exec -it namenode-0 -n hadoop -- /bin/bash
bash-4.2$ hdfs dfs -mkdir -p /test-namenode-persist
bash-4.2$ hdfs dfs -ls /
Found 1 items
drwxr-xr-x   - hadoop supergroup          0 2025-11-24 02:51 /test-namenode-persist
bash-4.2$ exit
exit

# 模拟namenode破坏性故障
root@master Mon Nov 24 [10:55:48] : /opt/bigdata2/hadoop/v3.1.1/v4
# kubectl delete pod namenode-0 -n hadoop
pod "namenode-0" deleted

# 等待namenode pod自动恢复后，查询历史数据
root@master Mon Nov 24 [10:57:28] : /opt/bigdata2/hadoop/v3.1.1/v4
# kubectl exec -it namenode-0 -n hadoop -- /bin/bash
bash-4.2$ hdfs dfs -ls /
Found 1 items
drwxr-xr-x   - hadoop supergroup          0 2025-11-24 02:51 /test-namenode-persist
bash-4.2$ exit
exit
root@master Mon Nov 24 [10:58:12] : /opt/bigdata2/hadoop/v3.1.1/v4
# kubectl exec -it namenode-0 -n hadoop -- ls -l /opt/hadoop/data/nn/current/
total 3108
-rw-r--r-- 1 hadoop hadoop     216 Nov 24 02:56 VERSION
-rw-r--r-- 1 hadoop hadoop      42 Nov 24 01:50 edits_0000000000000000001-0000000000000000002
-rw-r--r-- 1 hadoop hadoop      42 Nov 24 01:50 edits_0000000000000000003-0000000000000000004
-rw-r--r-- 1 hadoop hadoop      42 Nov 24 01:51 edits_0000000000000000005-0000000000000000006
-rw-r--r-- 1 hadoop hadoop 1048576 Nov 24 01:51 edits_0000000000000000007-0000000000000000007
-rw-r--r-- 1 hadoop hadoop 1048576 Nov 24 02:51 edits_0000000000000000008-0000000000000000009
-rw-r--r-- 1 hadoop users  1048576 Nov 24 02:56 edits_inprogress_0000000000000000010
-rw-r--r-- 1 hadoop hadoop     391 Nov 24 01:50 fsimage_0000000000000000000
-rw-r--r-- 1 hadoop hadoop      62 Nov 24 01:50 fsimage_0000000000000000000.md5
-rw-r--r-- 1 hadoop users      479 Nov 24 02:56 fsimage_0000000000000000009
-rw-r--r-- 1 hadoop users       62 Nov 24 02:56 fsimage_0000000000000000009.md5
-rw-r--r-- 1 hadoop users        3 Nov 24 02:56 seen_txid
```
### datanode
```bash
# 配置测试数据
root@master Mon Nov 24 [11:09:09] : /opt/bigdata2/hadoop/v3.1.1/v4
# kubectl exec -it datanode-0 -n hadoop -- /bin/bash
.txt-4.2$ echo "这是 Datanode 持久化测试数据" > /tmp/test-datanode-file 
bash-4.2$ hdfs dfs -put /tmp/test-datanode-file.txt /test-namenode-persist/
bash-4.2$ hdfs dfs -ls /test-namenode-persist/
Found 1 items
-rw-r--r--   1 hadoop supergroup         38 2025-11-24 03:10 /test-namenode-persist/test-datanode-file.txt
bash-4.2$ hdfs dfs -cat /test-namenode-persist/test-datanode-file.txt
这是 Datanode 持久化测试数据
bash-4.2$ exit
exit

# 模拟datanode破坏性故障
root@master Mon Nov 24 [11:10:35] : /opt/bigdata2/hadoop/v3.1.1/v4
# kubectl delete pod datanode-0 -n hadoop
pod "datanode-0" deleted

# 等待datanode pod自动恢复，查询历史数据
root@master Mon Nov 24 [11:11:41] : /opt/bigdata2/hadoop/v3.1.1/v4
# kubectl exec -it namenode-0 -n hadoop -- /bin/bash
bash-4.2$ hdfs dfs -cat /test-namenode-persist/test-datanode-file.txt
这是 Datanode 持久化测试数据
bash-4.2$ hdfs fsck /test-namenode-persist/test-datanode-file.txt -files -blocks
Connecting to namenode via http://namenode-0.namenode.hadoop:9870/fsck?ugi=hadoop&files=1&blocks=1&path=%2Ftest-namenode-persist%2Ftest-datanode-file.txt
FSCK started by hadoop (auth:SIMPLE) from /10.233.70.10 for path /test-namenode-persist/test-datanode-file.txt at Mon Nov 24 03:12:21 UTC 2025
/test-namenode-persist/test-datanode-file.txt 38 bytes, replicated: replication=1, 1 block(s):  OK
0. BP-461085605-10.233.70.200-1763949032615:blk_1073741825_1001 len=38 Live_repl=1


Status: HEALTHY
 Number of data-nodes:	1
 Number of racks:		1
 Total dirs:			0
 Total symlinks:		0

Replicated Blocks:
 Total size:	38 B
 Total files:	1
 Total blocks (validated):	1 (avg. block size 38 B)
 Minimally replicated blocks:	1 (100.0 %)
 Over-replicated blocks:	0 (0.0 %)
 Under-replicated blocks:	0 (0.0 %)
 Mis-replicated blocks:		0 (0.0 %)
 Default replication factor:	1
 Average block replication:	1.0
 Missing blocks:		0
 Corrupt blocks:		0
 Missing replicas:		0 (0.0 %)

Erasure Coded Block Groups:
 Total size:	0 B
 Total files:	0
 Total block groups (validated):	0
 Minimally erasure-coded block groups:	0
 Over-erasure-coded block groups:	0
 Under-erasure-coded block groups:	0
 Unsatisfactory placement block groups:	0
 Average block group size:	0.0
 Missing block groups:		0
 Corrupt block groups:		0
 Missing internal blocks:	0
FSCK ended at Mon Nov 24 03:12:21 UTC 2025 in 5 milliseconds


The filesystem under path '/test-namenode-persist/test-datanode-file.txt' is HEALTHY
bash-4.2$ exit
exit

root@master Mon Nov 24 [11:12:34] : /opt/bigdata2/hadoop/v3.1.1/v4
# kubectl exec -it datanode-0 -n hadoop -- ls -l /opt/hadoop/data/dn/current/
total 8
drwx------ 4 hadoop hadoop 4096 Nov 24 03:11 BP-461085605-10.233.70.200-1763949032615
-rw-r--r-- 1 hadoop hadoop  229 Nov 24 03:11 VERSION
```
# 配置更新
若需更新配置文件，需修改confimap.yaml中内容，
```bash
# 应用更新
kubectl apply -f hadoop-configmap.yaml

# 重启namenode
kubectl rollout restart statefulset namenode -n hadoop

# 重启 DataNode
kubectl rollout restart statefulset datanode -n hadoop

# 重启 ResourceManager
kubectl rollout restart statefulset resourcemanager -n hadoop

# 重启 NodeManager
kubectl rollout restart statefulset nodemanager -n hadoop
```

# 数据持久化

当云主机磁盘被破坏，或k8s底层存储数据文件损坏时，可通过备份数据恢复，确认备份数据真实存在。删除k8s残留数据（pod,pvc,pv），将备份数据的pv文件复制到新pv目录下，重启pod读取旧数据。

提醒：若先复制备份数据到新存储目录，新建pvc时，旧数据被被覆盖部分文件，导致新建pod无法正常运行，且读取不到pv内存储的数据。
