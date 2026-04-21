#!/bin/bash
# 用于拉取minio对象存储指定文件,拉取目录需使用`mc cp`

# 下载指定文件到指定目录
mc get myminio/kubesphere/configFile/deploy_v6.yaml /opt/
mc get myminio/kubesphere/kubesphere/kubekey/kubekey-v3.1.11-linux-amd64.tar.gz /opt/
mc get myminio/kubesphere/script/init_autoCli.sh /opt/
mc get myminio/kubesphere/script/init_harbor_v3.sh /opt/
mc get myminio/kubesphere/artifact/ks341offline_v5.tar.gz /opt/
