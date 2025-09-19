## In create_review_tasks.py around line 6 (and also update annotations at 53-54

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856199

- [review_comment] 2025-09-16T23:20:22Z by coderabbitai[bot] (create_review_tasks.py:6)

```text
In create_review_tasks.py around line 6 (and also update annotations at 53-54
and 70-75): replace the legacy typing.List import and usages with PEP 585 native
generics. Change the import to use Iterable from collections.abc (remove
typing.List), then update all type annotations to use built-in generics (e.g.,
list[str] instead of List[str], tuple[int, str] and Iterable[str] using the
modern syntax). Ensure any "typing.Tuple"/"typing.List" references are converted
to tuple[...] and list[...] and remove unused typing imports.
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
