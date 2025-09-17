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

| Criterion | Confidence | Rationale | Owner | Exit Criteria | How to improve | Status |
|---|---:|---|---|---|---|---|
| Functional completeness | 0.90 | All core roles implemented and tested; admin CLI present | @queue-core | Nightly `make smoke` run exercises producer/worker/all roles with zero failures | Add more E2E admin flows; edge-case docs | Done |
| Observability endpoints | 0.95 | /metrics, /healthz, /readyz live; used in CI/e2e | @obs-team | CI probe job hits `/metrics`, `/healthz`, `/readyz` in staging and returns 200 for each | Add probes to examples; alert rules | Done |
| CI health | 0.90 | CI green (build, vet, race, e2e, govulncheck) | @release-eng | `main` pipeline green across build/vet/race/e2e/govulncheck for 5 consecutive runs | Add matrix (OS/Go), flaky-test detector | Done |
| Coverage ≥ 80% | 0.75 | Gaps in admin/obs packages | @qa-team | `go test ./... -coverprofile=cover.out` shows ≥0.80 coverage for admin+obs packages in CI artifact | Add tests for admin + HTTP handlers | In progress |
| E2E determinism | 0.80 | CI runs e2e 5×; generally stable | @qa-team | E2E suite passes 5 consecutive scheduled runs without quarantined flakes | Gate on 5× passing and record | In progress |
| Security (govulncheck) | 0.95 | Go 1.25 stdlib; no critical findings | @security | `govulncheck ./...` returns no High/Critical findings in release workflow | Image scanning; pin base digest | Done |
| Performance baseline | 0.70 | Harness present; prelim results only | @perf-lab | Benchmark job processes 1k jobs at 500±10% jobs/sec on 4 vCPU runner and records report | Use Prom histograms; 4 vCPU run | In progress |
| Documentation completeness | 0.90 | PRD, runbook, deploy, perf, checklists | @docs-team | Docs review checklist signed off and README diff merged with approvals | Add alert rules + Helm usage | Done |

> _CI expects `Owner` values to start with a GitHub handle (e.g. `@team-name`) and `Exit Criteria` to describe an objective, machine-verifiable check._

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

| Criterion | Confidence | Rationale | Owner | Exit Criteria | How to improve | Status |
|---|---:|---|---|---|---|---|
| ≥1k jobs/min for ≥10m | 0.60 | Not yet run on dedicated 4 vCPU node | @perf-lab | Load test job logs show ≥1000 jobs/min sustained for 10 minutes on 4 vCPU runner | Schedule controlled benchmark; record env + metrics | In progress |
| p95 < 2s (<1MB) | 0.60 | Coarse sampling currently | @perf-lab | Prometheus histogram query reports p95 < 2s for 10-minute benchmark window | Use Prom histograms; sustained run | In progress |
| Chaos (outage/latency/crash) | 0.70 | Recovery logic exists; partial tests | @reliability | Chaos workflow simulating Redis outage/latency/crash completes with no lost jobs and breaker recovery events logged | Add chaos e2e (stop Redis; add latency) | In progress |
| Admin validation | 0.85 | Admin commands + tests | @queue-core | Admin CLI validation suite passes against staging cluster (stats, peek, purge) | Add e2e assertions for outputs | Ready for review |
| Gauges/breaker accuracy | 0.85 | Metrics wired; observed | @obs-team | Metric smoke test asserts queue gauges and breaker state match synthetic load within tolerance | Add metric assertions; dashboards | Ready for review |
| 24–48h soak | 0.50 | Not yet performed | @reliability | Completed 48h soak run with error rate <0.5% and no resource leaks documented | Stage soak; capture dashboards | Not started |
| Security and deps | 0.90 | govulncheck green; deps pinned | @security | Renovate/Dependabot run clean, `govulncheck` + container scan report no High/Critical issues | Add Renovate/Dependabot; image scan | Done |
| Issue hygiene | 0.90 | No open P0/P1 | @release-eng | Release triage dashboard shows 0 P0/P1, ≤3 P2 with documented workarounds | Enforce labels/triage automation | Done |

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

| Criterion | Confidence | Rationale | Owner | Exit Criteria | How to improve | Status |
|---|---:|---|---|---|---|---|
| Code freeze discipline | 0.80 | Process defined; branch protection enabled | @release-eng | Freeze checklist executed and branch protection prevents non-approved merges post-freeze | Require 1 review + passing checks; CODEOWNERS | Ready for review |
| Zero P0/P1; ≤2 P2 | 0.85 | Backlog clean at present | @release-eng | Issue tracker snapshot shows 0 P0/P1 and ≤2 P2 with documented workarounds | Maintain triage; add SLOs | Ready for review |
| Release workflow | 0.90 | GoReleaser + GHCR configured | @devops | GoReleaser dry-run publishes artifact to staging registry and rollback tag verified | Dry-run snapshot + pre-release tag | Done (configured) |
| Rollback rehearsal | 0.60 | Runbook documented | @ops-oncall | Staging rollback rehearsal executed with success log and MTTR < 10m | Execute rehearsal in staging | In progress |
| Backward compatibility | 0.80 | Config stable; validation added | @config-owners | Compatibility tests load previous release config set without errors | Versioned schema + migration notes | Ready for review |
| Docs completeness | 0.90 | Extensive docs + examples | @docs-team | Docs checklist signed; README examples re-run and pasted into PR evidence | Add alert rules + dashboards | Done |
| 7-day soak | 0.50 | Not yet executed | @reliability | 7-day soak report shows readiness >99.9% and DLQ <0.5% with dashboard links | Run RC soak; attach dashboards | Not started |
