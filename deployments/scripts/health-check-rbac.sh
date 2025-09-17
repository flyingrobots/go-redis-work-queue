#!/bin/bash
set -euo pipefail

# RBAC Token Service Health Check Script
# This script performs comprehensive health checks on the RBAC token service

NAMESPACE="${NAMESPACE:-work-queue}"
SERVICE_NAME="${SERVICE_NAME:-rbac-token-service}"
TIMEOUT="${TIMEOUT:-30}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

info() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] INFO: $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
}

# Check if kubectl is available and cluster is accessible
check_kubectl() {
    if ! command -v kubectl &> /dev/null; then
        error "kubectl not found"
        return 1
    fi

    if ! kubectl cluster-info &> /dev/null; then
        error "Cannot connect to Kubernetes cluster"
        return 1
    fi

    return 0
}

# Check if namespace exists
check_namespace() {
    if kubectl get namespace "$NAMESPACE" &> /dev/null; then
        log "Namespace '$NAMESPACE' exists"
        return 0
    else
        error "Namespace '$NAMESPACE' does not exist"
        return 1
    fi
}

# Check deployment status
check_deployment() {
    local deployment_status
    deployment_status=$(kubectl get deployment "$SERVICE_NAME" -n "$NAMESPACE" -o jsonpath='{.status.conditions[?(@.type=="Available")].status}' 2>/dev/null || echo "NotFound")

    if [[ "$deployment_status" == "True" ]]; then
        log "Deployment '$SERVICE_NAME' is available"
        return 0
    else
        error "Deployment '$SERVICE_NAME' is not available (status: $deployment_status)"
        return 1
    fi
}

# Check pod status
check_pods() {
    local pod_count ready_pods
    pod_count=$(kubectl get pods -n "$NAMESPACE" -l "app=$SERVICE_NAME" --no-headers 2>/dev/null | wc -l)
    ready_pods=$(kubectl get pods -n "$NAMESPACE" -l "app=$SERVICE_NAME" -o jsonpath='{.items[*].status.conditions[?(@.type=="Ready")].status}' 2>/dev/null | grep -o "True" | wc -l)

    if [[ $pod_count -eq 0 ]]; then
        error "No pods found for service '$SERVICE_NAME'"
        return 1
    fi

    if [[ $ready_pods -eq $pod_count ]]; then
        log "All $pod_count pods are ready"
        return 0
    else
        error "Only $ready_pods out of $pod_count pods are ready"

        # Show pod status for debugging
        info "Pod status:"
        kubectl get pods -n "$NAMESPACE" -l "app=$SERVICE_NAME" -o wide

        return 1
    fi
}

# Check service
check_service() {
    local service_ip
    service_ip=$(kubectl get service "$SERVICE_NAME" -n "$NAMESPACE" -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "NotFound")

    if [[ "$service_ip" != "NotFound" && "$service_ip" != "" ]]; then
        log "Service '$SERVICE_NAME' has cluster IP: $service_ip"
        return 0
    else
        error "Service '$SERVICE_NAME' not found or has no cluster IP"
        return 1
    fi
}

# Check configmap
check_configmap() {
    if kubectl get configmap rbac-config -n "$NAMESPACE" &> /dev/null; then
        log "ConfigMap 'rbac-config' exists"
        return 0
    else
        error "ConfigMap 'rbac-config' not found"
        return 1
    fi
}

# Check secrets
check_secrets() {
    if kubectl get secret rbac-secrets -n "$NAMESPACE" &> /dev/null; then
        log "Secret 'rbac-secrets' exists"
        return 0
    else
        error "Secret 'rbac-secrets' not found"
        return 1
    fi
}

# Check persistent volume claims
check_pvcs() {
    local pvc_errors=0

    for pvc in "rbac-keys-pvc" "rbac-audit-pvc"; do
        local pvc_status
        pvc_status=$(kubectl get pvc "$pvc" -n "$NAMESPACE" -o jsonpath='{.status.phase}' 2>/dev/null || echo "NotFound")

        if [[ "$pvc_status" == "Bound" ]]; then
            log "PVC '$pvc' is bound"
        else
            error "PVC '$pvc' is not bound (status: $pvc_status)"
            ((pvc_errors++))
        fi
    done

    return $pvc_errors
}

