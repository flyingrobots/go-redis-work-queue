# Renew Heartbeats While Processing

## Summary
Worker heartbeats are only written once when a job is claimed, so long-running tasks get reaped mid-flight. Introduce a heartbeat renewal loop that atomically refreshes the TTL while processing and shuts down cleanly when the worker finishes.

## Acceptance Criteria
- Heartbeat is set with `SETNX` when claiming work and renewed with `SETXX` on a jittered ticker until the job completes.
- Renewal loop cancels before job cleanup so no stale goroutine persists.
- Unit/integration tests cover long-running job processing without triggering the reaper.

## Dependencies / Inputs
- `internal/worker` heartbeat handling and processing loop.
- Configuration value `cfg.Worker.HeartbeatTTL`.

## Deliverables / Outputs
- Updated worker heartbeat implementation with renewal loop.
- Tests demonstrating heartbeat refresh for long jobs.
