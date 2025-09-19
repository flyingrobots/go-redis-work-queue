# JSON Payload Studio

- **Status:** BROKEN (`go build ./internal/json-payload-studio` fails)
- **Last checked:** 2025-09-18

## Why it is broken
- Handler layer still calls methods like `GetSchema`, `ListTemplates`, `ApplyTemplate`, but the backing `JSONPayloadStudio` type stub never implemented them.
- API surface needs a redesign; function signatures (e.g. `SaveTemplate`) changed during review and the handlers were not updated.

## Next steps
- Flesh out the studio core (schema catalogue, template persistence) and update the HTTP layer accordingly.
