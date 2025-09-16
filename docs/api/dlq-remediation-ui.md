# DLQ Remediation UI API Documentation

## Overview

The DLQ Remediation UI provides a comprehensive REST API for managing Dead Letter Queue (DLQ) entries. This API allows you to list, filter, peek, requeue, and purge failed jobs with advanced pattern analysis and bulk operations.

## Authentication

Currently, the API does not require authentication. In production deployments, it is recommended to implement proper authentication and authorization mechanisms.

## Base URL

All API endpoints are prefixed with `/api/dlq`.

## Endpoints

### List DLQ Entries

Retrieve a paginated list of DLQ entries with optional filtering and pattern analysis.

**Endpoint:** `GET /api/dlq/entries`

**Query Parameters:**

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `page` | integer | Page number (1-based) | 1 |
| `page_size` | integer | Number of entries per page (max 1000) | 50 |
| `queue` | string | Filter by queue name | - |
| `type` | string | Filter by job type | - |
| `error_pattern` | string | Filter by error message pattern | - |
| `start_time` | string | Filter entries failed after this time (RFC3339) | - |
| `end_time` | string | Filter entries failed before this time (RFC3339) | - |
| `min_attempts` | integer | Filter entries with at least N attempts | - |
| `max_attempts` | integer | Filter entries with at most N attempts | - |
| `sort_by` | string | Sort field: `failed_at`, `created_at`, `queue`, `type`, `attempts` | `failed_at` |
| `sort_order` | string | Sort order: `asc`, `desc` | `desc` |
| `include_patterns` | boolean | Include error pattern analysis | false |

**Example Request:**
```bash
GET /api/dlq/entries?page=1&page_size=20&queue=payment-processing&include_patterns=true
```

**Response:**
```json
{
  "entries": [
    {
      "id": "dlq_entry_12345",
      "job_id": "job_67890",
      "type": "process_payment",
      "queue": "payment-processing",
      "payload": {
        "user_id": 123,
        "amount": 99.99,
        "currency": "USD"
      },
      "error": {
        "message": "Payment gateway timeout",
        "code": "GATEWAY_TIMEOUT",
        "context": {
          "gateway": "stripe",
          "response_time": 30000
        }
      },
      "metadata": {
        "source": "api",
        "submitted_at": "2024-01-15T10:30:00Z",
        "priority": 1,
        "tags": ["payment", "urgent"]
      },
      "attempts": [
        {
          "attempt_number": 1,
          "started_at": "2024-01-15T10:30:05Z",
          "failed_at": "2024-01-15T10:30:35Z",
          "error": "Payment gateway timeout",
          "worker_id": "worker-001",
          "duration": 30000
        }
      ],
      "created_at": "2024-01-15T10:30:00Z",
      "failed_at": "2024-01-15T10:30:35Z"
    }
  ],
  "total_count": 156,
  "page": 1,
  "page_size": 20,
  "total_pages": 8,
  "has_next": true,
  "has_previous": false,
  "patterns": [
    {
      "id": "pattern_timeout_001",
      "pattern": "payment gateway timeout",
      "message": "Payment gateway timeout",
      "count": 23,
      "first_seen": "2024-01-15T08:00:00Z",
      "last_seen": "2024-01-15T10:30:35Z",
      "affected_queues": ["payment-processing"],
      "affected_types": ["process_payment"],
      "sample_entry_ids": ["dlq_entry_12345", "dlq_entry_12346"],
      "severity": "high",
      "suggested_action": "Consider increasing timeout values or investigating network latency"
    }
  ],
  "filter": {
    "queue": "payment-processing",
    "include_patterns": true
  }
}
```

### Get Single DLQ Entry

Retrieve detailed information about a specific DLQ entry.

**Endpoint:** `GET /api/dlq/entries/{id}`

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | string | DLQ entry ID |

**Example Request:**
```bash
GET /api/dlq/entries/dlq_entry_12345
```

