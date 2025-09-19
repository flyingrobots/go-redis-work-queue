## In deployments/docker/docker-compose.yaml around lines 20 to 56 (and also apply

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066702

- [review_comment] 2025-09-18T16:02:29Z by coderabbitai[bot] (deployments/docker/docker-compose.yaml:56)

```text
In deployments/docker/docker-compose.yaml around lines 20 to 56 (and also apply
same change to lines 57 to 84), the service blocks lack restart policies so
containers won’t automatically recover; add a restart policy to each application
service (for example restart: unless-stopped or restart: always) directly under
the service definition (align with other top-level keys such as ports/env_file)
and, if needed, add restart_policy options for finer control
(maximum_retry_count, window) to ensure services are automatically restarted on
failure or Docker daemon restarts.
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
