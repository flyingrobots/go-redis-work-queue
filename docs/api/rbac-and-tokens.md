# RBAC and Tokens API Documentation

## Overview

The RBAC and Tokens module provides role-based access control and JWT/PASETO token authentication for the Redis Work Queue Admin API. It implements secure token-based authentication with fine-grained authorization controls, comprehensive audit logging, and key management features.

## Features

- **Token-based Authentication**: JWT tokens with HMAC-SHA256 signatures
- **Role-based Access Control**: Four-tier role system (Admin, Maintainer, Operator, Viewer)
- **Permission System**: Fine-grained permissions for different operations
- **Key Management**: Key rotation with graceful expiration
- **Audit Logging**: Comprehensive logging of authentication and authorization events
- **Token Revocation**: Real-time token revocation support
- **Caching**: Authorization decision caching for performance

## Architecture

The RBAC system consists of several key components:

- **Manager**: Core logic for token generation, validation, and authorization
- **Handler**: HTTP endpoints for token operations
- **Middleware**: Request authentication and authorization
- **AuditLogger**: Audit trail for security events

## Roles and Permissions

### Roles

| Role | Description | Use Case |
|------|-------------|----------|
| `admin` | Full system access | System administrators |
| `maintainer` | Queue management and maintenance | Operations team |
| `operator` | Day-to-day operations | Application operators |
| `viewer` | Read-only access | Monitoring and reporting |

### Permissions

| Permission | Description |
|------------|-------------|
| `admin:all` | All permissions (admin role only) |
| `stats:read` | Read queue statistics |
| `queue:read` | Read queue contents |
| `queue:write` | Add jobs to queues |
| `queue:delete` | Delete queues or jobs |
| `job:read` | Read job details |
| `job:write` | Create or modify jobs |
| `job:delete` | Delete jobs |
| `worker:read` | Read worker status |
| `worker:manage` | Manage worker lifecycle |
| `bench:run` | Run performance benchmarks |

### Role-Permission Mapping

```yaml
admin:
  - admin:all

maintainer:
  - stats:read
  - queue:read
  - queue:write
  - queue:delete
  - job:read
  - job:write
  - job:delete
  - worker:read
  - worker:manage
  - bench:run

operator:
  - stats:read
  - queue:read
  - queue:write
  - job:read
  - job:write
  - worker:read
  - bench:run

viewer:
  - stats:read
  - queue:read
  - job:read
  - worker:read
```

## API Endpoints

### Generate Token

Generate a new authentication token for a user.

**POST** `/auth/token`

```json
{
  "subject": "user@example.com",
  "roles": ["operator"],
  "scopes": ["queue:read", "stats:read"],
  "ttl": "24h"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "subject": "user@example.com",
  "expires_at": "2025-09-15T13:00:00Z",
  "token_type": "bearer"
}
```

### Validate Token

Validate a token and return its claims.

**POST** `/auth/validate`

**Headers:**
```
Authorization: Bearer <token>
```

**Response:**
```json
{
  "valid": true,
  "subject": "user@example.com",
  "roles": ["operator"],
  "scopes": ["queue:read", "stats:read"],
  "expires_at": "2025-09-15T13:00:00Z",
  "issued_at": "2025-09-14T13:00:00Z",
  "token_type": "bearer",
  "key_id": "key-123"
}
```

### Get Token Info

Get information about the current token.

**GET** `/auth/token/info`

**Headers:**
```
Authorization: Bearer <token>
```

**Response:** Same as validate token response.

### Revoke Token

Revoke a specific token by its ID.

**POST** `/auth/token/revoke`

```json
{
  "token_id": "jti-123",
  "reason": "User requested revocation"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Token revoked successfully"
}
```

### Check Access

Check if the current token has permission for a specific action.

**POST** `/auth/check?action=queue:delete&resource=/api/v1/queues/dlq`

**Headers:**
```
Authorization: Bearer <token>
```

**Response:**
```json
{
  "allowed": true,
  "subject": "user@example.com",
  "roles": ["maintainer"],
  "scopes": ["queue:delete"],
  "reason": "granted by role: maintainer",
  "request_id": "req-123"
}
```

### Query Audit Log

Query the audit log for security events.

**GET** `/audit/query?subject=user@example.com&limit=10&start_time=2025-09-14T00:00:00Z`

**Response:**
```json
{
  "entries": [
    {
      "id": "audit-123",
      "timestamp": "2025-09-14T13:00:00Z",
      "subject": "user@example.com",
      "action": "DELETE /api/v1/queues/dlq",
      "resource": "/api/v1/queues/dlq",
      "result": "200",
      "reason": "Cleanup operation",
      "details": {
        "items_deleted": 5
      },
      "ip": "192.168.1.100",
      "user_agent": "curl/7.68.0",
      "request_id": "req-123"
    }
  ],
  "count": 1,
  "filter": {
    "subject": "user@example.com",
    "limit": 10
  }
}
```

## Integration with Admin API

