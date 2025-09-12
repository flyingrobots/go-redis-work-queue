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

- Copy config/config.example.yaml to config/config.yaml and adjust as needed
- Build (Go 1.24+): make build
- Run all-in-one: ./bin/job-queue-system --role=all --config=config/config.yaml
- Or run producer only: ./bin/job-queue-system --role=producer --config=config/config.yaml
- Or run worker only: ./bin/job-queue-system --role=worker --config=config/config.yaml
- Admin commands:
  - Stats: ./bin/job-queue-system --role=admin --admin-cmd=stats --config=config/config.yaml
  - Peek:  ./bin/job-queue-system --role=admin --admin-cmd=peek --queue=low --n=10 --config=config/config.yaml
  - Purge DLQ: ./bin/job-queue-system --role=admin --admin-cmd=purge-dlq --yes --config=config/config.yaml
 - Version: ./bin/job-queue-system --version

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

- Build: docker build -t job-queue-system:latest .
- Run: docker run --rm -p 9090:9090 --env-file env.list job-queue-system:latest --role=all
- Compose: see deploy/docker-compose.yml for multi-service setup (redis + worker/producer/all-in-one)

## Status

Scaffolding in place. Implementation, PRD, tests, and CI are coming next per plan.
