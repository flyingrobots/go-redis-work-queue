# Promotion Checklists

- Last updated: 2025-09-12

## Alpha → Beta Checklist

- [x] Functional completeness: producer, worker, all-in-one, reaper, breaker, admin CLI
- [x] Observability: /metrics, /healthz, /readyz live and correct
- [x] CI green on main (build, vet, race, unit, integration, e2e)
- [ ] Unit coverage ≥ 80% core packages (attach coverage report)
- [ ] E2E passes deterministically (≥ 5 runs)
- [x] govulncheck: no Critical/High in code paths/stdlib
- [ ] Performance baseline: 1k jobs at 500/s complete; limiter ±10%/60s
- [x] Docs: README, PRD, test plan, deployment, runbook updated
- Evidence links:
  - CI run URL: …
  - Bench JSON: …
  - Metrics snapshot(s): …
  - Issues list: …

### Confidence Scores (Alpha → Beta)

| Criterion | Confidence | Rationale | How to improve | Status |
|---|---:|---|---|---|
| Functional completeness | 0.90 | All core roles implemented and tested; admin CLI present | Add more E2E admin flows; edge-case docs | Done |
| Observability endpoints | 0.95 | /metrics, /healthz, /readyz live; used in CI/e2e | Add probes to examples; alert rules | Done |
| CI health | 0.90 | CI green (build, vet, race, e2e, govulncheck) | Add matrix (OS/Go), flaky-test detector | Done |
| Coverage ≥ 80% | 0.75 | Gaps in admin/obs packages | Add tests for admin + HTTP handlers | In progress |
| E2E determinism | 0.80 | CI runs e2e 5×; generally stable | Gate on 5× passing and record | In progress |
| Security (govulncheck) | 0.95 | Go 1.25 stdlib; no critical findings | Image scanning; pin base digest | Done |
| Performance baseline | 0.70 | Harness present; prelim results only | Use Prom histograms; 4 vCPU run | In progress |
| Documentation completeness | 0.90 | PRD, runbook, deploy, perf, checklists | Add alert rules + Helm usage | Done |

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

### Confidence Scores (Beta → RC)

| Criterion | Confidence | Rationale | How to improve | Status |
|---|---:|---|---|---|
| ≥1k jobs/min for ≥10m | 0.60 | Not yet run on dedicated 4 vCPU node | Schedule controlled benchmark; record env + metrics | In progress |
| p95 < 2s (<1MB) | 0.60 | Coarse sampling currently | Use Prom histograms; sustained run | In progress |
| Chaos (outage/latency/crash) | 0.70 | Recovery logic exists; partial tests | Add chaos e2e (stop Redis; add latency) | In progress |
| Admin validation | 0.85 | Admin commands + tests | Add e2e assertions for outputs | Ready for review |
| Gauges/breaker accuracy | 0.85 | Metrics wired; observed | Add metric assertions; dashboards | Ready for review |
| 24–48h soak | 0.50 | Not yet performed | Stage soak; capture dashboards | Not started |
| Security and deps | 0.90 | govulncheck green; deps pinned | Add Renovate/Dependabot; image scan | Done |
| Issue hygiene | 0.90 | No open P0/P1 | Enforce labels/triage automation | Done |

## RC → GA Checklist

- [ ] Code freeze; only showstopper fixes
- [ ] 0 P0/P1; ≤ 2 P2s with workarounds; no flakey tests across 10 runs
- [ ] Release workflow proven; rollback rehearsal complete
- [ ] Config/backcompat validated or migration guide
- [ ] Docs complete; README examples validated
- [ ] govulncheck clean; image scan no Critical
- [ ] 7-day RC soak: readiness > 99.9%, DLQ < 0.5%
- Evidence links as above

### Confidence Scores (RC → GA)

| Criterion | Confidence | Rationale | How to improve | Status |
|---|---:|---|---|---|
| Code freeze discipline | 0.80 | Process defined; branch protection enabled | Require 1 review + passing checks; CODEOWNERS | Ready for review |
| Zero P0/P1; ≤2 P2 | 0.85 | Backlog clean at present | Maintain triage; add SLOs | Ready for review |
| Release workflow | 0.90 | GoReleaser + GHCR configured | Dry-run snapshot + pre-release tag | Done (configured) |
| Rollback rehearsal | 0.60 | Runbook documented | Execute rehearsal in staging | In progress |
| Backward compatibility | 0.80 | Config stable; validation added | Versioned schema + migration notes | Ready for review |
| Docs completeness | 0.90 | Extensive docs + examples | Add alert rules + dashboards | Done |
| 7-day soak | 0.50 | Not yet executed | Run RC soak; attach dashboards | Not started |
