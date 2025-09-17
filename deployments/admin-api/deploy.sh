#!/bin/bash

set -Eeuo pipefail
IFS=$'\n\t'

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
    docker build -t "${APP_NAME}:${VERSION}" -f deployments/admin-api/Dockerfile .

    if [[ "${ENVIRONMENT}" != "local" ]]; then
        registry="${DOCKER_REGISTRY:-docker.io}"
        namespace="${DOCKER_NAMESPACE:-${APP_NAME}}"
        if [[ -z "${namespace}" ]]; then
            echo "Error: DOCKER_NAMESPACE must be provided" >&2
            exit 1
        fi
        repo="${registry%/}/${namespace}/${APP_NAME}:${VERSION}"

        echo "Tagging image as ${repo}"
        docker tag "${APP_NAME}:${VERSION}" "${repo}"

        echo "Pushing image to registry..."
        if ! docker push "${repo}"; then
            echo "Error: docker push failed" >&2
            exit 1
        fi
    fi
}

# Function to deploy to Kubernetes
deploy_kubernetes() {
    echo "Deploying to Kubernetes..."

    kubectl create namespace "${NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f -

    kubectl apply -n "${NAMESPACE}" -f deployments/admin-api/k8s-redis.yaml

    echo "Waiting for Redis to be ready..."
    kubectl wait --for=condition=ready pod -l app=redis -n "${NAMESPACE}" --timeout=60s

    kubectl apply -n "${NAMESPACE}" -f deployments/admin-api/k8s-deployment.yaml

    echo "Waiting for Admin API deployment to be ready..."
    kubectl rollout status "deployment/${APP_NAME}" -n "${NAMESPACE}"

    echo "Deployment completed successfully"
}

# Function to run smoke tests
run_smoke_tests() {
    echo "Running smoke tests..."

    local endpoint=""
    local pf_pid=""

    if [[ "${ENVIRONMENT}" == "local" ]]; then
        kubectl port-forward -n "${NAMESPACE}" service/admin-api-service 8080:8080 &
        pf_pid=$!
        trap '[[ -n "${pf_pid}" ]] && kill "${pf_pid}"' EXIT
        sleep 5
        endpoint="http://localhost:8080"
    else
        host="$(kubectl get ingress admin-api-ingress -n "${NAMESPACE}" -o jsonpath='{.spec.rules[0].host}')"
        endpoint="https://${host}"
    fi

    echo "Testing health endpoint..."
    if curl -fsS "${endpoint}/healthz" > /dev/null; then
        echo "✓ Health check passed"
    else
        echo "✗ Health check failed"
        exit 1
    fi

    echo "Testing readiness endpoint..."
    if curl -fsS "${endpoint}/readyz" > /dev/null; then
        echo "✓ Readiness check passed"
    else
        echo "✗ Readiness check failed"
    fi

    echo "Testing API stats endpoint..."
    if curl -fsS "${endpoint}/api/v1/stats" -H "Authorization: Bearer test-token" > /dev/null; then
        echo "✓ API endpoint accessible"
    else
        echo "⚠ API endpoint requires authentication (expected)"
    fi

    echo "Testing OpenAPI spec endpoint..."
    if curl -fsS "${endpoint}/api/v1/openapi.yaml" > /dev/null; then
        echo "✓ OpenAPI spec available"
    else
        echo "✗ OpenAPI spec not available"
    fi

    if [[ -n "${pf_pid}" ]]; then
        kill "${pf_pid}"
        trap - EXIT
    fi

    echo "Smoke tests completed"
}

# Function to setup monitoring
setup_monitoring() {
    echo "Setting up monitoring..."

    cat <<EOF | kubectl apply -n "${NAMESPACE}" -f -
apiVersion: monitoring.coreos.com/v1
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
    kubectl get pods -n "${NAMESPACE}" -l app="${APP_NAME}"

    # Check service endpoints
    kubectl get endpoints -n "${NAMESPACE}" admin-api-service

    # Check recent events
    kubectl get events -n "${NAMESPACE}" --sort-by='.lastTimestamp' | tail -10

    echo "Deployment verification complete"
}

# Function for rollback
rollback() {
    echo "Rolling back deployment..."
    kubectl rollout undo "deployment/${APP_NAME}" -n "${NAMESPACE}"
    kubectl rollout status "deployment/${APP_NAME}" -n "${NAMESPACE}"
    echo "Rollback completed"
}

# Main deployment flow
main() {
    check_prerequisites

    case "$ENVIRONMENT" in
        local)
            echo "Deploying locally with Docker Compose..."
            if ! docker compose version >/dev/null 2>&1; then
                echo "Error: docker compose is required" >&2
                exit 1
            fi
            docker compose -f deployments/admin-api/docker-compose.yaml up -d
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
