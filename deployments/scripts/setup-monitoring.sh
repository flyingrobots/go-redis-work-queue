#!/bin/bash
set -euo pipefail

# RBAC Token Service Monitoring Setup Script
# This script sets up monitoring and alerting for the RBAC token service

NAMESPACE="${NAMESPACE:-work-queue}"
MONITORING_NAMESPACE="${MONITORING_NAMESPACE:-monitoring}"

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

# Check prerequisites
check_prerequisites() {
    log "Checking monitoring prerequisites..."

    if ! command -v kubectl &> /dev/null; then
        error "kubectl not found"
        exit 1
    fi

    # Check if we can connect to the cluster
    if ! kubectl cluster-info &> /dev/null; then
        error "Cannot connect to Kubernetes cluster"
        exit 1
    fi

    # Check if monitoring namespace exists
    if ! kubectl get namespace "$MONITORING_NAMESPACE" &> /dev/null; then
        warn "Monitoring namespace '$MONITORING_NAMESPACE' does not exist"
        info "Creating monitoring namespace..."
        kubectl create namespace "$MONITORING_NAMESPACE"
    fi

    # Check if Prometheus operator is installed
    if ! kubectl get crd prometheusrules.monitoring.coreos.com &> /dev/null; then
        warn "Prometheus operator not found. Please install Prometheus operator first."
        info "You can install it with:"
        info "  helm repo add prometheus-community https://prometheus-community.github.io/helm-charts"
        info "  helm install prometheus-operator prometheus-community/kube-prometheus-stack -n $MONITORING_NAMESPACE"
        exit 1
    fi

    log "Prerequisites check passed"
}

# Deploy ServiceMonitor
deploy_service_monitor() {
    log "Deploying ServiceMonitor for RBAC token service..."

    kubectl apply -f - <<EOF
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: rbac-token-service
  namespace: $NAMESPACE
  labels:
    app: rbac-token-service
    component: auth
spec:
  selector:
    matchLabels:
      app: rbac-token-service
  endpoints:
  - port: http
    path: /metrics
    interval: 30s
    scrapeTimeout: 10s
    honorLabels: true
EOF

    log "ServiceMonitor deployed"
}

# Deploy PrometheusRule
deploy_prometheus_rules() {
    log "Deploying Prometheus rules for RBAC token service..."

    # Get the current directory to find the monitoring YAML file
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"

    kubectl apply -f "$PROJECT_ROOT/deployments/kubernetes/rbac-monitoring.yaml"

    log "Prometheus rules deployed"
}

# Create Grafana dashboard
create_grafana_dashboard() {
    log "Creating Grafana dashboard..."

    # Check if Grafana is installed
    if kubectl get deployment grafana -n "$MONITORING_NAMESPACE" &> /dev/null; then
        # Dashboard is included in the rbac-monitoring.yaml file
        log "Grafana dashboard ConfigMap created"
        log "Dashboard will be automatically imported if Grafana sidecar is configured"
    else
        warn "Grafana not found in $MONITORING_NAMESPACE namespace"
        info "Dashboard configuration saved to ConfigMap for manual import"
    fi
}

# Configure alert routing
configure_alerting() {
    log "Configuring alert routing..."

    # Create AlertManager configuration if it doesn't exist
    if ! kubectl get secret alertmanager-main -n "$MONITORING_NAMESPACE" &> /dev/null; then
        warn "AlertManager configuration not found"
        info "Creating basic AlertManager configuration..."

        kubectl create secret generic alertmanager-rbac-config \
            --namespace="$MONITORING_NAMESPACE" \
            --from-literal=alertmanager.yml="$(cat <<'EOF'
global:
  smtp_smarthost: 'localhost:587'
  smtp_from: 'alertmanager@company.com'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'default-receiver'
  routes:
  - match:
      service: rbac-token-service
      severity: critical
    receiver: 'rbac-critical'
  - match:
      service: rbac-token-service
      severity: warning
    receiver: 'rbac-warning'

receivers:
- name: 'default-receiver'
  webhook_configs:
  - url: 'http://localhost:9093/webhook'

- name: 'rbac-critical'
  email_configs:
  - to: 'admin@company.com'
    subject: 'CRITICAL: RBAC Token Service Alert'
    body: |
      Alert: {{ .CommonAnnotations.summary }}
      Description: {{ .CommonAnnotations.description }}
      Severity: {{ .CommonLabels.severity }}
      Time: {{ .CommonAnnotations.timestamp }}

- name: 'rbac-warning'
  email_configs:
  - to: 'ops@company.com'
    subject: 'WARNING: RBAC Token Service Alert'
    body: |
      Alert: {{ .CommonAnnotations.summary }}
      Description: {{ .CommonAnnotations.description }}
      Severity: {{ .CommonLabels.severity }}
      Time: {{ .CommonAnnotations.timestamp }}
EOF
)" --dry-run=client -o yaml | kubectl apply -f -

    else
        info "AlertManager configuration already exists"
    fi

    log "Alert routing configured"
}

