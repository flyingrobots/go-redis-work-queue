## In docs/api/calendar-view.md around lines 519 to 548, the error examples and

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033324

- [review_comment] 2025-09-18T15:55:21Z by coderabbitai[bot] (docs/api/calendar-view.md:548)

```text
In docs/api/calendar-view.md around lines 519 to 548, the error examples and
table mix numeric "code" values with string-style error identifiers; update the
documentation to use stable string error codes everywhere (e.g.
"ErrorCodeEventNotFound") instead of numeric codes in examples and the table,
keep numeric codes as internal implementation details only, and add a note/link
to a new docs/error_codes.md that lists the numeric->string mapping for humans;
ensure the JSON example uses "error_code" (string) consistently, update any
schema/response examples in this section to reflect the "error_code" field and
remove the numeric "code" field, and mention that services should still
translate to numeric codes internally but expose string codes in the public API
docs.
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
