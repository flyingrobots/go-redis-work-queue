# Long-Term Archives API

The Long-Term Archives module provides comprehensive job archiving capabilities with support for multiple storage backends, schema evolution, retention policies, and GDPR compliance.

## Overview

The archiving system supports:
- **Dual Storage**: ClickHouse for analytics and S3/Parquet for cost-effective long-term storage
- **Schema Versioning**: Backward-compatible schema evolution with automatic migrations
- **Retention Management**: Configurable retention policies with automated cleanup
- **GDPR Compliance**: Data deletion requests and audit trails
- **Export Operations**: Bulk data exports with progress tracking
- **Query Templates**: Predefined analytics queries with parameterization

## Configuration

```go
type Config struct {
    RedisAddr string
    RedisDB   int
    Archive   ArchiveConfig
}

type ArchiveConfig struct {
    Enabled         bool
    SamplingRate    float64
    BatchSize       int
    RedisStreamKey  string
    ClickHouse      ClickHouseConfig
    S3              S3Config
    Retention       RetentionConfig
    PayloadHandling PayloadHandlingConfig
}
```

## API Endpoints

### Archive Operations

#### Archive Job
```http
POST /api/v1/archive/jobs
Content-Type: application/json

{
  "job_id": "job-123",
  "queue": "high-priority",
  "priority": 1,
  "enqueued_at": "2024-01-15T10:30:00Z",
  "completed_at": "2024-01-15T10:35:00Z",
  "outcome": "success",
  "worker_id": "worker-1",
  "payload_snapshot": "eyJkYXRhIjoidGVzdCJ9"
}
```

#### Get Archived Job
```http
GET /api/v1/archive/jobs/{jobId}
```

#### Search Jobs
```http
POST /api/v1/archive/jobs/search
Content-Type: application/json

{
  "queue": "high-priority",
  "outcome": "success",
  "start_time": "2024-01-15T00:00:00Z",
  "end_time": "2024-01-15T23:59:59Z",
  "limit": 100,
  "order_by": "completed_at",
  "order_dir": "DESC"
}
```

### Export Operations

#### Export Jobs
```http
POST /api/v1/archive/export
Content-Type: application/json

{
  "query": {
    "queue": "analytics",
    "start_time": "2024-01-01T00:00:00Z",
    "end_time": "2024-01-31T23:59:59Z"
  },
  "type": "parquet"
}
```

Response:
```json
{
  "export_id": "exp_12345",
  "status": "running",
  "created_at": "2024-01-15T10:00:00Z",
  "estimated_completion": "2024-01-15T10:05:00Z"
}
```

#### Get Export Status
```http
GET /api/v1/archive/export/{exportId}
```

#### Cancel Export
```http
POST /api/v1/archive/export/{exportId}/cancel
```

#### List Exports
```http
GET /api/v1/archive/exports?limit=50&offset=0
```

### Statistics

#### Get Archive Statistics
```http
GET /api/v1/archive/stats?window=24h
```

Response:
```json
{
  "total_archived": 15420,
  "by_outcome": {
    "success": 14892,
    "failed": 528
  },
  "by_queue": {
    "high-priority": 8341,
    "normal": 6547,
    "low": 532
  },
  "storage_stats": {
    "clickhouse_records": 15420,
    "s3_objects": 342,
    "total_size_bytes": 45621890
  }
}
```

### Schema Management

#### Get Schema Version
```http
GET /api/v1/archive/schema/version
```

#### Upgrade Schema
```http
POST /api/v1/archive/schema/upgrade
Content-Type: application/json

{
  "version": 2
}
```

#### Get Schema Evolution
```http
GET /api/v1/archive/schema/evolution
```

### Retention Management

#### Cleanup Expired Records
```http
POST /api/v1/archive/retention/cleanup
```

#### Get Retention Policy
```http
GET /api/v1/archive/retention/policy
```

#### Update Retention Policy
```http
PUT /api/v1/archive/retention/policy
Content-Type: application/json

{
  "redis_stream_ttl": "24h",
  "archive_window": "720h",
  "delete_after": "8760h",
  "gdpr_compliant": true
}
```

#### Process GDPR Delete Request
```http
POST /api/v1/archive/retention/gdpr
Content-Type: application/json

{
  "user_id": "user-123",
  "reason": "User requested account deletion",
  "criteria": {
    "fields": ["user_id", "email"],
    "values": ["user-123", "user@example.com"]
  }
}
```

### Query Templates

#### Get Query Templates
```http
GET /api/v1/archive/templates
```

#### Add Query Template
```http
POST /api/v1/archive/templates
Content-Type: application/json

{
  "name": "queue_performance",
  "description": "Analyze queue performance metrics",
  "sql": "SELECT queue, AVG(duration) as avg_duration FROM archives WHERE completed_at >= ? AND completed_at < ? GROUP BY queue",
  "parameters": [
    {
      "name": "start_time",
      "type": "timestamp",
      "description": "Start time for analysis",
      "required": true
    },
    {
      "name": "end_time",
      "type": "timestamp",
      "description": "End time for analysis",
      "required": true
    }
  ],
  "tags": ["performance", "analytics"]
}
```

