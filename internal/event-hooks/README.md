# Event Hooks

- **Status:** BUILDS (`go build ./internal/event-hooks` passes; handler bodies still TODO)
- **Last checked:** 2025-09-18

## Notes
- Handlers compile, but they remain scaffolding—core replay/test implementations are still TODO until the manager layer is wired.
- Feature is unimplemented; build is green to unblock dependent modules.

## Next steps
- Flesh out webhook/NATS plumbing, replay/test endpoints, and manager wiring before enabling the feature.
