# Storage Backends

- **Status:** BUILDS (`go build ./internal/storage-backends` passes)
- **Last checked:** 2025-09-18

## Why it is broken
- Coverage tests previously duplicated the mock backend/factory helpers, causing compile failures when running the full test suite.
- Redis client code was migrated to go-redis v9; lingering tests still expect the v8 option set and need cleanup before we re-run them.

## Next steps
- Update the remaining tests for go-redis v9 semantics and re-enable them once the package’s mocks match the new API surface.