# Perform HTTP health check
check_http_health() {
    log "Starting HTTP health check..."

    # Start port forward in background
    kubectl port-forward service/"$SERVICE_NAME" 8081:80 -n "$NAMESPACE" &
    local port_forward_pid=$!

    # Function to cleanup port forward
    cleanup_port_forward() {
        if kill -0 "$port_forward_pid" 2>/dev/null; then
            kill "$port_forward_pid" 2>/dev/null || true
            wait "$port_forward_pid" 2>/dev/null || true
        fi
    }

    # Set trap for cleanup
    trap cleanup_port_forward EXIT

    # Wait for port forward to establish
    sleep 3

    local health_check_passed=0

    # Health endpoint check
    if timeout "$TIMEOUT" bash -c 'while ! curl -f http://localhost:8081/health &>/dev/null; do sleep 1; done'; then
        log "Health endpoint responded successfully"
    else
        error "Health endpoint did not respond within $TIMEOUT seconds"
        ((health_check_passed++))
    fi

    # Metrics endpoint check
    if timeout "$TIMEOUT" bash -c 'while ! curl -f http://localhost:8081/metrics &>/dev/null; do sleep 1; done'; then
        log "Metrics endpoint responded successfully"
    else
        warn "Metrics endpoint did not respond within $TIMEOUT seconds"
    fi

    # Check if service responds with expected status codes
    local health_status
    health_status=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8081/health 2>/dev/null || echo "000")

    if [[ "$health_status" == "200" ]]; then
        log "Health endpoint returned HTTP 200"
    else
        error "Health endpoint returned HTTP $health_status (expected 200)"
        ((health_check_passed++))
    fi

    cleanup_port_forward
    trap - EXIT

    return $health_check_passed
}

# Check resource usage
check_resource_usage() {
    log "Checking resource usage..."

    local pods
    pods=$(kubectl get pods -n "$NAMESPACE" -l "app=$SERVICE_NAME" -o name 2>/dev/null)

    if [[ -z "$pods" ]]; then
        warn "No pods found for resource usage check"
        return 1
    fi

    for pod in $pods; do
        local cpu_usage memory_usage
        cpu_usage=$(kubectl top "$pod" -n "$NAMESPACE" --no-headers 2>/dev/null | awk '{print $2}' || echo "N/A")
        memory_usage=$(kubectl top "$pod" -n "$NAMESPACE" --no-headers 2>/dev/null | awk '{print $3}' || echo "N/A")

        info "Pod $pod - CPU: $cpu_usage, Memory: $memory_usage"
    done

    return 0
}

# Check recent logs for errors
check_logs() {
    log "Checking recent logs for errors..."

    local error_count
    error_count=$(kubectl logs -n "$NAMESPACE" -l "app=$SERVICE_NAME" --since=5m 2>/dev/null | grep -i -c error || true)

    if [[ $error_count -eq 0 ]]; then
        log "No errors found in recent logs"
        return 0
    else
        warn "Found $error_count error(s) in recent logs"

        info "Recent error messages:"
        kubectl logs -n "$NAMESPACE" -l "app=$SERVICE_NAME" --since=5m 2>/dev/null | grep -i error | tail -5

        return 1
    fi
}

# Generate health report
generate_report() {
    local overall_status="HEALTHY"
    local failed_checks=0

    echo
    echo "=== RBAC Token Service Health Check Report ==="
    echo "Timestamp: $(date)"
    echo "Namespace: $NAMESPACE"
    echo "Service: $SERVICE_NAME"
    echo

    # Run all checks
    local checks=(
        "check_kubectl:Kubectl Connection"
        "check_namespace:Namespace"
        "check_deployment:Deployment"
        "check_pods:Pods"
        "check_service:Service"
        "check_configmap:ConfigMap"
        "check_secrets:Secrets"
        "check_pvcs:Persistent Volume Claims"
        "check_http_health:HTTP Health"
        "check_logs:Log Analysis"
    )

    for check_info in "${checks[@]}"; do
        local check_func="${check_info%%:*}"
        local check_name="${check_info##*:}"

        printf "%-30s: " "$check_name"

        if $check_func; then
            echo -e "${GREEN}PASS${NC}"
        else
            echo -e "${RED}FAIL${NC}"
            overall_status="UNHEALTHY"
            ((failed_checks++))
        fi
    done

    # Resource usage (informational only)
    printf "%-30s: " "Resource Usage"
    if check_resource_usage; then
        echo -e "${BLUE}INFO${NC}"
    else
        echo -e "${YELLOW}N/A${NC}"
    fi

    echo
    echo "=== Summary ==="
    echo "Overall Status: $overall_status"
    echo "Failed Checks: $failed_checks"

    if [[ "$overall_status" == "HEALTHY" ]]; then
        log "All health checks passed!"
        return 0
    else
        error "$failed_checks health check(s) failed"
        return 1
    fi
}

# Show usage
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo
    echo "Options:"
    echo "  -n, --namespace NAMESPACE    Kubernetes namespace (default: work-queue)"
    echo "  -s, --service SERVICE        Service name (default: rbac-token-service)"
    echo "  -t, --timeout TIMEOUT        HTTP timeout in seconds (default: 30)"
    echo "  -h, --help                   Show this help message"
    echo
    echo "Environment variables:"
    echo "  NAMESPACE    Override default namespace"
    echo "  SERVICE_NAME Override default service name"
    echo "  TIMEOUT      Override default timeout"
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -n|--namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            -s|--service)
                SERVICE_NAME="$2"
                shift 2
                ;;
            -t|--timeout)
                TIMEOUT="$2"
                shift 2
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *)
                error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done
}

# Main function
main() {
    parse_args "$@"

    log "Starting RBAC Token Service health check..."
    log "Namespace: $NAMESPACE"
    log "Service: $SERVICE_NAME"
    log "Timeout: ${TIMEOUT}s"

    generate_report
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
