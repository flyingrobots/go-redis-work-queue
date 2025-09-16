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

- Persist `origin_queue` in job metadata at enqueue time. Reaper uses this to identify the destination when re-queuing work.
- Periodically `SCAN`/`SSCAN` `jobqueue:worker:*:processing` with `COUNT=N` (default 100) and abort each pass after ~200ms to avoid starving workers.
- Between SCAN pages, sleep for `base_delay` plus ±50% jitter so fleets do not synchronize.
- For each candidate job:
  - Skip when `worker:hb:<id>` exists and is newer than `heartbeat_ttl`.
  - Execute a Lua script that atomically verifies heartbeat absence, removes the list entry, and `LPUSH`es the payload back to `origin_queue` (or DLQ once retries exhausted).
- Record `reaper_jobs_moved_total` and structured logs for observability, then wait `reaper_interval` with ±25% jitter before the next pass.

### Circuit Breaker

- Sliding window of recent results; failure rate `fails/total` compared to `failure_threshold`.
- States: Closed → Open on threshold; Open → HalfOpen on cooldown; HalfOpen → Closed on success, else Open.

## Metrics

- Counters (labels in parentheses, capped at ≤50 values):
  - `jobs_produced_total{queue}`
  - `jobs_consumed_total{queue}`
  - `jobs_completed_total{queue}`
  - `jobs_failed_total{queue,reason}` (reason from bounded enum)
  - `jobs_retried_total{queue}`
  - `jobs_dead_letter_total{queue}`
- Histogram:
  - `job_processing_duration_seconds{queue}` (bucket unit: seconds; suffix already `_seconds`).
- Gauges:
  - `queue_length{queue}` (queue names validated against configured allowlist; all others map to `other`).
  - `circuit_breaker_state{queue}` (0 Closed, 1 Half-Open, 2 Open).
- Validation: `promtool check metrics` in CI plus unit tests ensure label sets stay bounded.

## Logging and Tracing

- Logs are JSON with canonical keys: `level`, `ts`, `msg`, `trace_id`, `span_id`, `job_id`, `queue`, `worker_id`, `request_id`, `namespace`. Secrets, payloads, or PII are forbidden; the logging helper rejects non-allowlisted keys.
- Tracing: create a span per job processing; propagate `trace_id/span_id` if present; otherwise create new, tagging spans with `queue`, `job_id`, and outcome. Linting enforces presence of these attributes in PR review.
