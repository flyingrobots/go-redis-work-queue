## In .githooks/pre-commit around lines 19 to 25, the script unconditionally runs

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360032927

- [review_comment] 2025-09-18T15:55:17Z by coderabbitai[bot] (.githooks/pre-commit:25)

```text
In .githooks/pre-commit around lines 19 to 25, the script unconditionally runs
"git add docs/features-ledger.md README.md" which force-stages changes without
an opt-out; modify the script to gate the auto-stage behind an environment flag
(for example AUTO_STAGE_DOCS) so that by default it auto-stages but if
AUTO_STAGE_DOCS is set to false/0 it skips staging; implement a simple
conditional around the git add that echoes a message when skipping and preserves
the existing "|| true" behavior to avoid breaking the hook.
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
