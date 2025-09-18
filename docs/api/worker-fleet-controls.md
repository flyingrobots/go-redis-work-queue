# Worker Fleet Controls API Documentation

## Overview

The Worker Fleet Controls API provides comprehensive management capabilities for worker fleets, including listing workers, controlling worker states (pause/resume/drain/stop), rolling restarts, health monitoring, and audit logging.

## Authentication

Currently, the API does not require authentication. In production deployments, it is recommended to implement proper authentication and authorization mechanisms.

## Base URL

All API endpoints are prefixed with `/api/workers`.

## Core Concepts

### Worker States

- `running`: Worker is actively processing jobs
- `paused`: Worker has stopped accepting new jobs but may finish current job
- `draining`: Worker will finish current job and then stop
- `stopped`: Worker has completely stopped
- `offline`: Worker has not sent heartbeat within timeout period
- `unknown`: Worker state is undetermined

### Safety Features

- **Confirmation Requirements**: Destructive actions require explicit confirmation
- **Fleet Health Checks**: Prevent actions that would compromise fleet availability
- **Percentage Limits**: Limit the percentage of workers that can be affected
- **Minimum Healthy Workers**: Ensure minimum number of healthy workers remain

## Endpoints

### List Workers

Retrieve a paginated list of workers with filtering and sorting capabilities.

**Endpoint:** `GET /api/workers`

**Query Parameters:**

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `page` | integer | Page number (1-based) | 1 |
| `page_size` | integer | Number of workers per page (max 1000) | 50 |
| `states` | string | Comma-separated worker states to filter by | - |
| `hostname` | string | Filter by hostname | - |
| `version` | string | Filter by worker version | - |
| `sort_by` | string | Sort field: `id`, `state`, `last_heartbeat`, `hostname`, etc. | `last_heartbeat` |
| `sort_order` | string | Sort order: `asc`, `desc` | `desc` |

**Example Request:**
```bash
GET /api/workers?page=1&page_size=20&states=running,paused&sort_by=last_heartbeat
```

**Response:**
```json
{
  "workers": [
    {
      "id": "worker-001",
      "state": "running",
      "last_heartbeat": "2024-01-15T12:30:00Z",
      "started_at": "2024-01-15T08:00:00Z",
      "version": "1.2.3",
      "hostname": "worker-node-01",
      "pid": 12345,
      "current_job": {
        "id": "job-789",
        "type": "data_processing",
        "queue": "high-priority",
        "started_at": "2024-01-15T12:25:00Z",
        "estimated_eta": "2024-01-15T12:35:00Z",
        "progress": {
          "percentage": 75.0,
          "stage": "processing",
          "message": "Processing batch 3 of 4",
          "updated_at": "2024-01-15T12:28:00Z"
        }
      },
      "capabilities": ["golang", "redis", "postgres"],
      "stats": {
        "jobs_processed": 1250,
        "jobs_successful": 1238,
        "jobs_failed": 12,
        "total_runtime": 14400000000000,
        "average_job_time": 30000000000,
        "memory_usage": 536870912,
        "cpu_usage": 25.5,
        "goroutine_count": 15
      },
      "config": {
        "max_concurrent_jobs": 3,
        "queues": ["high-priority", "default"],
        "job_types": ["data_processing", "email_sending"],
        "heartbeat_interval": 30000000000,
        "graceful_timeout": 60000000000
      },
      "labels": {
        "env": "production",
        "role": "data-worker",
        "zone": "us-east-1a"
      },
      "health": {
        "status": "healthy",
        "last_check": "2024-01-15T12:29:00Z",
        "checks": {
          "redis": {
            "name": "redis",
            "status": "healthy",
            "message": "Connected",
            "timestamp": "2024-01-15T12:29:00Z",
            "duration": 5000000
          }
        },
        "error_count": 0,
        "recovery_count": 2
      }
    }
  ],
  "total_count": 156,
  "page": 1,
  "page_size": 20,
  "total_pages": 8,
  "has_next": true,
  "has_previous": false,
  "filter": {
    "states": ["running", "paused"]
  },
  "summary": {
    "total_workers": 156,
    "state_distribution": {
      "running": 145,
      "paused": 8,
      "draining": 2,
      "stopped": 1
    },
    "health_distribution": {
      "healthy": 150,
      "degraded": 4,
      "unhealthy": 2
    },
    "active_jobs": 87,
    "average_load": 22.3,
    "updated_at": "2024-01-15T12:30:00Z"
  }
}
```

### Get Single Worker

Retrieve detailed information about a specific worker.

**Endpoint:** `GET /api/workers/{id}`

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | string | Worker ID |

