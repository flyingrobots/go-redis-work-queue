## In deployments/kubernetes/rbac-monitoring.yaml around lines 45 to 51 (and also

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360038916

- [review_comment] 2025-09-18T15:56:32Z by coderabbitai[bot] (deployments/kubernetes/rbac-monitoring.yaml:51)

```text
In deployments/kubernetes/rbac-monitoring.yaml around lines 45 to 51 (and also
lines 88 to 100), the alert uses job="rbac-token-service" while other
rules/dashboards use app="rbac-token-service", causing alerts to miss metrics;
pick one label (recommended: app) and update the PromQL selectors to the chosen
label consistently (e.g., replace job="rbac-token-service" with
app="rbac-token-service" in this alert and the other affected rules), and verify
any relabeling rules export that label so both metrics and alerts match.
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
