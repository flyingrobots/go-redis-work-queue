## In .markdownlint.yaml around lines 10 to 13, the MD026 punctuation list includes

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353792378

- [review_comment] 2025-09-16T22:30:41Z by coderabbitai[bot] (.markdownlint.yaml:13)

```text
In .markdownlint.yaml around lines 10 to 13, the MD026 punctuation list includes
".,;:!" but omits "?", so either add the question mark to the allowed
punctuation string or explicitly document that question marks are intentionally
banned; to allow question marks update the punctuation value to include "?"
(e.g., add ? to the string), or if you intend to ban them, add a clarifying
comment above MD026 stating that "?" is intentionally excluded.
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
