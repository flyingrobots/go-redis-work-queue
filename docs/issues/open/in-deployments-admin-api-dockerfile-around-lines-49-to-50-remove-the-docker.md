## In deployments/admin-api/Dockerfile around lines 49 to 50, remove the Docker

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360044839

- [review_comment] 2025-09-18T15:57:46Z by coderabbitai[bot] (deployments/admin-api/Dockerfile:50)

```text
In deployments/admin-api/Dockerfile around lines 49 to 50, remove the Docker
HEALTHCHECK and the wget command so the image no longer includes a
container-side HTTP probe; delete the two HEALTHCHECK lines and any installation
of wget/alpine packages used solely for that check so the image is leaner and
rely on Kubernetes liveness/readiness probes instead.
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
