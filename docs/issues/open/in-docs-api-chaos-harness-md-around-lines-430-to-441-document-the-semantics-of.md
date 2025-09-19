## In docs/api/chaos-harness.md around lines 430 to 441, document the semantics of

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061302

- [review_comment] 2025-09-18T16:01:17Z by coderabbitai[bot] (docs/api/chaos-harness.md:441)

```text
In docs/api/chaos-harness.md around lines 430 to 441, document the semantics of
the wildcard used in examples: explicitly state that scope_value: "*" means “all
workers” (and similarly for queue scope), clarify its precedence relative to
specific IDs/names (e.g., exact matches take precedence over the wildcard),
state whether other wildcard patterns or full regex are supported or not (and if
supported, give syntax and matching rules), and add a short example and a note
right after the “Worker Scope” section showing usage and precedence so readers
aren’t left to guess.
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
