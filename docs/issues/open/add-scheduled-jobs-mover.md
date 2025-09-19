# Add Scheduled Jobs Mover

## Summary
Queue lacks delayed execution. Implement a scheduled jobs mover that atomically promotes entries from a sorted set to the ready queue when due.

## Acceptance Criteria
- New enqueue path stores delayed jobs in `jobqueue:sched:<queue>` with ready timestamps.
- Background mover (Lua or optimistic batching) promotes ready jobs using `ZPOPMIN` or equivalent.
- Unit/integration tests cover delay, promotion, and idempotence.

## Dependencies / Inputs
- Redis sorted set operations.
- Existing retry/backoff handling.

## Deliverables / Outputs
- Code for scheduling API and mover.
- Documentation detailing how to schedule and monitor delayed jobs.
