#!/bin/bash
set -euo pipefail

# RBAC Token Service Staging Deployment Script
# This script deploys the RBAC token service to staging environment

NAMESPACE="work-queue"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"

source "${SCRIPT_DIR}/lib/logging.sh"

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."

    if ! command -v kubectl &> /dev/null; then
        error "kubectl not found. Please install kubectl."
        exit 1
    fi

    if ! command -v docker &> /dev/null; then
        error "docker not found. Please install docker."
        exit 1
    fi

    # Check if we can connect to the cluster
    if ! kubectl cluster-info &> /dev/null; then
        error "Cannot connect to Kubernetes cluster. Please check your kubeconfig."
        exit 1
    fi

    log "Prerequisites check passed"
}

# Build RBAC token service image
build_image() {
    log "Building RBAC token service image..."

    cd "$PROJECT_ROOT"
    docker build -f deployments/docker/Dockerfile.rbac-token-service -t work-queue/rbac-token-service:latest .

    # Tag for staging
    docker tag work-queue/rbac-token-service:latest work-queue/rbac-token-service:staging

    log "Image built successfully"
}

# Generate secure secrets
generate_secrets() {
    log "Generating secure secrets..."

    # Generate random signing key (256 bits = 32 bytes = 64 hex chars)
    RBAC_SIGNING_KEY=$(openssl rand -hex 32)

    # Generate random encryption key (256 bits)
    RBAC_ENCRYPTION_KEY=$(openssl rand -hex 32)

    # Generate secure Redis password
    REDIS_PASSWORD=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-25)

    # Generate bootstrap admin token
    ADMIN_BOOTSTRAP_TOKEN=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-25)

    export RBAC_SIGNING_KEY RBAC_ENCRYPTION_KEY REDIS_PASSWORD ADMIN_BOOTSTRAP_TOKEN

    log "Secrets generated successfully"
}

# Deploy namespace and basic resources
deploy_namespace() {
    log "Creating namespace and basic resources..."

    kubectl apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: $NAMESPACE
  labels:
    environment: staging
    component: work-queue
EOF

    log "Namespace created"
}

# Deploy secrets with generated values
deploy_secrets() {
    log "Deploying secrets..."

    kubectl create secret generic rbac-secrets \
        --namespace="$NAMESPACE" \
        --from-literal=rbac-signing-key="$RBAC_SIGNING_KEY" \
        --from-literal=rbac-encryption-key="$RBAC_ENCRYPTION_KEY" \
        --from-literal=redis-password="$REDIS_PASSWORD" \
        --from-literal=admin-bootstrap-token="$ADMIN_BOOTSTRAP_TOKEN" \
        --dry-run=client -o yaml | kubectl apply -f -

    log "Secrets deployed"
}

# Deploy RBAC configuration
deploy_config() {
    log "Deploying RBAC configuration..."

    kubectl create configmap rbac-config \
        --namespace="$NAMESPACE" \
        --from-file=roles.yaml="$PROJECT_ROOT/deployments/docker/rbac-configs/roles.yaml" \
        --from-file=resources.yaml="$PROJECT_ROOT/deployments/docker/rbac-configs/resources.yaml" \
        --from-file=token-service.yaml="$PROJECT_ROOT/deployments/docker/rbac-configs/token-service.yaml" \
        --dry-run=client -o yaml | kubectl apply -f -

    log "Configuration deployed"
}

# Deploy the main RBAC service
deploy_rbac_service() {
    log "Deploying RBAC token service..."

    kubectl apply -f "$PROJECT_ROOT/deployments/kubernetes/rbac-token-service-deployment.yaml"

    log "RBAC service deployment applied"
}

# Wait for deployment to be ready
wait_for_deployment() {
    log "Waiting for RBAC token service to be ready..."

    kubectl wait --for=condition=available --timeout=300s deployment/rbac-token-service -n "$NAMESPACE"

    log "RBAC token service is ready"
}

# Run health checks
run_health_checks() {
    log "Running health checks..."

    # Port forward for testing
    kubectl port-forward service/rbac-token-service 8081:80 -n "$NAMESPACE" &
    PORT_FORWARD_PID=$!

    sleep 5

    # Health check
    if curl -f http://localhost:8081/health &> /dev/null; then
        log "Health check passed"
    else
        error "Health check failed"
        kill $PORT_FORWARD_PID 2>/dev/null || true
        exit 1
    fi

    # Metrics check
    if curl -f http://localhost:8081/metrics &> /dev/null; then
        log "Metrics endpoint accessible"
    else
        warn "Metrics endpoint not accessible"
    fi

    # Clean up port forward
    kill $PORT_FORWARD_PID 2>/dev/null || true

    log "Health checks completed"
}

# Print deployment info
print_deployment_info() {
    log "Deployment completed successfully!"
    echo
    echo "=== Deployment Information ==="
    echo "Namespace: $NAMESPACE"
    echo "Service: rbac-token-service"
    echo
    echo "To access the service:"
    echo "  kubectl port-forward service/rbac-token-service 8081:80 -n $NAMESPACE"
    echo "  curl http://localhost:8081/health"
    echo
    echo "To view logs:"
    echo "  kubectl logs -f deployment/rbac-token-service -n $NAMESPACE"
    echo
    echo "To view pod status:"
    echo "  kubectl get pods -n $NAMESPACE -l app=rbac-token-service"
    echo
    echo "Bootstrap admin token is stored in secret 'rbac-secrets'"
    echo "To retrieve it:"
    echo "  kubectl get secret rbac-secrets -n $NAMESPACE -o jsonpath='{.data.admin-bootstrap-token}' | base64 -d"
}

# Cleanup function for error handling
cleanup() {
    if [[ ${PORT_FORWARD_PID:-} ]]; then
        kill $PORT_FORWARD_PID 2>/dev/null || true
    fi
}

# Set trap for cleanup
trap cleanup EXIT

# Main deployment flow
main() {
    log "Starting RBAC Token Service deployment to staging..."

    check_prerequisites
    build_image
    generate_secrets
    deploy_namespace
    deploy_secrets
    deploy_config
    deploy_rbac_service
    wait_for_deployment
    run_health_checks
    print_deployment_info

    log "Deployment completed successfully!"
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
