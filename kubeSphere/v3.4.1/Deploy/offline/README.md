存放离线部署部署阶段的配置文件。提醒：文件随时可能更新导致名字产生少量差异。需自行分辨。  
# 说明：将部署所需文件放入本目录，上传至集群主节点使用。
# 需注意，全量部署，单个节点最低配置是8C16G500GB，至少3个节点，奇数的master节点，任意的work节点。

每次部署务必更新此处的文件。  
kubekey-v3.1.11-linux-amd64.tar.gz | 部署集群所需工具  
deploy.yaml | 部署集群软件所需配置文件 |   
ks341offline_v5.tar.gz | 用于部署的离线制品 | 可用 | 最新  
init_harbor_v3.sh | 初始化离线集群自有的harbor仓库 | 可用 | 最新  
init_autoCli.sh | 配置kubectl命令自动补齐功能 | 可用  
deploy-offline-v1.tar | 离线部署资源打包 | 可用 | 部署过程中故障较多 | 部署成功后故障最少  
deploy-offline-v2.tar | 离线部署资源打包 | 可用 | 部署过程中故障最少 | 部署成功后偶发POD故障    
完整配置文件参考`https://github.com/kubesphere/kubekey/blob/master/docs/manifest-example.md`

# 常用命令
按顺序执行可完成软件部署。

## 解压kubekey工具
```bash
tar -zxf kubekey-v3.1.11-linux-amd64.tar.gz
```
## 部署软件仓库
```bash
./kk init registry -f deploy_v5.yaml -a ks341offline_v5.tar.gz
```
## 初始化软件仓库，需修改文件中的url为自己机器的真实IP，一般采用master node1 IP
```bash
bash init_harbor_v3.sh
```
## 推送容器镜像到本地仓库
```bash
./kk artifact image push -f deploy_v5.yaml -a ks341offline_v5.tar.gz
```
## 跳过harbor初始化步骤。自动安装kubesphere v3.4.1和k8s。
```bash
./kk create cluster -f deploy_v5.yaml -a ks341offline_v5.tar.gz --with-packages --skip-push-images
```
## 配置kubectl自动补齐功能。在kubectl命令可使用时做配置。
```bash
bash init_autoCli.sh
```
## POD故障问题
部分POD拉取docker.io/busybox。需手动修改POD的控制器的image部分为本地仓库的busybox地址。
以下两个是必改。
```bash
kubectl edit -n kubesphere-logging-system statefulsets.apps opensearch-cluster-master 
kubectl edit -n kubesphere-logging-system statefulsets.apps opensearch-cluster-data 
```
