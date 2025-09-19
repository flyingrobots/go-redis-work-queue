## In deployments/admin-api/k8s-redis.yaml around lines 31-34 (and also lines

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814692

- [review_comment] 2025-09-16T22:46:56Z by coderabbitai[bot] (deployments/admin-api/k8s-redis.yaml:34)

```text
In deployments/admin-api/k8s-redis.yaml around lines 31-34 (and also lines
49-51), the securityContext uses runAsNonRoot: true with fsGroup: 1000 which
collides with common host UIDs; choose a high, non‑host UID/GID (e.g. >= 100000
or from your cluster's allocated range) and set runAsUser and runAsGroup to that
UID/GID and update fsGroup to the same high GID; ensure the chosen ID is used
consistently for both container specs so the pod runs as non‑root without
conflicting with host users.
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
