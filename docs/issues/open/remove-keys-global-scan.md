# Eliminate Global KEYS/SCAN Usage

## Summary
Admin and reaper flows still iterate Redis with broad `SCAN jobqueue:*`, which is unsafe at scale. Replace global scans with a worker registry set and hash-tagged keys.

## Acceptance Criteria
- Workers register themselves in `jobqueue:workers` (SET/SADD) on heartbeat/startup and remove on shutdown.
- Reaper/admin logic iterates the registry set and operates on per-worker processing keys (e.g., `jobqueue:{workerID}:processing`).
- Processing keys are hash-tagged consistently for cluster friendliness.
- No remaining use of `KEYS` or slot-crossing `SCAN` in codebase.

## Dependencies / Inputs
- Current reaper implementation.
- Worker heartbeat/registration logic.

## Deliverables / Outputs
- Updated admin/reaper code using registry iteration.
- Tests validating the new registry-based reaper.
