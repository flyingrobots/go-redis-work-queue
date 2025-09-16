# Risk Register

- Last updated: 2025-09-12

## Executive Summary

Top project risks with probability, impact, mitigation strategies, and contingency plans.

## Table of Contents

- [Risk Matrix](#risk-matrix)

## Risk Matrix

| # | Risk | Prob. | Impact (1-5) | Mitigation | Contingency |
|---|------|-------|--------------|------------|-------------|
| 1 | Priority dequeue semantics disagreement | Medium | 4 | Document guarantees; test; feature flag strategies | Offer alt queue design doc |
| 2 | Redis outages or high latency | Medium | 5 | Retries, breaker, backoff | Fail-fast mode; admin runbook |
| 3 | Throughput below targets | Low | 4 | Profiling, pool tuning | Scale out; shard queues |
| 4 | Metric cardinality growth | Low | 3 | Limit labels; sampling | Configurable scrape/update intervals |
| 5 | Tracing instability or cost | Low | 2 | Optional; timeouts | Disable tracing |
| 6 | Config misuse in prod | Medium | 3 | Validation + examples | Safe defaults; startup guardrails |
| 7 | Reaper over-scan in large fleets | Low | 3 | SCAN pacing; page size limits | Spread schedules; shard processing keys |
| 8 | CI flakiness (timing-sensitive tests) | Medium | 3 | Retries; timeouts; use service containers | Quarantine and fix tests |
| 9 | Security vulnerabilities in deps | Medium | 4 | `govulncheck` in CI | Pin versions; patch on release |
| 10 | Operator error on DLQ purge | Low | 4 | Confirmation prompts; export-first guidance | Restore from backup/export |
