#!/bin/bash
# 处理kubectl命令自动补齐.部署完成后执行该脚本，重新开启终端窗口。
apt-get install bash-completion -y
echo 'source <(kubectl completion bash)' >>~/.bashrc
kubectl completion bash >/etc/bash_completion.d/kubectl
