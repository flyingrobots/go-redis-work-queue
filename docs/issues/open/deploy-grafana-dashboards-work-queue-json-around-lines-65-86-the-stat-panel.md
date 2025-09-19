## deploy/grafana/dashboards/work-queue.json around lines 65-86: the Stat panel

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814685

- [review_comment] 2025-09-16T22:46:55Z by coderabbitai[bot] (deploy/grafana/dashboards/work-queue.json:86)

```text
deploy/grafana/dashboards/work-queue.json around lines 65-86: the Stat panel
currently computes a single value across multiple series unpredictably (it picks
a sample); change the reduceOptions to aggregate across series by using the
"sum" calculation (set reduceOptions.calcs to ["sum"]) so the panel shows total
active workers, and optionally add a value text override "{{__value.raw}}
workers" and explicit thresholds with 0 -> red and >0 -> green.
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
