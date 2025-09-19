# Long-Term Archives

- **Status:** BROKEN (`go build ./internal/long-term-archives` fails)
- **Last checked:** 2025-09-18

## Why it is broken
- Archive drivers depend on storage abstractions that are still being redesigned; code does not compile without the missing packages.

## Next steps
- Finalise the archive driver interfaces and re-enable the package when they are stable.
