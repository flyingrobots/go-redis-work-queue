## In deployments/admin-api/monitoring.yaml around lines 73-76 (and also lines

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856229

- [review_comment] 2025-09-16T23:20:23Z by coderabbitai[bot] (deployments/admin-api/monitoring.yaml:84)

```text
In deployments/admin-api/monitoring.yaml around lines 73-76 (and also lines
125-127), the dashboard JSON is nested under a top-level "dashboard" key but
Grafana expects the dashboard object at the root; move the inner object out so
the root contains the dashboard fields directly, and add basic metadata keys for
import parity (schemaVersion, version, time, uid) at the root of that object;
ensure the final YAML places the dashboard object at the file root without the
extra "dashboard" wrapper and includes the recommended metadata fields.
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
