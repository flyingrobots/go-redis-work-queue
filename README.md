# Go Redis Work Queue

> Redis job queue system in Go. 

Provides producer, worker, and all-in-one modes with robust resilience, observability, and configurable behavior via YAML.

- Single binary with multi-role execution
- Priority queues with reliable processing and retries
- Graceful shutdown, reaper for stuck jobs, circuit breaker
- Prometheus metrics, structured logging, optional tracing

See the [Feature Matrix](docs/features-ledger.md) for the latest capability status (stable, experimental, deprecated).

See `docs/` to learn more. A sample configuration is provided in `config/config.example.yaml`.

Developer tools and automation: see `docs/tools/README.md` for:
- Progress automation for the Features Ledger and README
- Local pre-commit hook and CI auto-update
- Script to extract CodeRabbit PR comments and “Prompt for AI Agents” sections

----

## Progress

For full details, see the Features Ledger at [docs/features-ledger.md](docs/features-ledger.md).

<!-- progress:begin -->
```text
██████████████████████▓░░░░░░░░░░░░░░░░░ 56%
---------|---------|---------|---------|
        MVP      Alpha     Beta  v1.0.0 
```
<!-- progress:end -->

----

## Quick start

1. Clone the repo
2. Ensure Redis is available (e.g., Docker container `redis:latest` on port `6379`)
3. Follow the instructions to run in producer, worker, or all-in-one modes

### Build and run

1. Copy example config

```bash
cp config/config.example.yaml config/config.yaml
```

2. Build (Go 1.25+)

```bash
make build
```

3. Run in one of the following modes:

Run all-in-one

```bash
./bin/job-queue-system --role=all --config=config/config.yaml
```

Run producer only

```bash
./bin/job-queue-system --role=producer --config=config/config.yaml
```

Run worker only

```bash
./bin/job-queue-system --role=worker --config=config/config.yaml
```

### TUI (Bubble Tea)

An interactive TUI is available for observing and administering the job queue. It uses `Charmbracelet`’s Bubble Tea stack and renders queue stats, keys, peeks, a simple benchmark, and charts.

Before the first run, download dependencies:

```bash
go mod download
```

Run it:

```
go run ./cmd/tui --config config/config.yaml
```

Or build it:

```
go build -o bin/tui ./cmd/tui
./bin/tui --config config/config.yaml
```

Flags:

- `--config`: Path to YAML config (defaults to `config/config.yaml`).
- `--refresh`: Stats refresh interval (default `2s`).

Keybindings:

- `q`: quit
- `esc`: toggle help overlay (when not in a modal/input)
- `tab`: switch Queues/Keys
- `r`: refresh
- `j/k`: move selection
- `p`: peek selected queue
- `b`: open bench form (tab cycles fields, enter runs, esc exits)
- `c`: charts view (time-series for queue lengths)
- `f` or `/`: filter queues (fuzzy, case-insensitive); `esc` clears
- `D`: purge dead-letter queue (modal confirm)
- `A`: purge ALL managed keys (modal confirm)

Mouse:

- Wheel scrolls, hover highlights row, left-click selects, right-click peeks.

Notes:

- The TUI calls internal admin APIs, so it reflects the same Redis keys as the CLI admin mode.
- When a confirmation modal is open, the background dims and a full-screen scrim appears for focus.

Screenshots (examples):

![Queues View](docs/images/tui-queues.png)

![Peek Modal](docs/images/tui-peek.png)

![Charts View](docs/images/tui-charts.png)

### Admin Commands

The CLI provides `--admin-cmd` flags that help you inspect the system.

```bash
# Stats
./bin/job-queue-system --role=admin --admin-cmd=stats --config=config/config.yaml

# Peek
./bin/job-queue-system --role=admin --admin-cmd=peek --queue=low --n=10 --config=config/config.yaml

# Purge DLQ
./bin/job-queue-system --role=admin --admin-cmd=purge-dlq --yes --config=config/config.yaml

# Purge all (test keys) — DEV ONLY; requires explicit --dev + --yes safeguards
./bin/job-queue-system --role=admin --admin-cmd=purge-all --dev --yes --config=config/config.yaml

# Stats (keys)
./bin/job-queue-system --role=admin --admin-cmd=stats-keys --config=config/config.yaml

# Version
./bin/job-queue-system --version
```

### Metrics

Prometheus metrics exposed at <http://localhost:9091/metrics> by default (override via `observability.metrics_port` to avoid conflicts with local Prometheus).

### Health and Readiness

- Liveness: <http://localhost:9090/healthz> returns 200 when the process is up
- Readiness: <http://localhost:9090/readyz> returns 200 only when Redis is reachable

### Priority Fetching

