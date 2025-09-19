## In docs/api/_index.md around lines 7–13, the versioning policy is vague; update

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033294

- [review_comment] 2025-09-18T15:55:20Z by coderabbitai[bot] (docs/api/_index.md:13)

```text
In docs/api/_index.md around lines 7–13, the versioning policy is vague; update
it to explicitly require the following: list required HTTP headers and semantics
— a Sunset header with an absolute RFC‑1123 timestamp, Link headers with
rel="sunset" and rel="deprecation" pointing to the deprecation/remove notices,
and an optional Deprecation header containing version/date; define explicit
LTS/support windows per major (e.g., state “Each major is maintained for 18
months after the next major GA” or replace with your org’s chosen N months),
require that deprecated endpoints include an explicit removal date in both the
API docs and error response bodies once past deprecation, and mandate that all
path examples across docs use the versioned prefix /api/v1/... so route examples
are consistent.
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
