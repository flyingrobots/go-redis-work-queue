# Sprint Plans

- Last updated: 2025-09-12

## Executive Summary

Four bi-weekly sprints lead to v1.0.0. Each sprint contains user stories with acceptance criteria, tasks, and estimates.

## Table of Contents

- [Sprint 1 (2025-09-12 → 2025-09-25)](#sprint-1-2025-09-12--2025-09-25)
- [Sprint 2 (2025-09-26 → 2025-10-09)](#sprint-2-2025-09-26--2025-10-09)
- [Sprint 3 (2025-10-10 → 2025-10-23)](#sprint-3-2025-10-10--2025-10-23)
- [Sprint 4 (2025-10-24 → 2025-11-06)](#sprint-4-2025-10-24--2025-11-06)

## Sprint 1 (2025-09-12 → 2025-09-25)

Stories (points):

1) As an operator, I need `/healthz` and `/readyz` so I can probe liveness/readiness. (5)
   - Acceptance: `/healthz`=200 always after start; `/readyz`=200 only when Redis reachable and metrics server running.
   - Tasks: add handlers, wire checks, tests, docs.
2) As an SRE, I need queue length gauges updated periodically. (3)
   - Acceptance: `queue_length{queue}` updated every 2s; tested with miniredis.
   - Tasks: background updater, config interval, tests.
3) As a developer, I want config validation errors to be explicit. (3)
   - Acceptance: invalid keys/values produce descriptive errors; unit tests.
   - Tasks: validation function, schema doc, tests.
4) As an engineer, I want TraceID/SpanID to create spans and enrich logs. (5)
   - Acceptance: spans created per job; logs include IDs; toggled by config.
   - Tasks: inject ctx from job, use otel tracer, tests.
5) As a maintainer, I want the prioritized dequeue strategy fully documented. (2)
   - Acceptance: README/PRD updated with guarantees and tradeoffs.

## Sprint 2 (2025-09-26 → 2025-10-09)

Stories (points):

1) As an operator, I can run admin commands to inspect and manage queues. (8)
   - Acceptance: `stats`, `peek`, `purge-dlq` subcommands; safe operations with confirmation; tests.
2) As a producer, my rate limiter smooths bursts with jitter and precise sleep. (5)
   - Acceptance: limiter respects per-second cap within ±10%; tests verify.
3) As a tester, I can run e2e tests against real Redis in CI. (5)
   - Acceptance: GH Actions spawns Redis service; e2e tests pass consistently.
4) As a user, I see clear operational docs and examples. (3)
   - Acceptance: ops runbook sections added; example configs and compose.

## Sprint 3 (2025-10-10 → 2025-10-23)

Stories (points):

1) As a platform team, I need performance guidance and validated numbers. (8)
   - Acceptance: doc with 1k jobs/min baseline and tuning steps; reproducible script.
2) As an SRE, the system recovers from injected failures automatically. (8)
   - Acceptance: chaos tests covering Redis down, latency; no lost jobs; tests pass.
3) As security, I need CI to run `govulncheck`. (3)
   - Acceptance: CI fails on critical vulns; allowlist documented.
4) As a maintainer, I need clear CHANGELOG and versioning. (2)
   - Acceptance: conventional commits and CHANGELOG.md generated.

## Sprint 4 (2025-10-24 → 2025-11-06)

Stories (points):

1) As a user, I want GA-quality docs and samples. (5)
   - Acceptance: comprehensive README, PRD, ops runbook, examples.
2) As a release manager, I want RC stabilization and GA. (5)
   - Acceptance: all blocking bugs fixed; v1.0.0 tagged and released.
3) As DevOps, I need rollback procedures validated. (3)
   - Acceptance: rollback SOP tested; documented step-by-step.
