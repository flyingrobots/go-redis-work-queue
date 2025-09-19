## In deployments/admin-api/monitoring.yaml around lines 11 to 18, the alert

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856220

- [review_comment] 2025-09-16T23:20:23Z by coderabbitai[bot] (deployments/admin-api/monitoring.yaml:18)

```text
In deployments/admin-api/monitoring.yaml around lines 11 to 18, the alert
expression only checks individual target up==0 and will not fire when all
targets vanish; replace the expr with a sum/absent check such as: use
sum(up{job="admin-api"}) == 0 or absent(up{job="admin-api"}) so the alert
triggers when the total is zero or the metric is missing, leaving the
for/labels/annotations unchanged.
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
