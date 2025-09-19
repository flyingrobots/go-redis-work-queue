## In docs/api/capacity-planning-api.md around lines 289 to 308, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912871

- [review_comment] 2025-09-18T12:12:39Z by coderabbitai[bot] (docs/api/capacity-planning-api.md:308)

```text
In docs/api/capacity-planning-api.md around lines 289 to 308, the
TrafficSpike/TrafficPattern enum values are inconsistently named (docs say
“instant, linear, exp, bell” while examples use `SpikeBell`) and the actual
constant names are missing; update the doc to list the exact constant names used
in the code (include the PatternType and SpikeShape constant values as defined
earlier), replace informal names with the precise enum identifiers used in the
type defs, and add a short code block showing the constants so the documentation
and examples match exactly.
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
