# 更新计划

改造为compose.yaml可一键部署的。

改造k8s部署方法，一键部署关联组件。

# 前端修复

`start_frontend_dev*.sh`是duomino的前端启动脚本，用于解决云平台环境部署domino出现network erro的问题。  
执行脚本之前，应修改容器配置部分。配置说明如下，  

```bash
CONTAINER_NAME是容器的名字，需当前环境全局唯一。  
API_URL是domino-rest容器的IP和对应的物理机端口。  
IMAGE是定制容器镜像，可替换为自己定制的前端镜像。  
HOST_PORT是前端容器的物理机端口，部署后可通过IP:端口在浏览器访问。
```
`.env`是环境变量，应存放在执行domino部署命令时的目录路径。  
重要参数说明，

```bash
DOMINO_DEFAULT_PIECES_REPOSITORY_TOKEN是自己的GitHub的token。在https://github.com/settings/personal-access-tokens配置
```



# 说明

Domino项目是一个开源任务调度平台，结合Apache Airflow使用，使Airflow的web界面更好看。

项目代码：`https://github.com/Tauffer-Consulting/domino`

