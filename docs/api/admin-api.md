# Admin API Documentation

The Admin API provides secure HTTP endpoints for managing Redis work queues. It includes authentication, rate limiting, audit logging, and confirmation requirements for destructive operations.

## Features

- **Secure by Default**: JWT authentication with deny-by-default policy
- **Rate Limiting**: Token bucket algorithm with configurable limits
- **Audit Logging**: Comprehensive logging of all destructive operations
- **Double Confirmation**: Required confirmation phrases for dangerous operations
- **OpenAPI Spec**: Full OpenAPI 3.0 specification available at `/api/v1/openapi.yaml`

## Configuration

The API is configured through environment variables or config file:

```yaml
admin_api:
  listen_addr: ":8080"

  # Authentication
  jwt_secret: "your-secret-key"
  require_auth: true
  deny_by_default: true

  # Rate Limiting
  rate_limit_enabled: true
  rate_limit_per_minute: 100
  rate_limit_burst: 10

  # Audit Logging
  audit_enabled: true
  audit_log_path: "/var/log/admin-api/audit.log"
  audit_rotate_size: 104857600  # 100MB
  audit_max_backups: 10

  # Security
  cors_enabled: false
  cors_allow_origins: ["*"]

  # Confirmations
  require_double_confirm: true
  confirmation_phrase: "CONFIRM_DELETE"
```

## Authentication

The API uses JWT Bearer tokens for authentication. Include the token in the Authorization header:

```http
Authorization: Bearer <your-jwt-token>
```

### JWT Token Structure

```json
{
  "sub": "user@example.com",
  "roles": ["admin"],
  "exp": 1234567890,
  "iat": 1234567880
}
```

## Endpoints

### Statistics

#### GET /api/v1/stats
Returns queue statistics including queue lengths, processing lists, and worker heartbeats.

**Response:**
```json
{
  "queues": {
    "high(jobqueue:high)": 42,
    "low(jobqueue:low)": 123
  },
  "processing_lists": {
    "jobqueue:worker:1:processing": 5
  },
  "heartbeats": 10,
  "timestamp": "2025-01-14T10:30:00Z"
}
```

#### GET /api/v1/stats/keys
Returns detailed information about all managed Redis keys.

**Response:**
```json
{
  "queue_lengths": {
    "high(jobqueue:high)": 42,
    "low(jobqueue:low)": 123
  },
  "processing_lists": 3,
  "processing_items": 15,
  "heartbeats": 10,
  "rate_limit_key": "jobqueue:rate_limit",
  "rate_limit_ttl": "45s",
  "timestamp": "2025-01-14T10:30:00Z"
}
```

### Queue Management

#### GET /api/v1/queues/{queue}/peek
View jobs in a queue without removing them.

**Parameters:**
- `queue`: Queue name (high, low, completed, dead_letter, or full Redis key)
- `count`: Number of items to peek (1-100, default 10)

**Example:**
```http
GET /api/v1/queues/high/peek?count=5
```

**Response:**
```json
{
  "queue": "jobqueue:high",
  "items": [
    "{\"id\":\"job-1\",\"filepath\":\"/data/file1.txt\"}",
    "{\"id\":\"job-2\",\"filepath\":\"/data/file2.txt\"}"
  ],
  "count": 2,
  "timestamp": "2025-01-14T10:30:00Z"
}
```

#### DELETE /api/v1/queues/dlq
Purge the dead letter queue. Requires confirmation.

**Request Body:**
```json
{
  "confirmation": "CONFIRM_DELETE",
  "reason": "Clearing failed jobs after investigation"
}
```

**Response:**
```json
{
  "success": true,
  "items_deleted": 42,
  "message": "Successfully purged 42 items from dead letter queue",
  "timestamp": "2025-01-14T10:30:00Z"
}
```

#### DELETE /api/v1/queues/all
Purge ALL queues. Requires double confirmation.

**Request Body:**
```json
{
  "confirmation": "CONFIRM_DELETE_ALL",
  "reason": "System reset for testing environment"
}
```

**Response:**
```json
{
  "success": true,
  "items_deleted": 15,
  "message": "Successfully purged 15 keys from all queues",
  "timestamp": "2025-01-14T10:30:00Z"
}
```

