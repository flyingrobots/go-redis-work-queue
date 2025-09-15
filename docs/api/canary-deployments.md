# Canary Deployments API Documentation

## Overview

The Canary Deployments API provides comprehensive functionality for safely rolling out new worker versions through traffic splitting, health monitoring, and automated promotion/rollback capabilities. This system enables zero-downtime deployments with built-in safety controls and real-time monitoring.

## Base URL

```
http://localhost:8080/api/v1/canary
```

## Authentication

The API supports multiple authentication methods:

- **Bearer Token**: Include `Authorization: Bearer <token>` header
- **API Key**: Include `X-API-Key: <key>` header

## Content Type

All requests and responses use `application/json` content type unless otherwise specified.

## Error Handling

All error responses follow this format:

```json
{
  "error": "Error description",
  "code": "ERROR_CODE",
  "details": {
    "field": "specific error details"
  }
}
```

### Error Codes

| Code | Description | HTTP Status |
|------|-------------|-------------|
| `DEPLOYMENT_NOT_FOUND` | Deployment does not exist | 404 |
| `DEPLOYMENT_EXISTS` | Deployment already exists | 409 |
| `INVALID_PERCENTAGE` | Percentage must be 0-100 | 400 |
| `VALIDATION_FAILED` | Request validation failed | 400 |
| `CONCURRENCY_LIMIT` | Too many active deployments | 429 |
| `DEPLOYMENT_NOT_ACTIVE` | Deployment is not in active state | 400 |

## API Endpoints

### Deployment Management

#### List Deployments

```http
GET /deployments
```

Retrieves all canary deployments with their current status.

**Response:**
```json
{
  "deployments": [
    {
      "id": "canary_123e4567-e89b-12d3-a456-426614174000",
      "queue_name": "payment-processing",
      "tenant_id": "acme-corp",
      "stable_version": "v1.2.0",
      "canary_version": "v1.3.0",
      "current_percent": 25,
      "target_percent": 25,
      "status": "active",
      "start_time": "2025-09-14T12:00:00Z",
      "last_update": "2025-09-14T12:15:00Z",
      "config": {
        "routing_strategy": "split_queue",
        "sticky_routing": true,
        "auto_promotion": false,
        "max_canary_duration": "2h0m0s",
        "min_canary_duration": "5m0s"
      },
      "created_by": "ops-team@acme.com"
    }
  ],
  "count": 1
}
```

#### Create Deployment

```http
POST /deployments
```

Creates a new canary deployment with specified configuration.

**Request Body:**
```json
{
  "queue_name": "payment-processing",
  "tenant_id": "acme-corp",
  "stable_version": "v1.2.0",
  "canary_version": "v1.3.0",
  "routing_strategy": "split_queue",
  "sticky_routing": true,
  "auto_promotion": false,
  "max_duration": "2h",
  "min_duration": "5m",
  "metrics_window": "5m",
  "created_by": "ops-team@acme.com",
  "profile": "conservative"
}
```

**Parameters:**
- `queue_name` (string, required): Target queue for canary deployment
- `tenant_id` (string, optional): Tenant identifier for multi-tenant setups
- `stable_version` (string, required): Current stable version identifier
- `canary_version` (string, required): New version to deploy as canary
- `routing_strategy` (string, optional): `split_queue`, `stream_group`, or `hash_ring`
- `sticky_routing` (boolean, optional): Enable consistent job routing
- `auto_promotion` (boolean, optional): Enable automatic promotion based on metrics
- `max_duration` (string, optional): Maximum canary duration (e.g., "2h", "30m")
- `min_duration` (string, optional): Minimum canary duration before promotion
- `metrics_window` (string, optional): Metrics collection window
- `profile` (string, optional): Configuration profile: `default`, `conservative`, `aggressive`

**Response:** Returns the created deployment object (201 Created).

#### Get Deployment

```http
GET /deployments/{id}
```

Retrieves detailed information about a specific deployment.

**Parameters:**
- `id` (path, required): Deployment ID

**Response:** Returns the deployment object with current metrics and status.

#### Update Traffic Percentage

```http
PUT /deployments/{id}/percentage
```

