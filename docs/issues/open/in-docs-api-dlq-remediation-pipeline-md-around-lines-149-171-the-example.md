## In docs/api/dlq-remediation-pipeline.md around lines 149-171, the example

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683113

- [review_comment] 2025-09-16T14:20:39Z by coderabbitai[bot] (docs/api/dlq-remediation-pipeline.md:171)

```text
In docs/api/dlq-remediation-pipeline.md around lines 149-171, the example
response uses a string duration and lacks a clear dry-run and idempotency
contract; change the response to expose duration_ms as an integer (milliseconds)
instead of a string "125ms", explicitly state dry_run is a boolean that
guarantees no state changes when true, and update the POST
/pipeline/process-batch docs to add an Idempotency-Key header (string, optional
but required for at‑least‑once safe retries) and a semantics note that requests
with the same Idempotency-Key must return the original 200 response with an
identical body for 24 hours to prevent duplicate execution.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | No | No | - | - |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Pending review. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:1163
>
> **Alternatives Considered**
> Not documented.
>
> **Lesson(s) Learned**
> None recorded.
