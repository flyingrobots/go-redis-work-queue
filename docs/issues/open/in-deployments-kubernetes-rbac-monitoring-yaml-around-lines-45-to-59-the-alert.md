## In deployments/kubernetes/rbac-monitoring.yaml around lines 45 to 59, the alert

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066756

- [review_comment] 2025-09-18T16:02:30Z by coderabbitai[bot] (deployments/kubernetes/rbac-monitoring.yaml:59)

```text
In deployments/kubernetes/rbac-monitoring.yaml around lines 45 to 59, the alert
expression mixes different label keys (job vs app) which can cause silent
mismatches; update the rule to consistently use job="rbac-token-service"
everywhere in the expression (both the error rate numerator and the total
request denominator), and then search and update dashboard panels/targets to use
the same job="rbac-token-service" selector (or switch all to a chosen canonical
selector such as service=<name> across alerts, recording rules and dashboards)
so all queries use the identical label key/value.
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
