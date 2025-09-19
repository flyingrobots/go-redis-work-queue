## In docs/api/canary-deployments.md around lines 84-90 (and likewise at 112-118,

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039213

- [review_comment] 2025-09-18T15:56:37Z by coderabbitai[bot] (docs/api/canary-deployments.md:90)

```text
In docs/api/canary-deployments.md around lines 84-90 (and likewise at 112-118,
120-131, 508-519, 527-553, 579-589), the duration fields and their example
values are inconsistent (mixing max_duration/min_duration with
max_canary_duration/min_canary_duration and non-canonical formats). Standardize
to a single canonical schema (use max_canary_duration and min_canary_duration
everywhere), convert all example duration values to Go time.Duration canonical
strings (e.g., "2h0m0s", "5m0s"), and update the Parameters section text and
each profile example to reference the exact field names and formats
consistently.
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
