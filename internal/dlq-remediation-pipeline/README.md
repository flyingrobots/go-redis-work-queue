# DLQ Remediation Pipeline

- **Status:** BROKEN (`go build ./internal/dlq-remediation-pipeline` fails)
- **Last checked:** 2025-09-18

## Why it is broken
- Redis client configuration still references go-redis v8 fields (`MaxConnAge`, `IdleTimeout`) that were removed in v9.
- Action loops leave temporary variables unused; build fails with `declared and not used` errors.

## Next steps
- Update Redis options to v9 equivalents (`ConnMaxIdleTime`, etc.) and finish wiring the action handlers.
- Clean up unused variables and ensure the package builds before adding tests.
