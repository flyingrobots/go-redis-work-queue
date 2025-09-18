# DLQ Remediation Pipeline API

The DLQ Remediation Pipeline provides automated classification and remediation of dead letter queue jobs using configurable rules and actions.

## Overview

The remediation pipeline automates DLQ cleanup through:
- **Intelligent Classification**: Pattern-based job classification with optional external ML models
- **Configurable Actions**: Requeue, transform, redact, drop, route, delay, tag, and notify actions
- **Safety Controls**: Rate limiting, circuit breakers, dry-run mode, and audit trails
- **Real-time Monitoring**: Comprehensive metrics, health checks, and audit logging

## Base URL

All API endpoints are prefixed with `/api/v1/dlq-remediation`.

## Authentication

All endpoints require authentication via JWT or PASETO tokens:

```http
Authorization: Bearer <token>
```

Audit payloads are redacted by default. Configure behaviour via `audit_redaction` (see `config/audit.yaml`). The default mask list includes: `ssn`, `email`, `phone`, `full_name`, `address`, `credit_card`, `payment_token`, `auth_token`, and `password`. Operators can extend the list or select `redaction_level: none|default|strict` per environment.

## Pipeline Management

### Start Pipeline

Start the remediation pipeline.

```http
POST /pipeline/start
```

**Response:**
```json
{
  "status": "started"
}
```

### Stop Pipeline

Stop the remediation pipeline.

```http
POST /pipeline/stop
```

**Response:**
```json
{
  "status": "stopped"
}
```

### Pause Pipeline

Pause the remediation pipeline.

```http
POST /pipeline/pause
```

**Response:**
```json
{
  "status": "paused"
}
```

### Resume Pipeline

Resume the paused pipeline.

```http
POST /pipeline/resume
```

**Response:**
```json
{
  "status": "resumed"
}
```

### Get Pipeline Status

Get current pipeline status and statistics.

```http
GET /pipeline/status
```

**Response:**
```json
{
  "status": "running",
  "started_at": "2024-01-15T10:00:00Z",
  "last_run_at": "2024-01-15T10:05:00Z",
  "next_run_at": "2024-01-15T10:05:30Z",
  "total_processed": 15423,
  "total_successful": 14892,
  "total_failed": 531,
  "rules_enabled": 8,
  "rules_disabled": 2,
  "current_batch_size": 50,
  "last_error": null,
  "last_error_at": null
}
```

### Get Pipeline Metrics

Get detailed pipeline performance metrics.

```http
GET /pipeline/metrics
```

**Response:**
```json
{
  "timestamp": "2024-01-15T10:05:30Z",
  "jobs_processed": 15423,
  "jobs_matched": 12341,
  "actions_executed": 18765,
  "actions_successful": 18234,
  "actions_failed": 531,
  "classification_time_ms": 45,
  "action_time_ms": 124,
  "end_to_end_time_ms": 169,
  "rate_limit_hits": 12,
  "circuit_breaker_trips": 3,
  "cache_hit_rate": 0.85
}
```

**Field schema:**

- `timestamp` — string (RFC3339)
- `jobs_processed`, `jobs_matched`, `actions_executed`, `actions_successful`, `actions_failed`, `rate_limit_hits`, `circuit_breaker_trips` — integer
- `classification_time_ms`, `action_time_ms`, `end_to_end_time_ms` — integer (milliseconds)
- `cache_hit_rate` — number (0.0–1.0)

## Batch Processing

### Process Batch

Process a batch of DLQ jobs using current rules.

```http
POST /pipeline/process-batch
```

**Headers:**
- `Idempotency-Key` (string, optional but required for safe retries): repeated values within 24 hours return the original `200` response body without re-executing actions.

**Response:**
```json
{
  "started_at": "2024-01-15T10:05:00Z",
  "completed_at": "2024-01-15T10:05:12Z",
  "total_jobs": 50,
  "processed_jobs": 45,
  "successful_jobs": 42,
  "failed_jobs": 3,
  "skipped_jobs": 5,
  "results": [
    {
      "job_id": "job_123",
      "rule_id": "rule_abc",
      "success": true,
      "actions": ["requeue"],
      "duration_ms": 125,
      "dry_run": false
    }
  ],
  "errors": []
}
```

