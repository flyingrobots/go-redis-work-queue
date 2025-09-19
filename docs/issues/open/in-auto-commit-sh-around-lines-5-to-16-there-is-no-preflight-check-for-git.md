## In auto_commit.sh around lines 5 to 16, there is no preflight check for git

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814673

- [review_comment] 2025-09-16T22:46:55Z by coderabbitai[bot] (auto_commit.sh:16)

```text
In auto_commit.sh around lines 5 to 16, there is no preflight check for git
configuration so the loop will repeatedly fail if git user.name or user.email
are not set; add a startup preflight function that verifies git is available and
inside a git repo (git rev-parse --is-inside-work-tree), then checks git config
--get user.name and git config --get user.email and exits immediately with a
non-zero status and an explanatory stderr message if any check fails; call this
preflight function once at script startup before entering the main loop so the
script fails fast instead of churning on commit errors.
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
