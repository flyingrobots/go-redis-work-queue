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

## Status

Scaffolding in place. Implementation, PRD, tests, and CI are coming next per plan.
