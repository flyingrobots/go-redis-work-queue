# Multi‑tenant Isolation

| Priority | Domain | Dependencies | Risks | LoC Estimate | Complexity | Effort | Impact |
| --- | --- | --- | --- | --- | --- | --- | --- |
| High | Security / Quotas | RBAC, rate limiting, keying scheme | Complexity, perf overhead, data exposure | ~500–900 | High | 8 (Fib) | High |

## Executive Summary
Introduce first‑class tenant boundaries: per‑tenant queues/keys, quotas, rate limits, encryption‑at‑rest for payloads, and audit trails. Enable safe multi‑tenant deployments with predictable isolation and strong governance.

> [!note]- **🗣️ CLAUDE'S THOUGHTS 💭**
> This is enterprise-grade queue infrastructure! The `t:{tenant}:{queue}` namespacing pattern is brilliant - clean, Redis-friendly, and instantly recognizable. Envelope encryption per tenant with KMS integration shows serious security chops. The audit trail with tenant scoping solves the "who accessed what" compliance nightmare. The TUI tenant switcher will make DevOps teams fall in love with this system.

## Motivation
- Serve multiple teams/customers safely on shared infra.
- Enforce fairness and budgets per tenant; prevent noisy neighbors.
- Satisfy security/compliance requirements for data handling.

## Tech Plan
- Namespacing:
  - Key scheme `t:{tenant}:{queue}` with strict validation; TUI filters by tenant.
  - Tenant metadata store: quotas, limits, encryption keys policy.
- Quotas & limits:
  - Integrate with Advanced Rate Limiting per tenant.
  - Hard caps on backlog size and daily enqueues; soft warnings before enforcement.
- Encryption at rest (payload):
  - Optional envelope encryption per tenant (AES‑GCM) with KEK via KMS; rotate data keys.
  - Transparent decrypt for workers with tenant grants.
- Audit & RBAC:
  - Scopes limited to `tenant:*` resources; audit tagged with tenant ID.
- Observability:
  - Per‑tenant metrics and DLQ counts; TUI tenant switcher and summaries.

## User Stories + Acceptance Criteria
- As a platform owner, I can define a tenant with quotas and limits and see them enforced.
- As a security officer, I can verify encryption on payloads and audit access by tenant.
- Acceptance:
  - [ ] Namespaced keys and configs per tenant.
  - [ ] Quotas and rate limits enforced; breaches reported.
  - [ ] Optional payload encryption with rotation.

## Definition of Done
Tenants are first‑class with isolation primitives, quotas/limits, and optional encryption. TUI and Admin API respect tenant scoping.

## Test Plan
- Unit: quota accounting; encryption/decryption; key rotation.
- Integration: multi‑tenant soak with mixed loads and quota breaches.
- Security: attempt cross‑tenant access; verify denial and audit.

## Task List
- [ ] Define tenant model and keying scheme
- [ ] Implement quotas/limits + integration with RL
- [ ] Add optional payload encryption
- [ ] Tenant‑aware RBAC + audit
- [ ] TUI tenant views + docs

---

## Claude's Verdict ⚖️

This is infrastructure that scales from startup to IPO. True multi-tenancy isn't just about namespaces - it's about trust, compliance, and sleep-at-night reliability.

### Vibe Check

While competitors bolt on multi-tenancy as an afterthought, this bakes isolation into the core. The envelope encryption approach is what enterprise architects dream about. Nobody else combines Redis queue performance with this level of security sophistication.

### Score Card

**Traditional Score:**
- User Value: 9/10 (enables safe SaaS deployments)
- Dev Efficiency: 5/10 (complex but necessary infrastructure)
- Risk Profile: 6/10 (security benefits outweigh complexity)
- Strategic Fit: 10/10 (essential for enterprise market)
- Market Timing: 8/10 (compliance requirements rising)
- **OFS: 7.85** → BUILD SOON