**Example Request:**
```bash
GET /api/workers/worker-001
```

**Response:**
```json
{
  "id": "worker-001",
  "state": "running",
  "last_heartbeat": "2024-01-15T12:30:00Z",
  // ... full worker details as shown in list response
}
```

### Register Worker

Register a new worker with the fleet.

**Endpoint:** `POST /api/workers/register`

**Request Body:**
```json
{
  "id": "worker-005",
  "hostname": "worker-node-05",
  "version": "1.2.3",
  "pid": 23456,
  "capabilities": ["golang", "redis"],
  "config": {
    "max_concurrent_jobs": 3,
    "queues": ["default"],
    "heartbeat_interval": 30000000000
  },
  "labels": {
    "env": "production",
    "role": "worker"
  }
}
```

**Response:**
```json
{
  "success": true,
  "worker_id": "worker-005",
  "message": "Worker registered successfully"
}
```

### Update Worker Heartbeat

Update a worker's heartbeat timestamp and current job status.

**Endpoint:** `POST /api/workers/{id}/heartbeat`

**Request Body:**
```json
{
  "timestamp": "2024-01-15T12:30:00Z",
  "current_job": {
    "id": "job-789",
    "type": "data_processing",
    "queue": "high-priority",
    "started_at": "2024-01-15T12:25:00Z",
    "progress": {
      "percentage": 75.0,
      "stage": "processing",
      "message": "Processing batch 3 of 4"
    }
  }
}
```

**Response:**
```json
{
  "success": true,
  "message": "Heartbeat updated successfully"
}
```

### Pause Workers

Pause one or more workers. Paused workers stop accepting new jobs but may finish their current job.

**Endpoint:** `POST /api/workers/actions/pause`

**Request Body:**
```json
{
  "worker_ids": ["worker-001", "worker-002"],
  "reason": "Maintenance window"
}
```

**Response:**
```json
{
  "request_id": "req-12345",
  "action": "pause",
  "total_requested": 2,
  "successful": ["worker-001", "worker-002"],
  "failed": [],
  "in_progress": [],
  "started_at": "2024-01-15T12:30:00Z",
  "completed_at": "2024-01-15T12:30:02Z",
  "status": "completed"
}
```

### Resume Workers

Resume one or more paused workers.

**Endpoint:** `POST /api/workers/actions/resume`

**Request Body:**
```json
{
  "worker_ids": ["worker-001", "worker-002"],
  "reason": "Maintenance complete"
}
```

**Response:**
```json
{
  "request_id": "req-12346",
  "action": "resume",
  "total_requested": 2,
  "successful": ["worker-001", "worker-002"],
  "failed": [],
  "in_progress": [],
  "started_at": "2024-01-15T12:35:00Z",
  "completed_at": "2024-01-15T12:35:01Z",
  "status": "completed"
}
```

### Drain Workers

Drain one or more workers. Drained workers finish their current job and then stop.

**Endpoint:** `POST /api/workers/actions/drain`

**Request Body:**
```json
{
  "worker_ids": ["worker-001", "worker-002"],
  "reason": "Deployment",
  "timeout_seconds": 300,
  "confirmation": "CONFIRM"
}
```

**Response:**
```json
{
  "request_id": "req-12347",
  "action": "drain",
  "total_requested": 2,
  "successful": ["worker-001"],
  "failed": [
    {
      "worker_id": "worker-002",
      "error": "Worker not responding",
      "code": "TIMEOUT"
    }
  ],
  "in_progress": [],
  "started_at": "2024-01-15T12:40:00Z",
  "completed_at": "2024-01-15T12:43:30Z",
  "estimated_eta": "2024-01-15T12:45:00Z",
  "status": "completed"
}
```

### Stop Workers

Stop one or more workers immediately or gracefully.

**Endpoint:** `POST /api/workers/actions/stop`

**Request Body:**
```json
{
  "worker_ids": ["worker-003"],
  "reason": "Emergency stop",
  "force": false,
  "confirmation": "CONFIRM"
}
```

**Response:**
```json
{
  "request_id": "req-12348",
  "action": "stop",
  "total_requested": 1,
  "successful": ["worker-003"],
  "failed": [],
  "in_progress": [],
  "started_at": "2024-01-15T12:45:00Z",
  "completed_at": "2024-01-15T12:45:05Z",
  "status": "completed"
}
```

### Restart Workers

Restart one or more workers.

**Endpoint:** `POST /api/workers/actions/restart`

**Request Body:**
```json
{
  "worker_ids": ["worker-001", "worker-002"],
  "reason": "Version update",
  "confirmation": "CONFIRM"
}
```

