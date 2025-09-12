# Testing Guide

This guide describes how to build and run the test suites, what each test verifies, and copy/paste commands to run any test in isolation.

## Prerequisites
- Go 1.25+
- Git
- Docker (only for the e2e test that talks to a real Redis)

## Quick build
```bash
make build
```

## Run all tests (race detector)
```bash
go test ./... -race -count=1
```

Notes
- Use `-count=1` to avoid cached results during iteration.
- Add `-v` for verbose output.

## Package-by-package suites

### internal/config
- Tests: configuration defaults and validation logic
  - `TestLoadDefaults` — Ensures reasonable defaults load without a file
  - `TestValidateFails` — Asserts invalid configs produce descriptive errors

Run the whole package:
```bash
go test ./internal/config -race -count=1 -v
```
Run a single test:
```bash
go test ./internal/config -run '^TestLoadDefaults$' -race -count=1 -v
```

### internal/breaker
- Tests: circuit breaker state machine and HalfOpen semantics
  - `TestBreakerTransitions` — Closed → Open; HalfOpen probe; Close
  - `TestBreakerHalfOpenSingleProbeUnderLoad` — Under heavy concurrency, HalfOpen admits one probe only

Run the whole package:
```bash
go test ./internal/breaker -race -count=1 -v
```
Run a single test:
```bash
go test ./internal/breaker -run '^TestBreakerHalfOpenSingleProbeUnderLoad$' -race -count=1 -v
```

### internal/queue
- Tests: job serialization round-trip
  - `TestMarshalUnmarshal` — JSON encode/decode preserves fields

Run:
```bash
go test ./internal/queue -race -count=1 -v
```

### internal/producer
- Tests: priority mapping and rate limiter behavior
  - `TestPriorityForExt` — `.pdf` → high, others → default (low)
  - `TestRateLimit` — Exceeding the fixed-window cap sleeps until TTL expiry

Run:
```bash
go test ./internal/producer -race -count=1 -v
```
Run a single test:
```bash
go test ./internal/producer -run '^TestRateLimit$' -race -count=1 -v
```

### internal/reaper
- Tests: requeue without heartbeat using miniredis
  - `TestReaperRequeuesWithoutHeartbeat` — Orphans in processing list are moved back to source queue

Run:
```bash
go test ./internal/reaper -race -count=1 -v
```

### internal/worker
- Tests: backoff, success/retry/DLQ paths, and breaker integration
  - `TestBackoffCaps` — Exponential backoff caps at configured max
  - `TestProcessJobSuccess` — Happy-path processing
  - `TestProcessJobRetryThenDLQ` — Retry then move to DLQ after threshold
  - `TestWorkerBreakerTripsAndPausesConsumption` — Failures trip breaker; consumption pauses while Open

Run the whole package:
```bash
go test ./internal/worker -race -count=1 -v
```
Run a single test:
```bash
go test ./internal/worker -run '^TestWorkerBreakerTripsAndPausesConsumption$' -race -count=1 -v
```

### test/e2e (real Redis)
End-to-end test that talks to a real Redis server. Start Redis locally with Docker, set an env var, then run.

Start Redis:
```bash
docker run --rm -d --name jobq-redis -p 6379:6379 redis:7-alpine
```
Run the e2e suite:
```bash
E2E_REDIS_ADDR=localhost:6379 go test ./test/e2e -race -count=1 -v
```
Run the single e2e test:
```bash
E2E_REDIS_ADDR=localhost:6379 go test ./test/e2e -run '^TestE2E_WorkerCompletesJobWithRealRedis$' -race -count=1 -v
```
Stop Redis:
```bash
docker rm -f jobq-redis
```

## Common flags and tips
- `-race` — enable the race detector (recommended)
- `-v` — verbose output
- `-run '^TestName$'` — run a single test by name (regex)
- `-count=1` — disable caching
- `-coverprofile=coverage.out` — generate coverage report

Coverage example:
```bash
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | sort -k3 -r
```

## Troubleshooting
- “connection refused” in e2e: ensure Redis is running and `E2E_REDIS_ADDR` points to it
- Flaky timings: add `-v` and rerun with `-count=1`; CI also runs the e2e test 5× to catch flakiness
- Missing Go tools: ensure `go version` reports 1.25+

