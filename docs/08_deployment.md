# Deployment Documentation

- Last updated: 2025-09-12

## Executive Summary

Defines CI/CD stages, environments, rollback procedures, monitoring metrics, and alert thresholds for safe operations.

## Table of Contents

- [CI/CD Pipeline](#cicd-pipeline)
- [Environments](#environments)
- [Rollback Procedures](#rollback-procedures)
- [Monitoring and Alerts](#monitoring-and-alerts)

## CI/CD Pipeline

Stages (GitHub Actions):

1) Lint & Vet: run `go vet` (optionally `golangci-lint`)
2) Build: compile all packages
3) Unit + Race: `go test -race ./...`
4) Integration/E2E: Redis service container; full flow tests
5) Security: `govulncheck`
6) Package: Docker image build (main branch tags and releases)
7) Release: tag + changelog on releases

## Environments

- Dev: local machine or CI; Redis via Docker `redis:latest`.
- Prod: managed Redis or self-hosted cluster; binary orchestrated via systemd/K8s.

Config overrides via env vars. Example:

```bash
./your-application-worker --config=config.yaml
```

Application workers embed `pkg/queueworker` and register a non-nil handler.
The repository binary's `--role=worker --bench-worker` mode is only the legacy
benchmark simulator and must not be used to process application payloads.

## Rollback Procedures

1) Identify the target rollback version (last known-good tag).
2) Redeploy binary or container with the previous version.
3) Verify `/healthz` and `/readyz` return 200.
4) Check metrics: breaker state 0, DLQ rate stable, job completion steady.
5) If needed, drain DLQ using `purge-dlq` (with backup/export first).
6) Document incident and root cause.

## Monitoring and Alerts

- Alerts (suggested thresholds):
  - `circuit_breaker_state > 0` for > 60s → WARN
  - `rate(jobs_failed_total[5m]) > 0.1 * rate(jobs_consumed_total[5m])` → CRITICAL
  - `queue_length{queue="jobqueue:dead_letter"} > 100` → WARN
  - `/readyz` non-200 for > 30s → CRITICAL
