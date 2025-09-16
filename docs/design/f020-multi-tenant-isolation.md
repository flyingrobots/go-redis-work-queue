# Multi-Tenant Isolation - Architecture Design

## Executive Summary

This document describes the architecture for implementing comprehensive multi-tenant isolation in the Redis Work Queue system. The design provides cryptographic isolation, fine-grained access control, resource quotas, and complete audit visibility while maintaining high performance and operational simplicity.

## System Architecture

### Overview

The multi-tenant isolation system transforms a shared queue infrastructure into a secure, compliant platform that multiple organizations can trust with sensitive workloads. Each tenant operates in complete isolation with their own namespace, encryption keys, quotas, and access controls.

### Architecture Diagram

```mermaid
graph TB
    subgraph "Client Layer"
        C1[Tenant A Client]
        C2[Tenant B Client]
        C3[Admin Client]
    end

    subgraph "API Gateway"
        AG[Authentication/Authorization]
        RL[Rate Limiter]
        TM[Tenant Middleware]
    end

    subgraph "Core Services"
        QM[Queue Manager]
        EM[Encryption Manager]
        QUM[Quota Manager]
        AM[Access Manager]
        AL[Audit Logger]
    end

    subgraph "Storage Layer"
        subgraph "Redis Cluster"
            R1[Tenant A Namespace]
            R2[Tenant B Namespace]
            R3[System Metadata]
        end

        subgraph "KMS"
            K1[Tenant A KEK]
            K2[Tenant B KEK]
        end

        ADB[(Audit Database)]
    end

    C1 --> AG
    C2 --> AG
    C3 --> AG

    AG --> RL
    RL --> TM
    TM --> QM

    QM --> EM
    QM --> QUM
    QM --> AM
    QM --> AL

    EM --> K1
    EM --> K2
    QM --> R1
    QM --> R2
    AM --> R3
    AL --> ADB
```

### Component Architecture

```mermaid
classDiagram
    class TenantManager {
        +CreateTenant(tenantID, config)
        +UpdateTenant(tenantID, config)
        +DeleteTenant(tenantID)
        +GetTenant(tenantID)
        +ListTenants()
        -validateTenantID()
        -enforceQuotas()
    }

    class EncryptionManager {
        +EncryptPayload(tenantID, data)
        +DecryptPayload(tenantID, encrypted)
        +RotateKeys(tenantID)
        -generateDEK()
        -encryptDEK(kekID, dek)
        -getKEK(tenantID)
    }

    class QuotaManager {
        +CheckEnqueueAllowed(tenantID, size)
        +IncrementUsage(tenantID, metric)
        +GetUsage(tenantID)
        +ResetDailyQuotas()
        -enforceHardLimits()
        -emitWarnings()
    }

    class AccessController {
        +ValidateAccess(userID, tenantID, resource, action)
        +GrantAccess(userID, tenantID, permissions)
        +RevokeAccess(userID, tenantID)
        -checkPermissions()
        -auditAccess()
    }

    class AuditLogger {
        +Log(event)
        +Query(filters)
        +Export(tenantID, format)
        -storeEvent()
        -indexEvent()
    }

    class TenantContext {
        +tenantID: string
        +quotas: TenantQuotas
        +encryption: EncryptionConfig
        +permissions: []Permission
    }

    TenantManager --> EncryptionManager
    TenantManager --> QuotaManager
    TenantManager --> AccessController
    TenantManager --> AuditLogger
    TenantManager --> TenantContext
```

### Data Flow

```mermaid
sequenceDiagram
    participant Client
    participant API
    participant Auth
    participant TenantMgr
    participant QuotaMgr
    participant EncryptMgr
    participant Redis
    participant Audit

    Client->>API: Enqueue Job Request
    API->>Auth: Validate Token
    Auth->>API: User Context
    API->>TenantMgr: Get Tenant Context
    TenantMgr->>API: Tenant Config

    API->>QuotaMgr: Check Quota
    alt Quota Exceeded
        QuotaMgr->>API: Deny (Quota Exceeded)
        API->>Client: 429 Too Many Requests
        API->>Audit: Log Denial
    else Quota Available
        QuotaMgr->>API: Allow
        API->>EncryptMgr: Encrypt Payload
        EncryptMgr->>API: Encrypted Data
        API->>Redis: Store Job (t:tenant:queue)
        API->>QuotaMgr: Update Usage
        API->>Audit: Log Success
        API->>Client: 201 Created
    end
```

