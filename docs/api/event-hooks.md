# Event Hooks API

Event Hooks provide real-time notifications for job lifecycle events through webhooks and NATS messaging. This allows external systems to react to job state changes without polling.

## Overview

The Event Hooks system emits events for:
- `job_enqueued` - Job added to queue
- `job_started` - Worker begins processing
- `job_succeeded` - Job completed successfully
- `job_failed` - Job failed (will retry or go to DLQ)
- `job_dlq` - Job moved to dead letter queue
- `job_retried` - Job retry attempt initiated

## Webhook Subscriptions

### Create Webhook Subscription

```http
POST /api/v1/event-hooks/webhooks
Content-Type: application/json

{
  "name": "payment-notifications",
  "url": "https://api.example.com/webhooks/payments",
  "secret": "your-webhook-secret",
  "events": ["job_succeeded", "job_failed"],
  "queues": ["payments", "orders"],
  "min_priority": 5,
  "max_retries": 5,
  "timeout": "30s",
  "rate_limit": 60,
  "headers": [
    {"key": "Authorization", "value": "Bearer token123"}
  ],
  "include_payload": true,
  "redact_fields": ["user_id", "payment_info"]
}
```

### List Webhook Subscriptions

```http
GET /api/v1/event-hooks/webhooks
```

### Get Webhook Subscription

```http
GET /api/v1/event-hooks/webhooks/{id}
```

### Update Webhook Subscription

```http
PUT /api/v1/event-hooks/webhooks/{id}
Content-Type: application/json

{
  "disabled": false,
  "rate_limit": 120
}
```

### Delete Webhook Subscription

```http
DELETE /api/v1/event-hooks/webhooks/{id}
```

### Test Webhook Delivery

```http
POST /api/v1/event-hooks/webhooks/{id}/test
```

### Enable/Disable Webhook

```http
POST /api/v1/event-hooks/webhooks/{id}/enable
POST /api/v1/event-hooks/webhooks/{id}/disable
```

## Webhook Payload Format

```json
{
  "event": "job_succeeded",
  "timestamp": "2025-01-15T10:30:00Z",
  "job_id": "job_abc123",
  "queue": "payments",
  "priority": 8,
  "attempt": 1,
  "duration": "1.5s",
  "worker": "worker-001",
  "payload": {
    "order_id": "order_456",
    "amount": 99.99
  },
  "trace_id": "trace_xyz789",
  "request_id": "req_123456",
  "_links": {
    "job_details": "queue://localhost:8080/jobs/job_abc123",
    "queue_dashboard": "queue://localhost:8080/queues/payments",
    "retry_job": "queue://localhost:8080/jobs/job_abc123/retry"
  }
}
```

## Webhook Headers

The following headers are included with webhook deliveries:

- `Content-Type: application/json`
- `User-Agent: go-redis-work-queue/1.0`
- `X-Webhook-Delivery: uuid` - Unique delivery ID
- `X-Webhook-Event: job_succeeded` - Event type
- `X-Webhook-Timestamp: 1642248600` - Unix timestamp
- `X-Webhook-Job-ID: job_abc123` - Job identifier
- `X-Webhook-Queue: payments` - Queue name
- `X-Webhook-Signature: sha256=abc123...` - HMAC signature computed over `"<timestamp>." + body` (if secret provided)
- `X-Trace-ID: trace_xyz789` - Distributed tracing ID (if available)
- `X-Request-ID: req_123456` - Request correlation ID (if available)

## HMAC Signature Verification

If a webhook secret is provided, payloads are signed with HMAC-SHA256 using a canonical string that binds the timestamp:

```python
import hmac
import hashlib
import time

def verify_signature(payload: bytes, timestamp: str, signature: str, secret: str, freshness_seconds: int = 300) -> bool:
    """Return True if the signature and timestamp are valid."""

    # Enforce freshness window to block replays
    try:
        ts = int(timestamp)
    except ValueError:
        return False
    if abs(time.time() - ts) > freshness_seconds:
        return False

    canonical = f"{timestamp}.{payload.decode('utf-8')}".encode('utf-8')
    expected_hex = hmac.new(secret.encode('utf-8'), canonical, hashlib.sha256).hexdigest()
    expected_signature = f"sha256={expected_hex}"
    return hmac.compare_digest(signature, expected_signature)
```

Receivers must reject requests with stale or missing `X-Webhook-Timestamp` headers even if the HMAC validates.

## Dead Letter Hooks (DLH)

Failed webhook deliveries are stored in the Dead Letter Hooks queue for manual replay.

### List Dead Letter Hooks

```http
GET /api/v1/event-hooks/dlh?subscription_id={id}&limit=50
```

### Get Dead Letter Hook

```http
GET /api/v1/event-hooks/dlh/{id}
```

### Replay Dead Letter Hook

```http
POST /api/v1/event-hooks/dlh/{id}/replay
```

### Replay All Dead Letter Hooks

