## In docs/api/capacity-planning-api.md around lines 322-324, the example import

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360067105

- [review_comment] 2025-09-18T16:02:34Z by coderabbitai[bot] (docs/api/capacity-planning-api.md:324)

```text
In docs/api/capacity-planning-api.md around lines 322-324, the example import
references an internal package
("github.com/flyingrobots/go-redis-work-queue/internal/automatic-capacity-planning")
which cannot be imported outside its module; either change the import to a
public package path (move the package out of internal or point to a published
public module) or add a clear note immediately above the snippet that this
example must be placed inside that repository/module tree (so readers know it
won’t work from external modules). Ensure the doc shows the correct public
import path or the placement caveat, and remove the misleading “Replace the
import above…” line if you opt for the placement note.
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
