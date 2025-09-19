## In deployments/admin-api/k8s-redis.yaml around lines 66 to 79, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814695

- [review_comment] 2025-09-16T22:46:56Z by coderabbitai[bot] (deployments/admin-api/k8s-redis.yaml:79)

```text
In deployments/admin-api/k8s-redis.yaml around lines 66 to 79, the
liveness/readiness probes use exec redis-cli ping; replace these with tcpSocket
probes against the Redis port (typically containerPort 6379) to avoid relying on
an external binary. For both probes remove the exec block and add tcpSocket:
port: 6379, preserving or adjusting initialDelaySeconds and periodSeconds as
appropriate; ensure probe entries remain under the container spec and validate
YAML indentation.
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
