## In deployments/docker/grafana/datasources/prometheus.yaml around lines 3 to 9,

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360038885

- [review_comment] 2025-09-18T15:56:32Z by coderabbitai[bot] (deployments/docker/grafana/datasources/prometheus.yaml:9)

```text
In deployments/docker/grafana/datasources/prometheus.yaml around lines 3 to 9,
the Prometheus datasource is missing a uid so dashboards that reference uid
"Prometheus" will 404; add a fixed uid field (set uid: Prometheus to match the
dashboards) to the datasource definition so it can be reliably referenced, keep
the rest of the fields unchanged.
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