Workers emulate prioritized multi-queue blocking fetch by looping priorities (e.g., high then low) and issuing `BRPOPLPUSH` per-queue with a short timeout (default 1s). This preserves atomic move semantics within each queue, prefers higher priority at sub-second granularity, and avoids job loss. Lower-priority jobs may incur up to the timeout in extra latency when higher-priority queues are empty.

### Rate Limiting

Producer rate limiting uses a fixed-window counter (`INCR` + 1s `EXPIRE`) and sleeps precisely until the end of the window (`TTL`), with small jitter to avoid thundering herd.

### Docker

To make it easy to use, a `Dockerfile` has been provided. The following demonstrate how to use it:

#### Build

```bash
docker build -t job-queue-system:latest .
```

#### Run

```bash
docker run --rm \
  -p 9091:9091 \
  -v $(pwd)/config/config.yaml:/app/config/config.yaml:ro \
  --env-file env.list \
  job-queue-system:latest --role=all --config=/app/config/config.yaml
```

Ensure `config/config.yaml` exists locally and `env.list` provides credentials (see `config/config.example.yaml` for keys).

#### Compose

```bash
docker compose -f deploy/docker-compose.yml up --build
```

----

## `v0.4.0-alpha` Coming Soon

Release branch open for `v0.4.0-alpha`: see PR <https://github.com/flyingrobots/go-redis-work-queue/pull/1>

Promotion gates and confidence summary (details in `docs/15_promotion_checklists.md`):

### Release Roadmap

- **Alpha → Beta**: overall confidence `~0.85` (functional/observability/CI strong; perf and coverage improvements planned)
- **Beta → RC**: overall confidence `~0.70` (needs controlled perf run, chaos tests, soak)
- **RC → GA**: overall confidence `~0.70` (release flow ready; soak and rollback rehearsal pending)

### Evidence artifacts (`docs/evidence/`):

- `ci_run.json` (CI URL), 
- `bench.json` (throughput/latency), 
- `metrics_before.txt`/`metrics_after.txt`, 
- `config.alpha.yaml`

To reproduce evidence locally, see `docs/evidence/README.md`.

----

## Testing

See `docs/testing-guide.md` for a package-by-package overview and copy/paste commands to run individual tests or the full suite with the race detector.

## Module Status Map

| Module | Status | Summary |
| --- | --- | --- |
| [internal/automatic-capacity-planning](internal/automatic-capacity-planning/README.md) | BROKEN | Forecasting/simulator tests expect legacy behaviour; new algorithms need history seeding. |
| [internal/canary-deployments](internal/canary-deployments/README.md) | BUILDS | Compiles; API endpoints still stubbed for worker/rollback flows. |
| [internal/collaborative-session](internal/collaborative-session/README.md) | BUILDS | Compiles after restoring handoff helpers; no tests yet. |
| [internal/distributed-tracing-integration](internal/distributed-tracing-integration/README.md) | BUILDS | Compiles; tests still expect legacy OpenTelemetry helpers. |
| [internal/dlq-remediation-pipeline](internal/dlq-remediation-pipeline/README.md) | BUILDS | Compiles; remediation actions still need implementation. |
| [internal/event-hooks](internal/event-hooks/README.md) | BUILDS | Compiles; handlers remain TODO for replay/test flows. |
| [internal/exactly_once](internal/exactly_once/README.md) | BROKEN | Outbox manager tests panic due to retry bookkeeping regressions. |
| [internal/forecasting](internal/forecasting/README.md) | BROKEN | Holt-Winters and recommendation tests fail with new defaults. |
| [internal/job-budgeting](internal/job-budgeting/README.md) | BUILDS | Compiles; budgeting enforcement still TODO. |
| [internal/json-payload-studio](internal/json-payload-studio/README.md) | BUILDS | Compiles; handlers still rely on in-memory stubs. |
| [internal/kubernetes-operator](internal/kubernetes-operator/README.md) | BUILDS | Compiles; controllers/webhooks still stubbed out. |
| [internal/long-term-archives](internal/long-term-archives/README.md) | BROKEN | Archive drivers unfinished; depends on pending storage abstractions. |
| [internal/storage-backends](internal/storage-backends/README.md) | BUILDS | Compiles; remaining test harness needs go-redis v9 cleanup. |
| [internal/tui](internal/tui/README.md) | BUILDS | Legacy view builds cleanly; enhanced view remains behind `tui_experimental`. |
| [internal/trace-drilldown-log-tail](internal/trace-drilldown-log-tail/README.md) | BUILDS | Compiles; log streaming endpoints still placeholders. |
| [internal/worker](internal/worker/README.md) | BUILDS | Runtime compiles; lacks dedicated unit tests. |

----

## Contributing

Want to help? Here's how:

1. Please report issues that you discover. 
2. If you solve any problems, PRs are welcome. 

### DX Tools

Please be sure to enable the following developer tools to enhance your development experience and align with the project's standard development practices.

#### Enable Git hooks

```bash
make hooks
```
