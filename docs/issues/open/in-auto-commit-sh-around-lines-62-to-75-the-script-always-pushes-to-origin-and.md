## In auto_commit.sh around lines 62 to 75, the script always pushes to origin and

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814676

- [review_comment] 2025-09-16T22:46:55Z by coderabbitai[bot] (auto_commit.sh:75)

```text
In auto_commit.sh around lines 62 to 75, the script always pushes to origin and
uses an unescaped @{u} which triggers SC1083; change it to detect the actual
upstream remote/branch using an escaped ref (e.g. capture upstream_ref=$(git
rev-parse --abbrev-ref --symbolic-full-name '\@{u}' 2>/dev/null)), if
upstream_ref is non-empty split it into upstream_remote and upstream_branch and
push to that remote/branch (git push "$upstream_remote"
"$current_branch:$upstream_branch"), otherwise fall back to creating an upstream
with --set-upstream (e.g. git push --set-upstream origin "$current_branch"), and
keep the existing success/failure logging.
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