## Key Management System

### Namespace Design

```
# Queue Resources
t:{tenant_id}:{queue_name}           # Queue metadata
t:{tenant_id}:{queue_name}:jobs      # Job list
t:{tenant_id}:{queue_name}:dlq       # Dead letter queue
t:{tenant_id}:{queue_name}:workers   # Worker registry
t:{tenant_id}:{queue_name}:metrics   # Queue metrics
t:{tenant_id}:{queue_name}:locks     # Distributed locks

# Tenant Configuration
tenant:{tenant_id}:config            # Tenant configuration
tenant:{tenant_id}:quotas           # Quota tracking
tenant:{tenant_id}:keys             # Encryption key metadata
tenant:{tenant_id}:audit            # Audit log indices
tenant:{tenant_id}:users            # User access list

# System Resources
system:tenants                      # Tenant registry
system:metrics:{tenant_id}          # Per-tenant system metrics
system:quotas:hourly:{tenant_id}    # Hourly quota counters
system:quotas:daily:{tenant_id}     # Daily quota counters
```

### Tenant ID Validation

- Length: 3-32 characters
- Format: lowercase alphanumeric with hyphens
- Pattern: `^[a-z0-9][a-z0-9-]*[a-z0-9]$`
- Reserved: system, admin, default, test

## Security Model

### Encryption Architecture

```mermaid
graph LR
    subgraph "Key Hierarchy"
        MK[Master Key<br/>AWS KMS]
        KEK[Key Encryption Key<br/>Per Tenant]
        DEK[Data Encryption Key<br/>Per Job/Rotation]
    end

    subgraph "Data Protection"
        P[Plain Text Payload]
        EP[Encrypted Payload<br/>AES-256-GCM]
    end

    MK -->|Protects| KEK
    KEK -->|Protects| DEK
    DEK -->|Encrypts| P
    P -->|Becomes| EP
```

### Access Control Matrix

| Resource | Admin | Tenant Owner | Queue Manager | Worker | Viewer |
|----------|-------|--------------|---------------|--------|--------|
| Create Tenant | ✓ | - | - | - | - |
| Manage Quotas | ✓ | ✓ | - | - | - |
| Create Queue | ✓ | ✓ | ✓ | - | - |
| Enqueue Job | ✓ | ✓ | ✓ | ✓ | - |
| Dequeue Job | ✓ | ✓ | ✓ | ✓ | - |
| View Metrics | ✓ | ✓ | ✓ | ✓ | ✓ |
| View Audit Log | ✓ | ✓ | - | - | - |
| Rotate Keys | ✓ | ✓ | - | - | - |

### Threat Model

```mermaid
graph TB
    subgraph "Threats"
        T1[Cross-Tenant Access]
        T2[Quota Exhaustion]
        T3[Key Compromise]
        T4[Audit Tampering]
        T5[Side Channel]
    end

    subgraph "Mitigations"
        M1[Namespace Isolation]
        M2[Rate Limiting]
        M3[Envelope Encryption]
        M4[Immutable Logs]
        M5[Resource Limits]
    end

    T1 -->|Prevented by| M1
    T2 -->|Prevented by| M2
    T3 -->|Mitigated by| M3
    T4 -->|Prevented by| M4
    T5 -->|Mitigated by| M5
```

## Quota Management

### Quota Types

```yaml
tenant_quotas:
  # Job Limits
  max_jobs_per_hour: 10000
  max_jobs_per_day: 200000
  max_backlog_size: 50000
  max_job_size_bytes: 1048576  # 1MB

  # Resource Limits
  max_queues_per_tenant: 100
  max_workers_per_queue: 50
  max_storage_bytes: 10737418240  # 10GB

  # Rate Limits
  enqueue_rate_limit: 100  # per second
  dequeue_rate_limit: 200  # per second

  # Soft Limits (warnings at 80%)
  soft_limit_threshold: 0.8
```

### Quota Enforcement Flow

