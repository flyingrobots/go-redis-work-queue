## In docs/api/anomaly-radar-slo-budget.md around lines 51 to 58, the docs present

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583592

- [review_comment] 2025-09-16T03:22:09Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:58)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 51 to 58, the docs present
two conflicting collector APIs (closures-based SimpleMetricsCollector and an
interface-based QueueMetricsCollector); remove the closures-based
SimpleMetricsCollector snippet and keep the interface-based approach, add an
explicit MetricCollector interface signature description immediately before the
QueueMetricsCollector example so readers see the expected methods and types;
repeat the same cleanup for the other occurrence around lines 446–470 by
deleting the closure example and ensuring the interface signature is documented
prior to the QueueMetricsCollector sample.
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
> Pending review. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:674
>
> **Alternatives Considered**
> Not documented.
>
> **Lesson(s) Learned**
> None recorded.
