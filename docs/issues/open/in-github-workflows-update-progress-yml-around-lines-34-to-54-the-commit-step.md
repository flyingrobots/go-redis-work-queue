## In .github/workflows/update-progress.yml around lines 34 to 54, the commit step

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360044726

- [review_comment] 2025-09-18T15:57:45Z by coderabbitai[bot] (.github/workflows/update-progress.yml:54)

```text
In .github/workflows/update-progress.yml around lines 34 to 54, the commit step
should explicitly mark the repo safe to avoid “dubious ownership” errors and
must handle empty staging reliably; add a git config --global --add
safe.directory "$(pwd)" (or "$GITHUB_WORKSPACE") before any git commands, ensure
the files_to_add array is only passed to git add when non-empty (as you already
guard) and replace the cached-diff check with a robust staged-change check such
as using git diff --cached --quiet || git diff --quiet to detect staged or
unstaged changes or use git diff --name-only --cached | grep -q . to determine
if there are staged files before committing, then only run git commit when there
is at least one staged change and set the changed output accordingly.
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
