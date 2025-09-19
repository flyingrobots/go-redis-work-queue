## In .github/workflows/changelog.yml around lines 36-38, the workflow pushes

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567131

- [review_comment] 2025-09-16T03:15:01Z by coderabbitai[bot] (.github/workflows/changelog.yml:42)

```text
In .github/workflows/changelog.yml around lines 36-38, the workflow pushes
directly to the default branch and can have concurrent runs collide; add a
top-level concurrency stanza to serialize runs (use a stable group key tied to
the repository/workflow, e.g. "changelog-${{ github.repository }}-${{
github.workflow }}" or include default_branch) and set cancel-in-progress: false
so only one push to the default branch runs at a time; add this concurrency
block at the top of the workflow YAML to prevent race-prone pushes.
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
