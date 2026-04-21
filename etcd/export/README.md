说明  
> 前置条件：应已经执行`etcd/etcd_init.sh` 完成环境初始化。  
> 提示：导出的数据是加密的，解密需参考：`https://github.com/etcd-io/auger`

导出etcd数据文件内指定资源为yaml文件，可导入到其他k8s或etcd环境使用。  

`export.sh`
- 读取`export.conf`中配置，指定数据导出存放的目录，指定其他配置。详见文件中说明。  
- 读取`resources.txt`中的资源类型，导出资源类型下所有资源为yaml文件，存放在指定位置。可使用 `--help`查看使用方法。

查询旧数据中可选的资源类型（精简）:  
```bash
ETCDCTL_API=3 etcdctl \
  --endpoints=http://127.0.0.1:2389 \
  get --prefix "" --keys-only | awk -F '/' '{if(NF >= 2) print $(NF-1)}' | sort | uniq -c
```

查询旧数据中可选的资源类型（完整路径）: 
```bash
ETCDCTL_API=3 etcdctl \
  --endpoints=http://127.0.0.1:2389 \
  get --prefix "" --keys-only | sort
```
