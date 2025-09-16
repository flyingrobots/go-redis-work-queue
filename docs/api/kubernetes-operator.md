# Kubernetes Operator API Reference

## Overview

The go-redis-work-queue Kubernetes Operator provides Custom Resource Definitions (CRDs) to manage queues, worker pools, and policies in a Kubernetes environment with native autoscaling and observability.

## Custom Resources

### Queue

A Queue represents a job queue configuration managed by the operator.

```yaml
apiVersion: queue.example.com/v1
kind: Queue
metadata:
  name: high-priority-jobs
  namespace: production
spec:
  name: "high-priority-jobs"
  priority: "high"
  rateLimit:
    requestsPerSecond: 500
    burstCapacity: 1000
    enabled: true
  deadLetterQueue:
    enabled: true
    maxRetries: 5
    retryBackoff:
      initialDelay: "2s"
      maxDelay: "60s"
      multiplier: 2.0
  retention:
    completedJobs: "24h"
    failedJobs: "72h"
    maxJobs: 50000
  redis:
    addresses:
      - "redis-cluster:6379"
    database: 0
    passwordSecret:
      name: "redis-credentials"
      key: "password"
    tls:
      enabled: true
      insecureSkipVerify: false
      caSecret:
        name: "redis-ca"
        key: "ca.crt"
```

#### Queue Spec Fields

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `name` | string | Queue identifier (immutable) | ✓ |
| `priority` | string | Queue priority: "critical", "high", "medium", "low" | ✓ |
| `rateLimit` | object | Rate limiting configuration | ✗ |
| `deadLetterQueue` | object | Dead letter queue configuration | ✗ |
| `retention` | object | Job retention policy | ✗ |
| `redis` | object | Redis connection configuration | ✗ |

#### Queue Status Fields

| Field | Type | Description |
|-------|------|-------------|
| `phase` | string | Current phase: "pending", "active", "failed" |
| `observedGeneration` | int64 | Generation observed by controller |
| `conditions` | []Condition | Status conditions |
| `metrics` | object | Current queue metrics |

### WorkerPool

A WorkerPool manages a group of workers for processing queues with intelligent autoscaling.

```yaml
apiVersion: queue.example.com/v1
kind: WorkerPool
metadata:
  name: batch-processors
  namespace: production
spec:
  queueName: "batch-jobs"
  replicas: 5
  image: "myapp/worker:v1.2.3"
  command: ["/app/worker"]
  args: ["--queue", "batch-jobs", "--concurrency", "10"]
  env:
    - name: "LOG_LEVEL"
      value: "info"
    - name: "REDIS_URL"
      valueFrom:
        secretKeyRef:
          name: "redis-credentials"
          key: "url"
  resources:
    requests:
      cpu: "100m"
      memory: "128Mi"
    limits:
      cpu: "500m"
      memory: "512Mi"
  autoscaling:
    enabled: true
    minReplicas: 2
    maxReplicas: 20
    metrics:
      backlogThreshold: 100
      latencyTarget: "5s"
      scaleUpCooldown: "30s"
      scaleDownCooldown: "300s"
      scaleUpFactor: 2.0
      scaleDownFactor: 0.5
  podDisruptionBudget:
    enabled: true
    minAvailable: 2
  strategy:
    type: "RollingUpdate"
    rollingUpdate:
      maxUnavailable: 1
      maxSurge: 2
  gracefulShutdown:
    enabled: true
    timeout: "30s"
    drainTimeout: "60s"
```

#### WorkerPool Spec Fields

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `queueName` | string | Target queue name | ✓ |
| `replicas` | int32 | Desired number of replicas | ✓ |
| `image` | string | Container image | ✓ |
| `command` | []string | Container command | ✗ |
| `args` | []string | Container arguments | ✗ |
| `env` | []EnvVar | Environment variables | ✗ |
| `resources` | ResourceRequirements | Resource requirements | ✗ |
| `autoscaling` | object | Autoscaling configuration | ✗ |
| `podDisruptionBudget` | object | PDB configuration | ✗ |
| `strategy` | object | Deployment strategy | ✗ |
| `gracefulShutdown` | object | Graceful shutdown configuration | ✗ |

#### WorkerPool Status Fields

| Field | Type | Description |
|-------|------|-------------|
| `phase` | string | Current phase: "pending", "running", "scaling", "failed" |
| `observedGeneration` | int64 | Generation observed by controller |
| `replicas` | int32 | Current number of replicas |
| `readyReplicas` | int32 | Number of ready replicas |
| `updatedReplicas` | int32 | Number of updated replicas |
| `conditions` | []Condition | Status conditions |
| `autoscaling` | object | Autoscaling status |

### Policy

A Policy defines operational rules and constraints for queue and worker pool management.

