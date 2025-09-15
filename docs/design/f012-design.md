# Kubernetes Operator (F012) - Design Document

**Version:** 1.0
**Date:** 2025-09-14
**Status:** Draft
**Author:** Claude (Worker 6)
**Reviewers:** TBD

## Executive Summary

The Kubernetes Operator transforms the go-redis-work-queue from a standalone application into a cloud-native, Kubernetes-first platform. This operator enables declarative queue and worker management through Custom Resource Definitions (CRDs), providing GitOps-friendly operations, intelligent autoscaling, and production-ready lifecycle management.

Built on the proven controller-runtime framework, the operator reconciles desired state from YAML manifests, automatically scales workers based on queue metrics and SLA targets, and orchestrates safe rolling deployments with zero downtime. This approach aligns with modern cloud-native practices while providing enterprise-grade operational capabilities.

The operator addresses the critical need for standardized, repeatable, and version-controlled infrastructure management in Kubernetes environments. By providing native Kubernetes resources for queue operations, teams can leverage existing GitOps workflows, RBAC controls, and monitoring infrastructure while gaining queue-aware autoscaling capabilities not available in generic solutions.

### Key Benefits

- **Kubernetes-Native Operations**: First-class CRDs integrate seamlessly with existing K8s tooling
- **Intelligent Autoscaling**: Queue-aware scaling based on backlog depth and latency SLOs
- **GitOps Ready**: Declarative configuration enables version-controlled infrastructure
- **Zero-Downtime Deployments**: Sophisticated drain and readiness hooks ensure safe updates
- **Enterprise Security**: Comprehensive RBAC, validation webhooks, and namespace isolation
- **Operational Excellence**: Rich observability with conditions, events, and Prometheus metrics

### Architecture Overview

```mermaid
graph TB
    subgraph "Kubernetes Control Plane"
        A[API Server]
        B[etcd]
        C[Scheduler]
    end

    subgraph "Operator Components"
        D[Queue Controller]
        E[WorkerPool Controller]
        F[Policy Controller]
        G[Validation Webhooks]
        H[Metrics Collector]
    end

    subgraph "Custom Resources"
        I[Queue CRDs]
        J[WorkerPool CRDs]
        K[Policy CRDs]
    end

    subgraph "Managed Resources"
        L[Deployments]
        M[ConfigMaps]
        N[Services]
        O[HPA/VPA]
        P[ServiceMonitor]
    end

    subgraph "Queue Infrastructure"
        Q[Redis Cluster]
        R[Worker Pods]
        S[Admin API]
        T[Metrics Endpoint]
    end

    subgraph "GitOps Pipeline"
        U[Git Repository]
        V[ArgoCD/Flux]
        W[CI/CD Pipeline]
    end

    A --> D
    A --> E
    A --> F
    A --> G

    I --> D
    J --> E
    K --> F

    D --> L
    D --> M
    E --> L
    E --> O
    F --> M

    L --> R
    M --> R
    R --> Q
    R --> S

    H --> T
    H --> P

    U --> V
    V --> I
    V --> J
    V --> K
    W --> U

    style D fill:#e1f5fe
    style E fill:#e1f5fe
    style F fill:#e1f5fe
    style G fill:#e1f5fe
    style H fill:#e1f5fe
    style I fill:#f3e5f5
    style J fill:#f3e5f5
    style K fill:#f3e5f5
    style L fill:#fff3e0
    style M fill:#fff3e0
    style N fill:#fff3e0
    style O fill:#fff3e0
    style P fill:#fff3e0
    style Q fill:#e8f5e8
    style R fill:#e8f5e8
    style S fill:#e8f5e8
    style T fill:#e8f5e8
    style U fill:#ffebee
    style V fill:#ffebee
    style W fill:#ffebee
```

## System Architecture

### Core Components

#### 1. Custom Resource Definitions (CRDs)

The operator defines three primary CRDs that represent the queue system's domain model in Kubernetes.

**Queue CRD**: Defines queue configuration, behavior, and policies.

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: queues.redis-queue.io
spec:
  group: redis-queue.io
  versions:
  - name: v1
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              name:
                type: string
                pattern: "^[a-z0-9-]+$"
              priorities:
                type: array
                items:
                  type: integer
                  minimum: 1
                  maximum: 10
              rateLimits:
                type: object
                properties:
                  maxJobsPerSecond:
                    type: integer
                    minimum: 1
                  burstCapacity:
                    type: integer
                    minimum: 1
              dlqPolicy:
                type: object
                properties:
                  enabled:
                    type: boolean
                  maxRetries:
                    type: integer
                    minimum: 0
                  retentionDays:
                    type: integer
                    minimum: 1
              retention:
                type: object
                properties:
                  completedJobs:
                    type: string
                    pattern: "^[0-9]+[hdm]$"
                  failedJobs:
                    type: string
                    pattern: "^[0-9]+[hdm]$"
