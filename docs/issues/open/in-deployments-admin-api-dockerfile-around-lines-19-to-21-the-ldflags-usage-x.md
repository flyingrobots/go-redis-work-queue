## In deployments/admin-api/Dockerfile around lines 19 to 21, the ldflags usage -X

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360044814

- [review_comment] 2025-09-18T15:57:46Z by coderabbitai[bot] (deployments/admin-api/Dockerfile:21)

```text
In deployments/admin-api/Dockerfile around lines 19 to 21, the ldflags usage -X
main.version=${VERSION} will fail because there is no var version string in
package main (cmd/admin-api); either add a top-level declaration in
cmd/admin-api (package main) like a var version string so -X main.version can
link, or change the Dockerfile ldflags to reference the exact exported variable
by its full package path and identifier (case‑sensitive), e.g. -X
'github.com/your/repo/cmd/admin-api.VarName=${VERSION}', and ensure proper
quoting/escaping in the Dockerfile.
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
