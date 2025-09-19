## In deployments/admin-api/k8s-redis.yaml around lines 7 to 12, the ServiceAccount

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814690

- [review_comment] 2025-09-16T22:46:56Z by coderabbitai[bot] (deployments/admin-api/k8s-redis.yaml:12)

```text
In deployments/admin-api/k8s-redis.yaml around lines 7 to 12, the ServiceAccount
is created without disabling token automounting; add
automountServiceAccountToken: false to the ServiceAccount spec (i.e., under
metadata add the automountServiceAccountToken field at the same indentation
level) so pods using this SA do not automatically get a token mounted.
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
