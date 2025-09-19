## In docs/api/capacity-planning-api.md around lines 365 to 373 the example calls

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912889

- [review_comment] 2025-09-18T12:12:39Z by coderabbitai[bot] (docs/api/capacity-planning-api.md:373)

```text
In docs/api/capacity-planning-api.md around lines 365 to 373 the example calls
planner.GeneratePlan and panics on error — we should not teach panicking in
examples; instead handle the error properly. Replace the panic with structured
error handling: if this is a main/demo show a graceful exit using log.Fatalf
with a clear message and the error (or os.Exit after logging), otherwise return
the error up the call stack (or wrap it with context and return). Ensure the
example imports/uses the chosen logger or returns the error so the sample
demonstrates safe, production-appropriate error handling.
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
