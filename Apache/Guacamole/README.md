# 说明

官网`https://guacamole.apache.org/`

容器化部署时，整体需要三个重要组件。

`guacamole/guacd`，包含guacd守护进程，由guacamole-server构建。

`guacamole/guacamole`，为运行在 Tomcat 9.x 中的 Guacamole Web 应用程序提供 WebSocket 支持。

`mysql\postgresql`，提供 Guacamole 将用于身份验证和存储连接配置数据的数据库。

# 部署

仅测试容器化部署，区分docker环境和kubernetes环境部署。分别在对应目录。