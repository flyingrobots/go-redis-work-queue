# Calendar View

- **Status:** BUILDS (`go build ./internal/calendar-view` passes; calendar UX still needs wiring to the TUI workers tab)
- **Last checked:** 2025-09-18

## Notes
- HTTP handlers, data source plumbing, and domain types compile; calendar rendering/zooming remains stubbed behind API responses.
- Tests cover cache/validator logic; end-to-end integration with the scheduler is still pending.

## Next steps
- Hook the calendar data endpoints into the TUI calendar panel and implement zoom-level navigation flows.
- Add integration tests once the scheduler emits real events at multiple horizons (hour/day/week/month/year).
