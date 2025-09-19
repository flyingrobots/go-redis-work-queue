## In deployments/kubernetes/monitoring.yaml around lines 49-56, the alert expr

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353924984

- [review_comment] 2025-09-17T00:25:55Z by coderabbitai[bot] (deployments/kubernetes/monitoring.yaml:56)

```text
In deployments/kubernetes/monitoring.yaml around lines 49-56, the alert expr
only matches when an individual target reports up==0 and will miss the case
where all targets disappear; change the expr to cover both absence and zero sum,
e.g. replace the current expr with a compound that uses sum and absent such as:
absent(up{app="admin-api"}) OR sum(up{app="admin-api"}) == 0, leaving the
for/labels/annotations intact so the alert fires when the app is fully missing
or all instances are down.
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
