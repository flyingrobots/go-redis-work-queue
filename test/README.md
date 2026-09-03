# test/ Directory Map

This catalog enumerates the remaining artifacts under `test/`. Phase II
actively runs the default suite and restores valuable gated coverage without
enabling unwired feature suites prematurely.

## Default Coverage Canary

Core worker, queue, producer, reaper, and breaker tests run under
`go test ./... -race -count=1`. Run
`./scripts/check_test_package_count.sh` to assert that at least 25 packages
still contribute tests to the default build. CI runs this canary after the
default suite.

## External Client Contract (`test/external-queueclient/`)

This nested Go module imports `pkg/queueclient` through the public module path,
then enqueues and peeks a job against `miniredis`. The separate `go.mod` is the
compile-time proof that the supported API does not leak Go `internal/`
boundaries. CI runs it independently because `go test ./...` does not recurse
into nested modules.

Run it with:

```bash
(cd test/external-queueclient && go test ./... -race -count=1)
```

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
| `e2e/e2e_test.go` | Proves byte-exact handler delivery, long-handler heartbeat renewal, and completion against real Redis. | `E2E_REDIS_ADDR=host:port go test -tags e2e_tests -v ./test/e2e -run '^TestE2E_WorkerCompletesJobWithRealRedis$'`; CI asserts the verbose PASS line. | Reachable Redis, `zap`. |
| `e2e/per_key_fifo_test.go` | Proves strict same-key order, cross-key concurrency, crash recovery, lease renewal, fairness, key safety, single-winner recovery races, and O(1) claim behavior at 10k keys. | `E2E_REDIS_ADDR=host:port go test -tags e2e_tests -v ./test/e2e -run '^TestE2E_PerKeyFIFO' -race -count=1`; CI asserts every scenario runs. | Redis 7, `zap`. |
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

The external-client module is the only standalone fixture under `test/`; the
event-hooks test plan that used to live here now resides under
`docs/testing/event-hooks-test-plan.md`.
