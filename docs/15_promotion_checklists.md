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

### Confidence Scores (Alpha → Beta)

| Criterion | Confidence | Rationale | How to improve |
|---|---:|---|---|
| Functional completeness | 0.9 | All core roles implemented and tested; admin CLI present | Add more end-to-end tests for admin flows; document edge cases |
| Observability endpoints | 0.95 | Live and exercised in CI/e2e; stable | Add /healthz readiness probes to example manifests; alert rules examples |
| CI health | 0.9 | CI green with race, vet, e2e, govulncheck | Increase matrix (Go versions, OS); add flaky-test detection |
| Coverage ≥ 80% | 0.75 | Core packages covered; gaps in admin/obs | Add tests for admin and HTTP server handlers |
| E2E determinism | 0.8 | E2E with Redis service stable locally and in CI | Add retries and timing buffers; run 5x in workflow and gate |
| Security (govulncheck) | 0.95 | Using Go 1.24; no critical findings | Add image scanning; pin base image digest |
| Performance baseline | 0.7 | Bench harness exists; sample run meets ~960 jobs/min, latency sampling coarse | Improve latency measurement via metrics; run on 4 vCPU node and document env |
| Documentation completeness | 0.9 | PRD, runbook, deployment, perf, checklists present | Add Helm usage examples and alert rules |

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
### Confidence Scores (Beta → RC)

| Criterion | Confidence | Rationale | How to improve |
|---|---:|---|---|
| ≥1k jobs/min for ≥10m | 0.6 | Not yet run on dedicated 4 vCPU node | Schedule controlled benchmark; record metrics and environment |
| p95 < 2s (<1MB) | 0.6 | Latency sampling method is coarse | Use Prometheus histogram quantiles on /metrics; run sustained test |
| Chaos (outage/latency/crash) | 0.7 | Logic supports recovery; tests cover happy-path and reaper | Add chaos e2e in CI (stop Redis container; tc latency); verify no loss |
| Admin validation | 0.85 | Admin commands tested manually; unit tests for helpers | Add e2e assertions for stats and peek outputs |
| Gauges/breaker accuracy | 0.85 | Metrics wired; observed locally | Add metric assertions in e2e; dashboards and alerts validate |
| 24–48h soak | 0.5 | Not yet executed | Run soak in staging and record dashboards |
| Security and deps | 0.9 | govulncheck in CI; deps pinned | Add Renovate/Dependabot; image scanning stage |
| Issue hygiene | 0.9 | No open P0/P1 | Enforce labels and triage automation |
### Confidence Scores (RC → GA)

| Criterion | Confidence | Rationale | How to improve |
|---|---:|---|---|
| Code freeze discipline | 0.8 | Process defined; branch protection enabled | Require 1 review and passing checks (enabled); add CODEOWNERS |
| Zero P0/P1; ≤2 P2 | 0.85 | Current backlog clean | Maintain triage; add SLOs for bug classes |
| Release workflow | 0.9 | GoReleaser + GHCR configured; test via pre-release | Dry-run snapshot and tag a pre-release on branch |
| Rollback rehearsal | 0.6 | Procedure documented | Execute runbook in staging and document proof |
| Backward compatibility | 0.8 | Config stable; validation added | Add versioned config schema and migration notes |
| Docs completeness | 0.9 | Extensive docs present | Add Grafana/Prometheus import snippets and examples (added dashboard) |
| 7-day soak | 0.5 | Not yet executed | Run RC soak with dashboard snapshots and attach to evidence |
