## In README.md lines 44-48 the build section targets Go 1.25, but go.mod (line 3)

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683159

- [review_comment] 2025-09-16T14:20:40Z by coderabbitai[bot] (README.md:55)

```text
In README.md lines 44-48 the build section targets Go 1.25, but go.mod (line 3)
still declares `go 1.24.0`; update go.mod line 3 to `go 1.25` (or `1.25.0`) so
it matches README and the CI workflow files (ci.yml, release.yml,
goreleaser.yml, changelog.yml which use go-version: '1.25.x'), commit the change
and re-run CI to verify everything passes.
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
