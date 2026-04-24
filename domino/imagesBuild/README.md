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
docker build -t domino-frontend:dev -f Dockerfile.prod .
```

# 构建后端镜像

```bash
cd rest
docker build -t domino-rest:dev .
```

# 其他镜像

直接在代码根目录下构建，根据需要使用的Dockerfile选择。



# 快速测试

使用docker部署所需的`compose.yaml`快速部署测试。
