## In create_postmortem_tasks.py around lines 18-19 (and also apply the same change

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360072243

- [review_comment] 2025-09-18T16:03:43Z by coderabbitai[bot] (create_postmortem_tasks.py:19)

```text
In create_postmortem_tasks.py around lines 18-19 (and also apply the same change
to lines 72-73), the timestamps currently include varying sub-second precision;
normalize them to seconds precision and a stable Z suffix by removing
microseconds and formatting the datetime in UTC with a trailing "Z". Update the
code that builds those "created_at" values to zero out microseconds (or
otherwise format to seconds) and emit an ISO-8601 string with a literal "Z"
timezone indicator so all timestamps are stable and consistent.
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
