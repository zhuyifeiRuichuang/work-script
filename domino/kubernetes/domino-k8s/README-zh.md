# 更新

存在故障，暂停更新

# Domino K8s 部署方案

[Domino](https://github.com/Tauffer-Consulting/domino) 工作流编排平台的完整 Kubernetes 部署配置。

> **English**: [README.md](./README.md)

## 系统架构

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Kubernetes 集群                               │
│                                                                      │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐           │
│  │  NodePort    │    │  NodePort    │    │  NodePort    │           │
│  │  前端 :随机端口│    │  API :随机端口│    │  Airflow UI  │           │
│  └──────┬───────┘    └──────┬───────┘    └──────┬───────┘           │
│         │                   │                   │                    │
│  ┌──────▼───────┐    ┌──────▼───────┐    ┌──────▼───────┐           │
│  │    前端      │    │   REST API  │    │  Airflow     │           │
│  │   (Nginx)    │    │  (FastAPI)  │    │  Webserver   │           │
│  └──────────────┘    └──────┬───────┘    └──────────────┘           │
│                             │                                        │
│                             │ DAG 同步 (共享 PVC)                     │
│                    ┌────────▼────────┐                               │
│                    │   共享存储卷     │                               │
│                    │  /opt/airflow/  │                               │
│                    │     dags        │                               │
│                    └────────┬────────┘                               │
│                             │                                        │
│              ┌──────────────┼──────────────┐                        │
│              │              │              │                         │
│       ┌──────▼───────┐ ┌───▼────────┐ ┌───▼────────┐              │
│       │  Scheduler   │ │  Worker 1  │ │  Worker 2  │              │
│       │  (调度器)     │ │ (Celery)   │ │ (Celery)   │              │
│       └──────────────┘ └─────┬──────┘ └─────┬──────┘              │
│                              │              │                       │
│                              └──────┬───────┘                       │
│                                     │                               │
│                              ┌──────▼───────┐                       │
│                              │    Redis     │                       │
│                              │  (消息队列)   │                       │
│                              └──────────────┘                       │
│                                                                      │
│       ┌──────────────┐          ┌──────────────┐                    │
│       │  Domino DB   │          │ Airflow DB   │                    │
│       │  (Postgres)  │          │ (Postgres)   │                    │
│       └──────────────┘          └──────────────┘                    │
│                                                                      │
│       ┌──────────────┐ ┌────────────┐                               │
│       │  Triggerer   │ │  Init Job  │ (一次性初始化)                   │
│       └──────────────┘ └────────────┘                               │
└─────────────────────────────────────────────────────────────────────┘
```

## 组件清单

| 组件 | 镜像 | 作用 |
|------|------|------|
| domino-frontend | `ghcr.io/tauffer-consulting/domino-frontend:latest` | React 前端 (Nginx) |
| domino-rest | `ghcr.io/tauffer-consulting/domino-rest:latest` | FastAPI REST API |
| airflow-webserver | `ghcr.io/tauffer-consulting/domino-airflow:latest` | Airflow 界面和 API |
| airflow-scheduler | `ghcr.io/tauffer-consulting/domino-airflow:latest` | DAG 调度器 |
| airflow-worker | `ghcr.io/tauffer-consulting/domino-airflow:latest` | Celery 任务执行器 |
| airflow-triggerer | `ghcr.io/tauffer-consulting/domino-airflow:latest` | 异步操作触发器 |
| airflow-redis | `redis:7-alpine` | Celery 消息队列 |
| domino-postgres | `postgres:13` | Domino 元数据库 |
| airflow-postgres | `postgres:13` | Airflow 元数据库 |

## 前置要求

1. **Kubernetes 集群** (v1.25+)
2. **kubectl** 已配置并连接到集群
3. **存储类** 支持 `ReadWriteMany`（NFS、CephFS、EFS、Longhorn 等）
   - 单节点集群通常可用默认存储类配合 `ReadWriteOnce`
4. **网关**（可选）：[Higress](https://higress.io/)、APISIX、Traefik 或其他 L7 网关，用于代理 NodePort 服务

## 快速部署

### 第一步：配置密钥

编辑 `01-secrets.yaml`，替换所有 `CHANGE_ME` 值：

```bash
# 生成 Airflow 加密密钥
python3 -c "from cryptography.fernet import Fernet; print(Fernet.generate_key().decode())"

# 生成随机密钥
python3 -c "import secrets; print(secrets.token_hex(32))"
```

**必须修改的字段：**

| 字段 | 说明 |
|------|------|
| `DOMINO_DB_PASSWORD` | Domino 数据库密码 |
| `AIRFLOW_DB_PASSWORD` | Airflow 数据库密码 |
| `AIRFLOW_ADMIN_PASSWORD` | Airflow 管理员密码 |
| `DOMINO_GITHUB_ACCESS_TOKEN_WORKFLOWS` | GitHub Token（REST API 推送 DAG 用） |
| `DOMINO_DEFAULT_PIECES_REPOSITORY_TOKEN` | GitHub Token（读取 piece 仓库） |
| `AIRFLOW_FERNET_KEY` | Airflow Fernet 加密密钥 |
| `AIRFLOW_SECRET_KEY` | Airflow Webserver 密钥 |
| `AUTH_SECRET_KEY` | Domino 认证密钥 |

### 第二步：配置 GitHub 仓库

DAG 同步需要一个 GitHub 仓库，REST API 会将工作流 DAG 推送到此仓库：

1. 创建仓库（例如 `your-org/domino-dags`）
2. 生成 GitHub Personal Access Token（需要 `repo` 权限）
3. 在 `01-secrets.yaml` 中更新：
   - `DOMINO_GITHUB_ACCESS_TOKEN_WORKFLOWS` → 你的 token
4. 在 `08-domino-rest.yaml` 中更新：
   - `DOMINO_GITHUB_WORKFLOWS_REPOSITORY` → `your-org/domino-dags`

### 第三步：执行部署

```bash
chmod +x deploy.sh
./deploy.sh apply
```

### 第四步：访问服务

部署完成后，脚本会显示 NodePort 访问地址：

```
前端:          http://<节点IP>:<前端端口>/
REST API:      http://<节点IP>:<API端口>/api
Airflow 界面:  http://<节点IP>:<Airflow端口>/
```

随时查看端口信息：

```bash
./deploy.sh ports
```

本地测试可用端口转发：

```bash
kubectl port-forward -n domino svc/domino-frontend 3000:80 &
kubectl port-forward -n domino svc/domino-rest 8000:8000 &
kubectl port-forward -n domino svc/airflow-webserver 8080:8080 &
```

## 默认账号

| 服务 | 用户名 | 密码 |
|------|--------|------|
| Domino 界面 | admin@email.com | admin |
| Airflow 界面 | admin | 见 `01-secrets.yaml` |

⚠️ **生产环境请立即修改默认密码！**

## 网关集成（Higress / APISIX / Traefik）

NodePort 服务设计为由网关前置代理。以 [Higress](https://higress.io/) 为例：

**方式一：静态路由**

在 Higress 控制台配置 HTTP 路由：

| 路径 | 上游地址 |
|------|----------|
| `/api/*` | `http://<节点IP>:<REST端口>` |
| `/airflow/*` | `http://<节点IP>:<Airflow端口>` |
| `/*` | `http://<节点IP>:<前端端口>` |

**方式二：K8s Service 直接引用**

Higress 支持直接引用 K8s Service，无需关心 NodePort：

```yaml
# Higress HTTP 路由示例
apiVersion: networking.higress.io/v1
kind: McpBridge
metadata:
  name: domino-bridge
spec:
  registries:
    - name: domino-rest
      type: static
      domains:
        - "<节点IP>:<REST端口>"
    - name: domino-frontend
      type: static
      domains:
        - "<节点IP>:<前端端口>"
    - name: airflow-webserver
      type: static
      domains:
        - "<节点IP>:<Airflow端口>"
```

## 配置说明

### 使用外部数据库

如果有外部 PostgreSQL 实例：

1. 删除 `02-domino-postgres.yaml` 和 `03-airflow-postgres.yaml`
2. 在 `01-secrets.yaml` 中更新外部数据库凭据
3. 更新 ConfigMap 中的连接字符串

### 扩容 Worker

编辑 `06-airflow-deployment.yaml` 中的 `replicas` 字段：

```yaml
# Airflow Worker
spec:
  replicas: 4  # ← 增加并发能力
```

### 共享存储方案

部署需要 `ReadWriteMany` 的 PVC。如果集群不支持：

**方案 A：NFS Provisioner**

```bash
helm install nfs-provisioner stable/nfs-server-provisioner
```

**方案 B：Longhorn**

```bash
helm install longhorn longhorn/longhorn --namespace longhorn-system --create-namespace
```

**方案 C：单节点方案**

将 PVC 访问模式改为 `ReadWriteSingle`，并确保所有 Airflow Pod 调度到同一节点。

## 文件结构

```
domino-k8s/
├── deploy.sh                    # 一键部署脚本
├── README.md                    # 英文文档
├── README-zh.md                 # 中文文档（本文件）
├── 00-namespace.yaml            # 命名空间定义
├── 01-secrets.yaml              # 密钥配置（数据库密码、Token、密钥）
├── 02-domino-postgres.yaml      # Domino PostgreSQL（PVC + Deployment + Service）
├── 03-airflow-postgres.yaml     # Airflow PostgreSQL（PVC + Deployment + Service）
├── 04-airflow-redis.yaml        # Redis（Deployment + Service）
├── 05-airflow-storage.yaml      # 共享 PVC（DAGs、日志、共享存储）
├── 06-airflow-deployment.yaml   # Airflow 全套（ConfigMap、Init Job、Webserver、Scheduler、Worker、Triggerer）
├── 07-airflow-rbac.yaml         # Airflow k8s Operator RBAC 权限
├── 08-domino-rest.yaml          # Domino REST API（Deployment + Service）
├── 09-domino-frontend.yaml      # Domino 前端（Deployment + Service）
├── 10-nodeport-services.yaml    # NodePort 服务（外部访问）
└── helm-domino/                 # 更新后的 Helm Chart（含 Airflow 子 chart）
    ├── Chart.yaml
    └── values.yaml
```

## 常见问题

### Pod 无法启动

```bash
# 查看 Pod 状态
kubectl get pods -n domino

# 查看 Pod 事件
kubectl describe pod <pod-name> -n domino

# 查看日志
kubectl logs <pod-name> -n domino
```

### Airflow Init Job 失败

```bash
# 查看 Job 状态
kubectl get jobs -n domino

# 查看初始化日志
kubectl logs job/airflow-init -n domino
```

### DAG 未出现在 Airflow 中

1. 检查 GitHub Token 权限是否正确
2. 确认 `DOMINO_GITHUB_WORKFLOWS_REPOSITORY` 值
3. 查看 REST API 日志：`kubectl logs deployment/domino-rest -n domino`
4. 检查 DAG 是否在共享卷中：
   ```bash
   kubectl exec -it deployment/airflow-scheduler -n domino -- ls /opt/airflow/dags/
   ```

### 数据库连接问题

```bash
# 测试 Domino 数据库连接
kubectl exec -it deployment/domino-rest -n domino -- python -c "
import psycopg2
conn = psycopg2.connect(host='domino-postgres', dbname='domino', user='postgres', password='YOUR_PASSWORD')
print('连接成功!')
conn.close()
"
```

## 清理

```bash
./deploy.sh delete
```
