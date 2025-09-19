# Job Budgeting

- **Status:** BUILDS (`go build ./internal/job-budgeting` passes; feature still unimplemented)
- **Last checked:** 2025-09-18

## Why it is broken
- Core error helpers (`ErrInvalidJobData`, `IsTemporary`) were removed during the refactor; remaining code still references them.
- TUI scaffolding imports components that were never implemented, leaving unused variables.

## Next steps
- Restore the shared error helpers and finish the budget manager API before re-running builds.
