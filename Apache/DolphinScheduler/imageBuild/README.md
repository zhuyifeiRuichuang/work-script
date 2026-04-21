# 说明
构建专用容器镜像。

# 参考资料
代码。`https://github.com/apache/dolphinscheduler`  

参考：`https://hub.docker.com/r/apache/dolphinscheduler`

# 构建镜像

浏览器访问`https://dolphinscheduler.apache.org` ，下载指定版本源码包。



下载插件

安装插件一定要在打包容器镜像阶段进行。

方法1：

根据项目业务需求，浏览器访问`https://repo.maven.apache.org/maven2/org/apache/dolphinscheduler/` ，下载插件。若资源充足，建议全下载。

方法2：推荐。

修改容器中文件`conf/plugins_config`，添加希望下载的插件。

执行脚本`bash ./bin/install-plugins.sh 3.4.1`，自动下载插件。`3.4.1`是`dolphinscheduler`的版本。

插件会下载到容器中的`/opt/dolphinscheduler/plugins`



容器镜像更新。需在API组件容器安装全部插件，使平台支持各类功能。



容器组件角色

master

work

api 。需安装所有插件，提高功能兼容性。

alert

Schema Initializer 。用于初始化数据库。



分析容器镜像构建的方法

基于原版镜像做扩展。



# 测试镜像

