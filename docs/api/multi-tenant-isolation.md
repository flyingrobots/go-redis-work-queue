# Multi-Tenant Isolation API

The Multi-Tenant Isolation feature provides secure tenant boundaries with namespaced keys, quotas, rate limits, encryption-at-rest for payloads, and audit trails.

## Overview

This module enables safe multi-tenant deployments by:
- Providing strict tenant isolation through key namespacing
- Enforcing quotas and rate limits per tenant
- Optional payload encryption with key rotation
- Comprehensive audit logging
- Role-based access control (RBAC)

## Key Concepts

### Tenant ID Format
- Length: 3-32 characters
- Format: Lowercase alphanumeric characters and hyphens only
- Cannot start or end with hyphens
- Examples: `tenant-1`, `acme-corp`, `dev-environment`

### Key Namespacing
All tenant data is namespaced using the pattern `t:{tenant}:{resource}`:
- Queue: `t:tenant-1:queue-name`
- Jobs: `t:tenant-1:queue-name:jobs`
- Workers: `t:tenant-1:queue-name:workers`
- Metrics: `t:tenant-1:queue-name:metrics`
- Config: `tenant:tenant-1:config`

## API Endpoints

### Create Tenant
```http
POST /tenants
Content-Type: application/json
X-User-ID: {user_id}

{
  "id": "tenant-1",
  "name": "Tenant 1",
  "contact_email": "admin@tenant1.com",
  "metadata": {
    "environment": "production",
    "region": "us-west-2"
  },
  "quotas": {
    "max_jobs_per_hour": 10000,
    "max_jobs_per_day": 100000,
    "max_backlog_size": 50000,
    "max_job_size_bytes": 1048576,
    "max_queues_per_tenant": 10,
    "max_workers_per_queue": 50,
    "max_storage_bytes": 104857600,
    "enqueue_rate_limit": 100,
    "dequeue_rate_limit": 100,
    "soft_limit_threshold": 0.8
  },
  "encryption": {
    "enabled": true,
    "kek_provider": "aws-kms",
    "kek_key_id": "arn:aws:kms:us-west-2:123456789012:key/12345678-1234-1234-1234-123456789012",
    "dek_rotation_period": "168h",
    "algorithm": "AES-256-GCM"
  }
}
```

**Response (201 Created):**
```json
{
  "id": "tenant-1",
  "name": "Tenant 1",
  "status": "active",
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T10:30:00Z",
  "contact_email": "admin@tenant1.com",
  "quotas": { ... },
  "encryption": { ... },
  "rate_limiting": { ... },
  "metadata": { ... }
}
```

### Get Tenant
```http
GET /tenants/{tenant_id}
X-User-ID: {user_id}
```

**Response (200 OK):**
```json
{
  "id": "tenant-1",
  "name": "Tenant 1",
  "status": "active",
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T10:30:00Z",
  "quotas": { ... },
  "encryption": { ... },
  "rate_limiting": { ... }
}
```

### Update Tenant
```http
PUT /tenants/{tenant_id}
Content-Type: application/json
X-User-ID: {user_id}

{
  "id": "tenant-1",
  "name": "Updated Tenant Name",
  "status": "suspended",
  "quotas": { ... }
}
```

### Delete Tenant
```http
DELETE /tenants/{tenant_id}
X-User-ID: {user_id}
```

**Response (204 No Content)**

Note: This marks the tenant as deleted and cleans up all tenant data.

### List Tenants
```http
GET /tenants
X-User-ID: {user_id}
```

**Response (200 OK):**
```json
{
  "tenants": [
    {
      "id": "tenant-1",
      "name": "Tenant 1",
      "status": "active",
      "active_queues": 3,
      "quota_health": "good",
      "last_activity": "2025-01-15T10:45:00Z"
    }
  ],
  "count": 1
}
```

### Get Quota Usage
```http
GET /tenants/{tenant_id}/quota-usage
X-User-ID: {user_id}
```

**Response (200 OK):**
```json
{
  "tenant_id": "tenant-1",
  "jobs_this_hour": 150,
  "jobs_this_day": 2500,
  "current_backlog_size": 45,
  "storage_used_bytes": 1024000,
  "active_queues": 3,
  "active_workers": 12,
  "last_updated": "2025-01-15T10:45:00Z"
}
```

### Check Quota
```http
POST /tenants/{tenant_id}/check-quota
Content-Type: application/json
X-User-ID: {user_id}

{
  "quota_type": "jobs_per_hour",
  "amount": 50
}
```

**Response (200 OK):**
```json
{
  "allowed": true
}
```

**Response when quota exceeded:**
```json
{
  "allowed": false,
  "reason": "quota exceeded for tenant tenant-1: jobs_per_hour usage 9950 exceeds limit 10000"
}
```

## Tenant Status Values

- `active`: Tenant is operational and can process jobs
- `suspended`: Tenant is temporarily suspended, no job processing
- `warning`: Tenant is approaching quota limits
- `deleted`: Tenant is marked for deletion, data cleaned up

## Quota Types

- `jobs_per_hour`: Maximum jobs that can be enqueued per hour
- `jobs_per_day`: Maximum jobs that can be enqueued per day
- `backlog_size`: Maximum number of jobs in all queues
- `storage_bytes`: Maximum storage usage in bytes

## Quota Health Indicators

- `good`: Usage below soft limit threshold (default 80%)
- `warning`: Usage above soft limit but below hard limit
- `critical`: Usage at or above hard limit

## Encryption Configuration