```http
POST /api/v1/event-hooks/dlh/replay-all?subscription_id={id}
```

### Delete Dead Letter Hook

```http
DELETE /api/v1/event-hooks/dlh/{id}
```

## Health and Metrics

### Get Health Status

```http
GET /api/v1/event-hooks/health
```

Response:
```json
{
  "event_bus": {
    "running": true,
    "events_emitted": 15420,
    "dlh_size": 3
  },
  "webhook_subscriptions": [
    {
      "subscription_id": "sub_123",
      "success_rate": 0.98,
      "last_success": "2025-01-15T10:29:45Z",
      "consecutive_failures": 0,
      "total_deliveries": 1250
    }
  ]
}
```

### Get Metrics

```http
GET /api/v1/event-hooks/metrics
```

Response:
```json
{
  "events_emitted": 15420,
  "webhook_deliveries": 15100,
  "webhook_failures": 320,
  "retry_attempts": 180,
  "dlh_size": 3,
  "delivery_latency_p95": "150ms",
  "rate_limit_violations": 12,
  "circuit_breaker_trips": 2,
  "subscription_health": {
    "sub_123": 0.98,
    "sub_456": 1.0
  }
}
```

## Testing

### Emit Test Event

```http
POST /api/v1/event-hooks/emit-test
Content-Type: application/json

{
  "event_type": "job_succeeded",
  "queue": "test-queue",
  "job_id": "test-job-123",
  "priority": 5,
  "payload": {"test": true}
}
```

## Rate Limiting

Webhook subscriptions support rate limiting to protect external endpoints:

- `rate_limit`: Maximum requests per minute (default: 60)
- Burst allowance for temporary spikes
- 429 Too Many Requests response when exceeded
- Automatic retry with exponential backoff

## Retry Policy

Failed webhook deliveries are automatically retried with exponential backoff:

- Initial delay: 1 second
- Backoff multiplier: 2.0
- Maximum delay: 5 minutes
- Maximum retries: 5 (configurable)
- Jitter added to prevent thundering herd

## Idempotency & Replay Semantics

Webhooks may be delivered more than once, especially when DLH replays are triggered. Consumers must:

- Treat every delivery as potentially duplicated.
- Use `X-Webhook-Delivery` as the idempotency key and persist it (e.g., in a short-lived cache or datastore) with a retention window appropriate for your SLA—30 to 90 minutes is typical.
- Honor the optional `X-Webhook-Replay: true` header on DLH replays so you can provide custom logging or altered retry logic.
- Return a 2xx response for requests whose delivery ID has already been processed; the sender interprets duplicate IDs with 2xx as success and will not replay again.

**Recommendation:** store delivery IDs with an expiry in Redis or your application database. On duplicate IDs, skip side effects and log the duplicate rather than erroring. When possible, make downstream operations idempotent so retries and replays remain safe.

## Error Handling

Webhook delivery errors are categorized as:

**Retryable (5xx, timeouts, network errors):**
- 500-599 server errors
- 408 request timeout
- 429 too many requests
- Network/DNS failures

**Non-retryable (4xx client errors):**
- 400-499 client errors (except 408, 429)
- Invalid URL
- Authentication failures

## Security

- HMAC-SHA256 payload signing
- Field redaction for sensitive data
- Optional mTLS support
- Rate limiting and circuit breakers
- Audit logging for all operations

## TUI Integration

Access Event Hooks management through the TUI:

1. Navigate to "Event Hooks" tab
2. View subscription status and metrics
3. Monitor dead letter hooks
4. Access management commands

Keyboard shortcuts:
- `h` - Event hooks help
- `t` - Test webhook delivery
- `r` - Replay failed deliveries

## Examples

### Payment Processing Webhook

```bash
curl -X POST http://localhost:8080/api/v1/event-hooks/webhooks \
  -H "Content-Type: application/json" \
  -d '{
    "name": "payment-processor",
    "url": "https://api.payments.com/webhooks/jobs",
    "secret": "webhook-secret-key",
    "events": ["job_succeeded", "job_failed"],
    "queues": ["payments"],
    "min_priority": 7,
    "headers": [
      {"key": "Authorization", "value": "Bearer your-token"}
    ]
  }'
```

### Monitoring Dashboard

```bash
curl -X POST http://localhost:8080/api/v1/event-hooks/webhooks \
  -H "Content-Type: application/json" \
  -d '{
    "name": "monitoring-dashboard",
    "url": "https://dashboard.internal.com/webhooks/queue-events",
    "events": ["job_failed", "job_dlq"],
    "queues": ["*"],
    "rate_limit": 120,
    "include_payload": false
  }'
```

### Slack Notifications

```bash
curl -X POST http://localhost:8080/api/v1/event-hooks/webhooks \
  -H "Content-Type: application/json" \
  -d '{
    "name": "slack-alerts",
    "url": "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK",
    "events": ["job_failed", "job_dlq"],
    "queues": ["critical-jobs"],
    "min_priority": 8
  }'
```