Updates the traffic split percentage for a canary deployment.

**Parameters:**
- `id` (path, required): Deployment ID

**Request Body:**
```json
{
  "percentage": 50
}
```

**Response:**
```json
{
  "success": true,
  "percentage": 50,
  "updated_at": "2025-09-14T12:30:00Z"
}
```

#### Promote Deployment

```http
POST /deployments/{id}/promote
```

Promotes the canary to 100% traffic and marks deployment as completed.

**Parameters:**
- `id` (path, required): Deployment ID

**Response:**
```json
{
  "success": true,
  "message": "Deployment promoted to 100%",
  "timestamp": "2025-09-14T12:45:00Z"
}
```

#### Rollback Deployment

```http
POST /deployments/{id}/rollback
```

Rolls back the canary to 0% traffic and marks deployment as failed.

**Parameters:**
- `id` (path, required): Deployment ID

**Request Body:**
```json
{
  "reason": "High error rate detected in canary version"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Deployment rolled back: High error rate detected in canary version",
  "reason": "High error rate detected in canary version",
  "timestamp": "2025-09-14T12:50:00Z"
}
```

#### Delete Deployment

```http
DELETE /deployments/{id}
```

Deletes a completed or failed deployment. Active deployments cannot be deleted.

**Parameters:**
- `id` (path, required): Deployment ID

**Response:** 204 No Content on success.

### Health and Monitoring

#### Get Deployment Health

```http
GET /deployments/{id}/health
```

Retrieves the current health status of a canary deployment based on SLO thresholds.

**Parameters:**
- `id` (path, required): Deployment ID

**Response:**
```json
{
  "overall_status": "healthy",
  "error_rate_check": {
    "name": "Error Rate",
    "passing": true,
    "message": "Error rate increase: 0.5% (threshold: 5.0%)",
    "timestamp": "2025-09-14T12:45:00Z"
  },
  "latency_check": {
    "name": "P95 Latency",
    "passing": true,
    "message": "P95 latency increase: 8.2% (threshold: 50.0%)",
    "timestamp": "2025-09-14T12:45:00Z"
  },
  "throughput_check": {
    "name": "Throughput",
    "passing": true,
    "message": "Throughput decrease: -5.1% (threshold: 20.0%)",
    "timestamp": "2025-09-14T12:45:00Z"
  },
  "sample_size_check": {
    "name": "Sample Size",
    "passing": true,
    "message": "Sample size: 156 (required: 20)",
    "timestamp": "2025-09-14T12:45:00Z"
  },
  "duration_check": {
    "name": "Duration",
    "passing": true,
    "message": "Duration: 12m30s (minimum: 5m0s)",
    "timestamp": "2025-09-14T12:45:00Z"
  },
  "last_evaluation": "2025-09-14T12:45:00Z"
}
```

**Health Status Values:**
- `healthy`: All checks passing, deployment proceeding normally
- `warning`: Some degradation detected, monitoring closely
- `failing`: Critical issues detected, rollback recommended
- `unknown`: Insufficient data for health assessment

#### Get Deployment Metrics

```http
GET /deployments/{id}/metrics
```

Retrieves current performance metrics for both stable and canary versions.

**Parameters:**
- `id` (path, required): Deployment ID

**Response:**
```json
{
  "stable": {
    "timestamp": "2025-09-14T12:45:00Z",
    "window_start": "2025-09-14T12:40:00Z",
    "window_end": "2025-09-14T12:45:00Z",
    "job_count": 1250,
    "success_count": 1238,
    "error_count": 12,
    "error_rate": 0.96,
    "success_rate": 99.04,
    "avg_latency": 245.6,
    "p50_latency": 198.3,
    "p95_latency": 456.7,
    "p99_latency": 782.1,
    "max_latency": 1234.5,
    "jobs_per_second": 4.17,
    "avg_memory_mb": 128.4,
    "peak_memory_mb": 186.2,
    "queue_depth": 45,
    "worker_count": 8,
    "version": "v1.2.0"
  },
  "canary": {
    "timestamp": "2025-09-14T12:45:00Z",
    "window_start": "2025-09-14T12:40:00Z",
    "window_end": "2025-09-14T12:45:00Z",
    "job_count": 312,
    "success_count": 309,
    "error_count": 3,
    "error_rate": 0.96,
    "success_rate": 99.04,
    "avg_latency": 231.2,
    "p50_latency": 189.5,
    "p95_latency": 423.8,
    "p99_latency": 698.4,
    "max_latency": 987.6,
    "jobs_per_second": 1.04,
    "avg_memory_mb": 142.1,
    "peak_memory_mb": 201.3,
    "queue_depth": 12,
    "worker_count": 2,
    "version": "v1.3.0"
  }
}
```

