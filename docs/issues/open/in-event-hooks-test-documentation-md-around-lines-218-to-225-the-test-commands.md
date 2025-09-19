## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 218 to 225, the test commands

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358913042

- [review_comment] 2025-09-18T12:12:40Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:225)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 218 to 225, the test commands
do not enable the Go race detector (and optionally test shuffling) which misses
concurrency bugs; update the documented test commands to include the -race flag
on go test (e.g., go test -v -race ./... and go test -v -race
-coverprofile=coverage.out ./...) and optionally show adding test shuffling
(e.g., -shuffle=on) where supported, and update any examples or notes to mention
using -race (and -shuffle) for concurrency-sensitive suites.
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
