#!/bin/bash
# 查询当前k8s集群的kubeconfig，并导出为json格式至文件。
# 用于airflow填写kubeconfig（Json format）
python3 -c "import yaml, json; print(json.dumps(yaml.safe_load(open('/root/.kube/config')), indent=2))" > /opt/kubeconfig.json
