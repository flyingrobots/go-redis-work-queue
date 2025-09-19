# Unify Redis Client on go-redis/v9

## Summary
Multiple Redis client versions are present. Standardize on `github.com/redis/go-redis/v9`, wrap it behind a local interface, and remove duplicate client trees.

## Acceptance Criteria
- All modules import a single Redis client dependency (`go-redis/v9`).
- Shared interface (e.g., `type RedisCmdable interface`) added for mocking.
- go.mod/go.sum reflect the unified dependency tree; no legacy v8 modules remain.

## Dependencies / Inputs
- Existing Redis usage across worker, admin API, CLI, tests.
- Build tooling (`go mod tidy`).

## Deliverables / Outputs
- Updated code references to the unified client.
- Tests and builds passing with the new client setup.