```mermaid
stateDiagram-v2
    [*] --> CheckQuota
    CheckQuota --> QuotaAvailable: Under Limit
    CheckQuota --> SoftLimit: 80-100%
    CheckQuota --> HardLimit: Exceeded

    QuotaAvailable --> ProcessRequest
    SoftLimit --> EmitWarning
    EmitWarning --> ProcessRequest
    HardLimit --> RejectRequest

    ProcessRequest --> UpdateUsage
    RejectRequest --> LogDenial

    UpdateUsage --> [*]
    LogDenial --> [*]
```

## Performance Requirements

### Latency Targets

| Operation | P50 | P95 | P99 | P99.9 |
|-----------|-----|-----|-----|-------|
| Tenant Validation | 1ms | 5ms | 10ms | 50ms |
| Quota Check | 2ms | 10ms | 20ms | 100ms |
| Encryption (1KB) | 5ms | 20ms | 50ms | 200ms |
| Audit Logging | 10ms | 50ms | 100ms | 500ms |
| Key Rotation | 100ms | 500ms | 1s | 5s |

### Throughput Requirements

- Tenant Operations: 1,000 ops/sec per tenant
- Quota Checks: 10,000 ops/sec globally
- Encryption: 5,000 ops/sec per tenant
- Audit Events: 50,000 events/sec globally

### Resource Scaling

```mermaid
graph LR
    subgraph "Small (1-10 tenants)"
        S1[1 Redis Instance]
        S2[Shared KMS]
        S3[Local Audit]
    end

    subgraph "Medium (10-100 tenants)"
        M1[Redis Cluster]
        M2[Dedicated KMS]
        M3[Audit Database]
    end

    subgraph "Large (100+ tenants)"
        L1[Sharded Redis]
        L2[Multi-Region KMS]
        L3[Distributed Audit]
    end

    S1 --> M1 --> L1
    S2 --> M2 --> L2
    S3 --> M3 --> L3
```

## Testing Strategy

### Unit Tests

```go
// Test tenant validation
func TestTenantIDValidation(t *testing.T) {
    validIDs := []string{"acme-corp", "tenant-123", "a1b2c3"}
    invalidIDs := []string{"-invalid", "invalid-", "UPPERCASE", "too_long_tenant_id_exceeding_32_chars"}

    for _, id := range validIDs {
        assert.NoError(t, TenantID(id).Validate())
    }

    for _, id := range invalidIDs {
        assert.Error(t, TenantID(id).Validate())
    }
}

// Test quota enforcement
func TestQuotaEnforcement(t *testing.T) {
    mgr := NewQuotaManager()
    tenant := TenantID("test-tenant")

    // Set quota
    mgr.SetQuota(tenant, TenantQuotas{
        MaxJobsPerHour: 100,
    })

    // Exhaust quota
    for i := 0; i < 100; i++ {
        err := mgr.CheckEnqueueAllowed(tenant, 1024)
        assert.NoError(t, err)
    }

    // Should be rejected
    err := mgr.CheckEnqueueAllowed(tenant, 1024)
    assert.ErrorIs(t, err, ErrQuotaExceeded)
}
```

### Integration Tests

```go
// Test cross-tenant isolation
func TestCrossTenantIsolation(t *testing.T) {
    tenantA := setupTenant("tenant-a")
    tenantB := setupTenant("tenant-b")

    // Create job in tenant A
    jobID := tenantA.EnqueueJob("test-job")

    // Try to access from tenant B
    job, err := tenantB.GetJob(jobID)
    assert.ErrorIs(t, err, ErrAccessDenied)
    assert.Nil(t, job)
}

// Test encryption end-to-end
func TestEncryption(t *testing.T) {
    tenant := setupTenantWithEncryption("encrypted-tenant")
    plaintext := []byte("sensitive data")

    // Enqueue encrypted job
    jobID := tenant.EnqueueJob(plaintext)

    // Verify Redis contains encrypted data
    raw := getRedisValue(jobID)
    assert.NotEqual(t, plaintext, raw)

    // Verify decryption works
    job := tenant.GetJob(jobID)
    assert.Equal(t, plaintext, job.Payload)
}
```

### Security Tests

