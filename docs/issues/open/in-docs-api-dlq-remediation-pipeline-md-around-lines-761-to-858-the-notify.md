## In docs/api/dlq-remediation-pipeline.md around lines 761 to 858, the notify

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679617

- [review_comment] 2025-09-16T14:19:28Z by coderabbitai[bot] (docs/api/dlq-remediation-pipeline.md:927)

```text
In docs/api/dlq-remediation-pipeline.md around lines 761 to 858, the notify
action examples and description omit operational safeguards for outbound egress;
update the docs to define an explicit outbound allowlist for notification
destinations, default and configurable network timeouts, retry policy (attempt
count, backoff), a notification DLQ/escape hatch for failed deliveries, and
clear semantics for partial failures (e.g., per-channel success/failure
reporting, rollback/no-rollback behavior, and how the pipeline marks the job).
Also include recommended defaults and configuration keys, and an example showing
how to configure allowlist, timeouts, retries, and DLQ behavior.
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
