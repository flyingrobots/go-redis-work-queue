## In deployments/scripts/deploy-staging.sh around lines 184-185 you call

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039024

- [review_comment] 2025-09-18T15:56:34Z by coderabbitai[bot] (deployments/scripts/deploy-staging.sh:185)

```text
In deployments/scripts/deploy-staging.sh around lines 184-185 you call
register_port_forward "$PF_PID" twice which double-registers the same PID and
will attempt to kill it twice; remove the duplicated register_port_forward
invocation so the PID is registered only once (leave a single
register_port_forward "$PF_PID" call) and optionally add a sanity check that
PF_PID is non-empty before calling to avoid registering an empty value.
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
