# 说明
有安全加固或其他需求的，可自定义基础镜像用于Hadoop镜像构建。  

# 目录结构
```bash
.
├── Dockerfile.ubuntu25 # 基于Ubuntu25.10，python3，JDK11.
├── README.md
└── scripts # 已改造适配python3。
    ├── envtoconf.py
    ├── krb5.conf
    ├── starter.sh
    └── transformation.py
```

# 构建基础镜像
参考[Hadoop官方说明](https://apache.github.io/hadoop/hadoop-project-dist/hadoop-common/HadoopDocker.html)  
在本目录进行构建。因[Hadoop官方项目](https://github.com/apache/hadoop/tree/docker-hadoop-runner-latest)长期未更新Dockerfile，已做部分改动。请自行对比官方原代码判断是否要调整。  
请在联网环境构建镜像，推荐美国网络，规避软件下载异常问题。  

## JDK版本选择说明
参考[Hadoop官方说明](https://cwiki.apache.org/confluence/display/HADOOP/Hadoop+Java+Versions)，JDK8支持编译和运行hadoop当前所有版本，Hadoop v3.3及更高版本可使用JDK11运行。  
基础镜像默认使用JDK8，若使用JDK11，需编辑Dockerfile修改JDK版本。

## 执行构建命令
命令格式
```bash
docker build -t hadoop:runner-v1 -f Dockerfile.ubuntu25 .
````