**Response:**
```json
{
  "id": "dlq_entry_12345",
  "job_id": "job_67890",
  "type": "process_payment",
  "queue": "payment-processing",
  "payload": {
    "user_id": 123,
    "amount": 99.99,
    "currency": "USD"
  },
  "error": {
    "message": "Payment gateway timeout",
    "code": "GATEWAY_TIMEOUT",
    "context": {
      "gateway": "stripe",
      "response_time": 30000
    }
  },
  "metadata": {
    "source": "api",
    "submitted_at": "2024-01-15T10:30:00Z",
    "priority": 1,
    "tags": ["payment", "urgent"]
  },
  "attempts": [
    {
      "attempt_number": 1,
      "started_at": "2024-01-15T10:30:05Z",
      "failed_at": "2024-01-15T10:30:35Z",
      "error": "Payment gateway timeout",
      "worker_id": "worker-001",
      "duration": 30000
    }
  ],
  "created_at": "2024-01-15T10:30:00Z",
  "failed_at": "2024-01-15T10:30:35Z"
}
```

### Requeue Entries

Requeue selected DLQ entries back to their original queues for retry.

**Endpoint:** `POST /api/dlq/entries/requeue`

**Request Body:**
```json
{
  "ids": ["dlq_entry_12345", "dlq_entry_12346"]
}
```

**Response:**
```json
{
  "total_requested": 2,
  "successful": ["dlq_entry_12345", "dlq_entry_12346"],
  "failed": [],
  "started_at": "2024-01-15T11:00:00Z",
  "completed_at": "2024-01-15T11:00:02Z",
  "duration": 2000
}
```

### Purge Entries

Permanently delete selected DLQ entries.

**Endpoint:** `POST /api/dlq/entries/purge`

**Request Body:**
```json
{
  "ids": ["dlq_entry_12345", "dlq_entry_12346"]
}
```

**Response:**
```json
{
  "total_requested": 2,
  "successful": ["dlq_entry_12345", "dlq_entry_12346"],
  "failed": [],
  "started_at": "2024-01-15T11:00:00Z",
  "completed_at": "2024-01-15T11:00:01Z",
  "duration": 1000
}
```

### Purge All Entries

Permanently delete all DLQ entries matching the specified filter criteria.

**Endpoint:** `POST /api/dlq/entries/purge-all`

**Query Parameters:**

| Parameter | Type | Description | Required |
|-----------|------|-------------|----------|
| `confirm` | string | Must be "true" to confirm the operation | Yes |
| `queue` | string | Filter by queue name | No |
| `type` | string | Filter by job type | No |
| `error_pattern` | string | Filter by error message pattern | No |
| `start_time` | string | Filter entries failed after this time (RFC3339) | No |
| `end_time` | string | Filter entries failed before this time (RFC3339) | No |
| `min_attempts` | integer | Filter entries with at least N attempts | No |
| `max_attempts` | integer | Filter entries with at most N attempts | No |

**Example Request:**
```bash
POST /api/dlq/entries/purge-all?confirm=true&queue=payment-processing&error_pattern=timeout
```

**Response:**
```json
{
  "total_requested": 23,
  "successful": ["dlq_entry_12345", "dlq_entry_12346", "..."],
  "failed": [],
  "started_at": "2024-01-15T11:00:00Z",
  "completed_at": "2024-01-15T11:00:05Z",
  "duration": 5000
}
```

### Get DLQ Statistics

Retrieve statistics about DLQ entries.

**Endpoint:** `GET /api/dlq/stats`

**Response:**
```json
{
  "total_entries": 156,
  "by_queue": {
    "payment-processing": 89,
    "email-delivery": 34,
    "data-processing": 23,
    "notifications": 10
  },
  "by_type": {
    "process_payment": 89,
    "send_email": 34,
    "process_data": 23,
    "send_notification": 10
  },
  "updated_at": "2024-01-15T11:00:00Z"
}
```

## Error Responses

All endpoints return appropriate HTTP status codes and error details in case of failures.

### Error Response Format

```json
{
  "error": "Error message",
  "details": "Detailed error information",
  "timestamp": "2024-01-15T11:00:00Z"
}
```

