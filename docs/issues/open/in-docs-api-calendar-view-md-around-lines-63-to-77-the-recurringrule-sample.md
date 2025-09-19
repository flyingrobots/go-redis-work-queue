## In docs/api/calendar-view.md around lines 63 to 77, the RecurringRule sample

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061257

- [review_comment] 2025-09-18T16:01:17Z by coderabbitai[bot] (docs/api/calendar-view.md:77)

```text
In docs/api/calendar-view.md around lines 63 to 77, the RecurringRule sample
mixes Go's time.Duration with JSON string values (e.g. "300s"); update the
documentation and sample struct so JSON shows a string for Jitter (or introduce
a custom Duration type) — either change the Jitter field to string in the sample
JSON struct and use a string example like "300s", or, if the SDK should keep
strong typing, define a custom Duration type that marshals/unmarshals to/from a
JSON string and replace time.Duration with that type in the sample; ensure the
docs show the final chosen representation and include an example value formatted
as a string.
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
