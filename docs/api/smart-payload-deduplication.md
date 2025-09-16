# Smart Payload Deduplication API

## Overview

The Smart Payload Deduplication system provides intelligent storage optimization for job queues by detecting and eliminating duplicate content across payloads. Using content-addressable storage with Rabin fingerprinting, it can reduce memory usage by 50-90% while maintaining complete transparency to application code.

## Features

- **Content-Based Chunking**: Uses Rabin fingerprinting for optimal chunk boundary detection
- **Compression**: Zstandard compression with learned dictionaries
- **Reference Counting**: Automatic garbage collection of unused chunks
- **Similarity Detection**: MinHash LSH for near-duplicate detection
- **Transparent Integration**: Works seamlessly with existing producer/worker systems
- **Comprehensive Monitoring**: Detailed statistics and health metrics

## Core Components

### DeduplicationManager

The main interface for all deduplication operations.

```go
type DeduplicationManager interface {
    // Core operations
    DeduplicatePayload(jobID string, payload []byte) (*PayloadMap, error)
    ReconstructPayload(payloadMap *PayloadMap) ([]byte, error)

    // Chunk management
    StoreChunk(chunk *Chunk) error
    GetChunk(hash []byte) (*Chunk, error)
    DeleteChunk(hash []byte) error

    // Reference counting
    AddReference(chunkHash []byte) error
    RemoveReference(chunkHash []byte) error
    GetReferenceCount(chunkHash []byte) (int64, error)

    // Statistics and monitoring
    GetStats() (*DeduplicationStats, error)
    GetChunkStats(hash []byte) (*ChunkStats, error)
    GetPopularChunks(limit int) ([]ChunkStats, error)

    // Garbage collection
    RunGarbageCollection() error
    GetOrphanedChunks() ([][]byte, error)

    // Health and diagnostics
    ValidateIntegrity() error
    AuditReferences() error
    GetHealth() map[string]interface{}

    // Configuration
    UpdateConfig(config *Config) error
    GetConfig() *Config
}
```

### Data Structures

#### Chunk
Represents a deduplicated data chunk:

```go
type Chunk struct {
    Hash      []byte    `json:"hash"`       // SHA-256 hash
    Data      []byte    `json:"data"`       // Compressed chunk data
    Size      int       `json:"size"`       // Original size
    CompSize  int       `json:"comp_size"`  // Compressed size
    RefCount  int64     `json:"ref_count"`  // Reference count
    CreatedAt time.Time `json:"created_at"` // Creation timestamp
    LastUsed  time.Time `json:"last_used"`  // Last access time
}
```

#### PayloadMap
Represents a deduplicated payload as chunk references:

```go
type PayloadMap struct {
    JobID      string           `json:"job_id"`      // Original job ID
    OrigSize   int              `json:"orig_size"`   // Original size
    ChunkRefs  []ChunkReference `json:"chunk_refs"`  // Chunk references
    Checksum   []byte           `json:"checksum"`    // Integrity checksum
    CreatedAt  time.Time        `json:"created_at"`  // Creation timestamp
    Compressed bool             `json:"compressed"`  // Compression flag
}
```

#### ChunkReference
References a chunk within a payload:

```go
type ChunkReference struct {
    Hash   []byte `json:"hash"`   // Hash of referenced chunk
    Offset int    `json:"offset"` // Offset in original payload
    Size   int    `json:"size"`   // Size of chunk
}
```

## HTTP API Endpoints

### Payload Deduplication

**POST** `/api/v1/dedup/payload`

Deduplicates a job payload into chunk references.

**Request Body:**
```json
{
    "job_id": "job_12345",
    "payload": "base64_encoded_payload_data"
}
```

**Response:**
```json
{
    "job_id": "job_12345",
    "payload_map": {
        "job_id": "job_12345",
        "orig_size": 10240,
        "chunk_refs": [
            {
                "hash": "sha256_hash_bytes",
                "offset": 0,
                "size": 4096
            }
        ],
        "checksum": "payload_checksum_bytes",
        "created_at": "2025-01-15T10:30:00Z",
        "compressed": true
    },
    "original_size": 10240,
    "chunk_count": 3,
    "processing_time": "15ms",
    "compressed": true
}
```

### Payload Reconstruction

**POST** `/api/v1/dedup/reconstruct`

Reconstructs original payload from chunk references.

**Request Body:**
```json
{
    "payload_map": {
        "job_id": "job_12345",
        "orig_size": 10240,
        "chunk_refs": [...],
        "checksum": "payload_checksum_bytes",
        "created_at": "2025-01-15T10:30:00Z",
        "compressed": true
    }
}
```

**Response:**
```json
{
    "job_id": "job_12345",
    "payload": "base64_encoded_reconstructed_data",
    "reconstructed_size": 10240,
    "processing_time": "8ms"
}
```

### Statistics

**GET** `/api/v1/dedup/stats`

Returns comprehensive deduplication statistics.

