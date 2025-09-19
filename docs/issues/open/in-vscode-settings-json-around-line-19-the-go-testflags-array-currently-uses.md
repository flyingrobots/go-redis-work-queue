## In .vscode/settings.json around line 19, the go.testFlags array currently uses

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353792407

- [review_comment] 2025-09-16T22:30:42Z by coderabbitai[bot] (.vscode/settings.json:19)

```text
In .vscode/settings.json around line 19, the go.testFlags array currently uses
["-race", "-count=1"] but lacks a test timeout, which can allow hung tests to
consume CI time; add a sensible timeout flag (for example "-timeout=2m" or
another project-appropriate duration) to the array so it becomes ["-race",
"-count=1", "-timeout=2m"] to ensure tests fail fast on hangs.
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