The RBAC system integrates with the existing Admin API through middleware:

1. **AuthMiddleware**: Validates tokens and adds claims to request context
2. **AuthzMiddleware**: Checks permissions for each endpoint
3. **AuditMiddleware**: Logs destructive operations

### Protected Endpoints

| Endpoint | Required Permission | Notes |
|----------|-------------------|--------|
| `GET /api/v1/stats` | `stats:read` | Queue statistics |
| `GET /api/v1/stats/keys` | `stats:read` | Detailed key statistics |
| `GET /api/v1/queues/*/peek` | `queue:read` | Peek at queue contents |
| `DELETE /api/v1/queues/dlq` | `queue:delete` | Purge dead letter queue |
| `DELETE /api/v1/queues/all` | `queue:delete` + `admin:all` | Requires admin role |
| `POST /api/v1/bench` | `bench:run` | Run performance benchmark |

### Bypassed Endpoints

These endpoints bypass authentication:
- `GET /health` - Health checks
- `POST /auth/token` - Token generation

## Configuration

```yaml
token:
  format: "jwt"                    # Token format: jwt or paseto
  default_ttl: "24h"              # Default token lifetime
  max_ttl: "168h"                 # Maximum token lifetime (7 days)
  issuer: "redis-work-queue"      # Token issuer
  audience: "admin-api"           # Token audience
  allow_refresh: true             # Enable token refresh
  refresh_ttl: "168h"             # Refresh token lifetime

keys:
  rotation_interval: "720h"       # Key rotation interval (30 days)
  grace_period: "24h"             # Grace period after rotation
  algorithm: "HS256"              # Signing algorithm
  key_size: 256                   # Key size in bits
  storage:
    type: "file"                  # Storage backend: memory, file, redis, vault
    connection: "./keys"          # Storage connection string

audit:
  enabled: true                   # Enable audit logging
  log_path: "./audit.log"         # Audit log file path
  rotate_size: 104857600          # Log rotation size (100MB)
  max_backups: 10                 # Maximum backup files
  compress: true                  # Compress rotated logs
  retention_days: 90              # Log retention period
  filter_sensitive: true         # Filter sensitive data
  include_bodies: false           # Include request/response bodies

authz:
  default_deny: true              # Default deny policy
  cache_enabled: true             # Cache authorization decisions
  cache_ttl: "5m"                 # Cache TTL
  roles_file: "./roles.yaml"      # Role definitions file
  resources_file: "./resources.yaml" # Resource patterns file
  dynamic_roles: false            # Allow dynamic role assignment
```

## Security Considerations

### Token Security

- Tokens use HMAC-SHA256 signatures for integrity
- Tokens include expiration times (`exp`) and not-before times (`nbf`)
- Token revocation is tracked in memory (consider Redis for clustering)
- Key rotation is supported with graceful transition

### Authorization

- Default deny policy - explicit permissions required
- Admin role bypass for emergency access
- Audit logging for all destructive operations
- Rate limiting through existing middleware

### Key Management

- Keys are generated using crypto/rand for security
- Key rotation with configurable intervals
- Multiple active keys supported for zero-downtime rotation
- Keys should be stored securely (file system permissions, encryption at rest)

## Error Codes

| Code | Description |
|------|-------------|
| `TOKEN_MISSING` | No authentication token provided |
| `TOKEN_INVALID` | Token format or signature invalid |
| `TOKEN_EXPIRED` | Token has expired |
| `TOKEN_REVOKED` | Token has been revoked |
| `TOKEN_NOT_YET_VALID` | Token is not yet valid (nbf) |
| `ACCESS_DENIED` | Insufficient permissions |
| `KEY_NOT_FOUND` | Signing key not found |
| `SIGNATURE_MISMATCH` | Token signature verification failed |

## Usage Examples

### Generate a Token for an Operator

```bash
curl -X POST http://localhost:8080/auth/token \
  -H "Content-Type: application/json" \
  -d '{
    "subject": "operator@example.com",
    "roles": ["operator"],
    "ttl": "8h"
  }'
```

### Access Protected Endpoint

```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/stats
```

### Check Permissions

```bash
curl -X POST "http://localhost:8080/auth/check?action=queue:delete&resource=/api/v1/queues/dlq" \
  -H "Authorization: Bearer $TOKEN"
```

### Query Audit Log

```bash
curl "http://localhost:8080/audit/query?action=DELETE&limit=5" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

## Implementation Notes

- The system is designed for single-node deployments. For clustering, consider using Redis for revocation lists and key storage.
- Audit logs use JSON Lines format for easy parsing and analysis.
- Authorization decisions are cached for performance, with configurable TTL.
- The middleware can be selectively applied to different route groups.

## Testing

The implementation includes comprehensive tests covering:

- Token generation and validation
- Authorization checks for different roles
- Middleware integration
- Error conditions and edge cases
- Token revocation
- Audit logging

Run tests with:
```bash
go test ./internal/rbac-and-tokens/ -v
```