## In deploy/data/test.txt lines 1-1, this test file must never be included in

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814679

- [review_comment] 2025-09-16T22:46:55Z by coderabbitai[bot] (deploy/data/test.txt:1)

```text
In deploy/data/test.txt lines 1-1, this test file must never be included in
built images or releases; add entries to project ignore/config files so it’s
always excluded. Update .dockerignore and .helmignore to include
deploy/data/test.txt (and deploy/data/ as appropriate), and if you use
goreleaser, add exclusion patterns for test/ or deploy/data/test.txt (or fixture
paths) under the archives/exclude section in .goreleaser.yaml so packaging never
includes this file.
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