`dry_run: true` guarantees no state changes. When `dry_run` is `false`, the service persists outcomes before responding. Clients should cache successful responses for 24 hours and reuse the same `Idempotency-Key` if they must retry to avoid duplicate execution.

### Dry Run Batch

Run batch processing in dry-run mode (no actual changes).

```http
POST /pipeline/dry-run
```

**Response:** Same as process-batch but with `dry_run: true` in all results and `duration_ms` values reflecting simulated execution time.

## Rule Management

### Get Rules

List all remediation rules with optional filtering.

```http
GET /rules?enabled=true&tag=validation
```

**Query Parameters:**
- `enabled` (boolean): Filter by enabled status
- `tag` (string): Filter by tag

**Response:**
```json
{
  "rules": [
    {
      "id": "rule_abc123",
      "name": "Validation Error Remediation",
      "description": "Handles validation errors by redacting PII and retrying",
      "priority": 100,
      "enabled": true,
      "created_at": "2024-01-15T09:00:00Z",
      "updated_at": "2024-01-15T09:30:00Z",
      "created_by": "admin@example.com",
      "matcher": {
        "error_pattern": "validation.*failed|invalid.*format",
        "job_type": "user_registration",
        "retry_count": "< 3"
      },
      "actions": [
        {
          "type": "redact",
          "parameters": {
            "fields": ["ssn", "email", "phone"],
            "replacement": "[REDACTED]"
          }
        },
        {
          "type": "requeue",
          "parameters": {
            "target_queue": "user_registration_retry",
            "delay": "5m"
          }
        }
      ],
      "safety": {
        "max_per_minute": 10,
        "max_total_per_run": 100,
        "error_rate_threshold": 0.05,
        "backoff_on_failure": true
      },
      "tags": ["validation", "user", "pii"],
      "statistics": {
        "total_matches": 1247,
        "successful_actions": 1198,
        "failed_actions": 49,
        "success_rate": 0.96,
        "average_latency": 145.7,
        "last_matched_at": "2024-01-15T10:05:00Z",
        "last_success_at": "2024-01-15T10:05:00Z",
        "last_failure_at": "2024-01-15T09:45:00Z"
      }
    }
  ],
  "count": 1
}
```

### Create Rule

Create a new remediation rule.

```http
POST /rules
Content-Type: application/json

{
  "name": "Payment Timeout Remediation",
  "description": "Handles payment timeouts with exponential backoff",
  "priority": 90,
  "enabled": true,
  "matcher": {
    "error_pattern": {"regex": "timeout"},
    "job_type": {"wildcard": "payment_*"},
    "retry_count": {"operator": ">", "value": 0}
  },
  "actions": [
    {
      "type": "delay",
      "parameters": {
        "delay": "exponential:30s:5m"
      }
    },
    {
      "type": "requeue",
      "parameters": {
        "target_queue": "payment_retry",
        "priority": 3
      }
    }
  ],
  "safety": {
    "max_per_minute": 20,
    "max_total_per_run": 200,
    "error_rate_threshold": 0.1
  },
  "tags": ["payment", "timeout"]
}
```

**Response:**
```json
{
  "status": "created",
  "rule_id": "rule_def456"
}
```

**Matcher schema**

- `error_pattern` — `{ "regex": "..." }` (ECMAScript-compatible)
- `job_type` — one of `{ "equals": "queue" }`, `{ "wildcard": "pattern_*" }`, `{ "values": ["a", "b"] }`
- `retry_count` — `{ "operator": "<|<=|=|>=|>", "value": <integer> }`
- `time_window` (optional) — `{ "start": "09:00", "end": "17:00", "timezone": "America/Los_Angeles" }`

Requests failing schema validation return `400 Bad Request` with payload:

```json
{
  "status": "invalid_matcher",
  "errors": [
    {"field": "matcher.retry_count", "message": "operator must be one of <, <=, =, >=, >"}
  ]
}
```

### Get Rule

Get details of a specific rule.

