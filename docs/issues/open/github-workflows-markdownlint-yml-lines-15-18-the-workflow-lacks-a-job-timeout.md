## .github/workflows/markdownlint.yml lines 15-18: the workflow lacks a job timeout

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353724347

- [review_comment] 2025-09-16T21:42:54Z by coderabbitai[bot] (.github/workflows/markdownlint.yml:18)

```text
.github/workflows/markdownlint.yml lines 15-18: the workflow lacks a job timeout
which can cause infinite hangs; add a timeout-minutes setting for the lint job
(e.g., timeout-minutes: 10) under the job definition (right below runs-on or at
the job root) to cap execution time and fail fast if it runs too long.
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