**Response:**
```json
{
    "total_payloads": 15420,
    "total_chunks": 8932,
    "total_bytes": 1048576000,
    "deduplicated_bytes": 314572800,
    "compression_ratio": 0.65,
    "deduplication_ratio": 0.58,
    "memory_savings": 734003200,
    "savings_percent": 0.70,
    "chunk_hit_rate": 0.85,
    "avg_chunk_size": 4096.5,
    "popular_chunks": [
        {
            "hash": "abc123...",
            "ref_count": 1250,
            "size": 4096,
            "created_at": "2025-01-15T08:00:00Z",
            "last_used": "2025-01-15T10:29:45Z",
            "hit_count": 1250
        }
    ],
    "last_updated": "2025-01-15T10:30:00Z"
}
```

### Chunk Management

**GET** `/api/v1/dedup/chunks?limit=50`

Returns popular chunks by reference count.

**Response:**
```json
{
    "chunks": [
        {
            "hash": "abc123...",
            "ref_count": 1250,
            "size": 4096,
            "created_at": "2025-01-15T08:00:00Z",
            "last_used": "2025-01-15T10:29:45Z",
            "hit_count": 1250
        }
    ],
    "total": 50,
    "limit": 50
}
```

**DELETE** `/api/v1/dedup/chunks?hash=abc123...`

Deletes a specific chunk (decrements reference count).

**Response:**
```json
{
    "hash": "abc123...",
    "deleted": true
}
```

### Garbage Collection

**POST** `/api/v1/dedup/gc`

Manually triggers garbage collection.

**Response:**
```json
{
    "started": "2025-01-15T10:30:00Z",
    "completed": "2025-01-15T10:30:15Z",
    "duration": "15s",
    "success": true
}
```

### Health Check

**GET** `/api/v1/dedup/health`

Returns system health information.

**Response:**
```json
{
    "enabled": true,
    "total_chunks": 8932,
    "memory_savings": 734003200,
    "compression_ratio": 0.65,
    "last_updated": "2025-01-15T10:30:00Z",
    "gc_enabled": true,
    "compression_stats": {
        "total_compressed": 15420,
        "total_decompressed": 14850,
        "bytes_in": 1048576000,
        "bytes_out": 681574400,
        "compression_ratio": 0.65,
        "avg_compression_time": "2ms",
        "avg_decompression_time": "1ms",
        "dictionary_size": 65536,
        "last_updated": "2025-01-15T10:30:00Z"
    }
}
```

### Configuration

**GET** `/api/v1/dedup/config`

Returns current configuration.

**Response:**
```json
{
    "enabled": true,
    "redis_key_prefix": "go-redis-work-queue:dedup:",
    "redis_db": 0,
    "chunking": {
        "min_chunk_size": 1024,
        "max_chunk_size": 65536,
        "avg_chunk_size": 8192,
        "window_size": 64,
        "polynomial": 2179725432,
        "similarity_threshold": 0.8
    },
    "compression": {
        "enabled": true,
        "level": 3,
        "dictionary_size": 65536,
        "use_dictionary": true
    },
    "garbage_collection": {
        "enabled": true,
        "interval": "1h",
        "orphan_threshold": "24h",
        "batch_size": 1000,
        "concurrent_workers": 4
    },
    "safety_mode": true,
    "migration_ratio": 1.0,
    "max_memory_mb": 1024,
    "stats_interval": "5m"
}
```

**PUT** `/api/v1/dedup/config`

Updates system configuration.

**Request Body:** (Same structure as GET response)

**Response:**
```json
{
    "updated": true,
    "config": { /* Updated configuration */ }
}
```

## Integration Examples

### Producer Integration

```go
// Initialize deduplication
config := deduplication.DefaultConfig()
config.RedisKeyPrefix = "myapp:dedup:"

manager, err := deduplication.NewManager(config, redisClient, logger)
if err != nil {
    log.Fatal(err)
}

integration := deduplication.NewProducerIntegration(manager, config, logger)

// Enqueue job with deduplication
func enqueueJob(jobID string, payload []byte) error {
    // Deduplicate payload
    dedupedData, err := integration.EnqueueJob(jobID, payload)
    if err != nil {
        return err
    }

    // Store in queue (dedupedData is either PayloadMap or original payload)
    return queueClient.Enqueue(jobID, dedupedData)
}
```

### Worker Integration

```go
// Initialize worker integration
integration := deduplication.NewWorkerIntegration(manager, config, logger)

// Process job with reconstruction
func processJob(jobData interface{}) error {
    // Reconstruct payload
    payload, err := integration.DequeueJob(jobData)
    if err != nil {
        return err
    }

    // Process the job
    success := processPayload(payload)

    // Cleanup references
    return integration.CleanupJob(jobData, success)
}
```

## Configuration Parameters

### Chunking Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `min_chunk_size` | int | 1024 | Minimum chunk size in bytes |
| `max_chunk_size` | int | 65536 | Maximum chunk size in bytes |
| `avg_chunk_size` | int | 8192 | Target average chunk size |
| `window_size` | int | 64 | Rolling hash window size |
| `polynomial` | uint64 | 0x82f63b78 | Rabin polynomial for hashing |
| `similarity_threshold` | float64 | 0.8 | Threshold for similarity detection |

