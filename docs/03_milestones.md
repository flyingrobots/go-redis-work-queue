# Milestones

- Last updated: 2025-09-12

## Executive Summary

Milestones define concrete deliverables with success criteria, dependencies, and decision gates. These map to the alpha, beta, RC, and GA releases.

## Table of Contents

- [Milestone List](#milestone-list)
- [Dependencies](#dependencies)
- [Go/No-Go Decision Gates](#gono-go-decision-gates)

## Milestone List

1. Health + Validation (due 2025-09-20)
   - Deliverables: `/healthz` and `/readyz` endpoints; strict config validation with errors; docs updated
   - Success: endpoints return 200; malformed configs fail with descriptive messages
2. Observability Depth (due 2025-09-24)
   - Deliverables: queue length gauges updated every 2s; TraceID/SpanID used for spans and logs
   - Success: metrics visible under load; traces exported when enabled
3. Alpha Ship (due 2025-09-26)
   - Deliverables: v0.4.0 tag; CHANGELOG; release notes
   - Success: CI green; basic load test at 500 jobs/min sustained
4. Admin Tooling + Rate Limiter (due 2025-10-05)
   - Deliverables: `stats`, `peek`, `purge-dlq` CLI; limiter jitter + precise window sleep
   - Success: commands operate safely; limiter smooths bursts; docs updated
5. E2E + Perf Baseline (due 2025-10-08)
   - Deliverables: real Redis service container tests; perf doc; pool sizing guidance
   - Success: 1k jobs/min on 4 vCPU passes with p95 < 2s small files
6. RC Ship + Chaos (due 2025-10-24)
   - Deliverables: chaos tests; security checks in CI; RC release v0.9.0
   - Success: automatic recovery from injected failures; zero critical vulns
7. GA (due 2025-11-07)
   - Deliverables: docs and ops runbook; GA release; example deploy assets
   - Success: zero P0/P1 issues; CI green across matrix

## Dependencies

- 2 depends on 1 (observability requires validated config and endpoints)
- 5 depends on 4 (perf needs stabilized limiter and tools)
- 6 depends on 5 (chaos built on validated e2e)
- 7 depends on 6 (GA after RC validation)

## Go/No-Go Decision Gates

- Alpha (2025-09-26): must meet milestone 1–3 success criteria.
- Beta (2025-10-10): limiter behavior validated, admin tools safe; perf doc drafted.
- RC (2025-10-24): chaos tests pass; no critical security findings.
- GA (2025-11-07): p95 < 2s small files; SLOs defined; docs complete.
