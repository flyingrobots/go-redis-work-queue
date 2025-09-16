# Product Roadmap

- Last updated: 2025-09-12

## Executive Summary

This roadmap sequences the remaining work to reach v1.0.0 and beyond. Priorities are reliability, observability, and operational readiness. We map initiatives quarterly with explicit dependencies and a standing monthly review cadence.

## Table of Contents

- [Objectives](#objectives)
- [Quarterly Roadmap](#quarterly-roadmap)
- [Initiative Dependencies](#initiative-dependencies)
- [Business Priorities](#business-priorities)
- [Review and Update Cadence](#review-and-update-cadence)

## Objectives

- Ship v1.0.0 by 2025-11-07 with production readiness.
- Support sustained throughput of 1k jobs/min per 4 vCPU node.
- Provide actionable metrics, health endpoints, and robust recovery.

## Quarterly Roadmap

### Q3 2025 (Sep)

- Prioritized dequeue strategy finalized and documented
- Health/readiness endpoints
- Queue length gauges and config validation
- Tracing propagation from Job TraceID/SpanID
- Alpha release v0.4.0 (2025-09-26)

### Q4 2025 (Oct–Dec)

- Beta hardening, admin tooling (peek/purge/list)
- Performance tuning and load tests; pool sizing guidance
- E2E tests with real Redis; chaos testing scenarios
- RC release v0.9.0 (2025-10-24), GA v1.0.0 (2025-11-07)
  - Release gates (see `docs/15_promotion_checklists.md` and `.github/workflows/release.yml`):
    - CI jobs must be green: `unit`, `integration`, `e2e-with-redis`, `security-scan`, `performance-smoke`, `deploy-preview`
    - Branch protection enabled on `main`; no bypass merges
    - Promotion checklist signed by Release DRI and QA DRI
    - Verify artifacts published (container image, binary bundles, Terraform module)
- Post-1.0: Helm chart and Docker Compose examples (Dec)

### Q1 2026

- Horizontal sharding guidance and queue-partitioning patterns
- Optional Redis Streams backend as an alternative
- Advanced observability: exemplars, RED metrics, SLO dashboards

## Initiative Dependencies

- Tracing propagation — owner: @alice (Observability) — depends on Job struct update (PR #123) and `docs/06_technical_spec.md#job-schema` review.
- Reaper improvements — owner: @bob (Runtime) — depends on heartbeat semantics definition (`docs/06_technical_spec.md#reaper`) and implementation PR #145.
- Performance tuning — owner: @carol (Perf Guild) — depends on priority dequeue semantics (PR #130) and metrics completeness plan (`docs/12_performance_baseline.md`).

## Business Priorities

1) Reliability and data safety (DLQ, retries, reaper) — P0
2) Operational visibility (metrics, health, tracing) — P0
3) Performance and scale guidance — P1
4) Operational tooling (admin CLI) — P1
5) Packaging and deployment assets (Docker/Helm) — P2

## Review and Update Cadence

- Monthly roadmap review on first business day of each month.
- Sprint reviews bi-weekly; adjust scope based on findings.
