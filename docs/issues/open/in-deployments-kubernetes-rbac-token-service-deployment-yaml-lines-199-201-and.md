## In deployments/kubernetes/rbac-token-service-deployment.yaml lines 199-201 (and

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066815

- [review_comment] 2025-09-18T16:02:31Z by coderabbitai[bot] (deployments/kubernetes/rbac-token-service-deployment.yaml:201)

```text
In deployments/kubernetes/rbac-token-service-deployment.yaml lines 199-201 (and
also 263-267), the pod securityContext uses UID/GID 1000 which may collide with
host users. Update runAsUser and fsGroup to a high, non-host ID like 10001
(consistent across all containers and pod-level securityContext). Keep
runAsNonRoot: true. Verify any related runAsGroup or container-level overrides
also use 10001, and ensure both affected blocks are updated.
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
