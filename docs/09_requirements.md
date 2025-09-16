# Requirements

- Last updated: 2025-09-12

## Executive Summary

Functional and non-functional requirements with user stories, acceptance criteria, and definition of done for a production-ready release.

## Table of Contents

- [Functional Requirements](#functional-requirements)
- [Non-Functional Requirements](#non-functional-requirements)
- [User Stories](#user-stories)
- [Acceptance Criteria](#acceptance-criteria)
- [Definition of Done](#definition-of-done)

## Functional Requirements

- Producer scans directory, prioritizes by extension, and enqueues JSON jobs. (complexity: ~80 LoC; O(1) `LPUSH` per job)
- Worker pool consumes jobs by priority; atomic move to processing list; heartbeat; retries with backoff; DLQ. (~200 LoC; O(1) list ops)
- Reaper detects missing heartbeats and requeues abandoned jobs. (~100 LoC)
- Circuit breaker with Closed/Open/HalfOpen states; cooldown. (~120 LoC)
- Observability: Prometheus metrics, structured logs, optional tracing. (~150 LoC)
- Config: YAML + env overrides; validation with descriptive errors. (~120 LoC)
- Health endpoints: `/healthz` and `/readyz`. (~60 LoC)
- Admin tooling: `stats`, `peek`, `purge-dlq`. (~120 LoC)

## Non-Functional Requirements

- Performance: ≥1k jobs/min per 4 vCPU node; p95 < 2s small files.
- Reliability: no job loss; DLQ for failures; reaper recovers within 10s.
- Security: no secrets in logs; dependencies free of critical CVEs in CI.
- Usability: single binary; clear CLI; documented examples and configs.

## User Stories

- As a producer operator, I can limit enqueue rate to prevent flooding Redis.
- As a worker operator, I can scale worker count and observe breaker state.
- As an SRE, I can monitor queue lengths and processing latencies.
- As a developer, I can trace a job with its TraceID/SpanID across logs and traces.
- As a platform engineer, I can purge DLQ safely after exporting.

## Acceptance Criteria

- Each requirement has unit and/or integration tests.
- Metrics appear in `/metrics` with expected names and types.
- Health endpoints return correct status based on readiness.
- Admin commands operate without data loss and require confirmation when destructive.

## Definition of Done

- Code merged with passing CI (unit, race, integration).
- Documentation updated (README, PRD, relevant docs under `docs/`).
- Version bumped and CHANGELOG updated for releases.
