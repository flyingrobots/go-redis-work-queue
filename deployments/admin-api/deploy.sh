#!/bin/bash

set -e

# Configuration
NAMESPACE="work-queue"
APP_NAME="admin-api"
ENVIRONMENT="${1:-staging}"
VERSION="${2:-latest}"

echo "Deploying Admin API to ${ENVIRONMENT} environment (version: ${VERSION})"

# Function to check prerequisites
check_prerequisites() {
    echo "Checking prerequisites..."

    if ! command -v kubectl &> /dev/null; then
        echo "Error: kubectl is not installed"
        exit 1
    fi

    if ! command -v docker &> /dev/null; then
        echo "Error: docker is not installed"
        exit 1
    fi

    echo "Prerequisites check passed"
}

# Function to build Docker image
build_image() {
    echo "Building Docker image..."
    docker build -t ${APP_NAME}:${VERSION} -f deployments/admin-api/Dockerfile .

    if [ "$ENVIRONMENT" != "local" ]; then
        # Tag for registry
        REGISTRY_URL="${DOCKER_REGISTRY:-docker.io}"
        docker tag ${APP_NAME}:${VERSION} ${REGISTRY_URL}/${APP_NAME}:${VERSION}

        echo "Pushing image to registry..."
        docker push ${REGISTRY_URL}/${APP_NAME}:${VERSION}
    fi
}

# Function to deploy to Kubernetes
deploy_kubernetes() {
    echo "Deploying to Kubernetes..."

    # Create namespace if it doesn't exist
    kubectl create namespace ${NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -

    # Apply Redis deployment
    kubectl apply -f deployments/admin-api/k8s-redis.yaml

    # Wait for Redis to be ready
    echo "Waiting for Redis to be ready..."
    kubectl wait --for=condition=ready pod -l app=redis -n ${NAMESPACE} --timeout=60s

    # Apply Admin API deployment
    kubectl apply -f deployments/admin-api/k8s-deployment.yaml

    # Wait for deployment to be ready
    echo "Waiting for Admin API deployment to be ready..."
    kubectl rollout status deployment/${APP_NAME} -n ${NAMESPACE}

    echo "Deployment completed successfully"
}

# Function to run smoke tests
run_smoke_tests() {
    echo "Running smoke tests..."

    # Get service endpoint
    if [ "$ENVIRONMENT" == "local" ]; then
        # For local testing with port-forward
        kubectl port-forward -n ${NAMESPACE} service/admin-api-service 8080:8080 &
        PF_PID=$!
        sleep 5

        ENDPOINT="http://localhost:8080"
    else
        # Get ingress endpoint
        ENDPOINT=$(kubectl get ingress admin-api-ingress -n ${NAMESPACE} -o jsonpath='{.spec.rules[0].host}')
        ENDPOINT="https://${ENDPOINT}"
    fi

    # Test health endpoint
    echo "Testing health endpoint..."
    if curl -f "${ENDPOINT}/health" > /dev/null 2>&1; then
        echo "✓ Health check passed"
    else
        echo "✗ Health check failed"
        [ ! -z "$PF_PID" ] && kill $PF_PID
        exit 1
    fi

    # Test API endpoint
    echo "Testing API stats endpoint..."
    if curl -f "${ENDPOINT}/api/v1/stats" -H "Authorization: Bearer test-token" > /dev/null 2>&1; then
        echo "✓ API endpoint accessible"
    else
        echo "⚠ API endpoint requires authentication (expected)"
    fi

    # Test OpenAPI spec
    echo "Testing OpenAPI spec endpoint..."
    if curl -f "${ENDPOINT}/api/v1/openapi.yaml" > /dev/null 2>&1; then
        echo "✓ OpenAPI spec available"
    else
        echo "✗ OpenAPI spec not available"
    fi

    # Clean up port-forward if local
    [ ! -z "$PF_PID" ] && kill $PF_PID

    echo "Smoke tests completed"
}

# Function to setup monitoring
setup_monitoring() {
    echo "Setting up monitoring..."

    cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ServiceMonitor
metadata:
  name: admin-api-monitor
  namespace: ${NAMESPACE}
spec:
  selector:
    matchLabels:
      app: admin-api
  endpoints:
  - port: http
    interval: 30s
    path: /metrics
EOF

    echo "Monitoring configured"
}

# Function to verify deployment
verify_deployment() {
    echo "Verifying deployment..."

    # Check pod status
    kubectl get pods -n ${NAMESPACE} -l app=${APP_NAME}

    # Check service endpoints
    kubectl get endpoints -n ${NAMESPACE} admin-api-service

    # Check recent events
    kubectl get events -n ${NAMESPACE} --sort-by='.lastTimestamp' | tail -10

    echo "Deployment verification complete"
}

# Function for rollback
rollback() {
    echo "Rolling back deployment..."
    kubectl rollout undo deployment/${APP_NAME} -n ${NAMESPACE}
    kubectl rollout status deployment/${APP_NAME} -n ${NAMESPACE}
    echo "Rollback completed"
}

# Main deployment flow
main() {
    check_prerequisites

    case "$ENVIRONMENT" in
        local)
            echo "Deploying locally with Docker Compose..."
            docker-compose -f deployments/admin-api/docker-compose.yaml up -d
            ;;
        staging)
            build_image
            deploy_kubernetes
            run_smoke_tests
            setup_monitoring
            verify_deployment
            ;;
        rollback)
            rollback
            ;;
        *)
            echo "Usage: $0 [local|staging|rollback] [version]"
            exit 1
            ;;
    esac

    echo "Deployment script completed successfully"
}

# Run main function
main