```http
GET /rules/{ruleID}
```

**Response:** Single rule object (same format as in list).

### Update Rule

Update an existing rule.

```http
PUT /rules/{ruleID}
Content-Type: application/json

{
  "name": "Updated Payment Timeout Remediation",
  "description": "Updated description",
  "priority": 95,
  "enabled": true,
  "matcher": {
    "error_type": "timeout",
    "job_type": "payment_*",
    "retry_count": "> 1"
  },
  "actions": [
    {
      "type": "requeue",
      "parameters": {
        "target_queue": "payment_retry_v2",
        "delay": "10m"
      }
    }
  ],
  "safety": {
    "max_per_minute": 15,
    "max_total_per_run": 150,
    "error_rate_threshold": 0.08
  },
  "tags": ["payment", "timeout", "v2"]
}
```

**Response:**
```json
{
  "status": "updated"
}
```

### Delete Rule

Delete a rule.

```http
DELETE /rules/{ruleID}
```

**Response:**
```json
{
  "status": "deleted"
}
```

### Enable Rule

Enable a disabled rule.

```http
POST /rules/{ruleID}/enable
```

**Response:**
```json
{
  "status": "enabled"
}
```

### Disable Rule

Disable an enabled rule.

```http
POST /rules/{ruleID}/disable
```

**Response:**
```json
{
  "status": "disabled"
}
```

### Test Rule

Test a rule against a sample job.

```http
POST /rules/{ruleID}/test
Content-Type: application/json

{
  "job_id": "test_job_123",
  "job_type": "payment_processing",
  "queue": "payments",
  "error": "connection timeout after 30s",
  "error_type": "timeout",
  "retry_count": 2,
  "payload": {
    "user_id": 12345,
    "amount": 99.99,
    "currency": "USD"
  },
  "failed_at": "2024-01-15T10:00:00Z"
}
```

**Response:**
```json
{
  "rule_id": "rule_def456",
  "classification": {
    "job_id": "test_job_123",
    "category": "Payment Timeout Remediation",
    "confidence": 0.95,
    "rule_id": "rule_def456",
    "actions": ["delay", "requeue"],
    "reason": "Matched rule 'Payment Timeout Remediation' with confidence 0.95"
  },
  "execution": {
    "job_id": "test_job_123",
    "rule_id": "rule_def456",
    "success": true,
    "actions": ["delay", "requeue"],
    "duration": "45ms",
    "dry_run": true
  },
  "would_match": true
}
```

## Classification

### Classify Job

Classify a single job against all rules.

```http
POST /classify
Content-Type: application/json

{
  "job_id": "job_789",
  "job_type": "user_registration",
  "queue": "users",
  "error": "validation failed: email format invalid",
  "error_type": "validation_error",
  "retry_count": 1,
  "payload": {
    "user_id": 456,
    "email": "invalid-email",
    "name": "John Doe"
  },
  "failed_at": "2024-01-15T10:00:00Z"
}
```

**Response:**
```json
{
  "job_id": "job_789",
  "category": "Validation Error Remediation",
  "confidence": 0.92,
  "rule_id": "rule_abc123",
  "actions": ["redact", "requeue"],
  "reason": "Matched rule 'Validation Error Remediation' with confidence 0.92",
  "timestamp": "2024-01-15T10:05:30Z"
}
```

### Classify Batch

Classify multiple jobs at once.

```http
POST /classify/batch
Content-Type: application/json

[
  {
    "job_id": "job_1",
    "job_type": "payment",
    "error": "timeout",
    "error_type": "timeout",
    "retry_count": 3
  },
  {
    "job_id": "job_2",
    "job_type": "notification",
    "error": "invalid email",
    "error_type": "validation_error",
    "retry_count": 0
  }
]
```

**Response:**
```json
{
  "classifications": [
    {
      "job_id": "job_1",
      "category": "Payment Timeout Remediation",
      "confidence": 0.89,
      "rule_id": "rule_def456",
      "actions": ["delay", "requeue"],
      "reason": "Matched timeout pattern"
    },
    {
      "job_id": "job_2",
      "category": "unclassified",
      "confidence": 0.0,
      "actions": [],
      "reason": "No matching rules found"
    }
  ],
  "count": 2
}
```

