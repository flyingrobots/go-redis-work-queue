# Long-Term Archives

- **Status:** BUILDS (`go build ./internal/long-term-archives` passes; archival flows remain stubbed)
- **Last checked:** 2025-09-18

## Notes
- Redis stats, exporters, and retention helpers compile but still use in-memory or no-op paths.
- ClickHouse/S3 exporters remain placeholders; real storage integration is pending.

## Next steps
- Flesh out exporter implementations and wire actual storage writes before enabling in production.
- Add integration tests once storage backends stabilise.