**Response:**
```json
{
  "request_id": "req-12349",
  "action": "restart",
  "total_requested": 2,
  "successful": ["worker-001", "worker-002"],
  "failed": [],
  "in_progress": [],
  "started_at": "2024-01-15T12:50:00Z",
  "status": "in_progress"
}
```

### Rolling Restart

Perform a rolling restart of workers matching filter criteria.

**Endpoint:** `POST /api/workers/actions/rolling-restart`

**Request Body:**
```json
{
  "filter": {
    "labels": {
      "env": "production"
    },
    "states": ["running"]
  },
  "concurrency": 2,
  "drain_timeout": 300000000000,
  "restart_timeout": 120000000000,
  "max_unavailable": 3,
  "health_checks": true,
  "confirmation": "CONFIRM"
}
```

**Response:**
```json
{
  "request_id": "req-12350",
  "total_workers": 10,
  "phases": [
    {
      "phase_number": 1,
      "worker_ids": ["worker-001", "worker-002"],
      "status": "completed",
      "started_at": "2024-01-15T13:00:00Z",
      "completed_at": "2024-01-15T13:02:30Z",
      "errors": []
    },
    {
      "phase_number": 2,
      "worker_ids": ["worker-003", "worker-004"],
      "status": "in_progress",
      "started_at": "2024-01-15T13:02:30Z",
      "errors": []
    }
  ],
  "current_phase": 1,
  "status": "in_progress",
  "started_at": "2024-01-15T13:00:00Z",
  "estimated_eta": "2024-01-15T13:15:00Z",
  "success_count": 2,
  "failure_count": 0
}
```

### Get Action Status

Check the status of a running action.

**Endpoint:** `GET /api/workers/actions/{request_id}`

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `request_id` | string | Action request ID |

**Example Request:**
```bash
GET /api/workers/actions/req-12350
```

**Response:**
```json
{
  "request_id": "req-12350",
  "action": "restart",
  "total_requested": 2,
  "successful": ["worker-001"],
  "failed": [],
  "in_progress": ["worker-002"],
  "started_at": "2024-01-15T12:50:00Z",
  "estimated_eta": "2024-01-15T12:52:00Z",
  "status": "in_progress"
}
```

### Cancel Action

Cancel a running action.

**Endpoint:** `POST /api/workers/actions/{request_id}/cancel`

**Response:**
```json
{
  "success": true,
  "message": "Action cancelled successfully"
}
```

### Get Fleet Summary

Get high-level fleet statistics.

**Endpoint:** `GET /api/workers/summary`

**Response:**
```json
{
  "total_workers": 156,
  "state_distribution": {
    "running": 145,
    "paused": 8,
    "draining": 2,
    "stopped": 1
  },
  "health_distribution": {
    "healthy": 150,
    "degraded": 4,
    "unhealthy": 2
  },
  "active_jobs": 87,
  "average_load": 22.3,
  "updated_at": "2024-01-15T12:30:00Z"
}
```

### Get Audit Logs

Retrieve audit logs for worker actions.

**Endpoint:** `GET /api/workers/audit`

**Query Parameters:**

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `limit` | integer | Maximum number of logs to return | 100 |
| `offset` | integer | Number of logs to skip | 0 |
| `start_time` | string | Filter logs after this time (RFC3339) | - |
| `end_time` | string | Filter logs before this time (RFC3339) | - |
| `actions` | string | Comma-separated actions to filter by | - |
| `worker_ids` | string | Comma-separated worker IDs to filter by | - |

**Example Request:**
```bash
GET /api/workers/audit?limit=10&actions=pause,resume&start_time=2024-01-15T00:00:00Z
```

**Response:**
```json
{
  "audit_logs": [
    {
      "id": "audit-12345",
      "timestamp": "2024-01-15T12:30:00Z",
      "action": "pause",
      "worker_ids": ["worker-001", "worker-002"],
      "user_id": "admin",
      "reason": "Maintenance window",
      "success": true,
      "duration": 2000000000,
      "metadata": {
        "request_id": "req-12345",
        "successful": 2,
        "failed": 0
      },
      "ip_address": "192.168.1.100",
      "user_agent": "WorkerFleetUI/1.0"
    }
  ],
  "filter": {
    "actions": ["pause", "resume"],
    "start_time": "2024-01-15T00:00:00Z",
    "limit": 10
  },
  "count": 1
}
```

## Safety and Confirmation

### Confirmation Requirements

The following actions require confirmation when they affect a significant portion of the fleet:

- **Drain/Stop**: Required when affecting ≥25% of workers or ≥5 workers
- **Pause**: Required when affecting ≥50% of workers or ≥10 workers
- **Restart**: Required when affecting ≥3 workers

