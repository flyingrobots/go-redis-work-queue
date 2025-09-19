# Event Hooks

- **Status:** BUILDS (`go build ./internal/event-hooks` passes; handler bodies still TODO)
- **Last checked:** 2025-09-18

## Notes
- Handlers compile, but several endpoints still return TODO errors until the manager layer is implemented.

## Next steps
- Flesh out webhook/NATS plumbing, replay/test endpoints, and manager wiring before enabling the feature.
