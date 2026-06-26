# 说明

官方`https://doris.apache.org/`

参考`https://doris.apache.org/docs/4.x/gettingStarted/quick-start`

# 配置

检查当前环境网络未使用配置文件中的IP地址段。

doris不支持域名解析，组件的环境变量里必须写指定IP。

确认端口与当前系统资源无冲突。

CPU和内存默认最低配置，低于此配置会出现容器崩溃。

# 部署

```bash
docker compose up -d
```



# 访问测试

8030端口的账户是admin，无密码

