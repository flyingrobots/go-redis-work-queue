# Distributed Tracing Integration

- **Status:** BROKEN (`go build ./internal/distributed-tracing-integration` fails)
- **Last checked:** 2025-09-18

## Why it is broken
- Tests target old OpenTelemetry APIs (`trace.WithAttributes`, direct `attribute.KeyValue` event options) that no longer exist after the otel v1.20 upgrade.
- Helper functions (`getErrorType`, context plumbing) were deleted during the refactor and never replaced.

## Next steps
- Port tracing helpers to the new otel APIs and reinstate the deleted utilities.
- Restore the missing helper functions (`getErrorType`, context plumbing) before re-running builds.
