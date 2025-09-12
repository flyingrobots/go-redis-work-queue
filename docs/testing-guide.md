# Testing Guide

This guide describes how to build and run the test suites, what each test verifies, and copy/paste commands to run any test in isolation.

## Prerequisites
- Go 1.25+
- Git
- Docker (only for the e2e test that talks to a real Redis)

## Quick build


## Run all tests (race detector)


Notes
- Use  to avoid cached results during iteration.
- Add  for verbose output.

## Package-by-package suites

### internal/config
- Tests: configuration defaults and validation logic
  -  — Ensures reasonable defaults load without a file
  -  — Asserts invalid configs produce descriptive errors

Run the whole package:

Run a single test:


### internal/breaker
- Tests: circuit breaker state machine and HalfOpen semantics
  -  — Closed → Open; HalfOpen probe; Close
  -  — Under heavy concurrency, HalfOpen admits one probe only

Run the whole package:

Run a single test:


### internal/queue
- Tests: job serialization round-trip
  -  — JSON encode/decode preserves fields

Run:


### internal/producer
- Tests: priority mapping and rate limiter behavior
  -  —  → high, others → default (low)
  -  — Exceeding the fixed-window cap sleeps until TTL expiry

Run:

Run a single test:


### internal/reaper
- Tests: requeue without heartbeat using miniredis
  -  — Orphans in processing list are moved back to source queue

Run:


### internal/worker
- Tests: backoff, success/retry/DLQ paths, and breaker integration
  -  — Exponential backoff caps at configured max
  -  — Happy-path processing
  -  — Retry then move to DLQ after threshold
  -  — Failures trip breaker; consumption pauses while Open

Run the whole package:

Run a single test:


### test/e2e (real Redis)
End-to-end test that talks to a real Redis server. Start Redis locally with Docker, set an env var, then run.

Start Redis:

Run the e2e suite:

Run the single e2e test:

Stop Redis:


## Common flags and tips
-  — enable the race detector (recommended)
-  — verbose output
-  — run a single test by name (regex)
-  — disable caching
-  — generate coverage report

Coverage example:


## Troubleshooting
- “connection refused” in e2e: ensure Redis is running and  points to it
- Flaky timings: add  and rerun with ; CI also runs the e2e test 5× to catch flakiness
- Missing Go tools: ensure go version go1.25.0 darwin/arm64 reports 1.25+

