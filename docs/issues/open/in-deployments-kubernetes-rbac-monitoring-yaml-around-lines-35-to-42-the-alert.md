## In deployments/kubernetes/rbac-monitoring.yaml around lines 35 to 42, the alert

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360038904

- [review_comment] 2025-09-18T15:56:32Z by coderabbitai[bot] (deployments/kubernetes/rbac-monitoring.yaml:42)

```text
In deployments/kubernetes/rbac-monitoring.yaml around lines 35 to 42, the alert
expr uses up{app="rbac-token-service"} == 0 which misses the case where all
targets disappear; replace the expr with a combined check using sum and absent,
e.g. use an expression that evaluates true when either the summed up is zero or
the series is absent (for example: sum(up{app="rbac-token-service"}) == 0 or
absent(up{app="rbac-token-service"})), leaving the for, labels and annotations
unchanged.
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
