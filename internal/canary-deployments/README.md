# Canary Deployments

- **Status:** BROKEN (`go build ./internal/canary-deployments` fails)
- **Last checked:** 2025-09-18

## Why it is broken
- Router rebuild stalled: `router.go` references an `r` mux that no longer exists.
- Redis client usage still targets go-redis v8 APIs (`XInfoGroup`, `IdleTimeout`, etc.) that were removed when we bumped to go-redis v9.
- Handler scaffolding imports utilities that are unused / not wired (e.g. feature flag plumbing), causing compile failures.

## Next steps
- Finish the router/handler rewrite and migrate to go-redis v9 equivalents (`XInfoGroups`, `ClientInfo` etc.).
- Flesh out rollback/abort workflows and audit logging before attempting another build.
