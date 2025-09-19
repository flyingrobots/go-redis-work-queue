## In README.md around lines 38-49 there is a mismatch between the documented Go

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569836

- [review_comment] 2025-09-16T03:16:21Z by coderabbitai[bot] (README.md:56)

```text
In README.md around lines 38-49 there is a mismatch between the documented Go
version (Go 1.25+), go.mod (go 1.24.0) and CI (go-version: 1.25.x); update
go.mod to "go 1.25" to match README and CI, run "go mod tidy" locally to refresh
module files, commit the updated go.mod and go.sum, and push so CI (still set to
1.25.x) can verify the build; alternatively, if you prefer 1.24, change README
and CI to 1.24.x and then run go mod tidy and commit.
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
