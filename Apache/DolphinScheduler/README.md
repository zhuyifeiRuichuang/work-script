# 更新

当前版本存在功能缺陷。有工作流时不展示工作流关系。暂时不测试。



# 说明

部署和配置DolphinScheduler集群。

## 部署架构

| 架构   | 说明                       | 用途               |
| ------ | -------------------------- | ------------------ |
| 单实例 | 单个云主机中部署单个实例   | 临时快速体检       |
| 伪集群 | 单个云主机中部署集群       | 开发，测试和预生产 |
| 集群   | 多个云主机中部署多个集群。 | 生产环境           |



# 参考资料
`https://dolphinscheduler.apache.org/zh-cn/docs/3.4.0/guide/installation/pseudo-cluster`
`https://github.com/apache/dolphinscheduler/blob/37d2dc3ec1a8ec31498005d364f9ac1b4668c04c/dolphinscheduler-dist/src/main/docker/worker-server.dockerfile`
`https://github.com/apache/dolphinscheduler/blob/37d2dc3ec1a8ec31498005d364f9ac1b4668c04c/dolphinscheduler-dist/src/main/docker/api-server.dockerfile`

# 构建容器镜像
根据业务需求构建专用容器镜像。参考目录`imageBuild`

# 部署
支持多种部署环境。

| 环境   | 说明                                               | 用途                     |
| ------ | -------------------------------------------------- | ------------------------ |
| 主机   | 云主机或主机                                       | 开发，测试，预生产，生产 |
| docker | 任意安装了最新版本docker的环境。或其他容器管理器。 | 测试，预生产，生产       |
| k8s    | k8s集群                                            | 预生产，生产             |