## Audit and Monitoring

### Get Audit Log

Retrieve audit log entries with filtering.

```http
GET /audit?job_id=job_123&start_time=2024-01-15T09:00:00Z&limit=100
```

**Query Parameters:**
- `job_id` (string): Filter by job ID
- `rule_id` (string): Filter by rule ID
- `action` (string): Filter by action type
- `user_id` (string): Filter by user ID
- `result` (string): Filter by result (success/failure)
- `dry_run` (boolean): Filter by dry run status
- `start_time` (ISO8601): Filter by start time
- `end_time` (ISO8601): Filter by end time
- `limit` (integer): Maximum entries to return (default: 100)
- `offset` (integer): Number of entries to skip
- `sort_by` (string): Sort field (timestamp/duration)
- `sort_order` (string): Sort order (asc/desc)

**Response:**
```json
{
  "entries": [
    {
      "id": "audit_12345",
      "timestamp": "2024-01-15T10:05:00Z",
      "job_id": "job_123",
      "rule_id": "rule_abc123",
      "rule_name": "Validation Error Remediation",
      "action": "classify_and_remediate",
      "parameters": {
        "rule_priority": 100,
        "classification_confidence": 0.92,
        "classification_category": "Validation Error Remediation",
        "actions_count": 2,
        "action_types": ["redact", "requeue"]
      },
      "result": "success",
      "dry_run": false,
      "user_id": "system",
      "duration": "145ms",
      "before_state": {
        "job_id": "<redacted>",
        "error": "<redacted>",
        "retry_count": "<redacted>"
      },
      "after_state": {
        "job_id": "<redacted>",
        "error": "<redacted>",
        "retry_count": "<redacted>",
        "queue": "<redacted>"
      }
    }
  ],
  "count": 1,
  "filter": {
    "job_id": "job_123",
    "limit": 100,
    "sort_by": "timestamp",
    "sort_order": "desc"
  }
}
```

### Get Audit Statistics

Get audit log statistics and analytics.

```http
GET /audit/stats?days=30
```

**Query Parameters:**
- `days` (integer): Number of days to analyze (default: 30)

**Response:**
```json
{
  "total_entries": 15423,
  "daily_counts": {
    "2024-01-15": 1247,
    "2024-01-14": 1156,
    "2024-01-13": 1089
  },
  "action_counts": {
    "requeue": 8734,
    "transform": 3421,
    "redact": 2456,
    "drop": 712,
    "route": 100
  }
}
```

## Health Check

### Get Health Status

Get overall system health and status.

```http
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:05:30Z",
  "pipeline": {
    "status": "running",
    "uptime": "5h32m15s",
    "last_run": "2024-01-15T10:05:00Z",
    "total_processed": 15423,
    "success_rate": 0.965
  },
  "rules": {
    "enabled": 8,
    "disabled": 2,
    "total": 10
  },
  "metrics": {
    "jobs_processed": 15423,
    "actions_successful": 14892,
    "actions_failed": 531,
    "average_latency_ms": 168.9,
    "rate_limit_hits": 12,
    "circuit_breaker_trips": 3
  }
}
```

## Error Responses

All endpoints return standardized error envelopes and emit an `X-Request-ID` header that matches the response body. Clients should log this identifier for support requests.

```json
{
  "code": "rule_not_found",
  "error": "Rule not found",
  "status": 404,
  "request_id": "3f2c0b0a-4b61-4e12-9a43-0c6af6a27b9d",
  "timestamp": "2024-01-15T10:05:30Z",
  "details": "Rule with ID 'invalid_rule_id' does not exist"
}
```

**Error codes**

- `rule_not_found` — referenced rule id or cursor does not exist
- `validation_error` — payload failed schema/semantic validation
- `internal_error` — unexpected server-side failure; contact support with the `request_id`

**Common HTTP Status Codes:**
- `200 OK` - Successful operation
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid request format or parameters
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource conflict (e.g., duplicate rule name)
- `422 Unprocessable Entity` - Validation errors
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