# Test monitoring setup
test_monitoring() {
    log "Testing monitoring setup..."

    # Check if ServiceMonitor is created
    if kubectl get servicemonitor rbac-token-service -n "$NAMESPACE" &> /dev/null; then
        log "ServiceMonitor is active"
    else
        error "ServiceMonitor not found"
        return 1
    fi

    # Check if PrometheusRule is created
    if kubectl get prometheusrule rbac-token-service-alerts -n "$NAMESPACE" &> /dev/null; then
        log "PrometheusRule is active"
    else
        error "PrometheusRule not found"
        return 1
    fi

    # Test if metrics endpoint is accessible
    log "Testing metrics endpoint accessibility..."

    # Port forward to test metrics
    kubectl port-forward service/rbac-token-service 8081:80 -n "$NAMESPACE" &
    local port_forward_pid=$!

    sleep 3

    if curl -f http://localhost:8081/metrics &> /dev/null; then
        log "Metrics endpoint is accessible"
    else
        warn "Metrics endpoint is not accessible"
    fi

    # Clean up port forward
    kill $port_forward_pid 2>/dev/null || true

    # Check if Prometheus is scraping the target
    if kubectl get pods -n "$MONITORING_NAMESPACE" -l app.kubernetes.io/name=prometheus &> /dev/null; then
        info "Prometheus is running and should be scraping RBAC metrics"
    else
        warn "Prometheus not found in $MONITORING_NAMESPACE"
    fi

    log "Monitoring test completed"
}

# Show monitoring status
show_monitoring_status() {
    log "Monitoring setup completed!"
    echo
    echo "=== Monitoring Status ==="
    echo "Namespace: $NAMESPACE"
    echo "Monitoring Namespace: $MONITORING_NAMESPACE"
    echo
    echo "Components:"
    echo "  ServiceMonitor: rbac-token-service"
    echo "  PrometheusRule: rbac-token-service-alerts"
    echo "  Grafana Dashboard: rbac-grafana-dashboard"
    echo
    echo "To view metrics:"
    echo "  kubectl port-forward service/rbac-token-service 8081:80 -n $NAMESPACE"
    echo "  curl http://localhost:8081/metrics"
    echo
    echo "To access Grafana (if installed):"
    echo "  kubectl port-forward service/grafana 3000:80 -n $MONITORING_NAMESPACE"
    echo "  Open http://localhost:3000"
    echo
    echo "To view Prometheus targets:"
    echo "  kubectl port-forward service/prometheus-operated 9090:9090 -n $MONITORING_NAMESPACE"
    echo "  Open http://localhost:9090/targets"
    echo
    echo "Alert rules configured for:"
    echo "  - Service availability"
    echo "  - Error rates and latency"
    echo "  - Token validation issues"
    echo "  - Security events"
    echo "  - Resource usage"
    echo "  - Storage space"
}

# Show usage
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo
    echo "Options:"
    echo "  -n, --namespace NAMESPACE           Target namespace (default: work-queue)"
    echo "  -m, --monitoring-ns NAMESPACE       Monitoring namespace (default: monitoring)"
    echo "  -h, --help                          Show this help message"
    echo
    echo "Environment variables:"
    echo "  NAMESPACE                Override default namespace"
    echo "  MONITORING_NAMESPACE     Override default monitoring namespace"
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -n|--namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            -m|--monitoring-ns)
                MONITORING_NAMESPACE="$2"
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

    log "Setting up monitoring for RBAC Token Service..."
    log "Target namespace: $NAMESPACE"
    log "Monitoring namespace: $MONITORING_NAMESPACE"

    check_prerequisites
    deploy_service_monitor
    deploy_prometheus_rules
    create_grafana_dashboard
    configure_alerting
    test_monitoring
    show_monitoring_status

    log "Monitoring setup completed successfully!"
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi