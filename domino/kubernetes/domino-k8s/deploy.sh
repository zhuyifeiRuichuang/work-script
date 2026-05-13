#!/bin/bash
# ============================================================
# Domino K8s One-Click Deployment Script
# ============================================================
# Usage:
#   ./deploy.sh [apply|delete|status|ports|logs]
# ============================================================

set -euo pipefail

NAMESPACE="domino"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

log()   { echo -e "${BLUE}[INFO]${NC}  $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC}  $*"; }
err()   { echo -e "${RED}[ERROR]${NC} $*"; }
ok()    { echo -e "${GREEN}[OK]${NC}    $*"; }

# ------ Preflight checks ------
preflight() {
    log "Running preflight checks..."

    if ! command -v kubectl &> /dev/null; then
        err "kubectl not found. Please install kubectl first."
        exit 1
    fi

    if ! kubectl cluster-info &> /dev/null; then
        err "Cannot connect to Kubernetes cluster. Check your kubeconfig."
        exit 1
    fi

    ok "Preflight checks passed."
}

# ------ Check secrets ------
check_secrets() {
    local secrets_file="${SCRIPT_DIR}/01-secrets.yaml"
    if grep -q "CHANGE_ME" "$secrets_file" 2>/dev/null; then
        warn "============================================"
        warn "  WARNING: Secrets contain CHANGE_ME values"
        warn "============================================"
        echo ""
        warn "Please edit 01-secrets.yaml before deploying."
        warn "Required changes:"
        warn "  - DOMINO_DB_PASSWORD / AIRFLOW_DB_PASSWORD"
        warn "  - DOMINO_GITHUB_ACCESS_TOKEN_WORKFLOWS"
        warn "  - AIRFLOW_FERNET_KEY (generate with Python)"
        warn "  - AIRFLOW_SECRET_KEY"
        warn "  - AUTH_SECRET_KEY"
        echo ""
        read -p "Continue anyway? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log "Aborted. Please edit 01-secrets.yaml first."
            exit 0
        fi
    fi
}

# ------ Apply all resources ------
apply_all() {
    preflight
    check_secrets

    log "Deploying Domino to Kubernetes..."
    echo ""

    # Apply in order
    for f in \
        00-namespace.yaml \
        01-secrets.yaml \
        02-domino-postgres.yaml \
        03-airflow-postgres.yaml \
        04-airflow-redis.yaml \
        05-airflow-storage.yaml \
        07-airflow-rbac.yaml \
        06-airflow-deployment.yaml \
        08-domino-rest.yaml \
        09-domino-frontend.yaml \
        10-nodeport-services.yaml \
    ; do
        if [ -f "${SCRIPT_DIR}/${f}" ]; then
            log "Applying ${f}..."
            kubectl apply -f "${SCRIPT_DIR}/${f}"
        else
            warn "File ${f} not found, skipping."
        fi
    done

    echo ""
    ok "All resources applied!"
    echo ""

    # Wait for pods
    log "Waiting for pods to become ready..."
    echo ""

    log "  [1/6] Domino PostgreSQL..."
    kubectl wait --for=condition=ready pod -l app=domino-postgres -n ${NAMESPACE} --timeout=120s 2>/dev/null || warn "  Domino postgres not ready yet"

    log "  [2/6] Airflow PostgreSQL..."
    kubectl wait --for=condition=ready pod -l app=airflow-postgres -n ${NAMESPACE} --timeout=120s 2>/dev/null || warn "  Airflow postgres not ready yet"

    log "  [3/6] Redis..."
    kubectl wait --for=condition=ready pod -l app=airflow-redis -n ${NAMESPACE} --timeout=60s 2>/dev/null || warn "  Redis not ready yet"

    log "  [4/6] Airflow init job..."
    kubectl wait --for=condition=complete job/airflow-init -n ${NAMESPACE} --timeout=180s 2>/dev/null || warn "  Airflow init job not complete yet"

    log "  [5/6] Domino REST API..."
    kubectl wait --for=condition=ready pod -l app=domino-rest -n ${NAMESPACE} --timeout=180s 2>/dev/null || warn "  REST API not ready yet"

    log "  [6/6] Domino Frontend..."
    kubectl wait --for=condition=ready pod -l app=domino-frontend -n ${NAMESPACE} --timeout=120s 2>/dev/null || warn "  Frontend not ready yet"

    echo ""
    ok "============================================"
    ok "  Domino deployed successfully!"
    ok "============================================"
    echo ""

    # Show access info
    show_ports
}

