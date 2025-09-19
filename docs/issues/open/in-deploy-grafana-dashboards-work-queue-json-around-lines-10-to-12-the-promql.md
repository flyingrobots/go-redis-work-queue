## In deploy/grafana/dashboards/work-queue.json around lines 10 to 12, the PromQL

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814681

- [review_comment] 2025-09-16T22:46:55Z by coderabbitai[bot] (deploy/grafana/dashboards/work-queue.json:12)

```text
In deploy/grafana/dashboards/work-queue.json around lines 10 to 12, the PromQL
currently computes a global p95 across all queues; change the query to aggregate
histograms by both le and queue (sum by (le, queue) (rate(...))) so
histogram_quantile(0.95, ...) is evaluated per-queue, and set the panel/metric
legendFormat to include the queue label (e.g. {{queue}}) so operators can
identify which queue the p95 belongs to.
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