## Rate Limiting

API endpoints are rate limited:
- **Standard operations**: 1000 requests per minute
- **Batch operations**: 100 requests per minute
- **Rule modifications**: 60 requests per minute

Rate limit headers are included in responses:
- `X-RateLimit-Limit`: Requests allowed per window
- `X-RateLimit-Remaining`: Remaining requests in the active window for the evaluated limit (per-principal or per-IP)
- `X-RateLimit-Reset`: Unix timestamp (seconds since epoch) when the evaluated window resets
- `Retry-After`: Present on `429` responses; integer seconds until the next request will be accepted

Limits are enforced per token (principal) and per source IP simultaneously. The stricter limit wins—whichever window is exhausted first returns `429` along with headers describing that window. Clients should back off using `Retry-After` and the `X-RateLimit-*` values from the response that triggered the limit.

## Pagination

List endpoints support both offset and cursor pagination.

### Offset pagination
- `limit`: Maximum items per page (default: 50, max: 1000)
- `offset`: Number of items to skip (default: 0)

Offset responses include:
```json
{
  "data": [...],
  "pagination": {
    "limit": 50,
    "offset": 0,
    "total": 1247,
    "has_next": true,
    "has_prev": false
  }
}
```

### Cursor pagination

- `cursor`: Opaque Base64 token representing the last item seen (optional)
- `limit`: Maximum items per page (default: 50, max: 500)

Results are ordered by `failed_at` ascending with `id` as a tie-breaker to guarantee stability under concurrent writes. The service returns the next cursor when more data is available and always includes the `X-Request-ID` header for tracing.

```http
GET /api/v1/dlq/entries?limit=50&cursor=eyJmYWlsZWRfYXQiOiIyMDI0LTAxLTE1VDEwOjA1OjMwWiIsImlkIjoiZGw5OTgifQ==
```

```json
{
  "data": [...],
  "page": {
    "limit": 50,
    "next_cursor": "eyJmYWlsZWRfYXQiOiIyMDI0LTAxLTE1VDEwOjA2OjAwWiIsImlkIjoiZGwxMDAifQ==",
    "prev_cursor": "eyJmYWlsZWRfYXQiOiIyMDI0LTAxLTE1VDEwOjA1OjMwWiIsImlkIjoiZGw5OTgifQ=="
  }
}
```

Clients should persist the returned `next_cursor` (and `prev_cursor` when present) and supply it on subsequent calls. To ease migration, offset parameters remain supported, but new integrations should prefer cursors.

## Rule Matcher Patterns

### Error Pattern Matching
- **Regex**: `"validation.*failed|invalid.*format"`
- **Substring**: `"connection timeout"`
- **Exact match**: `"user not found"`

### Numeric Conditions
- **Greater than**: `"> 3"`
- **Less than**: `"< 5"`
- **Equal**: `"= 0"` or `"0"`
- **Range**: `">= 1"`, `"<= 10"`

### Size Conditions
- **Bytes**: `"> 1024"`, `"< 100"`
- **KB/MB/GB**: `"> 1MB"`, `"< 500KB"`, `"> 2GB"`

### Time Patterns
- **Business hours**: `"business_hours"`
- **Weekends**: `"weekends"`
- **Nights**: `"nights"`
- **Peak hours**: `"peak_hours"`

### Duration Conditions
- **Seconds**: `"> 30s"`, `"< 5s"`
- **Minutes**: `"> 5m"`, `"< 30m"`
- **Hours**: `"> 1h"`, `"< 24h"`

## Action Types and Parameters

### Requeue Action
```json
{
  "type": "requeue",
  "parameters": {
    "target_queue": "retry_queue",
    "priority": 5,
    "delay": "5m",
    "reset_retry_count": true
  }
}
```

### Transform Action
```json
{
  "type": "transform",
  "parameters": {
    "set": {
      "processed_at": "2024-01-15T10:00:00Z",
      "retry_config.max_attempts": 5
    },
    "remove": ["debug_flag", "temporary_data"],
    "add_if_missing": {
      "timeout": 30000
    }
  }
}
```

