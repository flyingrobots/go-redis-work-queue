## In deployments/admin-api/Dockerfile around lines 18 to 21, the Go build step

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360044788

- [review_comment] 2025-09-18T15:57:46Z by coderabbitai[bot] (deployments/admin-api/Dockerfile:21)

```text
In deployments/admin-api/Dockerfile around lines 18 to 21, the Go build step
currently uses -trimpath but still embeds VCS metadata; to make builds
reproducible append -buildvcs=false to the ldflags so the -ldflags string
becomes "-s -w -X main.version=${VERSION} -buildvcs=false" (preserving existing
flags and output path) to disable VCS stamping.
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