- Penetration testing for cross-tenant access
- Fuzzing tenant ID validation
- Key rotation under load
- Quota exhaustion attacks
- Audit trail tampering attempts

### Performance Tests

```yaml
scenarios:
  - name: "Single Tenant Load"
    tenants: 1
    workers_per_tenant: 10
    enqueue_rate: 1000
    duration: 300s

  - name: "Multi-Tenant Fair Share"
    tenants: 10
    workers_per_tenant: 5
    enqueue_rate: 100
    duration: 600s

  - name: "Quota Enforcement"
    tenants: 5
    workers_per_tenant: 10
    enqueue_rate: 10000  # Exceed quota
    duration: 60s
    expected: quota_errors

  - name: "Encryption Overhead"
    tenants: 3
    encryption: enabled
    payload_size: 10KB
    enqueue_rate: 100
    measure: latency_impact
```

## Migration Plan

### Phase 1: Foundation (Week 1-2)
- Implement tenant ID validation
- Add namespace prefixing to keys
- Create tenant configuration storage

### Phase 2: Access Control (Week 3-4)
- Integrate with RBAC system
- Add tenant context to API
- Implement permission validation

### Phase 3: Quotas (Week 5-6)
- Build quota tracking system
- Add rate limiting per tenant
- Implement soft/hard limits

### Phase 4: Encryption (Week 7-8)
- Integrate with KMS
- Implement envelope encryption
- Add key rotation mechanism

### Phase 5: Audit & Monitoring (Week 9-10)
- Deploy audit logging
- Add tenant metrics
- Build compliance reports

## Compliance Considerations

### SOC 2 Type II
- Logical access controls per tenant
- Encryption of sensitive data
- Comprehensive audit logging
- Change management procedures

### GDPR
- Data isolation by tenant
- Right to erasure (crypto-shredding)
- Data portability via export
- Processing activity records

### HIPAA
- Encryption at rest and in transit
- Access controls and audit logs
- Business Associate Agreements
- Incident response procedures

## Deployment Architecture

```mermaid
graph TB
    subgraph "Production Environment"
        subgraph "Region 1"
            P1[Primary Redis]
            K1[KMS Instance]
            A1[Audit Store]
        end

        subgraph "Region 2"
            P2[Replica Redis]
            K2[KMS Instance]
            A2[Audit Replica]
        end
    end

    subgraph "DR Environment"
        DR1[Standby Redis]
        DRK[Backup KMS]
        DRA[Archive Audit]
    end

    P1 -.->|Replicate| P2
    P1 -.->|Backup| DR1
    A1 -.->|Sync| A2
    A2 -.->|Archive| DRA
```

## Monitoring & Observability

### Key Metrics

```yaml
tenant_metrics:
  - metric: tenant.jobs.enqueued
    type: counter
    labels: [tenant_id, queue_name]

  - metric: tenant.quota.usage
    type: gauge
    labels: [tenant_id, quota_type]

  - metric: tenant.encryption.operations
    type: counter
    labels: [tenant_id, operation]

  - metric: tenant.access.denied
    type: counter
    labels: [tenant_id, resource, reason]

  - metric: tenant.audit.events
    type: counter
    labels: [tenant_id, action, result]
```

### Alerting Rules

```yaml
alerts:
  - name: TenantQuotaWarning
    condition: tenant.quota.usage > 0.8 * quota_limit
    severity: warning

  - name: TenantQuotaExceeded
    condition: tenant.quota.usage >= quota_limit
    severity: critical

  - name: CrossTenantAccessAttempt
    condition: tenant.access.denied{reason="cross_tenant"} > 0
    severity: security

  - name: EncryptionKeyRotationDue
    condition: time() - tenant.key.last_rotation > 30d
    severity: warning
```

## Success Criteria

1. **Isolation**: Zero cross-tenant data leakage in 6 months of operation
2. **Performance**: <10ms overhead for tenant operations at P95
3. **Compliance**: Pass SOC 2 Type II audit
4. **Scalability**: Support 1000+ tenants on single cluster
5. **Reliability**: 99.99% availability for tenant operations

---

*Document Version: 1.0*
*Last Updated: 2025-09-14*
*Status: DESIGN PHASE*