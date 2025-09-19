## .github/workflows/ci.yml around line 27: CI is using go-version '1.25.x' while

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353724329

- [review_comment] 2025-09-16T21:42:54Z by coderabbitai[bot] (.github/workflows/ci.yml:27)

```text
.github/workflows/ci.yml around line 27: CI is using go-version '1.25.x' while
go.mod declares 'go 1.24.0'; pick one consistent version and update the
corresponding file: either change go.mod to "go 1.25" (and add/verify a
//go:build toolchain directive if your repo uses toolchain management) or change
.github/workflows/ci.yml to use '1.24.x'; after making the change run the full
test suite (and go mod tidy / go vet / go test ./...) on the chosen Go version
before merging.
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
