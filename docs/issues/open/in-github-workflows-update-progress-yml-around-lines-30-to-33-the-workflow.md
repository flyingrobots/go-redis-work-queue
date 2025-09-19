## In .github/workflows/update-progress.yml around lines 30 to 33, the workflow

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360038808

- [review_comment] 2025-09-18T15:56:31Z by coderabbitai[bot] (.github/workflows/update-progress.yml:33)

```text
In .github/workflows/update-progress.yml around lines 30 to 33, the workflow
currently runs python3 scripts/update_progress.py without verifying the script
exists; add a pre-check that verifies scripts/update_progress.py is present and,
if not, echoes a clear error and exits non-zero so the job fails fast instead of
attempting to run a non-existent script, then only invoke python3 when the file
check passes.
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
