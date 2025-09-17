# Test Plan

- Last updated: 2025-09-12

## Executive Summary

This plan defines coverage goals, scenarios, performance benchmarks, and security testing to ensure production readiness by v1.0.0.

## Table of Contents

- [Coverage Goals](#coverage-goals)
- [Test Types](#test-types)
- [Critical Path Scenarios](#critical-path-scenarios)
- [Performance Benchmarks](#performance-benchmarks)
- [Security Testing](#security-testing)

## Coverage Goals

- Unit: ≥ 80% on core packages (config, worker, reaper, breaker, producer)
- Integration: end-to-end flows with Redis service container
- Race detector: enabled in CI for all tests

## Test Types

- Unit: algorithms (breaker), backoff, job marshal/unmarshal, rate limiter math, config validation.
- Integration: produce→consume→complete/retry/DLQ; reaper resurrection; graceful shutdown.
- E2E: GitHub Actions job with Redis service container; real network timings.
- Chaos (deterministic, automated):
  1. **Redis SIGSTOP (30s stall):**
     - Command: `docker exec redis pkill -STOP redis-server` followed by `sleep 30 && docker exec redis pkill -CONT redis-server`.
     - Pass criteria: workers retry without data loss, queue depth stabilises within 60s after resume, no 5xx spike >1%.
  2. **Latency injection (tc netem):**
     - Command: `tc qdisc add dev eth0 root netem delay 200ms 50ms distribution normal` on the Redis container/host for 2 minutes; remove with `tc qdisc del dev eth0 root netem`.
     - Pass criteria: p95 worker latency stays <5s, requeues remain <1%, rate limiter continues to operate.
  3. **Connection resets (5% loss):**
     - Command: `iptables -A OUTPUT -p tcp --dport 6379 -m statistic --mode random --probability 0.05 -j REJECT --reject-with tcp-reset`; clean up with corresponding `iptables -D`.
     - Pass criteria: retry/backoff logic clears transient errors within 10s, breaker does not enter permanent open state, no job duplication observed.

## Critical Path Scenarios

1) Single worker: consume success path; completed recorded; heartbeat deleted.
2) Retry then requeue: failure increments retry, backoff, LPUSH back; processing cleaned.
3) DLQ after threshold: job moved to DLQ; counters updated.
4) Producer rate limit: per-second cap respected within ±10% under burst.
5) Reaper: missing heartbeat → processing list drained → requeued to original priority.
6) Circuit breaker: threshold exceeded → Open; cooldown → HalfOpen; single probe → Closed on success.
7) Graceful shutdown: no lost jobs; in-flight completes or is requeued.

## Performance Benchmarks

- Baseline: 1k jobs/min per 4 vCPU node; p95 < 2s for small files.
- Method: synthetic job generation via producer; worker-only mode on dedicated runner; capture metrics.
- Reporting: include CPU, memory, Redis CPU/latency, queue depths.

## Security Testing

- `govulncheck` in CI; fail on critical CVEs.
- Static checks: `go vet` and `golangci-lint` (optional) for code issues.
- Secrets: ensure no secrets in logs; validate config does not dump secrets.
