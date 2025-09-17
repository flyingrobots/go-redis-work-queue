#!/bin/bash
set -euo pipefail

# RBAC Token Service Staging Deployment Test Script
# This script performs comprehensive testing of the RBAC deployment in staging

NAMESPACE="${NAMESPACE:-work-queue}"
SERVICE_NAME="${SERVICE_NAME:-rbac-token-service}"
TIMEOUT="${TIMEOUT:-300}"

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

# Test results tracking
TEST_RESULTS=()
PASSED_TESTS=0
FAILED_TESTS=0

# Add test result
add_test_result() {
    local test_name="$1"
    local result="$2"
    local message="$3"

    TEST_RESULTS+=("$test_name:$result:$message")

    if [[ "$result" == "PASS" ]]; then
        ((PASSED_TESTS++))
        log "✓ $test_name: $message"
    else
        ((FAILED_TESTS++))
        error "✗ $test_name: $message"
    fi
}

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites for staging deployment test..."

    local all_passed=true

    # Check kubectl
    if command -v kubectl &> /dev/null; then
        add_test_result "kubectl-available" "PASS" "kubectl is available"
    else
        add_test_result "kubectl-available" "FAIL" "kubectl not found"
        all_passed=false
    fi

    # Check jq
    if command -v jq &> /dev/null; then
        add_test_result "jq-available" "PASS" "jq is available"
    else
        add_test_result "jq-available" "FAIL" "jq not found"
        all_passed=false
    fi

    # Check cluster connection
    if kubectl cluster-info &> /dev/null; then
        add_test_result "cluster-connection" "PASS" "Can connect to Kubernetes cluster"
    else
        add_test_result "cluster-connection" "FAIL" "Cannot connect to Kubernetes cluster"
        all_passed=false
    fi

    # Check if namespace exists
    if kubectl get namespace "$NAMESPACE" &> /dev/null; then
        add_test_result "namespace-exists" "PASS" "Namespace '$NAMESPACE' exists"
    else
        add_test_result "namespace-exists" "FAIL" "Namespace '$NAMESPACE' does not exist"
        all_passed=false
    fi

    if [[ "$all_passed" == "false" ]]; then
        error "Prerequisites check failed. Cannot proceed with testing."
        return 1
    fi

    return 0
}

# Test deployment status
test_deployment_status() {
    log "Testing deployment status..."

    # Check deployment exists and is available
    local deployment_status
    deployment_status=$(kubectl get deployment "$SERVICE_NAME" -n "$NAMESPACE" -o jsonpath='{.status.conditions[?(@.type=="Available")].status}' 2>/dev/null || echo "NotFound")

    if [[ "$deployment_status" == "True" ]]; then
        add_test_result "deployment-available" "PASS" "Deployment is available"
    else
        add_test_result "deployment-available" "FAIL" "Deployment not available (status: $deployment_status)"
    fi

    # Check replicas
    local desired_replicas ready_replicas
    desired_replicas=$(kubectl get deployment "$SERVICE_NAME" -n "$NAMESPACE" -o jsonpath='{.spec.replicas}' 2>/dev/null || echo "0")
    ready_replicas=$(kubectl get deployment "$SERVICE_NAME" -n "$NAMESPACE" -o jsonpath='{.status.readyReplicas}' 2>/dev/null || echo "0")

    if [[ "$desired_replicas" == "$ready_replicas" && "$ready_replicas" -gt 0 ]]; then
        add_test_result "deployment-replicas" "PASS" "$ready_replicas/$desired_replicas replicas ready"
    else
        add_test_result "deployment-replicas" "FAIL" "Only $ready_replicas/$desired_replicas replicas ready"
    fi
}

