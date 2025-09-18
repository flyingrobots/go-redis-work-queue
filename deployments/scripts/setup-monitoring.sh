#!/bin/bash
set -euo pipefail

# RBAC Token Service Monitoring Setup Script
# This script sets up monitoring and alerting for the RBAC token service

NAMESPACE="${NAMESPACE:-work-queue}"
MONITORING_NAMESPACE="${MONITORING_NAMESPACE:-monitoring}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

source "${SCRIPT_DIR}/lib/logging.sh"

# Alert routing feature flag / configuration.
ENABLE_ALERTMANAGER_SMTP="${ENABLE_ALERTMANAGER_SMTP:-false}"

if [[ "$ENABLE_ALERTMANAGER_SMTP" == "true" ]]; then
  : "${SMTP_HOST:?ENABLE_ALERTMANAGER_SMTP=true but SMTP_HOST not set (e.g. smtp.example.com)}"
  : "${SMTP_PORT:?ENABLE_ALERTMANAGER_SMTP=true but SMTP_PORT not set (e.g. 587)}"
  : "${SMTP_FROM_ADDRESS:?ENABLE_ALERTMANAGER_SMTP=true but SMTP_FROM_ADDRESS not set (e.g. alertmanager@example.com)}"
  : "${ALERT_CRITICAL_EMAILS:?ENABLE_ALERTMANAGER_SMTP=true but ALERT_CRITICAL_EMAILS not set (comma-separated list)}"
  : "${ALERT_WARNING_EMAILS:?ENABLE_ALERTMANAGER_SMTP=true but ALERT_WARNING_EMAILS not set (comma-separated list)}"
else
  SMTP_HOST="${SMTP_HOST:-}"
  SMTP_PORT="${SMTP_PORT:-}"
  SMTP_FROM_ADDRESS="${SMTP_FROM_ADDRESS:-}"
  ALERT_CRITICAL_EMAILS="${ALERT_CRITICAL_EMAILS:-}"
  ALERT_WARNING_EMAILS="${ALERT_WARNING_EMAILS:-}"
fi

ALERTMANAGER_WEBHOOK_URL="${ALERTMANAGER_WEBHOOK_URL:-http://localhost:9093/webhook}"

render_email_block() {
    local csv="$1"
    local subject="$2"
    local block=""
    IFS=',' read -r -a recipients <<< "$csv"
    for recipient in "${recipients[@]}"; do
        local trimmed
        trimmed=$(echo "$recipient" | xargs)
        block+="  - to: '${trimmed}'\\n"
        block+="    subject: '${subject}'\\n"
        block+="    body: |\\n"
        block+="      Alert: {{ .CommonAnnotations.summary }}\\n"
        block+="      Description: {{ .CommonAnnotations.description }}\\n"
        block+="      Severity: {{ .CommonLabels.severity }}\\n"
        block+="      Time: {{ .CommonAnnotations.timestamp }}\\n"
    done
    printf '%b' "$block"
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

        local alertmanager_config
        if [[ "$ENABLE_ALERTMANAGER_SMTP" == "true" ]]; then
            info "SMTP host: ${SMTP_HOST}:${SMTP_PORT} (from: ${SMTP_FROM_ADDRESS})"
            info "Critical alert recipients: ${ALERT_CRITICAL_EMAILS}"
            info "Warning alert recipients: ${ALERT_WARNING_EMAILS}"

            local critical_email_configs warning_email_configs
            critical_email_configs=$(render_email_block "$ALERT_CRITICAL_EMAILS" "CRITICAL: RBAC Token Service Alert")
            warning_email_configs=$(render_email_block "$ALERT_WARNING_EMAILS" "WARNING: RBAC Token Service Alert")
            alertmanager_config=$(cat <<EOF
global:
  smtp_smarthost: '${SMTP_HOST}:${SMTP_PORT}'
  smtp_from: '${SMTP_FROM_ADDRESS}'

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
  - url: '${ALERTMANAGER_WEBHOOK_URL}'

- name: 'rbac-critical'
  email_configs:
${critical_email_configs}
- name: 'rbac-warning'
  email_configs:
${warning_email_configs}
EOF
            )
        else
            info "ENABLE_ALERTMANAGER_SMTP=false; configuring webhook-only AlertManager receiver"
            alertmanager_config=$(cat <<EOF
route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'default-receiver'

receivers:
- name: 'default-receiver'
  webhook_configs:
  - url: '${ALERTMANAGER_WEBHOOK_URL}'
EOF
            )
        fi

        kubectl create secret generic alertmanager-rbac-config \
            --namespace="$MONITORING_NAMESPACE" \
            --from-literal=alertmanager.yml="$alertmanager_config" \
            --dry-run=client -o yaml | kubectl apply -f -

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
    kill "$port_forward_pid" 2>/dev/null || true

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
