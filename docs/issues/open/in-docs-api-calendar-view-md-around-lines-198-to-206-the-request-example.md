## In docs/api/calendar-view.md around lines 198 to 206, the request example

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061274

- [review_comment] 2025-09-18T16:01:17Z by coderabbitai[bot] (docs/api/calendar-view.md:206)

```text
In docs/api/calendar-view.md around lines 198 to 206, the request example
includes a client-supplied identity (user_id) and header examples reference
X-User-ID; remove the user_id field from the JSON example and remove any
X-User-ID header examples, and update any explanatory text to state that
identity is derived from the bearer token (Authorization: Bearer ...) instead;
apply the same changes to the other affected ranges (lines 225-241 and 666-681).
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
