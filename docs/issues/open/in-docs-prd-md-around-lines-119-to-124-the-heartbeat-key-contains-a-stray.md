## In docs/PRD.md around lines 119 to 124, the heartbeat key contains a stray

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061339

- [review_comment] 2025-09-18T16:01:18Z by coderabbitai[bot] (docs/PRD.md:124)

```text
In docs/PRD.md around lines 119 to 124, the heartbeat key contains a stray
backtick/space and the key formatting is inconsistent; remove the extra
backtick/space from `jobqueue:processing:worker:<ID> ` and make all keys
consistently formatted (e.g., wrap each key in backticks without trailing
spaces) so the list entries are uniform.
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
