# Performance Baseline and Tuning Guide

- Last updated: 2025-09-12

## Executive Summary

This guide provides a reproducible method to measure throughput and latency, and offers tuning steps to hit the target 1k jobs/min per 4 vCPU node with p95 < 2s for small files.

## Table of Contents

- [Methodology](#methodology)
- [Baseline Procedure](#baseline-procedure)
- [Tuning Levers](#tuning-levers)
- [Expected Results](#expected-results)

## Methodology

- Use the built-in admin bench command to enqueue N jobs at a target rate to a priority queue.
- Run worker-only mode on the same host (or separate host) and measure completion latency based on job creation_time vs. wall clock.
- Metrics are exposed at `/metrics`; verify `jobs_completed_total`, `job_processing_duration_seconds`, and `queue_length{queue}`.

## Baseline Procedure

### Test Environment

- Host: 4 vCPU (Intel Xeon Gold 6338, 16 GB RAM) running Ubuntu 24.04 LTS
- Redis: `redis:7.2.4-alpine` container with AOF disabled, `maxmemory-policy=noeviction`, `tcp-keepalive 60`
- Payload: synthetic NDJSON (1 KB per job) generated via the bench command’s `--bench-payload-size` flag

> Adjust the numbers to match your hardware; record CPU model, core count, RAM, and Redis config when sharing results.

### Procedure

1) Start Redis

```bash
docker run --rm -d --name jobq-redis -p 6379:6379 redis:7.2.4-alpine
```

When finished, tear down the container:

```bash
docker stop jobq-redis
```

2) Copy `config/config.example.yaml` to `config/config.yaml` and set:
   - `worker.count`: 16 on a 4 vCPU node (adjust as needed)
   - `redis.addr`: "localhost:6379" (matches the container mapping above)
3) In one shell, run the worker

```bash
./bin/job-queue-system --role=worker --config=config/config.yaml
```

4) In another shell, run the bench (enqueue and wait for completion)

```bash
./bin/job-queue-system --role=admin --admin-cmd=bench \
  --bench-count=2000 --bench-rate=1000 \
  --bench-priority=low --bench-payload-size=1024 \
  --bench-timeout=60s
```

5) Record the JSON result and capture Prometheus metrics (if scraping locally, curl /metrics).

## Tuning Levers

- Redis pool: `redis.pool_size_multiplier` (default 10*NumCPU). Increase for higher concurrency; monitor Redis CPU.
- Timeouts: `redis.read_timeout`/`write_timeout` (default 3s). Too low yields errors under load; too high slows failure detection.
- Worker concurrency: `worker.count`. Increase up to CPU saturation; watch goroutine scheduling and Redis ops.
- Backoff parameters: for retry behavior; not relevant for the baseline success path.
- Priority timeout: `worker.brpoplpush_timeout` (default 1s). Smaller values reduce low-priority latency but add Redis ops.

## Expected Results

- On a 4 vCPU node, `bench-count=2000`, `bench-rate=1000` should achieve ≥1k jobs/min throughput, with p95 latency < 2s for small files (<1MB).
- If results fall short, see tuning levers and ensure host/Redis are not CPU or I/O bound.
