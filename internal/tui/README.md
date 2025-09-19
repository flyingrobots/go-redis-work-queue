# Terminal UI

- **Status:** BUILDS (`go build ./cmd/tui` passes; experimental enhanced view gated by `tui_experimental` tag)
- **Last checked:** 2025-09-18

## Notes
- The "enhanced" view and style demo remain behind the `tui_experimental` build tag until those helpers are completed.
- Core TUI builds cleanly and continues to use the legacy view path by default.

## Next steps
- Finish the responsive view refactor (reintroduce build helpers) and remove the experimental tag once ready.
