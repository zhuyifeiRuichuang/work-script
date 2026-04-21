#!/usr/bin/env bash
# Harbor 仓库地址（写域名，默认配置为 https://dockerhub.kubekey.local）
url="https://172.16.0.11"

# 访问 Harbor 仓库的默认用户和密码（生产环境建议修改）
user="admin"
passwd="Harbor12345"

# 需要创建的项目名列表，按我们制作的离线包的镜像命名规则，实际上只需要创建一个 kubesphereio 即可，这里保留了所有可用值，各位可根据自己的离线仓库镜像名称规则修改。
harbor_projects=(
    kubesphere
	kubesphereio
	csiplugin
	minio
	osixia
	prom
	opensearchproject
	jimmidyson
	mirrorgooglecontainers
	jaegertracing
	istio
	argoproj
	dexidp
	thanosio
	weaveworks
	line
	openpolicyagent
)

for project in "${harbor_projects[@]}"; do
    echo "creating $project"
    curl -k -u "${user}:${passwd}" -X POST -H "Content-Type: application/json" "${url}/api/v2.0/projects" -d "{ \"project_name\": \"${project}\", \"public\": true}"
done
