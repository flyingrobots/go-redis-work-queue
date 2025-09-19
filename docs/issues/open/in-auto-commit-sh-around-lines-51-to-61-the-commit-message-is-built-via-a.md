## In auto_commit.sh around lines 51 to 61, the commit message is built via a

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814675

- [review_comment] 2025-09-16T22:46:55Z by coderabbitai[bot] (auto_commit.sh:61)

```text
In auto_commit.sh around lines 51 to 61, the commit message is built via a
subshell heredoc which can introduce trailing-newline quirks; replace that with
two explicit strings and pass them to git commit using two -m flags: build a
subject variable like "chore(slaps): auto-sync progress - $DONE done / $OPEN
open" and a body variable containing the Stats block without relying on
command-substitution heredoc, then run git commit -m "$subject" -m "$body" and
preserve the same multiline body formatting within the body string.
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