### Redact Action
```json
{
  "type": "redact",
  "parameters": {
    "fields": ["ssn", "email", "phone", "address"],
    "replacement": "[REDACTED]"
  }
}
```

### Drop Action
```json
{
  "type": "drop",
  "parameters": {
    "reason": "Invalid job format",
    "retain_for_audit": true
  }
}
```

### Route Action
```json
{
  "type": "route",
  "parameters": {
    "rules": [
      {
        "condition": "error_type = 'timeout'",
        "target_queue": "timeout_queue"
      },
      {
        "condition": "retry_count > 3",
        "target_queue": "manual_review"
      }
    ],
    "default_queue": "general_retry"
  }
}
```

### Tag Action
```json
{
  "type": "tag",
  "parameters": {
    "tags": {
      "remediated": "true",
      "remediation_time": "2024-01-15T10:00:00Z",
      "rule_applied": "validation_fix"
    }
  }
}
```

### Notify Action
```json
{
  "type": "notify",
  "parameters": {
    "channels": ["slack://ops-alerts", "email://team@company.com"],
    "message": "Job {{.JobID}} remediated by rule {{.RuleName}}"
  }
}
```

Outbound notifications honour the following safeguards:

- **Allowlist:** Destinations must appear in `notification.allowlist`. Non-listed endpoints are rejected.
- **Timeouts:** Each channel enforces `notification.default_timeout_ms` (default `3000`) unless overridden per channel.
- **Retries:** Failures retry up to `notification.retry.max_attempts` (default `3`) with exponential backoff starting at `notification.retry.initial_delay_ms` (default `500`).
- **Notification DLQ:** Exhausted attempts enqueue payloads to `notification.dlq_key` (default `rq:dlq:notification`) for later inspection.
- **Partial failures:** The pipeline reports per-channel success and failure counts. Successful channels are not rolled back when others fail; the job result includes a `notifications` array detailing channel outcomes.

Example configuration:

```yaml
notification:
  allowlist:
    - slack://ops-alerts
    - email://team@company.com
  default_timeout_ms: 5000
  retry:
    max_attempts: 4
    initial_delay_ms: 750
  dlq_key: rq:dlq:notification
```

Pipeline result payload excerpt with partial failure reporting:

```json
{
  "job_id": "job_123",
  "notifications": [
    {"channel": "slack://ops-alerts", "status": "sent", "attempts": 1},
    {"channel": "email://team@company.com", "status": "failed", "attempts": 4, "dlq_enqueued": true}
  ]
}
```

## WebSocket Events

Real-time pipeline events are available via WebSocket at `/ws/dlq-remediation/events`. Clients must authenticate with `Authorization: Bearer <token>` (or `?token=` for service accounts); tokens follow the DLQ admin scope and expire per RBAC policy.

- Server sends a ping frame every 20s; clients must reply with pong within 10s or the connection is closed with code `4001`.
- Clients should send an application heartbeat event (`{"type":"client_heartbeat"}`) at least every 30s to confirm liveness; missing two heartbeats closes the session.
- Each connection has a send buffer of 100 events. When full, the server drops the connection with code `4008` (slow consumer). Reconnect with exponential backoff starting at 5s.
- Monitor `dlq_ws_active_connections` and `dlq_ws_dropped_connections_total` Prometheus metrics to track health.

Example event payload:

```json
{
  "type": "job_processed",
  "timestamp": "2024-01-15T10:05:30Z",
  "data": {
    "job_id": "job_123",
    "rule_id": "rule_abc123",
    "success": true,
    "actions": ["redact", "requeue"],
    "duration_ms": 145
  }
}
```

**Event Types:**
- `pipeline_started` - Pipeline started
- `pipeline_stopped` - Pipeline stopped
- `pipeline_paused` - Pipeline paused
- `pipeline_resumed` - Pipeline resumed
- `job_processed` - Job processed by pipeline
- `rule_created` - New rule created
- `rule_updated` - Rule updated
- `rule_deleted` - Rule deleted
- `batch_completed` - Batch processing completed
- `error_occurred` - Error in pipeline processing