### KEK Providers
- `local`: Local encryption (development only)
- `aws-kms`: AWS Key Management Service
- `gcp-kms`: Google Cloud Key Management Service
- `azure-kv`: Azure Key Vault

### Encryption Process
1. Generate random Data Encryption Key (DEK) for each payload
2. Encrypt payload with AES-256-GCM using DEK
3. Encrypt DEK with Key Encryption Key (KEK) from KMS
4. Store encrypted payload with encrypted DEK

## Rate Limiting

Rate limiting uses a sliding window algorithm with Redis sorted sets:
- Window duration configurable per tenant
- Burst capacity for handling traffic spikes
- Custom limits per operation type
- Optional enforcement across all workers

## Error Responses

### 400 Bad Request
```json
{
  "error": "validation error",
  "status": 400,
  "timestamp": "2025-01-15T10:30:00Z",
  "details": "tenant ID length must be between 3 and 32 characters"
}
```

### 403 Forbidden
```json
{
  "error": "access denied",
  "status": 403,
  "timestamp": "2025-01-15T10:30:00Z",
  "details": "access denied for user user1: cannot write resource tenant in tenant tenant-1 (reason: insufficient permissions)"
}
```

### 404 Not Found
```json
{
  "error": "tenant not found",
  "status": 404,
  "timestamp": "2025-01-15T10:30:00Z",
  "details": "tenant not found: non-existent-tenant"
}
```

### 409 Conflict
```json
{
  "error": "tenant already exists",
  "status": 409,
  "timestamp": "2025-01-15T10:30:00Z",
  "details": "tenant already exists: existing-tenant"
}
```

### 429 Too Many Requests
```json
{
  "error": "rate limit exceeded",
  "status": 429,
  "timestamp": "2025-01-15T10:30:00Z",
  "details": "rate limit exceeded for tenant tenant-1: enqueue rate 101/s exceeds limit 100/s (retry after 1s)",
  "retry_after": 1
}
```

## Audit Events

All tenant operations are logged with the following structure:
```json
{
  "event_id": "1642248600123456789",
  "timestamp": "2025-01-15T10:30:00Z",
  "tenant_id": "tenant-1",
  "user_id": "user123",
  "action": "CREATE_TENANT",
  "resource": "tenant:tenant-1",
  "details": {
    "tenant_name": "Tenant 1"
  },
  "remote_ip": "192.168.1.100",
  "user_agent": "TenantManager/1.0",
  "result": "SUCCESS",
  "error_code": null
}
```

## Integration Examples

### Go Client
```go
import "internal/multi-tenant-isolation"

// Create tenant manager
redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
tm := multitenantiso.NewTenantManager(redisClient)

// Create tenant
config := &multitenantiso.TenantConfig{
    ID:   "my-tenant",
    Name: "My Tenant",
    Quotas: multitenantiso.DefaultQuotas(),
}
err := tm.CreateTenant(config)

// Check quota before enqueuing job
err = tm.CheckQuota("my-tenant", "jobs_per_hour", 1)
if err != nil {
    if multitenantiso.IsQuotaExceeded(err) {
        // Handle quota exceeded
    }
}

// Increment usage after successful enqueue
err = tm.IncrementQuotaUsage("my-tenant", "jobs_per_hour", 1)

// Get namespaced keys
ns := tm.GetTenantNamespace("my-tenant")
queueKey := ns.QueueKey("my-queue") // "t:my-tenant:my-queue"
```

### HTTP Client
```bash
# Create tenant
curl -X POST http://localhost:8080/tenants \
  -H "Content-Type: application/json" \
  -H "X-User-ID: admin" \
  -d '{
    "id": "test-tenant",
    "name": "Test Tenant"
  }'

# Check quota
curl -X POST http://localhost:8080/tenants/test-tenant/check-quota \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user1" \
  -d '{
    "quota_type": "jobs_per_hour",
    "amount": 10
  }'
```

## Security Considerations

1. **Authentication**: Always validate user identity before allowing tenant operations
2. **Authorization**: Implement proper RBAC to restrict tenant access
3. **Encryption**: Use strong KEK providers in production (AWS KMS, etc.)
4. **Audit**: Enable audit logging for compliance and security monitoring
5. **Rate Limiting**: Configure appropriate rate limits to prevent abuse
6. **Key Rotation**: Regularly rotate encryption keys according to policy

## Performance Considerations

1. **Redis Memory**: Monitor memory usage as tenant count grows
2. **Key Scanning**: Use specific key patterns to avoid expensive SCAN operations
3. **Audit Storage**: Configure appropriate retention periods for audit logs
4. **Rate Limit Windows**: Balance window size with memory usage
5. **Encryption Overhead**: Consider performance impact of encryption on large payloads

## Configuration

Default configuration in `config.yaml`:
```yaml
multi_tenant_isolation:
  enabled: true
  default_encryption_enabled: false
  default_kek_provider: "local"
  default_dek_rotation_period: "168h"
  default_rate_limiting_enabled: true
  default_window_duration: "1s"
  default_burst_capacity: 10
  audit_enabled: true
  audit_retention_days: 90
  require_encryption_for_sensitive_data: false
  allowed_kek_providers: ["local", "aws-kms", "gcp-kms", "azure-kv"]
  min_tenant_quota_limits:
    max_jobs_per_hour: 100
    max_jobs_per_day: 1000
    # ... other limits
  max_tenant_quota_limits:
    max_jobs_per_hour: 1000000
    max_jobs_per_day: 10000000
    # ... other limits
```