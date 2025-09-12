# Promotion Checklists

- Last updated: 2025-09-12

## Alpha → Beta Checklist
- [ ] Functional completeness: producer, worker, all-in-one, reaper, breaker, admin CLI
- [ ] Observability: /metrics, /healthz, /readyz live and correct
- [ ] CI green on main (build, vet, race, unit, integration, e2e)
- [ ] Unit coverage ≥ 80% core packages (attach coverage report)
- [ ] E2E passes deterministically (≥ 5 runs)
- [ ] govulncheck: no Critical/High in code paths/stdlib
- [ ] Performance baseline: 1k jobs at 500/s complete; limiter ±10%/60s
- [ ] Docs: README, PRD, test plan, deployment, runbook updated
- Evidence links:
  - CI run URL: …
  - Bench JSON: …
  - Metrics snapshot(s): …
  - Issues list: …

## Beta → RC Checklist
- [ ] Throughput ≥ 1k jobs/min for ≥ 10m; p95 < 2s (<1MB files)
- [ ] Chaos tests: Redis outage/latency/worker crash → no lost jobs; breaker transitions
- [ ] Admin CLI validated against live instance
- [ ] Queue gauges and breaker metric accurate under load
- [ ] 24–48h soak: error rate < 0.5%, no leaks
- [ ] govulncheck clean; deps pinned
- [ ] Docs: performance report and tuning
- [ ] No P0/P1; ≤ 3 P2s w/ workarounds
- Evidence links as above

## RC → GA Checklist
- [ ] Code freeze; only showstopper fixes
- [ ] 0 P0/P1; ≤ 2 P2s with workarounds; no flakey tests across 10 runs
- [ ] Release workflow proven; rollback rehearsal complete
- [ ] Config/backcompat validated or migration guide
- [ ] Docs complete; README examples validated
- [ ] govulncheck clean; image scan no Critical
- [ ] 7-day RC soak: readiness > 99.9%, DLQ < 0.5%
- Evidence links as above

