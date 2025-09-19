## In create_postmortem_tasks.py around lines 5, 18-19, 71-73, you are constructing

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912504

- [review_comment] 2025-09-18T12:12:36Z by coderabbitai[bot] (create_postmortem_tasks.py:5)

```text
In create_postmortem_tasks.py around lines 5, 18-19, 71-73, you are constructing
naive datetimes and appending "Z" manually; replace those with timezone-aware
UTC datetimes. Import timezone from datetime (or use datetime.timezone) and
replace datetime.now() (or naive constructions) with datetime.now(timezone.utc)
(or attach tzinfo=timezone.utc), then produce a proper Zulu-formatted string
either via .isoformat().replace('+00:00','Z') or format with
strftime('%Y-%m-%dT%H:%M:%SZ'); update all three locations accordingly so
timestamps are real UTC rather than naive times with a fake "Z".
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