```

**WorkerPool CRD**: Defines worker deployment configuration and scaling policies.

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: workerpools.redis-queue.io
spec:
  group: redis-queue.io
  versions:
  - name: v1
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              image:
                type: string
              version:
                type: string
              queues:
                type: array
                items:
                  type: string
              concurrency:
                type: integer
                minimum: 1
                maximum: 1000
              resources:
                type: object
                properties:
                  requests:
                    type: object
                  limits:
                    type: object
              autoscaling:
                type: object
                properties:
                  enabled:
                    type: boolean
                  minReplicas:
                    type: integer
                    minimum: 0
                  maxReplicas:
                    type: integer
                    minimum: 1
                  targetMetrics:
                    type: object
                    properties:
                      queueDepth:
                        type: integer
                      latencyP95:
                        type: string
                        pattern: "^[0-9]+[ms]$"
              drainPolicy:
                type: object
                properties:
                  gracePeriod:
                    type: string
                    pattern: "^[0-9]+[smh]$"
                  maxDrainTime:
                    type: string
                    pattern: "^[0-9]+[smh]$"
```

#### 2. Controller Architecture

The operator implements multiple controllers using the controller-runtime framework, each responsible for a specific CRD.

```go
type QueueController struct {
    client.Client
    Scheme    *runtime.Scheme
    Recorder  record.EventRecorder
    Redis     RedisClient
    AdminAPI  AdminAPIClient
}

func (r *QueueController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    queue := &redisqueuev1.Queue{}
    if err := r.Get(ctx, req.NamespacedName, queue); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // Apply queue configuration to Redis
    if err := r.reconcileQueueConfig(ctx, queue); err != nil {
        return ctrl.Result{}, err
    }

    // Update status
    return r.updateQueueStatus(ctx, queue)
}

type WorkerPoolController struct {
    client.Client
    Scheme   *runtime.Scheme
    Recorder record.EventRecorder
    Metrics  MetricsCollector
}

func (r *WorkerPoolController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    workerPool := &redisqueuev1.WorkerPool{}
    if err := r.Get(ctx, req.NamespacedName, workerPool); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // Reconcile Deployment
    if err := r.reconcileDeployment(ctx, workerPool); err != nil {
        return ctrl.Result{}, err
    }

    // Handle autoscaling
    if workerPool.Spec.Autoscaling.Enabled {
        if err := r.reconcileAutoscaling(ctx, workerPool); err != nil {
            return ctrl.Result{}, err
        }
    }

    return r.updateWorkerPoolStatus(ctx, workerPool)
}
```

#### 3. Intelligent Autoscaling Engine

The autoscaling engine monitors queue metrics and adjusts worker replicas based on configurable SLA targets.

```go
type AutoscalingEngine struct {
    metricsCollector MetricsCollector
    calculator       ScalingCalculator
    safetyLimits     SafetyLimits
}

type QueueMetrics struct {
    QueueDepth        int64         `json:"queue_depth"`
    ProcessingRate    float64       `json:"processing_rate"`
    LatencyP50        time.Duration `json:"latency_p50"`
    LatencyP95        time.Duration `json:"latency_p95"`
    LatencyP99        time.Duration `json:"latency_p99"`
    ErrorRate         float64       `json:"error_rate"`
    LastUpdated       time.Time     `json:"last_updated"`
}

type ScalingDecision struct {
    CurrentReplicas   int32     `json:"current_replicas"`
    DesiredReplicas   int32     `json:"desired_replicas"`
    ScalingReason     string    `json:"scaling_reason"`
    Confidence        float64   `json:"confidence"`
    CooldownUntil     time.Time `json:"cooldown_until"`
    SafetyLimitHit    bool      `json:"safety_limit_hit"`
}

func (e *AutoscalingEngine) CalculateDesiredReplicas(
    ctx context.Context,
    workerPool *redisqueuev1.WorkerPool,
    metrics *QueueMetrics,
) (*ScalingDecision, error) {
    currentReplicas := workerPool.Status.ReadyReplicas

    // Calculate target replicas based on queue theory
    targetReplicas := e.calculator.CalculateOptimalReplicas(
        metrics.QueueDepth,
        metrics.ProcessingRate,
        workerPool.Spec.Autoscaling.TargetMetrics,
    )

    // Apply safety limits and cooldown
    decision := e.applySafetyLimits(currentReplicas, targetReplicas, workerPool)

    return decision, nil
}
```

