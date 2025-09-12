# Release Plan

- Last updated: 2025-09-12

## Executive Summary
Three pre-GA releases (alpha, beta, RC) precede v1.0.0. Each release has specific feature scope, risks with mitigations, and acceptance criteria. Target GA date: 2025-11-07.

## Table of Contents
- [Release Cadence](#release-cadence)
- [Release Scope](#release-scope)
- [Risks and Mitigations](#risks-and-mitigations)
- [Acceptance Criteria](#acceptance-criteria)

## Release Cadence
- v0.4.0-alpha — 2025-09-26
- v0.7.0-beta — 2025-10-10
- v0.9.0-rc — 2025-10-24
- v1.0.0 — 2025-11-07

## Release Scope

### v0.4.0-alpha (2025-09-26)
- Health/readiness endpoints
- Queue length gauges; periodic updater
- Tracing propagation (TraceID/SpanID)
- Config validation and helpful error messages
- Document prioritized dequeue strategy and guarantees

### v0.7.0-beta (2025-10-10)
- Admin CLI subcommands: `stats`, `peek`, `purge-dlq`
- Rate limiter improvements (jitter, precise window sleep)
- E2E tests with real Redis via service container
- Initial performance baseline doc and tuning guidance

### v0.9.0-rc (2025-10-24)
- Hardening fixes from beta feedback
- Chaos tests (Redis unavailability, slow network)
- Security checks (govulncheck) wired in CI
- Documentation complete: ops runbook, dashboards guidance

### v1.0.0 (2025-11-07)
- Final polish, issue triage zero, CHANGELOG, version pinning
- Docker image published; example deployment assets

## Risks and Mitigations

| Risk | Prob. | Impact | Mitigation | Contingency |
|------|-------|--------|------------|-------------|
| Priority dequeue semantics disputed | Medium | 4 | Document guarantees; keep looped BRPOPLPUSH with low latency; proof via tests | Feature flag to choose strategy; offer alternate queue design |
| Redis outages longer than TTL | Medium | 5 | Robust retries, circuit breaker; reaper recovery | Backoff to fail-fast mode; admin docs to drain/retry DLQ |
| Throughput below target on modest nodes | Low | 4 | Tune pool sizes, pipelining where safe; profiling | Scale-out workers; document capacity planning |
| Metrics cardinality explosion | Low | 3 | Limit labels, provide sampling | Configurable metric collection intervals |
| Tracing exporter instability | Low | 2 | Make tracing optional; timeouts | Disable tracing by config |

## Acceptance Criteria
- All acceptance tests for the release pass in CI (unit, integration, race).
- Metrics present: job counters, histograms, queue length gauges, breaker state.
- Health endpoints return HTTP 200 and surface readiness correctly.
- On chaos tests: automatic recovery without manual intervention; no lost jobs.
- Documentation updated (README, PRD, ops guide) for each release.

