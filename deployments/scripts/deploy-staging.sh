#!/bin/bash
set -euo pipefail

# Deployment script for Admin API to staging environment
# Usage: ./deploy-staging.sh [--rollback]

NAMESPACE="work-queue"
APP_NAME="admin-api"
ENVIRONMENT="staging"
REGISTRY="${DOCKER_REGISTRY:-docker.io}"
IMAGE_TAG="${IMAGE_TAG:-latest}"
KUBECONFIG="${KUBECONFIG:-$HOME/.kube/config}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."

    # Check kubectl
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed"
        exit 1
    fi

    # Check docker
    if ! command -v docker &> /dev/null; then
        log_error "docker is not installed"
        exit 1
    fi

    # Check cluster connectivity
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster"
        exit 1
    fi

    # Check namespace exists
    if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
        log_warn "Namespace $NAMESPACE does not exist, creating..."
        kubectl create namespace "$NAMESPACE"
    fi

    log_info "Prerequisites check passed"
}

# Build Docker image
build_image() {
    log_info "Building Docker image..."

    IMAGE_NAME="$REGISTRY/$APP_NAME:$IMAGE_TAG"

    # Build the image
    if ! docker build -f deployments/docker/Dockerfile.admin-api -t "$IMAGE_NAME" .; then
        log_error "Docker build failed"
        exit 1
    fi

    log_info "Docker image built: $IMAGE_NAME"

    # Push to registry if not local
    if [ "$REGISTRY" != "local" ]; then
        log_info "Pushing image to registry..."
        if ! docker push "$IMAGE_NAME"; then
            log_error "Failed to push image to registry"
            exit 1
        fi
    fi
}

# Run tests
run_tests() {
    log_info "Running smoke tests..."

    # Run unit tests
    go test ./internal/admin-api/... -v

    if [ $? -ne 0 ]; then
        log_error "Unit tests failed"
        exit 1
    fi

    log_info "Tests passed"
}

# Deploy to Kubernetes
deploy_to_k8s() {
    log_info "Deploying to Kubernetes..."

    # Update image in deployment
    kubectl set image "deployment/$APP_NAME" "$APP_NAME=$REGISTRY/$APP_NAME:$IMAGE_TAG" -n "$NAMESPACE"

    # Apply configurations
    kubectl apply -f deployments/kubernetes/admin-api-deployment.yaml
    kubectl apply -f deployments/kubernetes/monitoring.yaml

    # Wait for rollout to complete
    log_info "Waiting for deployment rollout..."
    kubectl rollout status "deployment/$APP_NAME" -n "$NAMESPACE" --timeout=300s

    if [ $? -ne 0 ]; then
        log_error "Deployment rollout failed"
        rollback
        exit 1
    fi

    log_info "Deployment successful"
}

# Verify deployment
verify_deployment() {
    log_info "Verifying deployment..."

    # Check pod status
    READY_PODS=$(kubectl get pods -n "$NAMESPACE" -l app="$APP_NAME" -o jsonpath='{.items[*].status.containerStatuses[0].ready}' | tr ' ' '\n' | grep -c "true")
    TOTAL_PODS=$(kubectl get pods -n "$NAMESPACE" -l app="$APP_NAME" --no-headers | wc -l)

    if [ "$READY_PODS" -ne "$TOTAL_PODS" ]; then
        log_error "Not all pods are ready ($READY_PODS/$TOTAL_PODS)"
        exit 1
    fi

    log_info "All pods are ready ($READY_PODS/$TOTAL_PODS)"

    # Check service endpoint
    SERVICE_IP=$(kubectl get service "$APP_NAME" -n "$NAMESPACE" -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

    if [ -z "$SERVICE_IP" ]; then
        SERVICE_IP=$(kubectl get service "$APP_NAME" -n "$NAMESPACE" -o jsonpath='{.spec.clusterIP}')
    fi

    # Port forward for health check
    kubectl port-forward -n "$NAMESPACE" "service/$APP_NAME" 8080:80 &
    PF_PID=$!
    sleep 5

    # Health check
    HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health)

    if [[ -n "${PF_PID:-}" ]]; then
        kill "$PF_PID" 2>/dev/null || true
    fi

    if [ "$HTTP_STATUS" -ne 200 ]; then
        log_error "Health check failed (HTTP $HTTP_STATUS)"
        exit 1
    fi

    log_info "Health check passed"
}

# Rollback deployment
rollback() {
    log_warn "Rolling back deployment..."
    kubectl rollout undo "deployment/$APP_NAME" -n "$NAMESPACE"
    kubectl rollout status "deployment/$APP_NAME" -n "$NAMESPACE" --timeout=300s
    log_info "Rollback completed"
}

# Run smoke tests
run_smoke_tests() {
    log_info "Running smoke tests..."

    # Port forward
    kubectl port-forward -n "$NAMESPACE" "service/$APP_NAME" 8080:80 &
    PF_PID=$!
    sleep 5

    # Test endpoints
    ENDPOINTS=("/health" "/api/v1/stats" "/api/v1/stats/keys")

    for endpoint in "${ENDPOINTS[@]}"; do
        HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "X-API-Token: test-token" http://localhost:8080$endpoint)

        if [ "$HTTP_STATUS" -eq 200 ] || [ "$HTTP_STATUS" -eq 401 ]; then
            log_info "Endpoint $endpoint responded with $HTTP_STATUS"
        else
            log_error "Endpoint $endpoint failed with $HTTP_STATUS"
            if [[ -n "${PF_PID:-}" ]]; then
                kill "$PF_PID" 2>/dev/null || true
            fi
            exit 1
        fi
    done

    if [[ -n "${PF_PID:-}" ]]; then
        kill "$PF_PID" 2>/dev/null || true
    fi
    log_info "Smoke tests passed"
}

# Print deployment info
print_deployment_info() {
    log_info "Deployment Information:"
    echo "========================"
    echo "Namespace: $NAMESPACE"
    echo "Application: $APP_NAME"
    echo "Environment: $ENVIRONMENT"
    echo "Image: $REGISTRY/$APP_NAME:$IMAGE_TAG"
    echo ""

    kubectl get deployment "$APP_NAME" -n "$NAMESPACE"
    echo ""
    kubectl get pods -n "$NAMESPACE" -l app="$APP_NAME"
    echo ""
    kubectl get service "$APP_NAME" -n "$NAMESPACE"
    echo ""

    INGRESS_HOST=$(kubectl get ingress "$APP_NAME" -n "$NAMESPACE" -o jsonpath='{.spec.rules[0].host}')
    if [[ -n "$INGRESS_HOST" ]]; then
        echo "Access URL: https://$INGRESS_HOST"
    fi
}

# Main deployment flow
main() {
    log_info "Starting deployment of $APP_NAME to $ENVIRONMENT"

    # Check for rollback flag
    if [ "${1:-}" == "--rollback" ]; then
        rollback
        exit 0
    fi

    # Run deployment steps
    check_prerequisites
    run_tests
    build_image
    deploy_to_k8s
    verify_deployment
    run_smoke_tests
    print_deployment_info

    log_info "Deployment completed successfully!"
}

# Run main function
main "$@"
