## In deployments/scripts/deploy-staging.sh around lines 182 to 186 there is a

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066896

- [review_comment] 2025-09-18T16:02:32Z by coderabbitai[bot] (deployments/scripts/deploy-staging.sh:186)

```text
In deployments/scripts/deploy-staging.sh around lines 182 to 186 there is a
duplicate call to register_port_forward "$PF_PID"; remove the repeated line so
the port-forward PID is only registered once. Keep the single
register_port_forward "$PF_PID" immediately after PF_PID=$! (and before sleep 5)
to ensure the background PID is recorded exactly once; do not change the
surrounding port-forward or PID assignment logic.
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
