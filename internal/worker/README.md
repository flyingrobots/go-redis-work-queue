# Worker Runtime

- **Status:** BUILDS (`go test ./internal/worker` currently compiles; package has no unit tests)
- **Last checked:** 2025-09-18

## Notes
- Updated error logging to avoid format-string panics.
- Integration coverage still lives in the `internal/exactly_once` suite.

## Next steps
- Add unit tests around retry/backoff behaviour.
