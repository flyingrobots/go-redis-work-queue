## In create_review_tasks.py around lines 35 to 46, fromisoformat() currently fails

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856203

- [review_comment] 2025-09-16T23:20:22Z by coderabbitai[bot] (create_review_tasks.py:46)

```text
In create_review_tasks.py around lines 35 to 46, fromisoformat() currently fails
on ISO8601 strings that end with 'Z'; update parse_timestamp to normalize 'Z' to
an explicit offset (e.g. replace a trailing 'Z' with '+00:00') before calling
datetime.fromisoformat, keep the existing logic to set timezone to UTC when
missing, and wrap the fromisoformat call in a try/except that raises a clear
ValueError (e.g. "Invalid timestamp: <value>") if parsing still fails so the CLI
surfaces a helpful error message.
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
