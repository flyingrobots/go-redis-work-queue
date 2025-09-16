# Kubernetes Operator Examples

This directory contains example configurations for the Go Redis Work Queue Kubernetes Operator.

## Prerequisites

Before applying these examples, ensure you have:

1. A Kubernetes cluster with the operator installed
2. Redis instance accessible from the cluster
3. Proper RBAC permissions configured
4. Required secrets and ConfigMaps created

## Examples

### Basic Queue Configuration

**File**: `basic-queue.yaml`

A simple queue and worker pool configuration suitable for development and testing environments.

**Features**:
- Basic queue with rate limiting and DLQ
- Fixed replica count (2 workers)
- Standard resource requirements
- Basic security configuration

**Usage**:
```bash
kubectl apply -f basic-queue.yaml
```

### Autoscaling Production Setup

**File**: `autoscaling-queue.yaml`

A production-ready configuration with intelligent autoscaling, comprehensive security, and high availability.

**Features**:
- High-volume queue with Redis cluster support
- Intelligent autoscaling (2-50 replicas) based on:
  - Queue backlog per worker (target: 25 jobs/worker)
  - Latency SLO (p95 < 500ms)
- Production security hardening:
  - Pod Security Standards (restricted)
  - Non-root containers
  - Read-only root filesystem
  - Security context with dropped capabilities
- Advanced scheduling:
  - Node selectors for appropriate instance types
  - Pod anti-affinity for high availability
  - Tolerations for dedicated worker nodes
- Comprehensive monitoring with Prometheus metrics
- Global policies for circuit breaking and retry behavior

**Usage**:
```bash
# Create namespace
kubectl create namespace production

# Create required secrets (example)
kubectl create secret generic redis-cluster-connection \
  --from-literal=url="redis://redis-cluster.redis.svc.cluster.local:6379/1" \
  -n production

kubectl create secret generic redis-cluster-auth \
  --from-literal=password="your-redis-password" \
  -n production

# Apply configuration
kubectl apply -f autoscaling-queue.yaml
```

## Required RBAC

The operator requires the following RBAC permissions. Create a ServiceAccount and apply appropriate ClusterRole/Role bindings:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: queue-operator
  namespace: queue-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: queue-operator
rules:
  - apiGroups: ["queue.example.com"]
    resources: ["queues", "workerpools", "policies"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  - apiGroups: ["apps"]
    resources: ["deployments"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  - apiGroups: [""]
    resources: ["pods", "services", "secrets", "configmaps"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["create", "patch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: queue-operator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: queue-operator
subjects:
  - kind: ServiceAccount
    name: queue-operator
    namespace: queue-system
```

## Worker RBAC

Workers need minimal permissions to function:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: queue-worker
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: queue-worker
  namespace: default
rules:
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get"]
    resourceNames: ["redis-connection"]
  - apiGroups: ["queue.example.com"]
    resources: ["queues"]
    verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: queue-worker
  namespace: default
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: queue-worker
subjects:
  - kind: ServiceAccount
    name: queue-worker
    namespace: default
```

## Monitoring

The examples include Prometheus metrics annotations. Configure your Prometheus to scrape these endpoints:

```yaml
- job_name: 'queue-workers'
  kubernetes_sd_configs:
    - role: pod
  relabel_configs:
    - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
      action: keep
      regex: true
    - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
      action: replace
      target_label: __metrics_path__
      regex: (.+)
```

## Scaling Behavior

The autoscaling configuration uses the following logic:

1. **Backlog-based scaling**: Target 25 jobs per worker
   - If queue has 100 jobs and 2 workers → scale to 4 workers
   - Respects min/max replica limits

2. **Latency-based scaling**: Maintain p95 latency under 500ms
   - If latency exceeds target, scale up proportionally
   - Takes precedence over backlog-based scaling

3. **Cooldown periods**: Prevent rapid scaling oscillations
   - Scale-up cooldown: 2 minutes
   - Scale-down cooldown: 5 minutes

## Troubleshooting

### Common Issues

1. **Workers not starting**:
   ```bash
   kubectl describe workerpool autoscaling-workers -n production
   kubectl logs -l app.kubernetes.io/instance=autoscaling-workers -n production
   ```

2. **Autoscaling not working**:
   ```bash
   kubectl get workerpool autoscaling-workers -n production -o yaml
   # Check autoscaling status and conditions
   ```

3. **Queue not processing jobs**:
   ```bash
   kubectl describe queue high-volume-queue -n production
   # Check queue status and metrics
   ```

### Debug Commands

```bash
# Check operator logs
kubectl logs -l app.kubernetes.io/name=queue-operator -n queue-system

# View all queue resources
kubectl get queues,workerpools,policies --all-namespaces

# Check worker pod status
kubectl get pods -l app.kubernetes.io/component=worker

# View resource events
kubectl get events --sort-by=.metadata.creationTimestamp
```