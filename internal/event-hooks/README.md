# Event Hooks

- **Status:** BROKEN (`go build ./internal/event-hooks` fails)
- **Last checked:** 2025-09-18

## Why it is broken
- HTTP handlers were scaffolded but not wired; unused imports and placeholders stop compilation.
- Replay/test endpoints are unimplemented pending DLQ and webhook integration.

## Next steps
- Decide on the event transport (webhook/NATS) contract, then fill in handler implementations.
- Wire replay/test endpoints and clean up unused imports before attempting another build.