#### Execute Query Template
```http
POST /api/v1/archive/templates/{templateName}/execute
Content-Type: application/json

{
  "parameters": {
    "start_time": "2024-01-15T00:00:00Z",
    "end_time": "2024-01-15T23:59:59Z"
  }
}
```

### Health and Monitoring

#### Health Check
```http
GET /api/v1/archive/health
```

Response:
```json
{
  "status": "ok",
  "redis": "connected",
  "clickhouse": "connected",
  "s3": "connected",
  "version": 2,
  "uptime": "72h45m12s"
}
```

## Data Structures

### ArchiveJob
```go
type ArchiveJob struct {
    JobID           string                 `json:"job_id"`
    Queue           string                 `json:"queue"`
    Priority        int                    `json:"priority"`
    EnqueuedAt      time.Time              `json:"enqueued_at"`
    StartedAt       *time.Time             `json:"started_at,omitempty"`
    CompletedAt     time.Time              `json:"completed_at"`
    Duration        time.Duration          `json:"duration"`
    Outcome         JobOutcome             `json:"outcome"`
    RetryCount      int                    `json:"retry_count"`
    ErrorMessage    string                 `json:"error_message,omitempty"`
    WorkerID        string                 `json:"worker_id,omitempty"`
    PayloadSize     int64                  `json:"payload_size"`
    PayloadHash     string                 `json:"payload_hash,omitempty"`
    PayloadSnapshot []byte                 `json:"payload_snapshot,omitempty"`
    TraceID         string                 `json:"trace_id,omitempty"`
    Tags            map[string]string      `json:"tags,omitempty"`
    ArchivedAt      time.Time              `json:"archived_at"`
    SchemaVersion   int                    `json:"schema_version"`
    Tenant          string                 `json:"tenant,omitempty"`
}
```

### SearchQuery
```go
type SearchQuery struct {
    JobIDs      []string    `json:"job_ids,omitempty"`
    Queue       string      `json:"queue,omitempty"`
    Outcome     JobOutcome  `json:"outcome,omitempty"`
    WorkerID    string      `json:"worker_id,omitempty"`
    StartTime   *time.Time  `json:"start_time,omitempty"`
    EndTime     *time.Time  `json:"end_time,omitempty"`
    Tags        map[string]string `json:"tags,omitempty"`
    Limit       int         `json:"limit"`
    Offset      int         `json:"offset"`
    OrderBy     string      `json:"order_by"`
    OrderDir    string      `json:"order_dir"`
}
```

### ExportRequest
```go
type ExportRequest struct {
    Query         SearchQuery `json:"query"`
    Type          ExportType  `json:"type"`
    Compression   string      `json:"compression,omitempty"`
    IncludePayload bool       `json:"include_payload"`
}
```

## Usage Examples

### Basic Archiving
```go
manager, err := archives.NewManager(config, logger)
if err != nil {
    return err
}
defer manager.Close()

job := archives.ArchiveJob{
    JobID:       "job-123",
    Queue:       "processing",
    CompletedAt: time.Now(),
    Outcome:     archives.OutcomeSuccess,
    WorkerID:    "worker-1",
}

err = manager.ArchiveJob(ctx, job)
```

### Searching Archives
```go
query := archives.SearchQuery{
    Queue:     "high-priority",
    Outcome:   archives.OutcomeSuccess,
    StartTime: &startTime,
    EndTime:   &endTime,
    Limit:     100,
}

jobs, err := manager.SearchJobs(ctx, query)
```

### Schema Evolution
```go
// Get current schema version
version, err := manager.GetSchemaVersion(ctx)

// Upgrade to latest version
err = manager.UpgradeSchema(ctx, 2)

// Check evolution history
evolution, err := manager.GetSchemaEvolution(ctx)
```

### GDPR Compliance
```go
request := archives.GDPRDeleteRequest{
    UserID: "user-123",
    Reason: "Account deletion requested",
    Criteria: archives.DeleteCriteria{
        Fields: []string{"user_id"},
        Values: []string{"user-123"},
    },
}

err = manager.ProcessGDPRDelete(ctx, request)
```

## Error Handling

The API returns standard HTTP status codes:
- `200 OK` - Successful operation
- `400 Bad Request` - Invalid request format or parameters
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

Error responses include detailed error information:
```json
{
  "error": "Validation failed",
  "status": 400,
  "timestamp": "2024-01-15T10:30:00Z",
  "details": "job_id is required"
}
```

## Performance Considerations

- **Batch Size**: Configure appropriate batch sizes for exports based on memory constraints
- **Sampling Rate**: Use sampling for high-volume environments to reduce storage costs
- **Compression**: Enable compression for S3 exports to reduce storage costs
- **Indexing**: Ensure proper indexing on frequently queried fields in ClickHouse
- **Retention**: Configure retention policies to automatically clean up old data

## Security

- All API endpoints support authentication via JWT/PASETO tokens
- GDPR delete operations are logged and auditable
- Payload snapshots can be hashed instead of stored for privacy
- Schema migrations are tracked and reversible