# Product Requirements Document (PRD)

## Title

Go Redis Work Queue — Production-Ready File Processing System

## Summary

A single, multi-role Go binary that implements a robust file-processing job queue backed by Redis. It supports producer, worker, and all-in-one modes; priority queues; reliable processing with retries and a dead-letter queue; graceful shutdown; a reaper for stuck jobs; a circuit breaker; and comprehensive observability (metrics, logs, tracing). All behavior is driven by a YAML configuration with environment variable overrides.

## Goals

- Reliable and fault-tolerant job ingestion and processing
- Horizontally scalable worker pool with dynamic Redis connection pooling
- Strong operational visibility with Prometheus metrics and structured logs
- Minimal operational footprint: single binary, simple config, Docker-ready

## Non-goals

- Workflow orchestration or multi-step DAGs
- Persisting job payloads outside Redis
- UI dashboard (metrics only)

## User Stories

- As an operator, I can run the same binary in producer, worker, or all-in-one mode using a flag.
- As a developer, I can configure the system via a YAML file and environment overrides.
- As an SRE, I can observe queue depth, processing latency, throughput, failures, retries, and circuit breaker state via Prometheus.
- As a platform engineer, I can deploy the service with Docker/Kubernetes easily.

## Roles and Execution Modes

- `role=producer`: scans a directory and enqueues jobs with priority, rate-limited via Redis.
- `role=worker`: runs N worker goroutines consuming jobs by priority, with processing lists and heartbeats.
- `role=all`: runs both producer and worker in one process for development or small deployments.
- `role=admin`: provides operational commands: `stats` (print queue/processing/heartbeat counts), `peek` (inspect queue tail items), and `purge-dlq` (clear dead-letter queue with `--yes`).

## Configuration

All parameters are set via YAML with env var overrides. Example:

```yaml
redis:
  addr: "localhost:6379"
  username: ""
  password: ""
  db: 0
  pool_size_multiplier: 10     # PoolSize = multiplier * NumCPU
  min_idle_conns: 5
  dial_timeout: 5s
  read_timeout: 3s
  write_timeout: 3s
  max_retries: 3

worker:
  count: 16
  heartbeat_ttl: 30s
  max_retries: 3
  backoff:
    base: 500ms
    max: 10s
  priorities: ["high", "low"]
  queues:
    high: "jobqueue:high_priority"
    low:  "jobqueue:low_priority"
  processing_list_pattern: "jobqueue:worker:%s:processing"  # %s = workerID
  heartbeat_key_pattern:  "jobqueue:processing:worker:%s"   # %s = workerID
  completed_list: "jobqueue:completed"
  dead_letter_list: "jobqueue:dead_letter"
  brpoplpush_timeout: 1s  # per-priority poll timeout

producer:
  scan_dir: "./data"
  include_globs: ["**/*"]
  exclude_globs: ["**/*.tmp", "**/.DS_Store"]
  default_priority: "low"
  high_priority_exts: [".pdf", ".docx", ".xlsx", ".zip"]
  rate_limit_per_sec: 100
  rate_limit_key: "jobqueue:rate_limit:producer"

circuit_breaker:
  failure_threshold: 0.5
  window: 1m
  cooldown_period: 30s
  min_samples: 20

observability:
  metrics_port: 9090
  log_level: "info"   # debug, info, warn, error
  tracing:
    enabled: false
    endpoint: ""      # e.g. OTLP gRPC or HTTP endpoint
```

Environment overrides use upper snake case with dots replaced by underscores, e.g., `WORKER_COUNT`, `REDIS_ADDR`.

## Data Model

Job payload JSON:

```json
{
  "id": "uuid",
  "filepath": "/path/to/file",
  "filesize": 12345,
  "priority": "high|low",
  "retries": 0,
  "creation_time": "RFC3339",
  "trace_id": "",
  "span_id": ""
}
```

## Redis Keys and Structures

- Queues: `jobqueue:high_priority`, `jobqueue:low_priority` (List)
- Processing list per worker: `jobqueue:worker:<ID>:processing (List)`
- Heartbeat per worker: `jobqueue:processing:worker:<ID> `(String with EX)
- Completed list: `jobqueue:completed` (List)
- Dead letter list: `jobqueue:dead_letter` (List)
- Producer rate limit: `jobqueue:rate_limit:producer` (Counter with EX)