### Common HTTP Status Codes

| Status Code | Description |
|-------------|-------------|
| 200 | Success |
| 400 | Bad Request - Invalid parameters |
| 404 | Not Found - Entry not found |
| 500 | Internal Server Error |

## Data Types

### DLQEntry

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique DLQ entry identifier |
| `job_id` | string | Original job identifier |
| `type` | string | Job type |
| `queue` | string | Queue name |
| `payload` | object | Job payload data |
| `error` | ErrorDetails | Error information |
| `metadata` | JobMetadata | Job metadata |
| `attempts` | AttemptRecord[] | Retry attempt history |
| `created_at` | string | Entry creation timestamp (RFC3339) |
| `failed_at` | string | Last failure timestamp (RFC3339) |

### ErrorDetails

| Field | Type | Description |
|-------|------|-------------|
| `message` | string | Error message |
| `code` | string | Error code |
| `context` | object | Additional error context |

### JobMetadata

| Field | Type | Description |
|-------|------|-------------|
| `source` | string | Job source |
| `submitted_at` | string | Job submission timestamp (RFC3339) |
| `priority` | integer | Job priority |
| `tags` | string[] | Job tags |

### AttemptRecord

| Field | Type | Description |
|-------|------|-------------|
| `attempt_number` | integer | Attempt number |
| `started_at` | string | Attempt start timestamp (RFC3339) |
| `failed_at` | string | Attempt failure timestamp (RFC3339) |
| `error` | string | Error message |
| `worker_id` | string | Worker identifier |
| `duration` | integer | Attempt duration in milliseconds |

### ErrorPattern

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Pattern identifier |
| `pattern` | string | Normalized error pattern |
| `message` | string | Original error message |
| `count` | integer | Number of occurrences |
| `first_seen` | string | First occurrence timestamp (RFC3339) |
| `last_seen` | string | Last occurrence timestamp (RFC3339) |
| `affected_queues` | string[] | Affected queue names |
| `affected_types` | string[] | Affected job types |
| `sample_entry_ids` | string[] | Sample entry IDs |
| `severity` | string | Pattern severity: `low`, `medium`, `high`, `critical` |
| `suggested_action` | string | Suggested remediation action |

### BulkOperationResult

| Field | Type | Description |
|-------|------|-------------|
| `total_requested` | integer | Total number of entries requested |
| `successful` | string[] | Successfully processed entry IDs |
| `failed` | OperationError[] | Failed operations |
| `started_at` | string | Operation start timestamp (RFC3339) |
| `completed_at` | string | Operation completion timestamp (RFC3339) |
| `duration` | integer | Operation duration in milliseconds |

### OperationError

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Entry ID that failed |
| `error` | string | Error message |

## Rate Limiting

The API implements the following rate limits:

- List operations: 100 requests per minute
- Bulk operations: 10 requests per minute
- Individual operations: 1000 requests per minute

## Performance Considerations

- Use pagination for large result sets
- Enable pattern analysis only when needed (adds processing overhead)
- Consider using filters to reduce result set size
- Bulk operations are more efficient than individual operations

## Security Considerations

- Validate all input parameters
- Implement proper authentication in production
- Use HTTPS for all API communications
- Log all DLQ operations for audit purposes
- Consider implementing rate limiting and access controls

## Examples

### List recent timeout errors

```bash
curl -X GET "http://localhost:8080/api/dlq/entries?error_pattern=timeout&sort_by=failed_at&sort_order=desc&page_size=10"
```

### Requeue specific entries

```bash
curl -X POST "http://localhost:8080/api/dlq/entries/requeue" \
  -H "Content-Type: application/json" \
  -d '{"ids": ["dlq_entry_12345", "dlq_entry_12346"]}'
```

### Get statistics

```bash
curl -X GET "http://localhost:8080/api/dlq/stats"
```

### Purge all entries for a specific queue (with confirmation)

```bash
curl -X POST "http://localhost:8080/api/dlq/entries/purge-all?confirm=true&queue=test-queue"
```