## In docs/api/calendar-view.md around lines 465 to 491, the example response leaks

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033314

- [review_comment] 2025-09-18T15:55:21Z by coderabbitai[bot] (docs/api/calendar-view.md:491)

```text
In docs/api/calendar-view.md around lines 465 to 491, the example response leaks
numeric enum values (default_view: 0 and action numeric codes) which are
client‑hostile; update the JSON examples to emit enum names as strings (e.g.,
"default_view": "<EnumName>" and each "action": "<ActionName>") and update any
surrounding text to state that the server accepts and returns string enum names
(while noting the Go SDK may map those names to ints internally). Ensure all
key_bindings.action fields and default_view use the string names throughout the
example and add a short note clarifying server behavior and the Go SDK mapping.
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
