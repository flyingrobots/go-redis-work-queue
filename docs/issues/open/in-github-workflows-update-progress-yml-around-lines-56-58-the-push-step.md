## In .github/workflows/update-progress.yml around lines 56-58, the push step

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360044748

- [review_comment] 2025-09-18T15:57:46Z by coderabbitai[bot] (.github/workflows/update-progress.yml:58)

```text
In .github/workflows/update-progress.yml around lines 56-58, the push step
currently does a plain git push which will fail on non-fast-forward updates;
instead make the run step try to rebase the local changes onto the remote and
then push to avoid race failures — i.e., fetch the remote, perform a git pull
--rebase (or git rebase origin/<branch>) to incorporate upstream commits,
resolve/abort on conflicts as necessary, then git push; ensure the step still
only runs when changes exist.
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
