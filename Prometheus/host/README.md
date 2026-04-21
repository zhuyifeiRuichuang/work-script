# 说明

主机环境部署Prometheus，支持离线部署。

主机环境包含物理机和云主机，推荐使用云主机。



# 下载软件

浏览器访问`https://prometheus.io/download/` ，下载最新版软件，例如：`https://github.com/prometheus/prometheus/releases/download/v3.10.0/prometheus-3.10.0.linux-amd64.tar.gz`



# 部署软件

解压文件

```bash
tar -zxf prometheus-3.10.0.linux-amd64.tar.gz
```

在软件目录中启动软件，

```bash
cd prometheus-3.10.0.linux-amd64
./prometheus --config.file=prometheus.yml &
```



# 配置管理

`prometheus.yml`是全局配置文件。

# 访问测试

浏览器访问`IP:9090` 