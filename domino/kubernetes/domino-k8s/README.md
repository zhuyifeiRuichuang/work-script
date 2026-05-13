# Domino K8s Deployment

Complete Kubernetes deployment for the [Domino](https://github.com/Tauffer-Consulting/domino) workflow orchestration platform.

> **中文文档**: [README-zh.md](./README-zh.md)

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Kubernetes Cluster                            │
│                                                                      │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐           │
│  │  NodePort    │    │  NodePort    │    │  NodePort    │           │
│  │  Frontend    │    │  REST API    │    │  Airflow UI  │           │
│  │  :<rand>     │    │  :<rand>     │    │  :<rand>     │           │
│  └──────┬───────┘    └──────┬───────┘    └──────┬───────┘           │
│         │                   │                   │                    │
│  ┌──────▼───────┐    ┌──────▼───────┐    ┌──────▼───────┐           │
│  │  Frontend    │    │  REST API   │    │  Airflow     │           │
│  │  (Nginx)     │    │  (FastAPI)  │    │  Webserver   │           │
│  └──────────────┘    └──────┬───────┘    └──────────────┘           │
│                             │                                        │
│                             │ DAG sync (shared PVC)                  │
│                    ┌────────▼────────┐                               │
│                    │  Shared Volume  │                               │
│                    │  /opt/airflow/  │                               │
│                    │     dags        │                               │
│                    └────────┬────────┘                               │
│                             │                                        │
│              ┌──────────────┼──────────────┐                        │
│              │              │              │                         │
│       ┌──────▼───────┐ ┌───▼────────┐ ┌───▼────────┐              │
│       │  Scheduler   │ │  Worker 1  │ │  Worker 2  │              │
│       │              │ │  (Celery)  │ │  (Celery)  │              │
│       └──────────────┘ └─────┬──────┘ └─────┬──────┘              │
│                              │              │                       │
│                              └──────┬───────┘                       │
│                                     │                               │
│                              ┌──────▼───────┐                       │
│                              │    Redis     │                       │
│                              │  (Celery)    │                       │
│                              └──────────────┘                       │
│                                                                      │
│       ┌──────────────┐          ┌──────────────┐                    │
│       │  Domino DB   │          │ Airflow DB   │                    │
│       │  (Postgres)  │          │ (Postgres)   │                    │
│       └──────────────┘          └──────────────┘                    │
│                                                                      │
│       ┌──────────────┐ ┌────────────┐                               │
│       │  Triggerer   │ │  Init Job  │ (one-shot)                    │
│       └──────────────┘ └────────────┘                               │
└─────────────────────────────────────────────────────────────────────┘
```

## Components

| Component | Image | Purpose |
|-----------|-------|---------|
| domino-frontend | `ghcr.io/tauffer-consulting/domino-frontend:latest` | React UI (Nginx) |
| domino-rest | `ghcr.io/tauffer-consulting/domino-rest:latest` | FastAPI REST API |
| airflow-webserver | `ghcr.io/tauffer-consulting/domino-airflow:latest` | Airflow UI & API |
| airflow-scheduler | `ghcr.io/tauffer-consulting/domino-airflow:latest` | DAG scheduler |
| airflow-worker | `ghcr.io/tauffer-consulting/domino-airflow:latest` | Celery task executor |
| airflow-triggerer | `ghcr.io/tauffer-consulting/domino-airflow:latest` | Deferrable operator triggerer |
| airflow-redis | `redis:7-alpine` | Celery message broker |
| domino-postgres | `postgres:13` | Domino metadata DB |
| airflow-postgres | `postgres:13` | Airflow metadata DB |

## Prerequisites

1. **Kubernetes cluster** (v1.25+)
2. **kubectl** configured and connected to the cluster
3. **Storage class** supporting `ReadWriteMany` (NFS, CephFS, EFS, Longhorn, etc.)
   - For single-node clusters, the default storage class usually works with `ReadWriteOnce`
4. **Gateway** (optional): [Higress](https://higress.io/), APISIX, Traefik, or any L7 gateway that can proxy to NodePort services

## Quick Start

### 1. Configure Secrets

Edit `01-secrets.yaml` and replace ALL `CHANGE_ME` values:

```bash
# Generate Fernet key (Airflow encryption)
python3 -c "from cryptography.fernet import Fernet; print(Fernet.generate_key().decode())"

# Generate random secret key
python3 -c "import secrets; print(secrets.token_hex(32))"
```

### 2. Configure GitHub Repository

For DAG synchronization, you need a GitHub repository where the REST API pushes workflow DAGs:

1. Create a repository (e.g., `your-org/domino-dags`)
2. Generate a GitHub Personal Access Token with `repo` scope
3. Update in `01-secrets.yaml`:
   - `DOMINO_GITHUB_ACCESS_TOKEN_WORKFLOWS` → your token
4. Update in `08-domino-rest.yaml`:
   - `DOMINO_GITHUB_WORKFLOWS_REPOSITORY` → `your-org/domino-dags`

### 3. Deploy

```bash
chmod +x deploy.sh
./deploy.sh apply
```

### 4. Access

After deployment, the script will display NodePort access URLs:

```
Frontend:     http://<NODE_IP>:<FRONTEND_PORT>/
REST API:     http://<NODE_IP>:<REST_PORT>/api
Airflow UI:   http://<NODE_IP>:<AIRFLOW_PORT>/
```

You can also check ports later:

```bash
./deploy.sh ports
```

For local testing, use port-forward:

```bash
kubectl port-forward -n domino svc/domino-frontend 3000:80 &
kubectl port-forward -n domino svc/domino-rest 8000:8000 &
kubectl port-forward -n domino svc/airflow-webserver 8080:8080 &
```

## Default Credentials

| Service | Username | Password |
|---------|----------|----------|
| Domino UI | admin@email.com | admin |
| Airflow UI | admin | (see `01-secrets.yaml`) |

⚠️ **Change these immediately in production!**

## Gateway Integration (Higress / APISIX / Traefik)

NodePort services are designed to be fronted by a gateway. Example with [Higress](https://higress.io/):

```yaml
# Higress HTTP route example
apiVersion: networking.higress.io/v1
kind: McpBridge
metadata:
  name: domino-bridge
spec:
  registries:
    - name: domino-rest
      type: static
      domains:
        - "<NODE_IP>:<REST_PORT>"
    - name: domino-frontend
      type: static
      domains:
        - "<NODE_IP>:<FE_PORT>"
    - name: airflow-webserver
      type: static
      domains:
        - "<NODE_IP>:<AF_PORT>"
```

Or simply configure your gateway to proxy:

| Path | Upstream |
|------|----------|
| `/api/*` | `http://<NODE_IP>:<REST_PORT>` |
| `/airflow/*` | `http://<NODE_IP>:<AF_PORT>` |
| `/*` | `http://<NODE_IP>:<FE_PORT>` |

## Configuration

### Using External Database

If you have an external PostgreSQL instance:

1. Remove `02-domino-postgres.yaml` and `03-airflow-postgres.yaml`
2. Update `01-secrets.yaml` with external DB credentials
3. Update the connection strings in the ConfigMap

### Scaling Workers

Edit the `replicas` field in `06-airflow-deployment.yaml`:

```yaml
# Airflow Worker
spec:
  replicas: 4  # ← Increase for more parallelism
```

### Shared Storage (ReadWriteMany)

The deployment requires `ReadWriteMany` PVCs for DAGs and logs. If your cluster doesn't support it:

**Option A: Use NFS provisioner**

```bash
helm install nfs-provisioner stable/nfs-server-provisioner
```

**Option B: Use Longhorn**

```bash
helm install longhorn longhorn/longhorn --namespace longhorn-system --create-namespace
```

**Option C: Single-node workaround**

Change PVC access mode to `ReadWriteSingle` and ensure all Airflow pods run on the same node.

## File Structure

```
domino-k8s/
├── deploy.sh                    # One-click deployment script
├── README.md                    # English documentation
├── README-zh.md                 # Chinese documentation (中文文档)
├── 00-namespace.yaml            # Namespace definition
├── 01-secrets.yaml              # Secrets (DB creds, tokens, keys)
├── 02-domino-postgres.yaml      # Domino PostgreSQL (PVC + Deployment + Service)
├── 03-airflow-postgres.yaml     # Airflow PostgreSQL (PVC + Deployment + Service)
├── 04-airflow-redis.yaml        # Redis (Deployment + Service)
├── 05-airflow-storage.yaml      # Shared PVCs (DAGs, logs, shared storage)
├── 06-airflow-deployment.yaml   # Airflow (ConfigMap, Init Job, Webserver, Scheduler, Worker, Triggerer)
├── 07-airflow-rbac.yaml         # RBAC for Airflow k8s operator
├── 08-domino-rest.yaml          # Domino REST API (Deployment + Service)
├── 09-domino-frontend.yaml      # Domino Frontend (Deployment + Service)
├── 10-nodeport-services.yaml    # NodePort services for external access
└── helm-domino/                 # Updated Helm Chart (with Airflow subchart)
    ├── Chart.yaml
    └── values.yaml
```

## Troubleshooting

### Pods not starting

```bash
# Check pod status
kubectl get pods -n domino

# Check pod events
kubectl describe pod <pod-name> -n domino

# Check logs
kubectl logs <pod-name> -n domino
```

### Airflow init job failing

```bash
# Check job status
kubectl get jobs -n domino

# Check init logs
kubectl logs job/airflow-init -n domino
```

### DAGs not appearing in Airflow

1. Check that the GitHub token has correct permissions
2. Verify the `DOMINO_GITHUB_WORKFLOWS_REPOSITORY` value
3. Check REST API logs: `kubectl logs deployment/domino-rest -n domino`
4. Check if DAGs are in the shared volume:
   ```bash
   kubectl exec -it deployment/airflow-scheduler -n domino -- ls /opt/airflow/dags/
   ```

### Database connection issues

```bash
# Test Domino DB connectivity
kubectl exec -it deployment/domino-rest -n domino -- python -c "
import psycopg2
conn = psycopg2.connect(host='domino-postgres', dbname='domino', user='postgres', password='YOUR_PASSWORD')
print('Connected!')
conn.close()
"
```

## Cleanup

```bash
./deploy.sh delete
```
