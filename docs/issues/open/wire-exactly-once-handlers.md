# Use Idempotency Manager in Handlers

## Summary
Handlers ignore the existing Idempotency/Outbox module, so retries can double-execute side effects. Integrate the IdempotencyManager into worker handlers to guarantee at-least-once replay safety.

## Acceptance Criteria
- Handler template calls IdempotencyManager.Begin/End around side effects.
- Duplicate deliveries are recognized and short-circuited with no side-effect replays.
- Tests exercising retry scenarios confirm idempotent behavior.

## Dependencies / Inputs
- Existing IdempotencyManager and outbox implementation.
- Worker handler scaffolding.

## Deliverables / Outputs
- Updated handler code using idempotency primitives.
- Tests verifying exactly-once semantics for handled jobs.
