## In docs/api/admin-api.md around lines 10-11 (also apply same change at 41-46 and

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358974241

- [review_comment] 2025-09-18T12:25:37Z by coderabbitai[bot] (docs/api/admin-api.md:11)

```text
In docs/api/admin-api.md around lines 10-11 (also apply same change at 41-46 and
179-186), the phrase "Double Confirmation" is misleading because it implies two
inputs when the implementation uses distinct confirmation phrases per endpoint;
update the wording to explicitly state that each dangerous endpoint requires a
unique confirmation phrase (or, if intended, change the API docs to show two
separate confirmation fields), and revise the examples to either list the
distinct phrase required per endpoint or show two confirmation fields so
operators are not confused during incidents.
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
