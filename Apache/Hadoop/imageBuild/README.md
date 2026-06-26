# 说明
使用docker基于[Hadoop基础镜像](https://hub.docker.com/r/apache/hadoop-runner)构建Hadoop标准容器镜像。可用于docker和k8s环境部署。

可对任意版本Hadoop进行镜像构建。Hadoop[版本清单](https://archive.apache.org/dist/hadoop/common/)。

可在中国网络环境操作，容器镜像网络加速可参考`1ms.run`  

参考文档`https://hadoop.apache.org/docs/current/hadoop-project-dist/hadoop-common/HadoopDocker.html`

参考Dockerfile`https://github.com/apache/hadoop/tree/docker-hadoop-runner-latest`



# 思路
Hadoop官方指导的镜像构建流程是打包代码，构建基础镜像，构建Hadoop容器镜像。

# 资源说明

对当前目录的文件目录说明，

```bash
.
├── base # hadoop原版。用于构建基础镜像。默认python2,JDK8（可替换为JDK11）
│   ├── Dockerfile.centos7 # 基于centos7
│   ├── Dockerfile.ubuntu20 # 基于ubuntu20
│   ├── Dockerfile.ubuntu22 # 基于Ubuntu22
│   ├── README.md # 构建基础镜像的说明。
│   └── scripts # 需内置到基础镜像的脚本，来自Hadoop官方代码仓库，仅适配python2环境。
│       ├── envtoconf.py
│       ├── krb5.conf
│       ├── starter.sh
│       └── transformation.py
├── base.new # 改造版。用于构建基础镜像。默认python3，JDK11（可替换为JDK8）
│   ├── Dockerfile.ubuntu25 # 基于Ubuntu25
│   ├── README.md # 构建基础镜像的说明。
│   └── scripts # 需内置到基础镜像的脚本，基于Hadoop原版改造适配python3。
│       ├── envtoconf.py
│       ├── krb5.conf
│       ├── starter.sh
│       └── transformation.py
├── compose.yaml # 用于对Hadoop容器镜像快速测试。  
├── config # Hadoop快速测试使用的默认配置文件。
├── Dockerfile # 构建Hadoop容器镜像
├── hadoop.tar.gz # 需打包到容器镜像使用的Hadoop软件包。
├── log4j.properties # 可使`kubectl logs`可以直接查看到组件日志。可选。推荐。
└── README.md
```

# 二次开发
对Hadoop定制开发结束后，请打包项目为名称格式`hadoop-版本号.tar.gz`的文件存放在当前目录。

# 构建Hadoop基础镜像
可选。

有安全加固或定制需求，详见目录`base`和`base.new`。

若不构建基础镜像，在构建Hadoop容器镜像时，默认使用Hadoop官方的`apache/hadoop-runner:latest`，基于centos7，内置JDK8，python2。  

## JDK版本选择
请谨慎选择。若兼容性不匹配，构建的Hadoop镜像将无法启动。

参考[Hadoop官方说明](https://cwiki.apache.org/confluence/display/HADOOP/Hadoop+Java+Versions)，JDK8支持编译和运行Hadoop当前所有版本，Hadoop v3.3及更高版本可使用JDK11运行。  

# 构建Hadoop容器镜像

使用指定的Hadoop软件包构建容器镜像。

## 下载Hadoop软件
将需使用的hadoop.tar.gz文件存放到此目录。当前目录下只能存放一个Hadoop的软件包。例如下载[hadoop-3.1.1.tar.gz](https://archive.apache.org/dist/hadoop/common/hadoop-3.1.1/hadoop-3.1.1.tar.gz)  

## 构建镜像
修改`Dockerfile`中的`FROM`为期望使用的基础镜像。

构建Hadoop镜像。

```bash
docker build -t hadoop:dev1 .
```

# 快速测试镜像
可快速测试镜像可用性。确认容器均UP，且日志无异常，则测试成功。

修改配置文件，`image:`改为构建的镜像名字。

```bash
vim compose.yaml
```

启动Hadoop
```bash
docker compose up -d
docker compose ps -a
```

# 可用容器镜像

`https://cnb.cool/zhudev-2025/apache-hadoop-dev`