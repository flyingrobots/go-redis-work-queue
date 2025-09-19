## In .gitignore around lines 28-33, you need to prevent accidental commits of

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679347

- [review_comment] 2025-09-16T14:19:24Z by coderabbitai[bot] (.gitignore:34)

```text
In .gitignore around lines 28-33, you need to prevent accidental commits of
environment/secret files: add entries to ignore common env filenames (e.g. .env,
.env.* , .env.local, .env.production) while keeping any explicit templates like
.env.example tracked; update the file by appending those patterns (or placing
them before any allow rules) so env files are excluded from git commits.
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