#### 4. Validation Webhooks

Admission webhooks provide validation, defaulting, and mutation capabilities for CRDs.

```go
type QueueValidator struct {
    Client  client.Client
    decoder *admission.Decoder
}

func (v *QueueValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
    queue := &redisqueuev1.Queue{}
    if err := v.decoder.Decode(req, queue); err != nil {
        return admission.Errored(http.StatusBadRequest, err)
    }

    // Validate queue name uniqueness
    if err := v.validateQueueName(ctx, queue); err != nil {
        return admission.Denied(err.Error())
    }

    // Validate rate limits
    if err := v.validateRateLimits(queue.Spec.RateLimits); err != nil {
        return admission.Denied(err.Error())
    }

    // Apply defaults
    v.applyDefaults(queue)

    return admission.Allowed("")
}
```

### Data Flow Architecture

```mermaid
sequenceDiagram
    participant U as User
    participant K as kubectl/GitOps
    participant A as API Server
    participant W as Webhook
    participant C as Controller
    participant D as Deployment
    participant P as Pod
    participant R as Redis

    Note over U,R: Resource Creation Flow
    U->>K: Apply WorkerPool YAML
    K->>A: Create WorkerPool resource
    A->>W: Validate resource
    W-->>A: Validation result
    A-->>K: Resource created

    Note over U,R: Reconciliation Flow
    C->>A: Watch WorkerPool changes
    A-->>C: WorkerPool event
    C->>C: Calculate desired state
    C->>A: Create/Update Deployment
    A->>D: Deployment reconciliation
    D->>P: Create worker pods
    P->>R: Connect to Redis

    Note over U,R: Autoscaling Flow
    C->>R: Collect queue metrics
    R-->>C: Metrics data
    C->>C: Calculate scaling decision
    C->>A: Update Deployment replicas
    A->>D: Scale deployment
    D->>P: Adjust pod count
```

### Performance Requirements

#### Latency Requirements

- **Reconciliation Latency**: <30 seconds for resource changes to take effect
- **Autoscaling Response**: <60 seconds from metric change to scaling action
- **Webhook Validation**: <100ms for admission webhook processing
- **Status Updates**: <10 seconds for status condition updates

#### Throughput Requirements

- **Resource Reconciliation**: 100+ resources reconciled per minute
- **Metric Collection**: 1,000+ metrics points collected per minute
- **Webhook Processing**: 500+ admission requests per minute
- **Event Generation**: 10,000+ Kubernetes events per hour

#### Resource Requirements

- **Controller Memory**: <500MB base, +10MB per 100 managed resources
- **Controller CPU**: <100m base, +50m during active reconciliation
- **Webhook Memory**: <100MB for validation webhook
- **Storage**: <1MB per managed resource for status and configuration

#### Scalability Targets

- **Managed Resources**: Support 1,000+ Queue and WorkerPool resources per cluster
- **Concurrent Reconciliations**: Handle 50+ simultaneous reconciliation loops
- **Multi-Tenancy**: Support 100+ namespaces with isolated resource management
- **Cluster Scale**: Function effectively in clusters with 1,000+ nodes

## Testing Strategy

### Unit Testing

- Controller reconciliation logic with fake Kubernetes clients
- Autoscaling calculator algorithms with various queue scenarios
- Validation webhook logic with valid and invalid resource definitions
- Metrics collection and processing with mock data sources
- Resource generation and template rendering

### Integration Testing

- End-to-end resource lifecycle testing with real Kubernetes API
- Controller integration with live Redis clusters
- Webhook integration with Kubernetes admission controllers
- Metrics collection integration with Prometheus endpoints
- RBAC and security policy validation

### End-to-End Testing

- Complete operator deployment in kind/minikube clusters
- GitOps workflow testing with ArgoCD integration
- Autoscaling behavior under simulated load patterns
- Rolling update scenarios with drain and readiness validation
- Multi-namespace isolation and security boundary testing

### Performance Testing

- Controller performance under high resource churn
- Autoscaling responsiveness with rapid queue depth changes
- Webhook latency under concurrent admission requests
- Memory usage growth patterns with increasing resource counts
- Reconciliation loop efficiency and resource utilization

## Security Threat Model

### Threat Analysis Matrix

#### T1: Privilege Escalation Through CRDs

**Description**: Attackers exploit CRD permissions to gain unauthorized access to cluster resources or sensitive queue data.

**STRIDE Categories**: Elevation of Privilege, Information Disclosure

