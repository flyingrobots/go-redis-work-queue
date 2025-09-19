## In docs/api/anomaly-radar-slo-budget.md around lines 132-135 (and also apply

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583619

- [review_comment] 2025-09-16T03:22:09Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:198)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 132-135 (and also apply
same change at 247-249 and 335-337), the "duration" type for query parameters is
not defined; update the docs to state that durations use Go's time.ParseDuration
format and give a short example (e.g., "duration (Go time.ParseDuration format,
e.g., 30m, 1h, 24h, 7h30m)"). Insert this one-line clarification immediately
after each query-parameter list mentioned so callers know the expected format.
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