# Test pods status
test_pods_status() {
    log "Testing pods status..."

    local pods
    pods=$(kubectl get pods -n "$NAMESPACE" -l "app=$SERVICE_NAME" -o name 2>/dev/null)

    if [[ -z "$pods" ]]; then
        add_test_result "pods-exist" "FAIL" "No pods found for service '$SERVICE_NAME'"
        return 1
    fi

    add_test_result "pods-exist" "PASS" "Found pods for service '$SERVICE_NAME'"

    # Check each pod status
    local all_pods_ready=true
    for pod in $pods; do
        local pod_name="${pod##*/}"
        local pod_ready
        pod_ready=$(kubectl get "$pod" -n "$NAMESPACE" -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}' 2>/dev/null || echo "Unknown")

        if [[ "$pod_ready" == "True" ]]; then
            add_test_result "pod-ready-$pod_name" "PASS" "Pod $pod_name is ready"
        else
            add_test_result "pod-ready-$pod_name" "FAIL" "Pod $pod_name is not ready (status: $pod_ready)"
            all_pods_ready=false
        fi
    done

    if [[ "$all_pods_ready" == "true" ]]; then
        add_test_result "all-pods-ready" "PASS" "All pods are ready"
    else
        add_test_result "all-pods-ready" "FAIL" "Some pods are not ready"
    fi
}

# Test service and networking
test_service_networking() {
    log "Testing service and networking..."

    # Check service exists
    local service_ip
    service_ip=$(kubectl get service "$SERVICE_NAME" -n "$NAMESPACE" -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "NotFound")

    if [[ "$service_ip" != "NotFound" && "$service_ip" != "" ]]; then
        add_test_result "service-exists" "PASS" "Service exists with cluster IP: $service_ip"
    else
        add_test_result "service-exists" "FAIL" "Service not found or has no cluster IP"
        return 1
    fi

    # Check service endpoints
    local endpoints_count
    endpoints_count=$(kubectl get endpoints "$SERVICE_NAME" -n "$NAMESPACE" -o jsonpath='{.subsets[*].addresses[*].ip}' 2>/dev/null | wc -w || echo "0")

    if [[ "$endpoints_count" -gt 0 ]]; then
        add_test_result "service-endpoints" "PASS" "Service has $endpoints_count endpoint(s)"
    else
        add_test_result "service-endpoints" "FAIL" "Service has no endpoints"
    fi
}

# Test configuration resources
test_configuration() {
    log "Testing configuration resources..."

    # Check ConfigMap
    if kubectl get configmap rbac-config -n "$NAMESPACE" &> /dev/null; then
        add_test_result "configmap-exists" "PASS" "ConfigMap 'rbac-config' exists"

        # Validate ConfigMap contents
        local config_files
        config_files=$(kubectl get configmap rbac-config -n "$NAMESPACE" -o jsonpath='{.data}' | jq -r 'keys | .[]' 2>/dev/null || echo "")

        if echo "$config_files" | grep -q "roles.yaml"; then
            add_test_result "configmap-roles" "PASS" "roles.yaml found in ConfigMap"
        else
            add_test_result "configmap-roles" "FAIL" "roles.yaml missing from ConfigMap"
        fi

        if echo "$config_files" | grep -q "resources.yaml"; then
            add_test_result "configmap-resources" "PASS" "resources.yaml found in ConfigMap"
        else
            add_test_result "configmap-resources" "FAIL" "resources.yaml missing from ConfigMap"
        fi

        if echo "$config_files" | grep -q "token-service.yaml"; then
            add_test_result "configmap-token-service" "PASS" "token-service.yaml found in ConfigMap"
        else
            add_test_result "configmap-token-service" "FAIL" "token-service.yaml missing from ConfigMap"
        fi
    else
        add_test_result "configmap-exists" "FAIL" "ConfigMap 'rbac-config' does not exist"
    fi

    # Check Secret
    if kubectl get secret rbac-secrets -n "$NAMESPACE" &> /dev/null; then
        add_test_result "secret-exists" "PASS" "Secret 'rbac-secrets' exists"

        # Check required secret keys
        local secret_keys
        secret_keys=$(kubectl get secret rbac-secrets -n "$NAMESPACE" -o jsonpath='{.data}' | jq -r 'keys | .[]' 2>/dev/null || echo "")

        for key in "rbac-signing-key" "rbac-encryption-key" "redis-password" "admin-bootstrap-token"; do
            if echo "$secret_keys" | grep -q "$key"; then
                add_test_result "secret-key-$key" "PASS" "Secret key '$key' exists"
            else
                add_test_result "secret-key-$key" "FAIL" "Secret key '$key' missing"
            fi
        done
    else
        add_test_result "secret-exists" "FAIL" "Secret 'rbac-secrets' does not exist"
    fi
}

