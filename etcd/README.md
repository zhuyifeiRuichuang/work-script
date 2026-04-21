关于etcd的操作脚本。

> 业务场景1：  
k8s集群遇到毁灭性故障，`kubectl`和`etcdctl`软件均被破坏，磁盘数据中发现etcd的旧数据，希望从etcd旧数据导出部分内容到新k8s集群或etcd数据库继续使用。etcd数据存档默认路径`/var/lib/etcd/`，以环境真实配置为准。  
> ！！！警告！！！先备份，后操作。

`etcd_init.sh`
- 仅限联网环境使用。
- 自动检测etcd安装状态。
- 自动配置etcdctl环境变量。
- 自动安装指定版本etcd。
- 自动读取指定目录的etcd数据，旧数据指snap和wal后缀文件。
- 需编辑`--data-dir`,指定旧数据存放的目录。`ETCD_VERSION`指定使用的etcd版本。
- 应在全新操作环境处理，避免影响在用的etcd环境。

`export`目录。导出etcd存档数据中指定资源。
