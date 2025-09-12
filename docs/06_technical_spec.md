# Technical Specifications

- Last updated: 2025-09-12

## Executive Summary
This document details implementation approaches, contracts, schemas, and algorithms underpinning producer, worker, reaper, and observability subsystems.

## Table of Contents
- [Configuration Schema](#configuration-schema)
- [CLI and Process Contracts](#cli-and-process-contracts)
- [Job Schema](#job-schema)
- [Algorithms](#algorithms)
- [Metrics](#metrics)
- [Logging and Tracing](#logging-and-tracing)

## Configuration Schema
YAML keys (with env overrides using upper snake case):

```yaml
redis:
  addr: string
  username: string
  password: string
  db: int
  pool_size_multiplier: int
  min_idle_conns: int
  dial_timeout: duration
  read_timeout: duration
  write_timeout: duration
  max_retries: int
worker:
  count: int
  heartbeat_ttl: duration
  max_retries: int
  backoff: { base: duration, max: duration }
  priorities: [string]
  queues: { <priority>: string }
  processing_list_pattern: string # printf format, %s workerID
  heartbeat_key_pattern: string   # printf format, %s workerID
  completed_list: string
  dead_letter_list: string
  brpoplpush_timeout: duration
producer:
  scan_dir: string
  include_globs: [string]
  exclude_globs: [string]
  default_priority: string
  high_priority_exts: [string]
  rate_limit_per_sec: int
  rate_limit_key: string
circuit_breaker:
  failure_threshold: float
  window: duration
  cooldown_period: duration
  min_samples: int
observability:
  metrics_port: int
  log_level: string
  queue_sample_interval: duration
  tracing: { enabled: bool, endpoint: string }
```

Validation rules:
- `worker.count >= 1`
- `worker.priorities` non-empty; each has entry in `worker.queues`
- `worker.heartbeat_ttl >= 5s`, `brpoplpush_timeout <= heartbeat_ttl/2`
- `producer.rate_limit_per_sec >= 0`

## CLI and Process Contracts
- `--role={producer|worker|all}` selects the operational role.
- `--config=PATH` points to YAML. Missing file is allowed (defaults), invalid values are not.
- Process exits non-zero on fatal config errors or unrecoverable subsystem init.

## Job Schema

```json
{
  "id": "string",
  "filepath": "string",
  "filesize": 0,
  "priority": "high|low",
  "retries": 0,
  "creation_time": "RFC3339Nano",
  "trace_id": "string",
  "span_id": "string"
}
```

## Algorithms

### Prioritized Fetch
- Loop priorities (e.g., high then low), executing `BRPOPLPUSH src -> processing` with short timeout (e.g., 1s).
- Guarantees: atomic move per-queue; priority preference within timeout granularity; no job loss between queues and processing list.
- Tradeoffs: small added latency for lower-priority items; documented in README.

### Heartbeat
- On fetch, `SET heartbeatKey payload EX=heartbeat_ttl`.
- Heartbeat refreshed on each loop iteration for the active job (optional enhancement) or kept static (current).

### Completion
- Success: `LPUSH completed payload`; `LREM processing 1 payload`; `DEL heartbeatKey`.
- Failure: increment `Retries`; exponential backoff `min(base*2^(n-1), max)`; requeue or DLQ after `max_retries`.

### Reaper
- Periodically SCAN `jobqueue:worker:*:processing`. For each list:
  - Compose heartbeat key from worker id; if missing, `RPOP` items and `LPUSH` back to original priority queue inferred from payload.
  - Bounded per-scan to avoid long stalls; sleep between SCAN pages.

### Circuit Breaker
- Sliding window of recent results; failure rate `fails/total` compared to `failure_threshold`.
- States: Closed → Open on threshold; Open → HalfOpen on cooldown; HalfOpen → Closed on success, else Open.

## Metrics
- `jobs_produced_total`, `jobs_consumed_total`, `jobs_completed_total`, `jobs_failed_total`, `jobs_retried_total`, `jobs_dead_letter_total` (counters)
- `job_processing_duration_seconds` (histogram)
- `queue_length{queue}` (gauge)
- `circuit_breaker_state` (gauge: 0 Closed, 1 HalfOpen, 2 Open)

## Logging and Tracing
- Logs are JSON with keys: `level`, `ts`, `msg`, `trace_id`, `span_id`, `job_id`, `queue`, `worker_id`.
- Tracing: create a span per job processing; propagate `trace_id/span_id` if present; otherwise create new.
