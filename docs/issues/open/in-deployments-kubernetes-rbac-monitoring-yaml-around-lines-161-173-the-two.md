## In deployments/kubernetes/rbac-monitoring.yaml around lines 161–173, the two

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360038921

- [review_comment] 2025-09-18T15:56:32Z by coderabbitai[bot] (deployments/kubernetes/rbac-monitoring.yaml:173)

```text
In deployments/kubernetes/rbac-monitoring.yaml around lines 161–173, the two
alerts reference metrics (rbac_admin_actions_total and
rbac_key_last_rotation_timestamp) that are not exported by the RBAC service;
either implement and register those metrics in the RBAC service or
remove/replace these alerts with existing signals. To implement: add Prometheus
metric definitions (e.g., prometheus.NewCounter for admin actions and
prometheus.NewGauge or prometheus.NewGaugeFunc for last-rotation timestamp) in
the RBAC service code, register them with prometheus.MustRegister (see
internal/obs/metrics.go and internal/producer-backpressure/types.go for
examples), and update instrumentation to increment/set them; to remove/replace:
delete these alert blocks from rbac-monitoring.yaml or change expr to use valid
metrics already exported by the service.
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