### Benchmarking

#### POST /api/v1/bench
Run a performance benchmark by enqueuing test jobs.

**Request Body:**
```json
{
  "count": 1000,
  "priority": "high",
  "rate": 100,
  "timeout_seconds": 30
}
```

**Response:**
```json
{
  "count": 1000,
  "duration": "10.5s",
  "throughput_jobs_per_sec": 95.2,
  "p50_latency": "45ms",
  "p95_latency": "120ms",
  "timestamp": "2025-01-14T10:30:00Z"
}
```

## Rate Limiting

The API implements token bucket rate limiting:

- **Default Limit**: 100 requests per minute
- **Burst**: 10 requests
- **Headers**: Rate limit information is included in response headers

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1234567890
```

## Audit Logging

All destructive operations are logged to the audit log with the following information:

```json
{
  "id": "1234567890-123456",
  "timestamp": "2025-01-14T10:30:00Z",
  "user": "admin@example.com",
  "action": "PURGE_DLQ",
  "resource": "jobqueue:dead_letter",
  "result": "SUCCESS",
  "reason": "Clearing failed jobs after investigation",
  "details": {
    "items_deleted": 42
  },
  "ip": "192.168.1.100",
  "user_agent": "curl/7.68.0"
}
```

## Error Responses

All errors follow a consistent format:

```json
{
  "error": "Rate limit exceeded",
  "code": "RATE_LIMIT",
  "details": {
    "retry_after": "60"
  }
}
```

### Common Error Codes

- `AUTH_MISSING`: Authorization header not provided
- `AUTH_INVALID`: Invalid or expired JWT token
- `RATE_LIMIT`: Rate limit exceeded
- `CONFIRMATION_FAILED`: Invalid confirmation phrase
- `REASON_REQUIRED`: Reason not provided for destructive operation
- `INTERNAL_ERROR`: Internal server error

## Security Best Practices

1. **Always use HTTPS in production** - Enable TLS with proper certificates
2. **Rotate JWT secrets regularly** - Update the jwt_secret configuration
3. **Monitor audit logs** - Review logs for suspicious activity
4. **Use strong confirmation phrases** - Change default confirmation phrases
5. **Implement least privilege** - Grant minimal necessary permissions in JWT roles
6. **Set appropriate rate limits** - Adjust based on your usage patterns

## Integration Examples

### cURL

```bash
# Get stats
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/stats

# Peek at queue
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/queues/high/peek?count=10"

# Purge DLQ
curl -X DELETE \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"confirmation":"CONFIRM_DELETE","reason":"Clearing test data"}' \
  http://localhost:8080/api/v1/queues/dlq

# Run benchmark
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"count":100,"priority":"high","rate":50}' \
  http://localhost:8080/api/v1/bench
```

### Go Client

```go
import (
    "bytes"
    "encoding/json"
    "net/http"
)

func getStats(token string) (*StatsResponse, error) {
    req, _ := http.NewRequest("GET", "http://localhost:8080/api/v1/stats", nil)
    req.Header.Set("Authorization", "Bearer " + token)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var stats StatsResponse
    json.NewDecoder(resp.Body).Decode(&stats)
    return &stats, nil
}
```

## Monitoring

### Health Check

The `/health` endpoint provides a simple health check (no authentication required):

```http
GET /health

Response: {"status":"healthy"}
```

### Metrics

Monitor these key metrics:

- Request rate and latency per endpoint
- Authentication failures
- Rate limit violations
- Audit log volume
- Error rates by error code

## Troubleshooting

### Common Issues

1. **401 Unauthorized**
   - Check JWT token is valid and not expired
   - Verify jwt_secret configuration matches token signing

2. **429 Too Many Requests**
   - Rate limit exceeded, wait for reset
   - Consider increasing rate limits if legitimate

3. **400 Bad Request on Purge**
   - Verify confirmation phrase matches configuration
   - Ensure reason is provided and meets minimum length

4. **500 Internal Server Error**
   - Check server logs for details
   - Verify Redis connectivity
   - Check disk space for audit logs

## Version History

- **v1.0.0** - Initial release with core admin operations
  - Stats, Peek, Purge, and Benchmark endpoints
  - JWT authentication and rate limiting
  - Audit logging for destructive operations