#### Get Deployment Events

```http
GET /deployments/{id}/events
```

Retrieves the event history for a deployment, showing all significant actions and state changes.

**Parameters:**
- `id` (path, required): Deployment ID

**Response:**
```json
{
  "events": [
    {
      "id": "event_987fcdeb-51a2-43d8-b7e1-123456789abc",
      "deployment_id": "canary_123e4567-e89b-12d3-a456-426614174000",
      "type": "percentage_updated",
      "message": "Traffic split updated to 25%",
      "timestamp": "2025-09-14T12:15:00Z"
    },
    {
      "id": "event_456789ab-cdef-1234-5678-9abcdef01234",
      "deployment_id": "canary_123e4567-e89b-12d3-a456-426614174000",
      "type": "deployment_created",
      "message": "Canary deployment created",
      "timestamp": "2025-09-14T12:00:00Z"
    }
  ],
  "count": 2
}
```

**Event Types:**
- `deployment_created`: New canary deployment started
- `percentage_updated`: Traffic split percentage changed
- `deployment_promoted`: Canary promoted to 100%
- `deployment_rolled_back`: Canary rolled back due to issues
- `health_check_failed`: Automated health check detected problems
- `auto_promotion_triggered`: Automatic promotion rules activated

### Worker Management

#### List Workers

```http
GET /workers?lane={lane}
```

Retrieves information about registered workers, optionally filtered by lane.

**Query Parameters:**
- `lane` (string, optional): Filter by worker lane (`stable` or `canary`)

**Response:**
```json
{
  "workers": [
    {
      "id": "worker-001-stable",
      "version": "v1.2.0",
      "lane": "stable",
      "queues": ["payment-processing", "order-fulfillment"],
      "last_seen": "2025-09-14T12:44:30Z",
      "status": "healthy",
      "metrics": {
        "jobs_processed": 1250,
        "jobs_succeeded": 1238,
        "jobs_failed": 12,
        "avg_processing_time": 245.6,
        "memory_usage_mb": 128.4,
        "cpu_usage_percent": 45.2,
        "last_job_at": "2025-09-14T12:44:15Z"
      }
    }
  ],
  "count": 1
}
```

#### Register Worker

```http
POST /workers
```

Registers a new worker with the canary deployment system.

**Request Body:**
```json
{
  "id": "worker-003-canary",
  "version": "v1.3.0",
  "lane": "canary",
  "queues": ["payment-processing"],
  "metadata": {
    "hostname": "worker-host-03",
    "region": "us-west-2",
    "instance_type": "m5.large"
  }
}
```

**Parameters:**
- `id` (string, required): Unique worker identifier
- `version` (string, required): Worker version
- `lane` (string, optional): Worker lane (`stable` or `canary`), defaults to `stable`
- `queues` (array, required): List of queues this worker can process
- `metadata` (object, optional): Additional worker metadata

**Response:** Returns the registered worker object (201 Created).

#### Update Worker Status

```http
PUT /workers/{id}/status
```

Updates the health status of a worker.

**Parameters:**
- `id` (path, required): Worker ID

**Request Body:**
```json
{
  "status": "degraded"
}
```

**Status Values:**
- `healthy`: Worker operating normally
- `degraded`: Worker experiencing some issues but still functional
- `unhealthy`: Worker has significant problems
- `unreachable`: Worker is not responding

**Response:**
```json
{
  "success": true,
  "status": "degraded",
  "updated_at": "2025-09-14T12:50:00Z"
}
```

### Configuration

#### Get Configuration Profiles

```http
GET /config/profiles
```

