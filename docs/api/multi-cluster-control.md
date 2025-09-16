# Multi-Cluster Control API Documentation

## Overview

The Multi-Cluster Control module provides a unified interface for managing multiple Redis clusters from a single control plane. It enables side-by-side comparison, synchronized actions, and comprehensive monitoring across all configured clusters.

## Features

- **Multi-endpoint configuration** with hot-switching between clusters
- **Side-by-side comparison** views for Jobs and Workers
- **Multi-apply actions** with explicit confirmation and target listing
- **Per-cluster caching** with configurable TTL
- **Health monitoring** and anomaly detection
- **TUI integration** with keyboard shortcuts and visual indicators

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Multi-Cluster Manager                      │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │ Cluster1 │  │ Cluster2 │  │ Cluster3 │  │ ClusterN │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
│       ↓             ↓             ↓             ↓           │
│  ┌──────────────────────────────────────────────────────┐   │
│  │            Connection Pool & Health Monitor          │   │
│  └──────────────────────────────────────────────────────┘   │
│       ↓                                                      │
│  ┌──────────────────────────────────────────────────────┐   │
│  │         Stats Collector & Compare Engine             │   │
│  └──────────────────────────────────────────────────────┘   │
│       ↓                                                      │
│  ┌──────────────────────────────────────────────────────┐   │
│  │           Action Executor & Confirmation             │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## Configuration

### Example Configuration

```json
{
  "clusters": [
    {
      "name": "production",
      "label": "Production",
      "color": "green",
      "endpoint": "prod.redis.example.com:6379",
      "password": "secret",
      "db": 0,
      "enabled": true
    },
    {
      "name": "staging",
      "label": "Staging",
      "color": "yellow",
      "endpoint": "staging.redis.example.com:6379",
      "db": 0,
      "enabled": true
    }
  ],
  "default_cluster": "production",
  "polling": {
    "interval": "5s",
    "jitter": "1s",
    "timeout": "3s",
    "enabled": true
  },
  "compare_mode": {
    "enabled": false,
    "highlight_deltas": true,
    "delta_threshold": 10.0
  },
  "actions": {
    "require_confirmation": true,
    "allowed_actions": ["purge_dlq", "pause_queue", "resume_queue", "benchmark"],
    "max_concurrent": 5
  }
}
```

## API Endpoints

### Cluster Management

#### List Clusters
```
GET /api/v1/clusters
```

Returns all configured clusters.

**Response:**
```json
[
  {
    "name": "production",
    "label": "Production",
    "color": "green",
    "endpoint": "prod.redis.example.com:6379",
    "enabled": true
  }
]
```

#### Add Cluster
```
POST /api/v1/clusters
```

**Request Body:**
```json
{
  "name": "new-cluster",
  "label": "New Cluster",
  "color": "blue",
  "endpoint": "new.redis.example.com:6379",
  "password": "secret",
  "db": 0,
  "enabled": true
}
```

#### Get Cluster
```
GET /api/v1/clusters/{name}
```

Returns details for a specific cluster including connection status.

#### Update Cluster
```
PUT /api/v1/clusters/{name}
```

Updates cluster configuration.

#### Delete Cluster
```
DELETE /api/v1/clusters/{name}
```

Removes a cluster from configuration.

### Statistics & Monitoring

#### Get Stats
```
GET /api/v1/stats?cluster={name}
```

Returns statistics for a specific cluster or all clusters if no name provided.

**Response:**
```json
{
  "cluster_name": "production",
  "queue_sizes": {
    "default": 150,
    "high": 25,
    "low": 300
  },
  "processing_count": 45,
  "dead_letter_count": 3,
  "worker_count": 10,
  "job_rate": 125.5,
  "error_rate": 0.02
}
```

#### Compare Clusters
```
POST /api/v1/stats/compare
```

**Request Body:**
```json
{
  "clusters": ["production", "staging"]
}
```

**Response:**
```json
{
  "clusters": ["production", "staging"],
  "metrics": {
    "queue_size": {
      "name": "queue_size",
      "values": {
        "production": 475,
        "staging": 125
      },
      "delta": 350,
      "unit": "jobs"
    }
  },
  "anomalies": [
    {
      "type": "deviation",
      "cluster": "production",
      "description": "queue_size deviates significantly from average",
      "value": 475,
      "expected": 300,
      "severity": "warning"
    }
  ]
}
```

#### Health Check
```
GET /api/v1/health?cluster={name}
```

Returns health status for a specific cluster or all clusters.

**Response:**
```json
{
  "healthy": true,
  "issues": [],
  "metrics": {
    "latency_ms": 2.5,
    "worker_count": 10,
    "dead_letter_count": 3
  },
  "last_checked": "2025-01-14T10:30:00Z"
}
```

### Multi-Cluster Actions

