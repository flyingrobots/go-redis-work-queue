## In docs/api/admin-api.md around lines 360 to 366, the health endpoint is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360067091

- [review_comment] 2025-09-18T16:02:34Z by coderabbitai[bot] (docs/api/admin-api.md:366)

```text
In docs/api/admin-api.md around lines 360 to 366, the health endpoint is
documented as /health but the codebase and deployment use /healthz; update the
documentation to use /healthz (and similarly mention /readyz where appropriate)
so probes and runbooks match: replace occurrences of /health with /healthz in
this section and verify the example HTTP request and response remain unchanged.
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
