## In docs/api/dlq-remediation-pipeline.md around lines 168-176, the example

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856251

- [review_comment] 2025-09-16T23:20:24Z by coderabbitai[bot] (docs/api/dlq-remediation-pipeline.md:176)

```text
In docs/api/dlq-remediation-pipeline.md around lines 168-176, the example
response uses stringified duration values (e.g., "125ms") and is missing
idempotency guidance for write endpoints; replace duration strings with a
numeric duration_ms field (integer milliseconds) in all response examples
(including the other occurrences at 465-472 and 898-909) and update the POST
/pipeline/process-batch documentation to require/describe the Idempotency-Key
header on writes (explain header name, purpose, and that identical keys prevent
duplicate processing). Ensure examples and schema use duration_ms integers
consistently and add a short sentence in the process-batch endpoint docs
clarifying idempotency behavior.
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