#### Execute Action
```
POST /api/v1/actions
```

**Request Body:**
```json
{
  "type": "purge_dlq",
  "targets": ["production", "staging"],
  "parameters": {},
  "confirmations": [
    {
      "required": true,
      "message": "This will purge DLQ on 2 clusters. Continue?"
    }
  ]
}
```

**Response:**
```json
{
  "id": "action-123456",
  "type": "purge_dlq",
  "status": "completed",
  "results": {
    "production": {
      "success": true,
      "message": "Action completed successfully",
      "duration_ms": 45.2
    },
    "staging": {
      "success": true,
      "message": "Action completed successfully",
      "duration_ms": 38.7
    }
  }
}
```

#### Get Action Status
```
GET /api/v1/actions/{id}
```

Returns the status and results of a specific action.

#### Confirm Action
```
POST /api/v1/actions/{id}/confirm
```

**Request Body:**
```json
{
  "confirmed_by": "admin@example.com"
}
```

#### Cancel Action
```
POST /api/v1/actions/{id}/cancel
```

Cancels a pending action.

### Event Streaming

#### Subscribe to Events
```
GET /api/v1/events
```

Server-Sent Events (SSE) stream of cluster events.

**Event Format:**
```
data: {"id":"evt-123","type":"cluster_connected","cluster":"production","message":"Connected to cluster production","timestamp":"2025-01-14T10:30:00Z"}
```

### TUI Support

#### Get Tab Configuration
```
GET /api/v1/ui/tabs
```

**Response:**
```json
{
  "tabs": [
    {
      "index": 1,
      "cluster_name": "production",
      "label": "Production",
      "color": "green",
      "shortcut": "1"
    }
  ],
  "active_tab": 0,
  "compare_mode": false
}
```

#### Set Compare Mode
```
PUT /api/v1/ui/compare
```

**Request Body:**
```json
{
  "enabled": true,
  "clusters": ["production", "staging"]
}
```

## Usage Examples

### Go Client Example

```go
package main

import (
    "context"
    "fmt"
    multicluster "github.com/flyingrobots/go-redis-work-queue/internal/multi-cluster-control"
    "go.uber.org/zap"
)

func main() {
    // Load configuration
    config, err := multicluster.LoadConfig("config.json")
    if err != nil {
        panic(err)
    }

    // Create manager
    manager, err := multicluster.NewManager(config, zap.NewExample())
    if err != nil {
        panic(err)
    }
    defer manager.Close()

    ctx := context.Background()

    // Get stats for all clusters
    stats, err := manager.GetAllStats(ctx)
    if err != nil {
        panic(err)
    }

    for cluster, stat := range stats {
        fmt.Printf("Cluster %s: %d jobs in queue\n", cluster, stat.QueueSizes["default"])
    }

    // Execute multi-cluster action
    action := &multicluster.MultiAction{
        Type:    multicluster.ActionTypePurgeDLQ,
        Targets: []string{"production", "staging"},
    }

    if err := manager.ExecuteAction(ctx, action); err != nil {
        panic(err)
    }

    fmt.Printf("Action completed: %s\n", action.Status)
}
```

### TUI Integration

The multi-cluster control integrates seamlessly with the existing TUI:

1. **Tab Switching**: Use number keys (1-9) to quickly switch between clusters
2. **Compare Mode**: Press 'C' to enable side-by-side comparison
3. **Multi-Apply**: Select multiple clusters with space, then apply action with confirmation
4. **Health Indicators**: Color-coded tab headers show cluster health status

## Error Handling

The module provides structured error types for different scenarios:

- `ClusterError`: Errors specific to a cluster
- `MultiClusterError`: Errors from multiple clusters
- `ActionError`: Errors during action execution
- `ConnectionError`: Connection-related errors

Errors are classified by severity:
- **Critical**: System cannot continue (e.g., no enabled clusters)
- **Error**: Operation failed but system continues
- **Warning**: Degraded functionality (e.g., cluster disconnected)
- **Info**: Informational messages

## Performance Considerations

- **Caching**: Stats are cached with configurable TTL to reduce Redis calls
- **Polling**: Configurable intervals with jitter to avoid thundering herd
- **Connection Pooling**: Reuses connections across operations
- **Concurrent Actions**: Configurable concurrency limits for multi-cluster actions

## Security

- **Authentication**: Per-cluster password support
- **Confirmation**: Required confirmation for destructive actions
- **Audit Logging**: All actions are logged with user and timestamp
- **Action Allowlist**: Only configured action types are permitted

## Monitoring & Observability

The module integrates with existing observability infrastructure:

- **Metrics**: Prometheus metrics for cluster health and action performance
- **Tracing**: OpenTelemetry spans for action execution
- **Logging**: Structured logging with zap
- **Events**: Real-time event stream for monitoring tools