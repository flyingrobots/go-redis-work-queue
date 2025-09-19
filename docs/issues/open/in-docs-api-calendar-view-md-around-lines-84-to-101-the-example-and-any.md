## In docs/api/calendar-view.md around lines 84 to 101, the example and any

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033300

- [review_comment] 2025-09-18T15:55:21Z by coderabbitai[bot] (docs/api/calendar-view.md:101)

```text
In docs/api/calendar-view.md around lines 84 to 101, the example and any
documented endpoints omit the required API version prefix; update all path
references and example requests to use the /api/v1 prefix (e.g., change
/calendar/data to /api/v1/calendar/data) and apply the same change to all other
documented endpoints mentioned (/events, /reschedule, /rules, /config,
/timezones, /health, /debug/*) so every path in this file consistently uses
/api/v1.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD
