## In deployments/admin-api/monitoring.yaml around lines 29 to 36, the PromQL uses

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856225

- [review_comment] 2025-09-16T23:20:23Z by coderabbitai[bot] (deployments/admin-api/monitoring.yaml:36)

```text
In deployments/admin-api/monitoring.yaml around lines 29 to 36, the PromQL uses
histogram_quantile directly on per-series buckets which causes noisy per-series
quantiles; aggregate the bucket counts with sum by (le) (and any other desired
grouping like job/handler) over the rate window before passing to
histogram_quantile. Change the expression to sum the rate of
http_request_duration_seconds_bucket by (le) and then call
histogram_quantile(0.95, ...) so the 95th percentile is computed across
aggregated buckets rather than per-series.
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
