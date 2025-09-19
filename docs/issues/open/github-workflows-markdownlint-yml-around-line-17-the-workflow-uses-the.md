## .github/workflows/markdownlint.yml around line 17: the workflow uses the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353724349

- [review_comment] 2025-09-16T21:42:54Z by coderabbitai[bot] (.github/workflows/markdownlint.yml:17)

```text
.github/workflows/markdownlint.yml around line 17: the workflow uses the
floating runner "ubuntu-latest" which can change unexpectedly; replace it with a
specific, pinned runner version such as "ubuntu-22.04" (or your project's chosen
LTS like "ubuntu-20.04") by updating the runs-on value to that concrete label so
CI runs are reproducible.
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
