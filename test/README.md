# test/ Directory Map

This catalog enumerates the remaining artifacts under `test/`. Consistent with the standing guidance (“HALT ALL TESTING UNTIL BUILD IS GREEN”), I did not execute any tests or scripts while compiling these notes—run commands are documented purely for future reference.

## Integration Tests (`test/integration/`)

### integration/rbac_integration_test.go
- **Scope**: Spins up the Admin API and RBAC stack against `miniredis` to verify role-based permissions, token revocation, and audit logging.
- **Quality**: High-value cross-package coverage; it depends on both `internal/admin-api` and `internal/rbac-and-tokens`, so keeping it in this shared integration space makes sense.
- **Run hint**: `go test -tags integration_tests ./test/integration -run '^TestRBACIntegration'`
- **Dependencies**: `miniredis`, `go-redis`, `zap` (nop logger), `testify`.
- **Mocks**: None—uses real components with in-memory Redis.

## End-to-End Tests (`test/e2e/`)

_All E2E suites require the `e2e_tests` build tag; some also need environment vars or local services._

| File | Purpose | Runtime Notes | External Needs |
|------|---------|---------------|----------------|
| `e2e/e2e_test.go` | Smoke-test that a worker drains a queue against a real Redis instance. | `E2E_REDIS_ADDR=host:port go test -tags e2e_tests ./test/e2e -run '^TestE2E_WorkerCompletesJobWithRealRedis$'` | Reachable Redis, `zap`. |
| `e2e/migration_test.go` | Exercises `internal/storage-backends` migrations end-to-end using the registry/migrator APIs. | `go test -tags e2e_tests ./test/e2e -run MigrationE2ETestSuite` | Redis at `localhost:6379`, `testify/suite`. |
| `e2e/tracing_e2e_test.go` | Verifies distributed tracing across producer/worker with an OTLP collector stub. | `E2E_TESTS=true go test -tags 'e2e_tests integration' ./test/e2e -run '^TestE2EDistributedTracingFlow$'` | Redis, HTTP span collector (httptest inside the suite). |
| `e2e/rbac_e2e_test.go` | Walks complete RBAC workflows (tokens, destructive ops, audit logs). | `go test -tags e2e_tests ./test/e2e -run '^TestE2E'` | `miniredis`, Admin API stack, `testify`. |
| `e2e/multi_cluster_tui_test.go` | Simulates multi-cluster control via a mocked TUI facade talking to the real manager. | `go test -tags e2e_tests ./test/e2e -run '^TestMultiClusterTUI_'` | `miniredis`, `tview`, `tcell`, `testify`. |

## Acceptance Scripts (`test/*.sh`)

_These are human-facing acceptance checklists. None were executed._

- `test_p1.t022.sh`: Exactly-once patterns deployment checklist (`redis-cli`, app binary/`go run`, `curl`).
- `test_p2.t051.sh`: Multi-tenant isolation acceptance script (runs package unit tests; should be gated while testing is frozen).
- `test_p3.t035.sh`: Canary deployments design checklist (`grep`, `wc`).
- `test_p3.t046.sh`: Long-term archives design checklist.
- `test_p4.t029.sh`: Anomaly Radar SLO Budget design checklist (`grep`, `python3`, `wc`).
- `test_p4.t044.sh`: Job Genealogy Navigator design checklist.
- `test_p4.t065.sh`: Theme Playground design checklist.
- `test_p4.t073.sh`: Patterned Load Generator design checklist.
- `test_p4_t065_simple.sh`, `test_p4_t073_simple.sh`, `test_p4_simple.sh`: Lightweight variants that only verify file existence/line counts.

_Note_: several scripts still hard-code `/Users/james/...`; parameterize them before relying on the automation.

No standalone fixtures remain under `test/`; the helper package that used to live here now resides under `docs/testing/event-hooks-test-plan.md`.
