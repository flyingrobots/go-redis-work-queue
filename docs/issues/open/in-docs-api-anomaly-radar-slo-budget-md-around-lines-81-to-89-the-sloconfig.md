## In docs/api/anomaly-radar-slo-budget.md around lines 81 to 89, the SLOConfig

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583597

- [review_comment] 2025-09-16T03:22:09Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:89)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 81 to 89, the SLOConfig
struct references BurnRateThresholds but that type is not defined; add a new
BurnRateThresholds type immediately after the SLOConfig block with four fields:
FastBurnRate (float64) and FastBurnWindow (time.Duration) for the fast alert
threshold and its evaluation window, and SlowBurnRate (float64) and
SlowBurnWindow (time.Duration) for the slow alert threshold and its evaluation
window, each with brief inline comments explaining units (budget/hour for rates,
time.Duration for windows).
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
> Pending review. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:693
>
> **Alternatives Considered**
> Not documented.
>
> **Lesson(s) Learned**
> None recorded.
