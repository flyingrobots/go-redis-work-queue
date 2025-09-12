# Operations Runbook

- Last updated: 2025-09-12

## Executive Summary
Day-2 operations guide: deploy, scale, monitor, recover, and release/rollback procedures for the Go Redis Work Queue.

## Table of Contents
- [Deployment](#deployment)
- [Configuration](#configuration)
- [Health and Monitoring](#health-and-monitoring)
- [Scaling](#scaling)
- [Common Operations](#common-operations)
- [Troubleshooting](#troubleshooting)
- [Release and Rollback](#release-and-rollback)

## Deployment
- Docker: `docker build -t job-queue-system:local .` then run with flags.
- docker-compose: see `deploy/docker-compose.yml` (services: redis, app-all, app-worker, app-producer).
- Container image: `ghcr.io/<owner>/<repo>:<tag>` published on git tags (see release workflow).

## Configuration
- Primary: `config/config.yaml` (see `config/config.example.yaml`).
- Overrides: environment vars (upper snake case replaces dots, e.g., `WORKER_COUNT=32`).
- Validate: service fails to start with descriptive errors on invalid configs.

## Health and Monitoring
- Liveness: `/healthz` returns 200 when the process is up.
- Readiness: `/readyz` returns 200 when Redis is reachable.
- Metrics: `/metrics` exposes Prometheus counters/gauges/histograms:
  - jobs_* counters, job_processing_duration_seconds, queue_length{queue}, circuit_breaker_state, worker_active.

## Scaling
- Horizontal: run more worker instances; each instance can run N workers (`worker.count`).
- Redis: ensure adequate CPU and memory; monitor latency and ops/sec.
- Pooling: tune `redis.pool_size_multiplier`, `min_idle_conns` for throughput and latency.

## Common Operations
- Inspect stats:
  `./job-queue-system --role=admin --admin-cmd=stats --config=config.yaml`
- Peek queue items:
  `./job-queue-system --role=admin --admin-cmd=peek --queue=high --n=20 --config=config.yaml`
- Purge dead-letter queue:
  `./job-queue-system --role=admin --admin-cmd=purge-dlq --yes --config=config.yaml`
- Benchmark throughput/latency:
  `./job-queue-system --role=admin --admin-cmd=bench --bench-count=2000 --bench-rate=1000 --bench-priority=low --bench-timeout=60s`

## Troubleshooting
- High failures / breaker open:
  - Check Redis latency and CPU; verify timeouts.
  - Inspect logs for job-specific errors; consider reducing worker.count temporarily.
- Growing DLQ:
  - Peek/Dump items, assess causes; adjust max_retries/backoff; fix processing logic.
- Stuck processing lists:
  - Verify heartbeats; reaper should recover; run stats to confirm processing list sizes drop.
- Readiness failing:
  - Check Redis availability and credentials; verify network and firewall.

## Release and Rollback
- Versioning: SemVer; `--version` prints build version.
- Release: push tag `vX.Y.Z` to trigger release workflow; image published to GHCR.
- Rollback:
  1) Select previous good tag (e.g., `vX.Y.(Z-1)`).
  2) Deploy image `ghcr.io/<owner>/<repo>:vX.Y.(Z-1)`.
  3) Verify `/healthz` and `/readyz` return 200.
  4) Check metrics: `circuit_breaker_state=0`, `jobs_failed_total` steady, `queue_length` normalized.
  5) If DLQ large, export/inspect before purge.

