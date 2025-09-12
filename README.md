# Go Redis Work Queue

Production-ready Go-based job queue system backed by Redis. Provides producer, worker, and all-in-one modes with robust resilience, observability, and configurable behavior via YAML.

- Single binary with multi-role execution
- Priority queues with reliable processing and retries
- Graceful shutdown, reaper for stuck jobs, circuit breaker
- Prometheus metrics, structured logging, optional tracing

See docs/ for the Product Requirements Document (PRD) and detailed design. A sample configuration will be provided in config/config.example.yaml once the implementation lands.

## Quick start

- Clone the repo
- Ensure Redis is available (e.g., Docker container redis:latest on port 6379)
- Follow the instructions in the PRD to run in producer, worker, or all-in-one modes

### Build and run

- Copy example config
```bash
cp config/config.example.yaml config/config.yaml
```

- Build (Go 1.25+)
```bash
make build
```

- Run all-in-one
```bash
./bin/job-queue-system --role=all --config=config/config.yaml
```

- Run producer only
```bash
./bin/job-queue-system --role=producer --config=config/config.yaml
```

- Run worker only
```bash
./bin/job-queue-system --role=worker --config=config/config.yaml
```

- Admin commands
```bash
# Stats
./bin/job-queue-system --role=admin --admin-cmd=stats --config=config/config.yaml

# Peek
./bin/job-queue-system --role=admin --admin-cmd=peek --queue=low --n=10 --config=config/config.yaml

# Purge DLQ
./bin/job-queue-system --role=admin --admin-cmd=purge-dlq --yes --config=config/config.yaml

# Purge all (test keys)
./bin/job-queue-system --role=admin --admin-cmd=purge-all --yes --config=config/config.yaml

# Stats (keys)
./bin/job-queue-system --role=admin --admin-cmd=stats-keys --config=config/config.yaml

# Version
./bin/job-queue-system --version
```

### Metrics

- Prometheus metrics exposed at http://localhost:9090/metrics by default

### Health and Readiness

- Liveness: http://localhost:9090/healthz returns 200 when the process is up
- Readiness: http://localhost:9090/readyz returns 200 only when Redis is reachable

### Priority Fetching

- Workers emulate prioritized multi-queue blocking fetch by looping priorities (e.g., high then low) and issuing `BRPOPLPUSH` per-queue with a short timeout (default 1s). This preserves atomic move semantics within each queue, prefers higher priority at sub-second granularity, and avoids job loss. Lower-priority jobs may incur up to the timeout in extra latency when higher-priority queues are empty.

### Rate Limiting

- Producer rate limiting uses a fixed-window counter (`INCR` + 1s `EXPIRE`) and sleeps precisely until the end of the window (`TTL`), with small jitter to avoid thundering herd.

### Docker

- Build
```bash
docker build -t job-queue-system:latest .
```

- Run
```bash
docker run --rm -p 9090:9090 --env-file env.list job-queue-system:latest --role=all
```

- Compose
```bash
docker compose -f deploy/docker-compose.yml up --build
```

## Status

Release branch open for v0.4.0-alpha: see PR https://github.com/flyingrobots/go-redis-work-queue/pull/1

Promotion gates and confidence summary (details in docs/15_promotion_checklists.md):
- Alpha → Beta: overall confidence ~0.85 (functional/observability/CI strong; perf and coverage improvements planned)
- Beta → RC: overall confidence ~0.70 (needs controlled perf run, chaos tests, soak)
- RC → GA: overall confidence ~0.70 (release flow ready; soak and rollback rehearsal pending)

Evidence artifacts (docs/evidence/):
- ci_run.json (CI URL), bench.json (throughput/latency), metrics_before/after.txt, config.alpha.yaml

To reproduce evidence locally, see docs/evidence/README.md.

## Testing

See docs/testing-guide.md for a package-by-package overview and copy/paste commands to run individual tests or the full suite with the race detector.

## Contributing / Docs Linting

- Enable Git hooks (auto-fix Markdown on commit):
```bash
make hooks
```

- Run Markdown lint locally (optional):
```bash
# Using Node (autofix staged files happens on commit via hook)
npx -y markdownlint-cli2 "**/*.md" "!**/node_modules/**"
```

- CI runs markdownlint on every PR and on pushes to `main`.
