## In deployments/admin-api/k8s-redis.yaml around lines 40 to 45, the Redis command

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814693

- [review_comment] 2025-09-16T22:46:56Z by coderabbitai[bot] (deployments/admin-api/k8s-redis.yaml:49)

```text
In deployments/admin-api/k8s-redis.yaml around lines 40 to 45, the Redis command
is missing an explicit data directory; add the flag --dir /data to the command
array so Redis writes to /data, and ensure the Pod spec includes a volumeMount
for /data backed by a persistent volume (or emptyDir if ephemeral) and a
corresponding volume or PVC entry in the deployment.
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
