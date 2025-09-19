## In .github/workflows/goreleaser.yml around lines 13–16, the release job is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353724336

- [review_comment] 2025-09-16T21:42:54Z by coderabbitai[bot] (.github/workflows/goreleaser.yml:16)

```text
In .github/workflows/goreleaser.yml around lines 13–16, the release job is
missing concurrency and pinned action SHAs; add a top-level concurrency block
for the release job (group keyed by the ref or workflow and cancel-in-progress:
true) to serialize tag-triggered runs, and replace each external action version
(e.g., actions/checkout@vX, actions/setup-go@vX, goreleaser-action@vX) with an
explicit commit SHA to pin them to immutable references; ensure every uses:
entry in the job steps points to a specific SHA instead of a floating
major/minor tag.
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