Retrieves available canary deployment configuration profiles.

**Response:**
```json
{
  "default": {
    "routing_strategy": "split_queue",
    "sticky_routing": true,
    "auto_promotion": false,
    "max_canary_duration": "2h0m0s",
    "min_canary_duration": "5m0s",
    "drain_timeout": "5m0s",
    "metrics_window": "5m0s",
    "rollback_thresholds": {
      "max_error_rate_increase": 5.0,
      "max_latency_increase": 50.0,
      "max_throughput_decrease": 20.0,
      "min_success_rate": 95.0,
      "required_sample_size": 20
    }
  },
  "conservative": {
    "routing_strategy": "split_queue",
    "sticky_routing": true,
    "auto_promotion": false,
    "max_canary_duration": "4h0m0s",
    "min_canary_duration": "15m0s",
    "rollback_thresholds": {
      "max_error_rate_increase": 2.0,
      "max_latency_increase": 20.0,
      "max_throughput_decrease": 10.0,
      "min_success_rate": 95.0,
      "required_sample_size": 20
    }
  },
  "aggressive": {
    "routing_strategy": "split_queue",
    "sticky_routing": false,
    "auto_promotion": true,
    "max_canary_duration": "30m0s",
    "min_canary_duration": "2m0s",
    "rollback_thresholds": {
      "max_error_rate_increase": 20.0,
      "max_latency_increase": 200.0,
      "max_throughput_decrease": 75.0,
      "min_success_rate": 70.0,
      "required_sample_size": 5
    }
  }
}
```

## Data Models

### Deployment Object

```json
{
  "id": "string",
  "queue_name": "string",
  "tenant_id": "string",
  "stable_version": "string",
  "canary_version": "string",
  "current_percent": 0,
  "target_percent": 0,
  "status": "active|promoting|rolling_back|completed|failed|paused",
  "start_time": "2025-09-14T12:00:00Z",
  "last_update": "2025-09-14T12:00:00Z",
  "completed_at": "2025-09-14T12:00:00Z",
  "config": {
    "routing_strategy": "split_queue|stream_group|hash_ring",
    "sticky_routing": true,
    "auto_promotion": false,
    "max_canary_duration": "2h0m0s",
    "min_canary_duration": "5m0s",
    "drain_timeout": "5m0s",
    "metrics_window": "5m0s",
    "rollback_thresholds": {
      "max_error_rate_increase": 5.0,
      "max_latency_increase": 50.0,
      "max_throughput_decrease": 20.0,
      "min_success_rate": 95.0,
      "required_sample_size": 20
    }
  },
  "created_by": "string",
  "metadata": {}
}
```

### Metrics Snapshot

```json
{
  "timestamp": "2025-09-14T12:45:00Z",
  "window_start": "2025-09-14T12:40:00Z",
  "window_end": "2025-09-14T12:45:00Z",
  "job_count": 1250,
  "success_count": 1238,
  "error_count": 12,
  "error_rate": 0.96,
  "success_rate": 99.04,
  "avg_latency": 245.6,
  "p50_latency": 198.3,
  "p95_latency": 456.7,
  "p99_latency": 782.1,
  "max_latency": 1234.5,
  "jobs_per_second": 4.17,
  "avg_memory_mb": 128.4,
  "peak_memory_mb": 186.2,
  "avg_cpu_percent": 45.2,
  "queue_depth": 45,
  "dead_letters": 2,
  "worker_count": 8,
  "version": "v1.2.0"
}
```

### Worker Object

```json
{
  "id": "string",
  "version": "string",
  "lane": "stable|canary",
  "queues": ["string"],
  "last_seen": "2025-09-14T12:44:30Z",
  "status": "healthy|degraded|unhealthy|unreachable",
  "metrics": {
    "jobs_processed": 1250,
    "jobs_succeeded": 1238,
    "jobs_failed": 12,
    "avg_processing_time": 245.6,
    "memory_usage_mb": 128.4,
    "cpu_usage_percent": 45.2,
    "last_job_at": "2025-09-14T12:44:15Z"
  },
  "metadata": {}
}
```

## Usage Examples

### Basic Canary Deployment

