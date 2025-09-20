# Worker Fleet Controls

- **Status:** BUILDS (`go build ./internal/worker-fleet-controls` passes; control operations remain partially stubbed)
- **Last checked:** 2025-09-18

## Notes
- REST handlers, controller, and registry compile with Redis-backed scaffolding; long-running action flows still rely on in-memory tracking.
- Safety checker logic exists but requires real confirmations/workflow wiring before enabling destructive operations in production.

## Next steps
- Flesh out worker signal fan-out and persistence for pause/drain/stop actions, then add integration coverage.
- Connect the fleet controls surface into the TUI once the Workers tab ships.
