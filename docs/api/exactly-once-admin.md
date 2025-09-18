# Exactly-Once Patterns Admin API

This document describes the admin API endpoints for monitoring and managing exactly-once processing patterns in the Redis Work Queue system.

## Base URL

All endpoints are prefixed with: `/api/v1/exactly-once`

## Authentication

All endpoints require admin authentication (Bearer token or API key).

## Endpoints

### 1. Get Overall Statistics

**GET** `/api/v1/exactly-once/stats`

Returns comprehensive statistics about exactly-once processing.

#### Response

```json
{
  "deduplication": {
    "processed": 15000,
    "duplicates": 342,
    "hit_percent": 2.28,
    "storage_size": 1536000,
    "active_keys": 12000
  },
  "timestamp": "2025-09-14T12:00:00Z"
}
```

#### Fields

- `processed`: Total number of unique items processed
- `duplicates`: Number of duplicate attempts blocked
- `hit_percent`: Percentage of requests that were duplicates (e.g., `2.28` = `2.28%`)
- `storage_size`: Estimated storage size in bytes
- `active_keys`: Number of active idempotency keys

---

### 2. Get Deduplication Statistics

**GET** `/api/v1/exactly-once/dedup/stats`

Returns detailed deduplication statistics.

#### Response

```json
{
  "processed": 15000,
  "duplicates": 342,
  "hit_percent": 2.28,
  "storage_size": 1536000,
  "active_keys": 12000
}
```

---

### 3. Get Pending Outbox Events

**GET** `/api/v1/exactly-once/outbox/pending`

Retrieves pending events from the transactional outbox.

#### Query Parameters

- `limit` (optional): Maximum number of events to return (default: 100)

#### Response

```json
{
  "pending_events": [
    {
      "id": "evt_123",
      "event_type": "order.created",
      "aggregate_id": "order_456",
      "created_at": "2025-09-14T11:30:00Z",
      "retry_count": 0
    }
  ],
  "count": 1,
  "limit": 100,
  "timestamp": "2025-09-14T12:00:00Z"
}
```

---

### 4. Publish Outbox Events

**POST** `/api/v1/exactly-once/outbox/publish`

Manually triggers publishing of pending outbox events.

#### Response (Success)

```json
{
  "status": "success",
  "message": "Outbox events published successfully",
  "timestamp": "2025-09-14T12:00:00Z"
}
```

#### Response (Outbox Disabled)

```json
{
  "error": "Outbox is disabled",
  "message": "Enable outbox in configuration to use this feature"
}
```

Status Code: 400

---

### 5. Cleanup Outbox Events

**POST** `/api/v1/exactly-once/outbox/cleanup`

Triggers cleanup of old processed outbox events.

#### Response (Success)

```json
{
  "status": "success",
  "message": "Outbox cleanup completed successfully",
  "timestamp": "2025-09-14T12:00:00Z"
}
```

---

### 6. Get Configuration

**GET** `/api/v1/exactly-once/config`

Returns the current exactly-once configuration.

#### Response

```json
{
  "idempotency": {
    "enabled": true,
    "default_ttl": "24h0m0s",
    "key_prefix": "idempotency:",
    "storage_type": "redis"
  },
  "outbox": {
    "enabled": false,
    "storage_type": "redis",
    "batch_size": 50,
    "poll_interval": "5s",
    "max_retries": 5
  },
  "metrics": {
    "enabled": true
  }
}
```

---

### 7. Update Configuration

**PUT** `/api/v1/exactly-once/config`

Updates the exactly-once configuration.

#### Request Body

```json
{
  "idempotency": {
    "enabled": false
  },
  "outbox": {
    "batch_size": 100
  }
}
```

#### Response

```json
{
  "status": "success",
  "message": "Configuration update acknowledged",
  "timestamp": "2025-09-14T12:00:00Z"
}
```

---

### 8. Health Check

**GET** `/api/v1/exactly-once/health`

Performs health check on the exactly-once subsystem.

