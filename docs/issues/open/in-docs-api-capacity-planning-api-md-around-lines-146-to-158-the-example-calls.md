## In docs/api/capacity-planning-api.md around lines 146 to 158 the example calls

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572261

- [review_comment] 2025-09-16T03:17:34Z by coderabbitai[bot] (docs/api/capacity-planning-api.md:169)

```text
In docs/api/capacity-planning-api.md around lines 146 to 158 the example calls
calc.Calculate(..., metrics) but never declares metrics, causing copy-paste
compile errors; either add a minimal declaration such as metrics :=
capacityplanning.Metrics{ /* fill required fields */ } immediately before the
call, or remove the metrics parameter from the example call and adjust the
argument list accordingly so the snippet compiles.
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
