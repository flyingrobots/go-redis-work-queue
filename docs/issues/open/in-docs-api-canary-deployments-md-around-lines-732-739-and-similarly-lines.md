## In docs/api/canary-deployments.md around lines 732-739 (and similarly lines

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039258

- [review_comment] 2025-09-18T15:56:37Z by coderabbitai[bot] (docs/api/canary-deployments.md:739)

```text
In docs/api/canary-deployments.md around lines 732-739 (and similarly lines
13-21), clarify the ambiguous "per API key" rate limit by explicitly stating the
exact subject used for counting: whether it's the raw API token string for token
auth, or the JWT's subject claim (e.g., `sub`) or tenant identifier for JWT
auth; state both cases if both auth methods are supported. Update the rate limit
bullets to name the exact key used for each auth method, confirm that the
X-RateLimit-* headers are emitted and identical for both authentication methods,
and mirror the same precise language in the Authentication section so both
places describe the same subject and header behavior.
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
