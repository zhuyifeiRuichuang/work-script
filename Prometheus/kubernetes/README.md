# 说明

k8s环境部署Prometheus。建议在指定的namespace部署。可离线部署。可灵活调整各配置。Gemini提供参考。

# 文件说明

各yaml文件有详细说明。删除注释后使用。

# 部署

```bash
kubectl create namespace prometheus
kubectl apply -f . -n prometheus
```

# 查询检查

```bash
kubectl get all -n prometheus
```