1. **Create a canary deployment:**
```bash
curl -X POST http://localhost:8080/api/v1/canary/deployments \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "queue_name": "payment-processing",
    "stable_version": "v1.2.0",
    "canary_version": "v1.3.0",
    "profile": "default"
  }'
```

2. **Gradually increase traffic:**
```bash
curl -X PUT http://localhost:8080/api/v1/canary/deployments/{id}/percentage \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"percentage": 10}'
```

3. **Monitor health:**
```bash
curl -X GET http://localhost:8080/api/v1/canary/deployments/{id}/health \
  -H "Authorization: Bearer <token>"
```

4. **Promote or rollback:**
```bash
# Promote to 100%
curl -X POST http://localhost:8080/api/v1/canary/deployments/{id}/promote \
  -H "Authorization: Bearer <token>"

# Or rollback if issues detected
curl -X POST http://localhost:8080/api/v1/canary/deployments/{id}/rollback \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"reason": "High error rate detected"}'
```

### Conservative Deployment

For critical systems, use the conservative profile:

```bash
curl -X POST http://localhost:8080/api/v1/canary/deployments \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "queue_name": "critical-payments",
    "stable_version": "v2.1.0",
    "canary_version": "v2.1.1",
    "profile": "conservative",
    "max_duration": "4h",
    "min_duration": "15m"
  }'
```

### Automated Deployment

For non-critical systems with auto-promotion:

```bash
curl -X POST http://localhost:8080/api/v1/canary/deployments \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "queue_name": "analytics-processing",
    "stable_version": "v1.5.0",
    "canary_version": "v1.6.0",
    "profile": "aggressive",
    "auto_promotion": true
  }'
```

## Rate Limiting

The API implements rate limiting with the following defaults:
- 100 requests per minute per API key
- Burst size of 20 requests
- Rate limit headers are included in responses:
  - `X-RateLimit-Limit`: Request limit per window
  - `X-RateLimit-Remaining`: Requests remaining in window
  - `X-RateLimit-Reset`: Time when limit resets (Unix timestamp)

## Pagination

List endpoints support pagination using query parameters:
- `page`: Page number (default: 1)
- `page_size`: Items per page (default: 10, max: 100)

Paginated responses include metadata:
```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total_pages": 5,
    "total_items": 47,
    "has_next": true,
    "has_prev": false
  }
}
```

## WebSocket Events

Real-time updates are available via WebSocket connections:

```
ws://localhost:8080/api/v1/canary/deployments/{id}/events
```

**Event Types:**
- `deployment_status_changed`
- `percentage_updated`
- `health_status_changed`
- `metrics_updated`
- `promotion_triggered`
- `rollback_triggered`

**Event Format:**
```json
{
  "type": "percentage_updated",
  "deployment_id": "canary_123...",
  "timestamp": "2025-09-14T12:45:00Z",
  "data": {
    "old_percentage": 10,
    "new_percentage": 25
  }
}
```

## Best Practices

### Traffic Ramping Strategy

1. **Start Small**: Begin with 5-10% traffic
2. **Monitor Closely**: Watch metrics for 10-15 minutes at each level
3. **Gradual Increase**: Use increments of 10-25%
4. **Key Milestones**: Pay special attention at 25%, 50%, and 75%
5. **Business Hours**: Avoid major promotions outside business hours

### Health Monitoring

1. **Key Metrics**: Focus on error rate, latency (P95), and throughput
2. **Sample Size**: Ensure sufficient sample size before making decisions
3. **Duration**: Allow minimum duration for statistical significance
4. **Alerting**: Set up alerts for automatic notifications
5. **Manual Override**: Always have manual promotion/rollback capability

### Configuration Profiles

- **Conservative**: Use for critical systems, payments, user authentication
- **Default**: Use for most business logic and API endpoints
- **Aggressive**: Use for analytics, reporting, and non-critical features

### Worker Management

1. **Version Consistency**: Ensure workers properly report their version
2. **Health Checks**: Implement regular health reporting
3. **Graceful Shutdown**: Support graceful job completion during rollback
4. **Resource Monitoring**: Track memory and CPU usage
5. **Queue Affinity**: Consider worker capabilities when assigning queues