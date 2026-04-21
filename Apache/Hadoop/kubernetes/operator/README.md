# 说明

使用主流AI工具开发Hadoop cluster在生产环境可用的operator。

开发对比中，最终选择一个可用的标准方案。

参考`https://github.com/zncdatadev/hdfs-operator`



需求

Hadoop operator

- 代码本地打包
- 代码通过容器打包

- 单独镜像打包。
- 命令清晰不使用定制脚本。
- 可配置标准集群高可用。

Hadoop集群

单独镜像打包。