# Test persistent volumes
test_persistent_volumes() {
    log "Testing persistent volumes..."

    for pvc in "rbac-keys-pvc" "rbac-audit-pvc"; do
        local pvc_status
        pvc_status=$(kubectl get pvc "$pvc" -n "$NAMESPACE" -o jsonpath='{.status.phase}' 2>/dev/null || echo "NotFound")

        if [[ "$pvc_status" == "Bound" ]]; then
            add_test_result "pvc-$pvc" "PASS" "PVC '$pvc' is bound"
        else
            add_test_result "pvc-$pvc" "FAIL" "PVC '$pvc' not bound (status: $pvc_status)"
        fi
    done
}

# Test HTTP endpoints
test_http_endpoints() {
    log "Testing HTTP endpoints..."

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
    sleep 5

    # Test health endpoint
    if timeout 30 bash -c 'while ! curl -f http://localhost:8081/health &>/dev/null; do sleep 1; done'; then
        add_test_result "health-endpoint" "PASS" "Health endpoint responds"

        # Check health response
        local health_response
        health_response=$(curl -s http://localhost:8081/health 2>/dev/null || echo "")

        if echo "$health_response" | grep -q "status"; then
            add_test_result "health-response" "PASS" "Health endpoint returns valid response"
        else
            add_test_result "health-response" "FAIL" "Health endpoint response invalid"
        fi
    else
        add_test_result "health-endpoint" "FAIL" "Health endpoint not responding"
    fi

    # Test metrics endpoint
    if timeout 30 bash -c 'while ! curl -f http://localhost:8081/metrics &>/dev/null; do sleep 1; done'; then
        add_test_result "metrics-endpoint" "PASS" "Metrics endpoint responds"

        # Check metrics content
        local metrics_response
        metrics_response=$(curl -s http://localhost:8081/metrics 2>/dev/null || echo "")

        if echo "$metrics_response" | grep -q "# TYPE"; then
            add_test_result "metrics-content" "PASS" "Metrics endpoint returns Prometheus metrics"
        else
            add_test_result "metrics-content" "FAIL" "Metrics endpoint response invalid"
        fi
    else
        add_test_result "metrics-endpoint" "FAIL" "Metrics endpoint not responding"
    fi

    cleanup_port_forward
    trap - EXIT
}

# Test RBAC functionality
test_rbac_functionality() {
    log "Testing RBAC functionality..."

    # Start port forward for API testing
    kubectl port-forward service/"$SERVICE_NAME" 8081:80 -n "$NAMESPACE" &
    local port_forward_pid=$!

    cleanup_port_forward() {
        if kill -0 "$port_forward_pid" 2>/dev/null; then
            kill "$port_forward_pid" 2>/dev/null || true
            wait "$port_forward_pid" 2>/dev/null || true
        fi
    }

    trap cleanup_port_forward EXIT

    sleep 5

    # Get bootstrap token
    local bootstrap_token
    bootstrap_token=$(kubectl get secret rbac-secrets -n "$NAMESPACE" -o jsonpath='{.data.admin-bootstrap-token}' | base64 -d 2>/dev/null || echo "")

    if [[ -n "$bootstrap_token" ]]; then
        add_test_result "bootstrap-token-retrieved" "PASS" "Bootstrap token retrieved successfully"

        # Test token validation endpoint with bootstrap token
        local auth_response
        auth_response=$(curl -s -H "Authorization: Bearer $bootstrap_token" http://localhost:8081/api/v1/auth/validate 2>/dev/null || echo "")

        if [[ -n "$auth_response" ]]; then
            add_test_result "token-validation" "PASS" "Token validation endpoint accessible"
        else
            add_test_result "token-validation" "FAIL" "Token validation endpoint not accessible"
        fi

        # Test unauthorized access (without token)
        local unauth_status
        unauth_status=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8081/api/v1/admin/users 2>/dev/null || echo "000")

        if [[ "$unauth_status" == "401" ]]; then
            add_test_result "unauthorized-access" "PASS" "Unauthorized access properly rejected (HTTP $unauth_status)"
        else
            add_test_result "unauthorized-access" "FAIL" "Unauthorized access not properly handled (HTTP $unauth_status)"
        fi
    else
        add_test_result "bootstrap-token-retrieved" "FAIL" "Could not retrieve bootstrap token"
    fi

    cleanup_port_forward
    trap - EXIT
}

# Test resource limits and constraints
test_resource_constraints() {
    log "Testing resource constraints..."

    # Check resource requests and limits
    local pods
    pods=$(kubectl get pods -n "$NAMESPACE" -l "app=$SERVICE_NAME" -o name 2>/dev/null)

    if [[ -n "$pods" ]]; then
        for pod in $pods; do
            local pod_name="${pod##*/}"

            # Check CPU requests
            local cpu_requests
            cpu_requests=$(kubectl get "$pod" -n "$NAMESPACE" -o jsonpath='{.spec.containers[0].resources.requests.cpu}' 2>/dev/null || echo "")

            if [[ -n "$cpu_requests" ]]; then
                add_test_result "cpu-requests-$pod_name" "PASS" "CPU requests set: $cpu_requests"
            else
                add_test_result "cpu-requests-$pod_name" "FAIL" "CPU requests not set"
            fi

            # Check memory requests
            local memory_requests
            memory_requests=$(kubectl get "$pod" -n "$NAMESPACE" -o jsonpath='{.spec.containers[0].resources.requests.memory}' 2>/dev/null || echo "")

            if [[ -n "$memory_requests" ]]; then
                add_test_result "memory-requests-$pod_name" "PASS" "Memory requests set: $memory_requests"
            else
                add_test_result "memory-requests-$pod_name" "FAIL" "Memory requests not set"
            fi
        done
    fi

    # Check security context
    local security_context
    security_context=$(kubectl get deployment "$SERVICE_NAME" -n "$NAMESPACE" -o jsonpath='{.spec.template.spec.securityContext.runAsNonRoot}' 2>/dev/null || echo "")

    if [[ "$security_context" == "true" ]]; then
        add_test_result "security-context" "PASS" "Running as non-root user"
    else
        add_test_result "security-context" "FAIL" "Not configured to run as non-root"
    fi
}

# Test monitoring setup
test_monitoring_setup() {
    log "Testing monitoring setup..."

    # Check ServiceMonitor
    if kubectl get servicemonitor rbac-token-service -n "$NAMESPACE" &> /dev/null; then
        add_test_result "servicemonitor" "PASS" "ServiceMonitor exists"
    else
        add_test_result "servicemonitor" "FAIL" "ServiceMonitor does not exist"
    fi

    # Check PrometheusRule
    if kubectl get prometheusrule rbac-token-service-alerts -n "$NAMESPACE" &> /dev/null; then
        add_test_result "prometheus-rules" "PASS" "PrometheusRule exists"
    else
        add_test_result "prometheus-rules" "FAIL" "PrometheusRule does not exist"
    fi

    # Check if metrics are being scraped (if Prometheus is available)
    local monitoring_ns="monitoring"
    if kubectl get namespace "$monitoring_ns" &> /dev/null; then
        info "Monitoring namespace found, checking Prometheus setup"
        add_test_result "monitoring-namespace" "PASS" "Monitoring namespace exists"
    else
        add_test_result "monitoring-namespace" "FAIL" "Monitoring namespace not found"
    fi
}

# Test backup and recovery readiness
test_backup_readiness() {
    log "Testing backup and recovery readiness..."

    # Check if volumes are properly mounted
    local pods
    pods=$(kubectl get pods -n "$NAMESPACE" -l "app=$SERVICE_NAME" -o name 2>/dev/null)

    if [[ -n "$pods" ]]; then
        for pod in $pods; do
            local pod_name="${pod##*/}"

            # Check audit volume mount
            local audit_mount
            audit_mount=$(kubectl get "$pod" -n "$NAMESPACE" -o jsonpath='{.spec.containers[0].volumeMounts[?(@.name=="audit-logs")].mountPath}' 2>/dev/null || echo "")

            if [[ "$audit_mount" == "/app/audit" ]]; then
                add_test_result "audit-volume-mount-$pod_name" "PASS" "Audit volume properly mounted"
            else
                add_test_result "audit-volume-mount-$pod_name" "FAIL" "Audit volume not properly mounted"
            fi

            # Check keys volume mount
            local keys_mount
            keys_mount=$(kubectl get "$pod" -n "$NAMESPACE" -o jsonpath='{.spec.containers[0].volumeMounts[?(@.name=="rbac-keys")].mountPath}' 2>/dev/null || echo "")

            if [[ "$keys_mount" == "/app/keys" ]]; then
                add_test_result "keys-volume-mount-$pod_name" "PASS" "Keys volume properly mounted"
            else
                add_test_result "keys-volume-mount-$pod_name" "FAIL" "Keys volume not properly mounted"
            fi
        done
    fi
}

# Generate test report
generate_test_report() {
    echo
    echo "==============================================="
    echo "        RBAC STAGING DEPLOYMENT TEST REPORT"
    echo "==============================================="
    echo "Timestamp: $(date)"
    echo "Namespace: $NAMESPACE"
    echo "Service: $SERVICE_NAME"
    echo "Test Duration: ${SECONDS}s"
    echo

    printf "%-40s %-8s %s\n" "TEST NAME" "RESULT" "MESSAGE"
    echo "--------------------------------------------------------------------------------------------------------"

    for result in "${TEST_RESULTS[@]}"; do
        local test_name="${result%%:*}"
        local rest="${result#*:}"
        local test_result="${rest%%:*}"
        local message="${rest#*:}"

        if [[ "$test_result" == "PASS" ]]; then
            printf "%-40s ${GREEN}%-8s${NC} %s\n" "$test_name" "$test_result" "$message"
        else
            printf "%-40s ${RED}%-8s${NC} %s\n" "$test_name" "$test_result" "$message"
        fi
    done

    echo "--------------------------------------------------------------------------------------------------------"
    printf "SUMMARY: ${GREEN}%d PASSED${NC}, ${RED}%d FAILED${NC}, %d TOTAL\n" "$PASSED_TESTS" "$FAILED_TESTS" $((PASSED_TESTS + FAILED_TESTS))
    echo

    if [[ $FAILED_TESTS -eq 0 ]]; then
        echo -e "${GREEN}🎉 ALL TESTS PASSED! RBAC Token Service is ready for staging use.${NC}"
        echo
        echo "Next steps:"
        echo "1. Verify monitoring alerts are configured in your AlertManager"
        echo "2. Update your CI/CD pipeline to use the new RBAC endpoints"
        echo "3. Conduct user acceptance testing"
        echo "4. Schedule production deployment"
        return 0
    else
        echo -e "${RED}❌ SOME TESTS FAILED. Please address the issues before using in staging.${NC}"
        echo
        echo "Common troubleshooting steps:"
        echo "1. Check pod logs: kubectl logs -n $NAMESPACE -l app=$SERVICE_NAME"
        echo "2. Describe problematic resources: kubectl describe <resource> <name> -n $NAMESPACE"
        echo "3. Check cluster events: kubectl get events -n $NAMESPACE --sort-by='.lastTimestamp'"
        echo "4. Verify resource quotas and limits"
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
    echo "  -t, --timeout TIMEOUT        Test timeout in seconds (default: 300)"
    echo "  -h, --help                   Show this help message"
    echo
    echo "Environment variables:"
    echo "  NAMESPACE     Override default namespace"
    echo "  SERVICE_NAME  Override default service name"
    echo "  TIMEOUT       Override default timeout"
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

    log "Starting comprehensive RBAC Token Service staging deployment test..."
    log "Namespace: $NAMESPACE"
    log "Service: $SERVICE_NAME"
    log "Timeout: ${TIMEOUT}s"
    echo

    # Run all test suites
    check_prerequisites || exit 1
    test_deployment_status
    test_pods_status
    test_service_networking
    test_configuration
    test_persistent_volumes
    test_http_endpoints
    test_rbac_functionality
    test_resource_constraints
    test_monitoring_setup
    test_backup_readiness

    # Generate final report
    generate_test_report
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
