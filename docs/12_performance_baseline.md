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
1) Start Redis (e.g., docker run -p 6379:6379 redis:7-alpine)
2) Copy `config/config.example.yaml` to `config/config.yaml` and set:
   - `worker.count`: 16 on a 4 vCPU node (adjust as needed)
   - `redis.addr`: "localhost:6379"
3) In one shell, run the worker:
   - `./bin/job-queue-system --role=worker --config=config/config.yaml`
4) In another shell, run the bench (enqueue and wait for completion):
   - `./bin/job-queue-system --role=admin --admin-cmd=bench --bench-count=2000 --bench-rate=1000 --bench-priority=low --bench-timeout=60s`
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