# ------ Delete all resources ------
delete_all() {
    warn "Deleting all Domino resources from namespace '${NAMESPACE}'..."
    read -p "Are you sure? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        kubectl delete namespace ${NAMESPACE} --ignore-not-found
        ok "Namespace '${NAMESPACE}' deleted."
    else
        log "Cancelled."
    fi
}

# ------ Show status ------
show_status() {
    echo ""
    log "Domino deployment status:"
    echo ""
    kubectl get all -n ${NAMESPACE} 2>/dev/null || err "Namespace '${NAMESPACE}' not found"
    echo ""
    log "PVCs:"
    kubectl get pvc -n ${NAMESPACE} 2>/dev/null
    echo ""
    log "NodePort Services:"
    kubectl get svc -n ${NAMESPACE} -o wide 2>/dev/null | grep NodePort
}

# ------ Show ports ------
show_ports() {
    echo ""
    log "NodePort access information:"
    echo ""

    local NODE_IP
    NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}' 2>/dev/null)
    if [ -z "$NODE_IP" ]; then
        NODE_IP="<NODE_IP>"
        warn "Could not detect node IP. Replace <NODE_IP> below with your node's IP."
    fi

    # Get NodePort values
    local FE_PORT REST_PORT AF_PORT
    FE_PORT=$(kubectl get svc domino-frontend-np -n ${NAMESPACE} -o jsonpath='{.spec.ports[0].nodePort}' 2>/dev/null || echo "N/A")
    REST_PORT=$(kubectl get svc domino-rest-np -n ${NAMESPACE} -o jsonpath='{.spec.ports[0].nodePort}' 2>/dev/null || echo "N/A")
    AF_PORT=$(kubectl get svc airflow-webserver-np -n ${NAMESPACE} -o jsonpath='{.spec.ports[0].nodePort}' 2>/dev/null || echo "N/A")

    echo -e "  ${CYAN}Frontend:${NC}     http://${NODE_IP}:${FE_PORT}/"
    echo -e "  ${CYAN}REST API:${NC}     http://${NODE_IP}:${REST_PORT}/api"
    echo -e "  ${CYAN}Airflow UI:${NC}   http://${NODE_IP}:${AF_PORT}/"
    echo ""
    log "Default credentials:"
    echo "    Domino:  admin@email.com / admin"
    echo "    Airflow: admin / (see 01-secrets.yaml)"
    echo ""
    log "Port-forward for local testing:"
    echo "    kubectl port-forward -n domino svc/domino-frontend 3000:80 &"
    echo "    kubectl port-forward -n domino svc/domino-rest 8000:8000 &"
    echo "    kubectl port-forward -n domino svc/airflow-webserver 8080:8080 &"
}

# ------ Show logs ------
show_logs() {
    local component=${1:-domino-rest}
    log "Tailing logs for ${component}..."
    kubectl logs -f deployment/${component} -n ${NAMESPACE} --all-containers=true
}

# ------ Main ------
case "${1:-apply}" in
    apply)   apply_all ;;
    delete)  delete_all ;;
    status)  show_status ;;
    ports)   show_ports ;;
    logs)    show_logs "${2:-domino-rest}" ;;
    *)
        echo "Usage: $0 [apply|delete|status|ports|logs [component]]"
        echo ""
        echo "Commands:"
        echo "  apply   - Deploy all resources (default)"
        echo "  delete  - Delete all resources"
        echo "  status  - Show deployment status"
        echo "  ports   - Show NodePort access info"
        echo "  logs    - Tail logs (default: domino-rest)"
        exit 1
        ;;
esac