**Attack Scenarios**:
- Malicious CRD definitions grant excessive permissions
- Webhook bypasses enable creation of privileged resources
- Controller service account compromise leads to cluster-wide access

**Mitigations**:
- **Principle of Least Privilege**: Minimal RBAC permissions for controllers and webhooks
- **Resource Quotas**: Namespace-level limits on resource creation and scaling
- **Admission Controls**: Comprehensive validation webhooks prevent privilege escalation
- **Regular Audits**: Automated scanning of RBAC permissions and resource configurations

**Risk Level**: High
**Likelihood**: Medium
**Impact**: Critical

#### T2: Resource Exhaustion Through Autoscaling

**Description**: Malicious or misconfigured autoscaling policies cause resource exhaustion and cluster instability.

**STRIDE Categories**: Denial of Service

**Attack Scenarios**:
- Autoscaling policies configured with no upper limits
- Malicious metrics injection triggers excessive scaling
- Cascading failures cause simultaneous scaling across multiple WorkerPools

**Mitigations**:
- **Hard Scaling Limits**: Operator-enforced maximum replica counts per namespace
- **Rate Limiting**: Cooling periods between scaling operations
- **Resource Quotas**: Kubernetes ResourceQuota enforcement
- **Circuit Breakers**: Automatic scaling suspension on repeated failures

**Risk Level**: Medium
**Likelihood**: High
**Impact**: Medium

#### T3: Configuration Injection Through GitOps

**Description**: Attackers inject malicious configurations through compromised Git repositories or CI/CD pipelines.

**STRIDE Categories**: Tampering, Elevation of Privilege

**Attack Scenarios**:
- Compromised Git repository introduces malicious CRD configurations
- CI/CD pipeline injection modifies resource definitions during deployment
- Branch protection bypass allows direct malicious commits

**Mitigations**:
- **Signature Verification**: Cryptographic signing of Git commits and container images
- **Policy as Code**: OPA Gatekeeper policies validate resource configurations
- **Multi-Stage Validation**: Separate validation in CI/CD and admission webhooks
- **Immutable Infrastructure**: GitOps-only configuration changes with audit trails

**Risk Level**: Medium
**Likelihood**: Low
**Impact**: High

### Security Controls Framework

#### Preventive Controls

**Access Control**:
- Role-based access control with minimal required permissions
- Namespace isolation with NetworkPolicies and PodSecurityPolicies
- Service account token rotation and limited scope
- Multi-factor authentication for GitOps repository access

**Resource Protection**:
- ResourceQuota enforcement for compute and storage limits
- PodSecurityStandards compliance (restricted profile)
- Network segmentation between operator and managed workloads
- Container image scanning and signature verification

#### Detective Controls

**Monitoring and Alerting**:
- Real-time monitoring of controller health and performance
- Anomaly detection for unusual scaling patterns
- Security event correlation and SIEM integration
- Resource usage monitoring with threshold-based alerting

**Audit and Compliance**:
- Comprehensive audit logging of all operator actions
- GitOps audit trail for configuration changes
- Regular security scanning of operator container images
- Compliance reporting for regulatory requirements

#### Responsive Controls

**Incident Response**:
- Automated circuit breakers for runaway autoscaling
- Emergency operator shutdown procedures
- Rapid rollback capabilities for configuration changes
- Incident escalation and notification workflows

## Deployment Plan

### Phase 1: Core CRDs and Controllers (Weeks 1-2)
- Implement Queue and WorkerPool CRDs with basic validation
- Develop core reconciliation controllers for resource management
- Create deployment manifests and RBAC configurations
- Build basic validation webhooks for resource integrity

### Phase 2: Autoscaling and Metrics (Weeks 3-4)
- Implement intelligent autoscaling engine with SLA-based scaling
- Integrate with Prometheus for metrics collection and alerting
- Add Policy CRD for global configuration management
- Create comprehensive monitoring and observability features

### Phase 3: Production Hardening (Weeks 5-6)
- Implement advanced validation webhooks with security policies
- Add support for rolling updates with drain and readiness hooks
- Create comprehensive RBAC configurations and security boundaries
- Build disaster recovery and backup/restore capabilities

### Phase 4: Ecosystem Integration (Weeks 7-8)
- Create Helm charts and Kustomize bases for easy deployment
- Build ArgoCD/Flux integration examples and documentation
- Implement advanced features like multi-cluster support
- Conduct security review and penetration testing

---

This design document establishes the foundation for implementing the Kubernetes Operator as a cloud-native platform that transforms queue management into a declarative, GitOps-friendly experience. The focus on intelligent autoscaling, comprehensive security, and operational excellence ensures that the operator meets enterprise requirements while providing a superior developer experience.