## Core Algorithms

### Producer

- Scan directory recursively using include/exclude globs.
- Determine priority by extension list.
- Rate limiting: `INCR rate_limit_key; if first increment, set EX=1; if value > rate_limit_per_sec`, `TTL`-based precise sleep (with jitter) until window reset before enqueueing more.
- `LPUSH` job JSON to priority queue.

### Worker Fetch

- Unique worker ID: `"hostname-PID-idx"` for each goroutine.
- Prioritized fetch: loop priorities in order (e.g., high then low) and call `BRPOPLPUSH` per-queue with a short timeout (default 1s). Guarantees atomic move per-queue, priority preference within timeout granularity, and no job loss. Tradeoff: lower-priority jobs may wait up to the timeout when higher-priority queues are empty.
- On receipt, `SET heartbeat` key to job JSON with `EX=heartbeat_ttl`.

### Processing

- Create a span (if tracing enabled) using job trace/span IDs when present; log with IDs.
- Execute user-defined processing (stub initially: simulate processing with duration proportional to filesize; placeholder to plug real logic).
- On success: `LPUSH completed_list job JSON; LREM processing_list 1 job; DEL heartbeat key`.
- On failure: increment Retries in payload; exponential backoff; `if retries <= max_retries LPUSH` back to original priority queue; else `LPUSH dead_letter_list`; in both cases `LREM` from `processing_list` and `DEL heartbeat`.

### Graceful Shutdown

- Catch `SIGINT`/`SIGTERM`; cancel context; stop accepting new jobs; allow in-flight job to finish; ensure heartbeat and processing list cleanup as part of success/failure paths.

### Reaper

- Periodically scan all heartbeat keys matching pattern. For each missing/expired heartbeat, recover jobs lingering in processing lists:
  - For each worker processing list, if list has elements and the corresponding heartbeat key is absent, pop jobs (`LPOP`) one by one, inspect priority within payload, and `LPUSH` back to the appropriate priority queue.

### Circuit Breaker

- Closed: normal operation, track success/failure counts in rolling window.
- Open: `if failure_rate >= threshold and samples >= min_samples`, stop fetching jobs for `cooldown_period`.
- HalfOpen: probe with a single job; on success -> Closed; on failure -> Open.

## Observability

- HTTP server exposes `/metrics`, `/healthz`, and `/readyz`. Key metrics:
  - Counter: `jobs_produced_total`, `jobs_consumed_total`, `jobs_completed_total`, `jobs_failed_total`, `jobs_retried_total`, `jobs_dead_letter_total`
  - Histogram: `job_processing_duration_seconds`
  - Gauges: `queue_length{queue=...}`, `worker_active`, `circuit_breaker_state` (0=`Closed`,1=`HalfOpen`,2=`Open`)
- Logging (zap): structured, includes `trace_id`/`span_id` when present
- Tracing (OpenTelemetry): optional OTLP exporter, spans for produce/consume/process. Job `trace_id`/`span_id` are propagated as remote parent when present.

## CLI

- `--role=producer|worker|all`
- `--config=path/to/config.yaml`
- `--version`

## Performance & Capacity Targets

- Baseline throughput: O(1k) jobs/minute on modest 4 vCPU nodes
- End-to-end p95 latency under 2s for small files (<1MB)
- Sustained stability under brief Redis outages via retries and backoff

## Security & Reliability

- Redis credentials via config or environment
- Avoid plaintext logs of secrets
- Resilient to worker crashes with reaper resurrection

## Deployment

- Dockerfile builds single static binary
- Health probes: metrics endpoint (HTTP `200`), and optional /healthz (future)

## Testing Strategy

- Unit tests: job serialization, rate limiter, circuit breaker, worker loop logic
- Integration tests with Redis: enqueue/dequeue paths, retries, reaper behavior
- Race detector in CI; coverage target ≥80% for core packages

## Risks & Mitigations

- Multi-queue atomic move: emulate priority by sequential `BRPOPLPUSH` timeouts
- Large queues: monitor queue_length gauges; consider sharding per priority if needed

## Milestones

1) Scaffolding and PRD (this doc)
2) Core implementation (config, producer, worker, reaper, breaker, observability)
3) Tests and CI/CD (GitHub Actions)
4) Dockerfile and examples
