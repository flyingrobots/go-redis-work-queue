## In deployments/scripts/health-check-rbac.sh around lines 41 to 44, the kubectl

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066908

- [review_comment] 2025-09-18T16:02:32Z by coderabbitai[bot] (deployments/scripts/health-check-rbac.sh:44)

```text
In deployments/scripts/health-check-rbac.sh around lines 41 to 44, the kubectl
cluster-info call can hang CI; wrap it with a timeout (e.g. 10s) and fail if it
exceeds that. Implement: if command -v timeout >/dev/null use timeout 10s
kubectl cluster-info, else use kubectl cluster-info --request-timeout=10s; keep
the existing error message and return 1 when the guarded call fails or times
out.
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
