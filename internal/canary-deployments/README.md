# Canary Deployments

- **Status:** BUILDS (`go build ./internal/canary-deployments` passes; API still incomplete)
- **Last checked:** 2025-09-18

## Notes
- Router modules now compile against go-redis v9; handler stubs still return TODO errors.

## Next steps
- Flesh out rollback/abort workflows, auditing, and worker lookups before exposing the API.
- Add real implementations for manager methods that currently return `CodeSystemNotReady`.
