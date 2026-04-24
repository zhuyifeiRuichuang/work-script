# 说明

容器镜像构建。需要定制配置或定制开发时，需重新构建容器镜像。

项目代码：`https://github.com/Tauffer-Consulting/domino`



# 拉取代码

```bash
git clone https://github.com/Tauffer-Consulting/domino.git
```



# 进入项目目录

```bash
cd domino
```



# 构建前端镜像

```bash
cd frontend
docker build -t zhuyifeiruichuang/domino-frontend:dev -f Dockerfile.prod .
```

# 构建后端镜像

```bash
cd rest
docker build -t zhuyifeiruichuang/domino-rest:dev .
```

# 构建domino-airflow-base镜像

用于组件`airflow-scheduler`

```bash
在项目根目录下
docker build -t zhuyifeiruichuang/domino-airflow-base:dev -f Dockerfile-airflow-domino.prod .
```

# 快速测试

使用docker环境部署文件，快速测试镜像可用性。