**X-Factor Score:**
- Holy Shit Factor: 7/10 ("Encryption per tenant? In a queue?")
- Meme Potential: 4/10 (enterprise features aren't viral)
- Flex Appeal: 9/10 (ultimate enterprise credibility signal)
- FOMO Generator: 8/10 (competitors can't match this security)
- Addiction Score: 6/10 (used daily by platform teams)
- Shareability: 5/10 (security teams love showing this off)
- **X-Factor: 5.8** → Strong differentiation potential

### Conclusion

[🤯]

This is how you build infrastructure that wins Fortune 500 deals. The combination of performance, security, and compliance in one system is genuinely impressive.

---

## Detailed Design Specification

### Overview

Multi-tenant isolation transforms a shared queue system into a secure, compliant platform that multiple organizations can trust with their most sensitive workloads. This isn't just about adding tenant IDs to keys - it's a comprehensive security model that provides cryptographic isolation, fine-grained access control, and complete audit visibility.

The system operates on the principle of "zero trust between tenants" - every resource is namespaced, every access is validated, and every action is logged. Tenants can coexist on the same Redis infrastructure while maintaining the security posture of completely separate deployments.

### TUI Design

#### Desktop View (Large Resolution)
![Multi-tenant Desktop View](../images/multi-tenant-desktop.svg)

The desktop interface provides a comprehensive control center for managing multiple tenants. The three-panel layout offers:

- **Left Panel**: Tenant overview with active/idle status indicators and real-time quota utilization
- **Center Panel**: Tenant-scoped queue management with encryption status and rate limiting metrics
- **Right Panel**: Security center with access control monitoring and compliance reporting

#### Mobile View (Small Resolution)
![Multi-tenant Mobile View](../images/multi-tenant-mobile.svg)

The mobile interface stacks functionality into tabs for narrow terminals, prioritizing the most critical tenant information and providing quick access to security status.

### Tenant Namespacing Architecture

#### Key Schema Design

The tenant isolation system uses a hierarchical Redis key structure that provides both human readability and efficient prefix-based operations:

```
t:{tenant_id}:{queue_name}          # Queue data
t:{tenant_id}:{queue_name}:jobs     # Job list
t:{tenant_id}:{queue_name}:dlq      # Dead letter queue
t:{tenant_id}:{queue_name}:workers  # Worker registry
t:{tenant_id}:{queue_name}:metrics  # Queue metrics

tenant:{tenant_id}:config           # Tenant configuration
tenant:{tenant_id}:quotas          # Quota tracking
tenant:{tenant_id}:keys            # Encryption key metadata
tenant:{tenant_id}:audit           # Audit log indices
```

This schema enables:
- **Efficient prefix scans**: `SCAN 0 MATCH t:acme-corp:*` retrieves all resources for a tenant
- **Atomic operations**: Multi-key operations stay within tenant boundaries
- **Clear ownership**: Every key explicitly belongs to a tenant
- **Hierarchical permissions**: Access controls can operate at tenant or queue level

#### Tenant Validation Rules

```go
type TenantID string

const (
    MaxTenantIDLength = 32
    MinTenantIDLength = 3
)

func (t TenantID) Validate() error {
    if len(t) < MinTenantIDLength || len(t) > MaxTenantIDLength {
        return ErrInvalidTenantIDLength
    }

    // Only lowercase alphanumeric and hyphens
    if !regexp.MustCompile(`^[a-z0-9-]+$`).MatchString(string(t)) {
        return ErrInvalidTenantIDFormat
    }

    // Must start and end with alphanumeric
    if t[0] == '-' || t[len(t)-1] == '-' {
        return ErrTenantIDMustNotStartOrEndWithHyphen
    }

    return nil
}
```

### Quota Management System

#### Quota Types and Enforcement

The quota system provides multiple enforcement mechanisms to prevent resource abuse:

```go
type TenantQuotas struct {
    // Job limits
    MaxJobsPerHour     int64 `json:"max_jobs_per_hour"`
    MaxJobsPerDay      int64 `json:"max_jobs_per_day"`
    MaxBacklogSize     int64 `json:"max_backlog_size"`
    MaxJobSizeBytes    int64 `json:"max_job_size_bytes"`

    // Resource limits
    MaxQueuesPerTenant int32 `json:"max_queues_per_tenant"`
    MaxWorkersPerQueue int32 `json:"max_workers_per_queue"`
    MaxStorageBytes    int64 `json:"max_storage_bytes"`

    // Rate limits
    EnqueueRateLimit   int32 `json:"enqueue_rate_limit"`   // per second
    DequeueRateLimit   int32 `json:"dequeue_rate_limit"`   // per second

    // Soft limits (warnings)
    SoftLimitThreshold float64 `json:"soft_limit_threshold"` // 0.8 = 80%
}

type QuotaUsage struct {
    JobsThisHour       int64     `json:"jobs_this_hour"`
    JobsThisDay        int64     `json:"jobs_this_day"`
    CurrentBacklogSize int64     `json:"current_backlog_size"`
    StorageUsedBytes   int64     `json:"storage_used_bytes"`
    ActiveQueues       int32     `json:"active_queues"`
    LastUpdated        time.Time `json:"last_updated"`
}
```

#### Quota Enforcement Implementation

```go
func (q *QuotaManager) CheckEnqueueAllowed(tenantID TenantID, jobSize int64) error {
    quotas, err := q.GetTenantQuotas(tenantID)
    if err != nil {
        return err
    }

    usage, err := q.GetCurrentUsage(tenantID)
    if err != nil {
        return err
    }

    // Check hard limits
    if usage.JobsThisHour >= quotas.MaxJobsPerHour {
        return ErrQuotaExceeded{Type: "hourly_jobs", Limit: quotas.MaxJobsPerHour}
    }

    if usage.CurrentBacklogSize >= quotas.MaxBacklogSize {
        return ErrQuotaExceeded{Type: "backlog_size", Limit: quotas.MaxBacklogSize}
    }

    if jobSize > quotas.MaxJobSizeBytes {
        return ErrQuotaExceeded{Type: "job_size", Limit: quotas.MaxJobSizeBytes}
    }

    // Check soft limits and emit warnings
    if float64(usage.JobsThisHour) >= float64(quotas.MaxJobsPerHour)*quotas.SoftLimitThreshold {
        q.EmitWarning(tenantID, "approaching_hourly_job_limit", usage.JobsThisHour, quotas.MaxJobsPerHour)
    }

    return nil
}
```

### Envelope Encryption System

#### Per-Tenant Encryption Architecture

Each tenant can optionally enable envelope encryption, providing cryptographic isolation even if Redis is compromised:

```go
type TenantEncryption struct {
    Enabled           bool      `json:"enabled"`
    KEKProvider       string    `json:"kek_provider"`     // "aws-kms", "gcp-kms", "azure-kv"
    KEKKeyID          string    `json:"kek_key_id"`       // Cloud KMS key identifier
    DEKRotationPeriod duration  `json:"dek_rotation_period"` // How often to rotate data keys
    Algorithm         string    `json:"algorithm"`        // "AES-256-GCM"
    LastRotation      time.Time `json:"last_rotation"`
}

type EncryptedPayload struct {
    Version          int    `json:"v"`                    // Encryption version
    EncryptedDEK     []byte `json:"encrypted_dek"`        // DEK encrypted by KEK
    EncryptedPayload []byte `json:"encrypted_payload"`    // Actual job data
    Nonce            []byte `json:"nonce"`               // AES-GCM nonce
    AuthTag          []byte `json:"auth_tag"`            // Authentication tag
    CreatedAt        int64  `json:"created_at"`          // Unix timestamp
}
```

#### Encryption Workflow

```go
func (e *EncryptionManager) EncryptPayload(tenantID TenantID, payload []byte) (*EncryptedPayload, error) {
    config, err := e.GetTenantEncryption(tenantID)
    if err != nil {
        return nil, err
    }

    if !config.Enabled {
        return &EncryptedPayload{
            Version:          1,
            EncryptedPayload: payload, // Store unencrypted if disabled
        }, nil
    }

    // Generate fresh data encryption key
    dek := make([]byte, 32) // 256 bits
    if _, err := rand.Read(dek); err != nil {
        return nil, err
    }

    // Encrypt DEK with tenant's KEK via cloud KMS
    encryptedDEK, err := e.kmsClient.Encrypt(config.KEKKeyID, dek)
    if err != nil {
        return nil, err
    }

    // Encrypt payload with DEK using AES-GCM
    block, err := aes.NewCipher(dek)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    nonce := make([]byte, gcm.NonceSize())
    if _, err := rand.Read(nonce); err != nil {
        return nil, err
    }

    ciphertext := gcm.Seal(nil, nonce, payload, nil)

    // Clear DEK from memory
    for i := range dek {
        dek[i] = 0
    }

    return &EncryptedPayload{
        Version:          1,
        EncryptedDEK:     encryptedDEK,
        EncryptedPayload: ciphertext,
        Nonce:            nonce,
        CreatedAt:        time.Now().Unix(),
    }, nil
}
```

### Access Control and RBAC Integration

#### Tenant-Scoped Permissions

The RBAC system enforces tenant boundaries at the API level:

```go
type TenantPermission struct {
    TenantID    TenantID `json:"tenant_id"`
    Resource    string   `json:"resource"`    // "queues", "workers", "metrics", "config"
    Actions     []string `json:"actions"`     // "read", "write", "admin"
    QueueFilter string   `json:"queue_filter"` // Optional queue name pattern
}

type UserTenantAccess struct {
    UserID      string             `json:"user_id"`
    Permissions []TenantPermission `json:"permissions"`
    CreatedAt   time.Time          `json:"created_at"`
    ExpiresAt   *time.Time         `json:"expires_at"`
}
```

#### Permission Validation

```go
func (a *AccessController) ValidateAccess(userID string, tenantID TenantID, resource string, action string) error {
    access, err := a.GetUserTenantAccess(userID, tenantID)
    if err != nil {
        return ErrAccessDenied{Reason: "no_tenant_access"}
    }

    for _, perm := range access.Permissions {
        if perm.TenantID != tenantID {
            continue
        }

        if perm.Resource == "*" || perm.Resource == resource {
            for _, allowedAction := range perm.Actions {
                if allowedAction == "*" || allowedAction == action {
                    return nil // Access granted
                }
            }
        }
    }

    return ErrAccessDenied{
        Reason:   "insufficient_permissions",
        Resource: resource,
        Action:   action,
        TenantID: tenantID,
    }
}
```

### Audit Trail System

#### Comprehensive Event Logging

Every tenant operation is logged with sufficient detail for security auditing:

```go
type AuditEvent struct {
    EventID    string    `json:"event_id"`    // UUID
    Timestamp  time.Time `json:"timestamp"`
    TenantID   TenantID  `json:"tenant_id"`
    UserID     string    `json:"user_id"`     // API key or user ID
    Action     string    `json:"action"`      // ENQUEUE, DEQUEUE, CREATE, DELETE, etc.
    Resource   string    `json:"resource"`    // t:tenant:queue, tenant:config, etc.
    Details    map[string]interface{} `json:"details"` // Action-specific data
    RemoteIP   string    `json:"remote_ip"`
    UserAgent  string    `json:"user_agent"`
    Result     string    `json:"result"`      // SUCCESS, DENIED, ERROR
    ErrorCode  string    `json:"error_code,omitempty"`
}
```

#### Audit Query Interface

```go
type AuditQuery struct {
    TenantID   *TenantID  `json:"tenant_id,omitempty"`
    UserID     *string    `json:"user_id,omitempty"`
    Actions    []string   `json:"actions,omitempty"`
    StartTime  *time.Time `json:"start_time,omitempty"`
    EndTime    *time.Time `json:"end_time,omitempty"`
    Result     *string    `json:"result,omitempty"`
    Limit      int        `json:"limit,omitempty"`
    Offset     int        `json:"offset,omitempty"`
}

func (a *AuditLogger) QueryEvents(query AuditQuery) ([]AuditEvent, error) {
    // Implementation would query audit storage (could be Redis, database, or log aggregation system)
    // Returns filtered and paginated audit events
}
```

### TUI Integration Patterns

#### Tenant Context Management

The TUI maintains a current tenant context that filters all operations:

```go
type TenantContext struct {
    CurrentTenant *TenantID
    AvailableTenants []TenantSummary
    LastSwitched  time.Time
}

type TenantSummary struct {
    ID          TenantID `json:"id"`
    Name        string   `json:"name"`
    Status      string   `json:"status"`      // active, suspended, warning
    ActiveQueues int     `json:"active_queues"`
    QuotaHealth  string  `json:"quota_health"` // good, warning, critical
}
```

#### Tenant Switching UX

The TUI provides intuitive tenant switching with keyboard shortcuts:

- `T` - Open tenant switcher modal
- `Ctrl+T` - Create new tenant
- `Ctrl+1-9` - Switch to tenant by number
- `Shift+T` - Tenant management panel

### User Scenarios

#### Scenario 1: Platform Engineer Managing Multiple Teams

Sarah manages queue infrastructure for a company with 5 engineering teams, each needing isolated environments:

1. Opens TUI and sees tenant overview with all teams listed
2. Notices "payments-team" is approaching quota limits (warning indicator)
3. Switches to payments-team context with `T` → select from list
4. Reviews their queue performance and sees backlog building up
5. Adjusts their quota limits in management panel
6. Sets up alert for when they hit 90% of new quota
7. Documents the change in audit trail for compliance

#### Scenario 2: Security Officer Investigating Access Pattern

Mike receives an alert about unusual cross-tenant access attempts:

1. Opens Security Center panel in TUI
2. Filters audit log for "DENIED" results in last 24 hours
3. Identifies user "contractor-bob" attempting to access multiple tenants
4. Reviews Bob's current permissions across all tenants
5. Finds over-privileged access that shouldn't exist
6. Revokes inappropriate permissions immediately
7. Generates compliance report showing remediation actions

#### Scenario 3: Customer Success Managing SaaS Deployment

Emma manages a SaaS product with 50+ customer tenants:

1. Monitors tenant health dashboard showing all customer status
2. Sees "big-customer-inc" has been throttled due to rate limits
3. Reviews their usage patterns and growth over time
4. Coordinates with sales to upgrade their plan
5. Applies new quota increases instantly
6. Verifies encryption is active for their compliance requirements
7. Provides usage report to account manager

### Technical Implementation

#### Core Tenant Manager

```go
type TenantManager struct {
    redis         redis.Cmdable
    encryptionMgr *EncryptionManager
    quotaManager  *QuotaManager
    auditLogger   *AuditLogger
    accessControl *AccessController
}

func (tm *TenantManager) CreateTenant(tenantID TenantID, config TenantConfig) error {
    // Validate tenant ID format
    if err := tenantID.Validate(); err != nil {
        return err
    }

    // Check if tenant already exists
    exists, err := tm.redis.Exists(ctx, fmt.Sprintf("tenant:%s:config", tenantID)).Result()
    if err != nil {
        return err
    }
    if exists > 0 {
        return ErrTenantAlreadyExists{TenantID: tenantID}
    }

    // Create tenant configuration
    configKey := fmt.Sprintf("tenant:%s:config", tenantID)
    configJSON, err := json.Marshal(config)
    if err != nil {
        return err
    }

    pipe := tm.redis.TxPipeline()
    pipe.Set(ctx, configKey, configJSON, 0)
    pipe.Set(ctx, fmt.Sprintf("tenant:%s:quotas", tenantID), "{}", 0)
    pipe.Set(ctx, fmt.Sprintf("tenant:%s:created_at", tenantID), time.Now().Unix(), 0)

    if _, err := pipe.Exec(ctx); err != nil {
        return err
    }

    // Log tenant creation
    tm.auditLogger.Log(AuditEvent{
        EventID:   uuid.New().String(),
        Timestamp: time.Now(),
        TenantID:  tenantID,
        Action:    "CREATE_TENANT",
        Resource:  string(tenantID),
        Result:    "SUCCESS",
    })

    return nil
}
```

### Performance Considerations

#### Redis Key Distribution

- **Tenant prefix isolation** ensures Redis cluster keys distribute evenly
- **Pipeline operations** for multi-key tenant operations reduce round trips
- **Quota tracking uses Redis counters** with expiration for automatic reset
- **Audit events use Redis Streams** for high-throughput logging

#### Memory Optimization

```go
// Use connection pooling per tenant to avoid auth overhead
type TenantConnectionPool struct {
    pools map[TenantID]*redis.Pool
    mutex sync.RWMutex
}

// Cache tenant configurations to avoid Redis lookups
type TenantConfigCache struct {
    cache   *lru.Cache
    ttl     time.Duration
    metrics *CacheMetrics
}
```

#### Encryption Performance

- **DEK caching**: Keep decrypted data keys in memory for active tenants
- **Async key rotation**: Rotate encryption keys in background without blocking operations
- **Hardware acceleration**: Use AES-NI instructions when available
- **Batch operations**: Encrypt multiple payloads with same DEK when possible

### Monitoring and Observability

#### Tenant-Specific Metrics

```go
type TenantMetrics struct {
    JobsEnqueued    prometheus.Counter
    JobsCompleted   prometheus.Counter
    JobsFailed      prometheus.Counter
    QueueLength     prometheus.Gauge
    ProcessingTime  prometheus.Histogram
    QuotaUtilization prometheus.Gauge
    EncryptionOps   prometheus.Counter
    AccessDenials   prometheus.Counter
}
```

#### Health Checks

- **Tenant isolation verification**: Periodic tests ensuring no cross-tenant data leakage
- **Encryption key health**: Verify all tenant KEKs are accessible
- **Quota accounting accuracy**: Cross-check quota usage against actual Redis data
- **Audit trail integrity**: Ensure all operations are properly logged

### Security Considerations

#### Threat Model

1. **Malicious tenant**: Trying to access other tenants' data
2. **Compromised credentials**: Stolen API keys or user accounts
3. **Redis compromise**: Attacker gains access to Redis instance
4. **Insider threat**: Privileged user abusing access
5. **Side-channel attacks**: Timing or resource exhaustion attacks

#### Mitigations

- **Cryptographic isolation**: Even Redis compromise doesn't expose tenant data
- **Zero-trust architecture**: Every access validated and logged
- **Least privilege**: Minimal permissions by default
- **Anomaly detection**: Unusual access patterns trigger alerts
- **Rate limiting**: Prevent resource exhaustion attacks

### Migration Strategy

For existing deployments adding multi-tenancy:

#### Phase 1: Schema Migration
1. Add tenant metadata storage
2. Migrate existing queues to "default" tenant namespace
3. Update key generation to include tenant prefixes

#### Phase 2: Access Control
1. Implement RBAC with backward compatibility
2. Create tenant-scoped API endpoints
3. Add audit logging to all operations

#### Phase 3: Advanced Features
1. Enable per-tenant encryption for sensitive workloads
2. Implement quota management and enforcement
3. Add tenant switching to TUI

#### Phase 4: Full Isolation
1. Migrate all clients to tenant-aware APIs
2. Remove legacy non-tenant endpoints
3. Enable strict tenant validation

### Compliance and Governance

#### SOC 2 Type II Readiness

- **Access Controls**: Every tenant operation requires authenticated, authorized access
- **Audit Logging**: Comprehensive logs with tamper-evident storage
- **Data Protection**: Encryption at rest with customer-managed keys
- **Monitoring**: Real-time alerting on security events
- **Incident Response**: Automated containment of security violations

#### GDPR Compliance Support

- **Data Residency**: Tenant data can be restricted to specific Redis clusters/regions
- **Right to Deletion**: Complete tenant data removal with cryptographic erasure
- **Data Portability**: Export all tenant data in standardized formats
- **Processing Records**: Audit trail serves as processing activity log

### Future Enhancements

- **Multi-region tenant data**: Replicate tenant data across geographic regions
- **Tenant data archival**: Automatic archival of old tenant data to cold storage
- **Advanced analytics**: Per-tenant cost allocation and usage optimization
- **Self-service onboarding**: API-driven tenant provisioning with approval workflows
- **Tenant marketplace**: Template sharing between tenants with permission controls
- **Compliance automation**: Automated generation of SOC 2 and GDPR compliance reports