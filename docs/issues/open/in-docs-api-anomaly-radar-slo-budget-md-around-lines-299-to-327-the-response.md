## In docs/api/anomaly-radar-slo-budget.md around lines 299 to 327, the response

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583641

- [review_comment] 2025-09-16T03:22:10Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:327)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 299 to 327, the response
fields lack units/definitions; update the documentation by adding explicit unit
notes for budget_utilization, current_burn_rate, and time_to_exhaustion —
specify "budget_utilization: fraction [0,1]", "current_burn_rate: budget/hour
(fraction of total budget consumed per hour)", and "time_to_exhaustion: RFC3339
duration string" under the response example so readers know the domains and
units.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | No | No | - | - |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Pending review. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:995
>
> **Alternatives Considered**
> Not documented.
>
> **Lesson(s) Learned**
> None recorded.
