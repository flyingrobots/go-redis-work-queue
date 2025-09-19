## In deployments/kubernetes/admin-api-deployment.yaml around lines 134 to 147, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033213

- [review_comment] 2025-09-18T15:55:19Z by coderabbitai[bot] (deployments/kubernetes/admin-api-deployment.yaml:147)

```text
In deployments/kubernetes/admin-api-deployment.yaml around lines 134 to 147, the
liveness probe path uses /health while the standard is /healthz and readiness
should be /readyz; update the liveness httpGet.path to /healthz, confirm
readiness remains /readyz, and ensure both probe port and timing settings remain
unchanged; then search and align Dockerfile/Compose and other deployment
artifacts to use /healthz and /readyz consistently.
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
