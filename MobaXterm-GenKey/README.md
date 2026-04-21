# 上游项目

代码fork自`https://github.com/malaohu/MobaXterm-GenKey`



# 说明

工具名称：MobaXterm-GenKey

用途：生成mobaxterm激活文件，支持配置授权用户名，授权发行版本号，授权用户数量。



# 软件下载

`https://mobaxterm.mobatek.net/`

# 使用激活工具

可本地启动，也可容器化启动。

## 本地启动激活工具

需python3环境
```
pip install --no-cache-dir -r requirements.txt
python app.py
```

启动后可根据帮助提示生成激活文件。

## 构建容器镜像

```bash
docker build -t mobaxterm-genkey:dev1 .
```

## Docker运行软件

```
# 原版镜像malaohu/mobaxterm-genkey
# 备用镜像zhuyifeiruichuang/mobaxterm-genkey:dev2
docker run -d \
--name mobaxterm-genkey \
-p 5000:5000 \
mobaxterm-genkey:dev1
```



## 使用方法

浏览器访问`IP:5000` ，填写所有信息，获取认证文件。



### 激活方式
将认证文件放到mobaxterm软件根目录，运行软件。

