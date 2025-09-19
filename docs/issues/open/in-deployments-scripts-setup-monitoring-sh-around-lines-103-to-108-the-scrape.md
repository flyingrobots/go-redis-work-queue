## In deployments/scripts/setup-monitoring.sh around lines 103 to 108, the scrape

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039054

- [review_comment] 2025-09-18T15:56:34Z by coderabbitai[bot] (deployments/scripts/setup-monitoring.sh:108)

```text
In deployments/scripts/setup-monitoring.sh around lines 103 to 108, the scrape
config sets honorLabels: true which allows targets to override job/instance
labels; remove this line or set honorLabels: false to prevent targets from
clobbering labels and breaking grouping, then re-generate/validate the resulting
Prometheus config and restart/reload the monitoring stack to apply the change.
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
