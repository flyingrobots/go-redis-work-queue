# Exactly-Once Outbox

- **Status:** BROKEN (`go test ./internal/exactly_once` fails)
- **Last checked:** 2025-09-18

## Why it is broken
- `SQLOutboxManager` tests panic because the shutdown path closes channels twice and the retry bookkeeping no longer updates attempt counters after the refactor.

## Next steps
- Fix retry accounting (`attempts`, `last_error`) and guard the stop routine against double closes.
- Add regression tests before re-enabling in CI.