### Compression Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `enabled` | bool | true | Enable compression |
| `level` | int | 3 | Compression level (1-19) |
| `dictionary_size` | int | 65536 | Dictionary size in bytes |
| `use_dictionary` | bool | true | Use learned compression dictionary |

### Garbage Collection Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `enabled` | bool | true | Enable automatic GC |
| `interval` | duration | 1h | GC run interval |
| `orphan_threshold` | duration | 24h | Time before chunks are orphaned |
| `batch_size` | int | 1000 | Chunks processed per GC batch |
| `concurrent_workers` | int | 4 | Number of GC workers |

## Performance Characteristics

### Memory Savings

- **Typical workloads**: 50-70% memory reduction
- **Highly repetitive data**: 80-95% memory reduction
- **Unique data**: 0-20% reduction (compression only)

### Processing Overhead

- **Deduplication**: 5-15ms per payload (depending on size)
- **Reconstruction**: 2-8ms per payload
- **Memory overhead**: ~100 bytes per chunk for metadata

### Scalability

- **Chunk storage**: Scales linearly with unique content
- **Reference counting**: O(1) operations with Redis
- **Garbage collection**: Configurable batch processing

## Error Handling

### Error Types

| Type | Description | Recoverable |
|------|-------------|-------------|
| `configuration` | Configuration validation errors | No |
| `storage` | Redis storage errors | Yes |
| `compression` | Compression/decompression failures | Partial |
| `reference` | Reference counting corruption | Yes |
| `garbage_collection` | GC operation failures | Yes |
| `integrity` | Data integrity violations | No |

### Error Response Format

```json
{
    "error": {
        "type": "storage",
        "code": "CHUNK_NOT_FOUND",
        "message": "Chunk not found in storage",
        "recoverable": false,
        "retry_after": 0,
        "context": {
            "chunk_hash": "abc123...",
            "operation": "get_chunk"
        },
        "timestamp": 1642680000
    }
}
```

## Monitoring and Alerting

### Key Metrics

- **Deduplication ratio**: Percentage of chunks reused
- **Compression ratio**: Size reduction from compression
- **Memory savings**: Total bytes saved
- **Chunk hit rate**: Cache efficiency metric
- **Processing time**: Latency for operations
- **Reference count accuracy**: Integrity metric

### Health Checks

Regular health checks should monitor:

1. **Storage availability**: Redis connectivity
2. **Reference count integrity**: Consistency checks
3. **Memory usage**: Chunk storage size
4. **GC effectiveness**: Orphaned chunk cleanup
5. **Compression performance**: Dictionary effectiveness

## Security Considerations

### Data Safety

- **Checksums**: SHA-256 verification for all payloads
- **Reference counting**: Atomic operations prevent corruption
- **Graceful degradation**: Fallback to original payloads on errors
- **Audit trails**: Comprehensive logging and monitoring

### Access Control

- **Redis security**: Use Redis AUTH and TLS
- **API authentication**: Implement authentication for HTTP endpoints
- **Data isolation**: Use Redis key prefixes for multi-tenancy

## Migration Guide

### Gradual Rollout

1. **Start with low migration ratio** (e.g., 10%)
2. **Monitor performance and savings**
3. **Gradually increase ratio** as confidence builds
4. **Enable safety mode** for automatic fallback

### Zero-Downtime Migration

```go
// Configure gradual migration
config := deduplication.DefaultConfig()
config.MigrationRatio = 0.1  // Start with 10%
config.SafetyMode = true     // Enable fallback

// Gradually increase migration ratio
// config.MigrationRatio = 0.5  // 50%
// config.MigrationRatio = 1.0  // 100%
```

## Troubleshooting

### Common Issues

1. **High memory usage**: Check GC configuration and orphaned chunks
2. **Slow performance**: Verify Redis connectivity and chunk size tuning
3. **Reference count corruption**: Run audit and repair operations
4. **Low deduplication ratio**: Analyze payload patterns and similarity threshold

### Diagnostic Commands

```bash
# Check system health
curl http://localhost:8080/api/v1/dedup/health

# Get detailed statistics
curl http://localhost:8080/api/v1/dedup/stats

# Run manual garbage collection
curl -X POST http://localhost:8080/api/v1/dedup/gc

# Validate system integrity
curl -X POST http://localhost:8080/api/v1/dedup/audit
```

## Best Practices

### Configuration Tuning

1. **Chunk size**: Balance between deduplication effectiveness and overhead
2. **Compression level**: Higher levels for storage-critical applications
3. **GC frequency**: More frequent for high-churn workloads
4. **Memory limits**: Set appropriate limits based on available resources

### Operational Guidelines

1. **Monitor savings regularly**: Track deduplication effectiveness
2. **Run periodic audits**: Ensure reference count integrity
3. **Plan for growth**: Scale Redis capacity with chunk volume
4. **Test disaster recovery**: Validate fallback mechanisms

### Development Integration

1. **Use transparent APIs**: Minimal changes to existing code
2. **Handle errors gracefully**: Implement proper fallback logic
3. **Test with realistic data**: Use production-like payloads for testing
4. **Monitor performance impact**: Measure latency and throughput changes