#### Response (Healthy)

```json
{
  "status": "healthy",
  "components": {
    "redis": {
      "healthy": true,
      "error": ""
    },
    "deduplication": {
      "healthy": true,
      "error": ""
    },
    "outbox": {
      "healthy": true,
      "enabled": true
    }
  },
  "timestamp": "2025-09-14T12:00:00Z"
}
```

Status Code: 200

#### Response (Unhealthy)

```json
{
  "status": "unhealthy",
  "components": {
    "redis": {
      "healthy": false,
      "error": "connection refused"
    },
    "deduplication": {
      "healthy": false,
      "error": "redis unavailable"
    },
    "outbox": {
      "healthy": true,
      "enabled": false
    }
  },
  "timestamp": "2025-09-14T12:00:00Z"
}
```

Status Code: 503

## Error Responses

Errors follow the shared envelope and include an `X-Request-ID` header.

```json
{
  "code": "OUTBOX_DISABLED",
  "error": "Outbox is disabled for namespace default",
  "status": 400,
  "request_id": "b5790c1d-5c6a-41f2-a4fb-0d508593f664",
  "timestamp": "2025-09-14T12:05:00Z",
  "details": {
    "hint": "Enable outbox before publishing"
  }
}
```

Common error codes:

- `VALIDATION_ERROR` – payload failed validation
- `OUTBOX_DISABLED` – attempted to publish with the outbox turned off
- `AUTH_INVALID` / `AUTH_MISSING` – authentication failures
- `RATE_LIMIT` – exceeded API limits (check `details.retry_after`)
- `INTERNAL_ERROR` – unexpected failure; retry with the logged request ID

## Usage Examples

### cURL Examples

#### Get Statistics
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/exactly-once/stats
```

#### Trigger Outbox Publishing
```bash
curl -X POST -H "Authorization: Bearer ${API_TOKEN}" \
  http://localhost:8080/api/v1/exactly-once/outbox/publish
```

#### Update Configuration
```bash
curl -X PUT -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"idempotency": {"enabled": true}}' \
  http://localhost:8080/api/v1/exactly-once/config
```

#### Health Check
```bash
curl -H "Authorization: Bearer ${API_TOKEN}" \
  http://localhost:8080/api/v1/exactly-once/health
```

> Replace `${API_TOKEN}` with a real admin token or read it from your environment (e.g., `export API_TOKEN=...`) before invoking these commands.

## Integration with TUI

The TUI (Terminal User Interface) can use these endpoints to:

1. **Display real-time statistics** - Poll `/stats` endpoint
2. **Monitor outbox queue** - Check `/outbox/pending` for pending events
3. **Trigger manual operations** - Call `/outbox/publish` or `/outbox/cleanup`
4. **Health monitoring** - Regular health checks via `/health`
5. **Configuration management** - View and update settings via `/config`

## Metrics and Monitoring

These endpoints provide key metrics for monitoring:

- **Deduplication effectiveness**: `hit_percent` from `/stats` (percentage value; lower is better)
- **Processing volume**: `processed` count from `/dedup/stats`
- **Storage usage**: `storage_size` and `active_keys`
- **System health**: Component status from `/health`
- **Outbox backlog**: Event count from `/outbox/pending`

## Best Practices

1. **Regular Health Checks**: Poll `/health` every 30 seconds for monitoring
2. **Statistics Collection**: Collect `/stats` every minute for metrics
3. **Outbox Monitoring**: Check pending events if outbox is enabled
4. **Configuration Validation**: Always verify configuration changes via GET after PUT
5. **Error Handling**: Implement exponential backoff for failed requests

## Performance Considerations

- **Stats Endpoint**: Cached for 1 second, safe to poll frequently
- **Pending Events**: Limited by query parameter, use pagination for large datasets
- **Health Check**: Lightweight, suitable for frequent monitoring
- **Configuration Updates**: Changes may take up to 5 seconds to propagate

---

*Last Updated: 2025-09-14*
*API Version: 1.0.0*