```yaml
apiVersion: queue.example.com/v1
kind: Policy
metadata:
  name: production-policy
  namespace: production
spec:
  scope:
    namespaces: ["production", "staging"]
    queues: ["critical-*", "high-priority-*"]
    workerPools: ["*"]
  resourceLimits:
    maxQueuesPerNamespace: 50
    maxWorkerPoolsPerQueue: 5
    maxReplicasPerPool: 100
    totalMaxReplicas: 1000
  autoscaling:
    globalEnabled: true
    defaultMinReplicas: 1
    defaultMaxReplicas: 10
    maxScaleUpFactor: 3.0
    maxScaleDownFactor: 0.3
    minCooldownPeriod: "10s"
  security:
    allowedImages:
      - "myregistry.com/workers/*"
      - "trusted-registry.io/*"
    requiredSecurityContext:
      runAsNonRoot: true
      readOnlyRootFilesystem: true
    networkPolicies:
      enabled: true
      allowedEgress:
        - "redis-cluster"
        - "metrics-server"
  monitoring:
    metricsEnabled: true
    alerting:
      enabled: true
      slackWebhook:
        secretRef:
          name: "alert-config"
          key: "slack-webhook"
      criticalThresholds:
        errorRate: 0.05
        latency: "10s"
        backlogGrowthRate: 1000
```

#### Policy Spec Fields

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `scope` | object | Policy scope definition | ✓ |
| `resourceLimits` | object | Resource limit constraints | ✗ |
| `autoscaling` | object | Global autoscaling settings | ✗ |
| `security` | object | Security policies | ✗ |
| `monitoring` | object | Monitoring configuration | ✗ |

## API Operations

### Admin API Integration

The operator integrates with the queue system's Admin API for real-time management:

- **Queue Management**: Create, update, delete queues
- **Metrics Collection**: Real-time backlog, processing rate, error rate
- **Status Monitoring**: Queue health and operational state

### Autoscaling Algorithm

The WorkerPool controller implements intelligent autoscaling:

1. **Backlog-based scaling**: Scale up when backlog exceeds threshold
2. **Latency-based scaling**: Scale up when SLO latency is breached
3. **Predictive scaling**: Use historical data for proactive scaling
4. **Resource-aware**: Respect cluster resource constraints
5. **Graceful scaling**: Drain workers before scale-down

### Webhook Validation

Admission webhooks provide comprehensive validation:

- **Queue validation**: Name format, priority values, rate limits
- **WorkerPool validation**: Resource requirements, image security
- **Policy validation**: Scope conflicts, limit consistency
- **Immutable fields**: Prevent dangerous configuration changes

## Installation

### Prerequisites

- Kubernetes 1.20+
- Cert-manager (for webhook certificates)
- Redis cluster
- Prometheus (for metrics)

### Deployment

```bash
# Install CRDs
kubectl apply -f config/crd/bases/

# Install operator
kubectl apply -f config/manager/

# Configure RBAC
kubectl apply -f config/rbac/
```

### Configuration

The operator requires the following environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `ADMIN_API_URL` | Queue system Admin API endpoint | Required |
| `REDIS_URL` | Redis connection string | Required |
| `METRICS_ADDR` | Metrics server address | `:8080` |
| `WEBHOOK_PORT` | Webhook server port | `9443` |
| `LOG_LEVEL` | Logging level | `info` |

## Examples

### High-Throughput Queue with Autoscaling

```yaml
# High-performance queue
apiVersion: queue.example.com/v1
kind: Queue
metadata:
  name: high-throughput
spec:
  name: "high-throughput"
  priority: "high"
  rateLimit:
    requestsPerSecond: 10000
    burstCapacity: 20000
    enabled: true
---
# Auto-scaling worker pool
apiVersion: queue.example.com/v1
kind: WorkerPool
metadata:
  name: high-throughput-workers
spec:
  queueName: "high-throughput"
  replicas: 10
  image: "myapp/worker:latest"
  autoscaling:
    enabled: true
    minReplicas: 5
    maxReplicas: 50
    metrics:
      backlogThreshold: 500
      latencyTarget: "2s"
```

### Multi-Tenant Configuration

```yaml
# Tenant A policy
apiVersion: queue.example.com/v1
kind: Policy
metadata:
  name: tenant-a-policy
spec:
  scope:
    namespaces: ["tenant-a"]
  resourceLimits:
    maxQueuesPerNamespace: 10
    maxWorkerPoolsPerQueue: 3
    totalMaxReplicas: 100
  security:
    allowedImages:
      - "tenant-a-registry.com/*"
```

## Monitoring and Observability

### Prometheus Metrics

The operator exposes comprehensive metrics:

- `queue_operator_queues_total`: Total number of managed queues
- `queue_operator_workerpools_total`: Total number of managed worker pools
- `queue_operator_autoscaling_events_total`: Autoscaling events counter
- `queue_operator_reconcile_duration_seconds`: Controller reconciliation duration

### Health Checks

- `/healthz`: Overall operator health
- `/readyz`: Readiness probe for webhooks
- `/metrics`: Prometheus metrics endpoint

### Logging

Structured logging with configurable levels:
- Error: Critical failures
- Warn: Recoverable issues
- Info: Normal operations
- Debug: Detailed troubleshooting

## Troubleshooting

### Common Issues

1. **Queue Creation Fails**
   - Check Admin API connectivity
   - Verify Redis credentials
   - Review webhook validation errors

2. **Autoscaling Not Working**
   - Confirm metrics collection
   - Check resource quotas
   - Verify SLO thresholds

3. **Worker Pods Failing**
   - Review container logs
   - Check resource limits
   - Verify image pull secrets

### Debug Commands

```bash
# Check operator logs
kubectl logs -n queue-system deployment/queue-operator

# Describe queue status
kubectl describe queue my-queue

# Check worker pool events
kubectl get events --field-selector involvedObject.kind=WorkerPool

# View autoscaling metrics
kubectl top pods -l app=worker-pool
```