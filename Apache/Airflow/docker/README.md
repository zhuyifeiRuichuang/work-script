# 说明

docker环境部署Apache Airflow

参考`https://airflow.apache.org/docs/apache-airflow/stable/howto/docker-compose/index.html`



# gitsync

这是airflow的功能，需单独部署一个容器使用

参考配置如下

如果容器启动异常，就`cmod -R 777 /opt/domino/airflow/dags/git-data`

```bash
root@domino:/opt/domino# cat gitsync-compose.yaml

services:
  git-sync:
    image: zhuyifeiruichuang/git-sync:v4.3.0
    container_name: airflow-gitsync
    restart: always
    user: "50000:0"
    volumes:
      - /opt/domino/airflow/git-data:/sync
    environment:
      GITSYNC_REPO: https://github.com/代码仓库.git
      GITSYNC_BRANCH: main
      GITSYNC_PERIOD: 30s
      GITSYNC_DEPTH: 1
      GITSYNC_ROOT: /sync
      GITSYNC_DEST: repo
      GITSYNC_LINK: current
      GITSYNC_USERNAME: GitHub账户
      GITSYNC_PASSWORD: 认证令牌
      GITSYNC_KNOWN_HOSTS: "false"
      GITSYNC_ONE_TIME: "false"
root@domino:/opt/domino# cat gitsync-ssh-compose.yaml
services:
  git-sync:
    image: zhuyifeiruichuang/git-sync:v4.3.0
    container_name: airflow-git-sync
    user: "0:0"
    restart: always
    environment:
      - GITSYNC_REPO=git@github.com:代码仓库.git
      - GITSYNC_BRANCH=main
      - GITSYNC_PERIOD=60s
      - GITSYNC_SSH=true
      - GITSYNC_ROOT=/git
      - GITSYNC_DEST=repo
      - GITSYNC_LINK=dags
      - GITSYNC_SSH_KNOWN_HOSTS=false
      - GITSYNC_ADD_USER=true
    volumes:
      - /opt/domino/airflow/git-data:/git
      # privkey 文件保存ssh私钥
      - ./privkey:/etc/git-secret/ssh:ro
```

