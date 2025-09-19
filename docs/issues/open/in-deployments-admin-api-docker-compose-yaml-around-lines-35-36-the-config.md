## In deployments/admin-api/docker-compose.yaml around lines 35-36, the config

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033150

- [review_comment] 2025-09-18T15:55:19Z by coderabbitai[bot] (deployments/admin-api/docker-compose.yaml:36)

```text
In deployments/admin-api/docker-compose.yaml around lines 35-36, the config
volume is mounted to /root/configs but the app reads from /app/config (defaults
config/config.yaml and config/admin-api.yaml); change the mount to map your
local config directory into /app/config (for example ./config:/app/config:ro),
keep the audit-logs volume as-is, and ensure the ./config host directory exists
and contains the expected config/config.yaml and config/admin-api.yaml files.
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
