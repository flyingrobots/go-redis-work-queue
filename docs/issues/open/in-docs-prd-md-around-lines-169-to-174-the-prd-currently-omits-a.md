## In docs/PRD.md around lines 169 to 174, the PRD currently omits a

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575188

- [review_comment] 2025-09-16T03:18:48Z by coderabbitai[bot] (docs/PRD.md:188)

```text
In docs/PRD.md around lines 169 to 174, the PRD currently omits a
Kubernetes-ready readiness endpoint; add a /readyz readiness probe now and
document the exact checks and expected responses. Implement a /readyz HTTP
endpoint in the service that performs: a Redis ping (fail if unreachable or auth
fails), verification that required worker goroutines have started and are
processing (e.g., heartbeat or running flag), and the circuit-breaker state
check (fail if open or tripped); return 200 with JSON {status:"ok",
checks:{...}} when all pass and 500 with details when any fail. Update the docs
to list each check, the exact probe path, expected JSON schema, and include
example k8s readinessProbe snippet (httpGet path:/readyz,
initialDelaySeconds/periodSeconds) so deployments can use it immediately.
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