### Confirmation Response

When confirmation is required, the API returns HTTP 428 (Precondition Required):

```json
{
  "confirmation_required": true,
  "prompt": "This will affect 5 workers (31.3% of fleet) and may impact 3 active jobs. Type 'CONFIRM' to proceed.",
  "error": "invalid confirmation, expected 'CONFIRM', got ''"
}
```

Include the confirmation in subsequent requests:

```json
{
  "worker_ids": ["worker-001", "worker-002"],
  "confirmation": "CONFIRM"
}
```

## Error Responses

Errors conform to the shared Admin API envelope and include an `X-Request-ID` header for correlation.

```json
{
  "code": "CONFIRMATION_REQUIRED",
  "error": "confirmation token missing",
  "status": 428,
  "request_id": "d781c4f2-5f46-4c8c-9200-83b5f65a6d42",
  "timestamp": "2024-01-15T12:30:00Z",
  "details": {
    "expected": "CONFIRM"
  }
}
```

### Common HTTP Status Codes

| Status Code | Description |
|-------------|-------------|
| 200 | Success |
| 201 | Created successfully |
| 202 | Accepted (for async operations) |
| 400 | Bad Request - Invalid parameters |
| 404 | Not Found - Worker or action not found |
| 428 | Precondition Required - Confirmation needed |
| 500 | Internal Server Error |

Common error codes:

- `CONFIRMATION_REQUIRED` – destructive action attempted without the expected confirmation token
- `VALIDATION_ERROR` – payload failed validation (details include offending fields)
- `AUTH_INVALID` / `AUTH_MISSING` – authentication problems
- `RATE_LIMIT` – request exceeded the configured limit; respect `details.retry_after`
- `INTERNAL_ERROR` – unexpected failure; retry or contact support with the request ID

## Rate Limiting

The API implements the following rate limits:

- Worker registration: 10 requests per minute
- Action operations: 5 requests per minute per user
- Read operations: 100 requests per minute per user

## WebSocket Events (Future)

For real-time updates, the API will support WebSocket connections:

```javascript
const ws = new WebSocket('ws://localhost:8080/api/workers/events');
ws.onmessage = (event) => {
  const update = JSON.parse(event.data);
  console.log('Worker update:', update);
};
```

## SDK Examples

### Go SDK

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/your-org/worker-fleet-client"
)

func main() {
    client := fleetclient.New("http://localhost:8080")

    // List workers
    workers, err := client.ListWorkers(context.Background(), &fleetclient.ListOptions{
        States: []string{"running", "paused"},
        PageSize: 20,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d workers\n", len(workers.Workers))

    // Pause workers
    response, err := client.PauseWorkers(context.Background(), []string{"worker-001"}, "Maintenance")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Action %s started\n", response.RequestID)
}
```

### Python SDK

```python
import worker_fleet_client

client = worker_fleet_client.Client("http://localhost:8080")

# List workers
workers = client.list_workers(states=["running"], page_size=20)
print(f"Found {len(workers.workers)} workers")

# Drain workers with confirmation
try:
    response = client.drain_workers(
        worker_ids=["worker-001", "worker-002"],
        reason="Deployment",
        timeout_seconds=300
    )
except worker_fleet_client.ConfirmationRequired as e:
    print(f"Confirmation required: {e.prompt}")
    response = client.drain_workers(
        worker_ids=["worker-001", "worker-002"],
        reason="Deployment",
        timeout_seconds=300,
        confirmation="CONFIRM"
    )

print(f"Drain action {response.request_id} started")
```

## Best Practices

1. **Always check fleet health** before performing destructive actions
2. **Use confirmation prompts** to prevent accidental mass operations
3. **Monitor action progress** using the status endpoints
4. **Implement proper timeouts** for drain and stop operations
5. **Use rolling restarts** for zero-downtime deployments
6. **Review audit logs** regularly for operational insights
7. **Set up alerts** for worker health and fleet capacity
8. **Test safety mechanisms** in staging environments

## Troubleshooting

### Common Issues

1. **Workers not responding to signals**
   - Check worker signal handler implementation
   - Verify Redis connectivity
   - Check worker logs for errors

2. **Safety checks preventing operations**
   - Review fleet health and capacity
   - Use `force: true` for emergency situations
   - Adjust safety thresholds in configuration

3. **High action failure rates**
   - Check worker health status
   - Verify network connectivity
   - Review worker resource usage

4. **Rolling restart stuck**
   - Check individual worker health
   - Verify drain timeout settings
   - Monitor worker resource usage during restart
