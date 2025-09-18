# Release Plan

- Status: Actively maintained

## Executive Summary

Three pre-GA releases (alpha, beta, RC) precede v1.0.0. Each milestone ships when its gates clear; there is no fixed calendar, so freeze windows are triggered by progress rather than dates. Every release must exit code freeze with clean security scans (no High/Critical CVEs, `govulncheck` clean) before promotion.

## Table of Contents

- [Release Cadence](#release-cadence)
- [Release Scope](#release-scope)
- [Risks and Mitigations](#risks-and-mitigations)
- [Acceptance Criteria](#acceptance-criteria)

## Release Cadence

| Release | Status | Freeze Window Policy | Sign-off Owners |
|---------|--------|----------------------|-----------------|
| v0.4.0-alpha | Active | 72h code freeze before promotion once backlog stories move to "verify" | Morgan Lee (Release), Priya Shah (QA) |
| v0.7.0-beta | Queued | 72h freeze once alpha retrospective closes | Morgan Lee (Release), Priya Shah (QA), Alice Nguyen (Observability) |
| v0.9.0-rc | Planned | 72h freeze triggered after beta bug scrub completes | Priya Shah (QA), Daniel Reed (SRE) |
| v1.0.0 | Planned | 96h freeze following RC acceptance | Morgan Lee (Release), Jamie Patel (Platform) |

## Release Scope

### v0.4.0-alpha

- **Status:** Active build-up
- **Freeze window:** 72h code freeze before tagging; enforced via branch protection once checklist enters verification
- **Gate owners:** Morgan Lee (Release), Priya Shah (QA)
- **Rollback plan:** Revert to prior stable tag, redeploy previous manifests, replay smoke tests, and announce rollback in release channel
- **Scope:**
  - Health/readiness endpoints
  - Queue length gauges with periodic updater
  - Tracing propagation (TraceID/SpanID)
  - Config validation with helpful error messages
  - Document prioritized dequeue strategy and guarantees
- **Go/No-Go gates:**
  - CI green across unit, integration, race suites
  - `govulncheck ./...` clean
  - No High/Critical CVEs in dependency scans
  - Release checklist signed by gate owners

### v0.7.0-beta

- **Status:** Queued; starts after alpha promotion
- **Freeze window:** 72h freeze initiated once alpha retro actions are closed
- **Gate owners:** Morgan Lee (Release), Priya Shah (QA), Alice Nguyen (Observability)
- **Rollback plan:** Re-issue alpha tag, restore alpha manifests, and disable beta-only toggles
- **Scope:**
  - Admin CLI subcommands: `stats`, `peek`, `purge-dlq`
  - Rate limiter improvements (jitter, precise window sleep)
  - E2E tests with real Redis via service container
  - Initial performance baseline doc and tuning guidance
- **Go/No-Go gates:**
  - CI green including new E2E suite
  - `govulncheck ./...` clean
  - No High/Critical CVEs
  - Observability dashboards reviewed by Alice Nguyen

### v0.9.0-rc

- **Status:** Planned
- **Freeze window:** 72h freeze triggered after beta bug scrub completes
- **Gate owners:** Priya Shah (QA), Daniel Reed (SRE)
- **Rollback plan:** Re-deploy beta artifacts, keep RC branch open for fixes, announce rollback to stakeholders
- **Scope:**
  - Hardening fixes from beta feedback
  - Chaos tests (Redis unavailability, slow network)
  - Security checks (`govulncheck`, dependency scans) wired in CI
  - Documentation complete: ops runbook, dashboards guidance
- **Go/No-Go gates:**
  - Chaos suite passes without manual intervention
  - `govulncheck ./...` clean
  - No High/Critical CVEs
  - SRE sign-off on runbooks

### v1.0.0

- **Status:** Planned
- **Freeze window:** 96h freeze following RC acceptance to absorb final regression tests
- **Gate owners:** Morgan Lee (Release), Jamie Patel (Platform)
- **Rollback plan:** Revert to RC tag, pause GA announcements, and re-run acceptance once blockers resolved
- **Scope:**
  - Final polish, issue triage zero, CHANGELOG, version pinning
  - Docker image published; example deployment assets
  - GA checklist completed with docs and ops validation
- **Go/No-Go gates:**
  - CI matrix green
  - `govulncheck ./...` clean
  - No High/Critical CVEs
  - GA launch review signed by gate owners

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
- Chaos tests demonstrate automatic recovery without manual intervention; no lost jobs.
- Documentation updated (README, PRD, ops guide) for each release.
- Security gates: `govulncheck ./...` clean and no High/Critical CVEs prior to promotion.
