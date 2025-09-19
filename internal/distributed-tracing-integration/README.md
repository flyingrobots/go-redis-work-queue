# Distributed Tracing Integration

- **Status:** BUILDS (`go build ./internal/distributed-tracing-integration` passes; integration tests still pending)
- **Last checked:** 2025-09-18

## Notes
- Package builds cleanly; legacy tests still reference old OpenTelemetry APIs and helpers (`trace.WithAttributes`, `getErrorType`).

## Next steps
- Modernize the test harness and helper utilities before re-enabling test runs.
