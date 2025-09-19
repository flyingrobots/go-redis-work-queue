## In docs/api/capacity-planning-api.md around lines 508 to 561, the error-handling

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912912

- [review_comment] 2025-09-18T12:12:39Z by coderabbitai[bot] (docs/api/capacity-planning-api.md:561)

```text
In docs/api/capacity-planning-api.md around lines 508 to 561, the error-handling
example mixes direct type assertions with pointer/value forms and uses a bare
time.Sleep; replace the direct type assertion with errors.As to reliably extract
a *capacityplanning.PlannerError, and remove the magic time.Sleep by showing a
retry loop that respects context deadlines (e.g., loop with backoff and check
ctx.Done or a context.WithDeadline/WithTimeout) so the example demonstrates
safe, cancellable retries in library code.
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
