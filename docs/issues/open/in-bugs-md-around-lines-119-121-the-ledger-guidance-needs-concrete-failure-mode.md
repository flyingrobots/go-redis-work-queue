## In BUGS.md around lines 119–121, the ledger guidance needs concrete failure-mode

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033024

- [review_comment] 2025-09-18T15:55:17Z by coderabbitai[bot] (BUGS.md:121)

```text
In BUGS.md around lines 119–121, the ledger guidance needs concrete failure-mode
and redaction requirements: update the doc to require emitting ack/history
events to a durable sink (e.g., S3/Kafka) while retaining the existing LREM
procList 1 payload after success; mandate a bounded non‑blocking local channel
(with drop counter/metrics) when the sink is unreachable and explicit
alerts/backpressure so workers fail fast instead of silently dropping history;
specify local fallback must use an atomic appender with daily rotation, gzip
compression, size caps, documented retention policy and retention enforcement,
and list default redactions (JWTs/tokens, emails, PII IDs) plus rotation caps
and scrub rules to guarantee redaction; also call out instrumentation/alerts and
backoff semantics to surface sink outages.
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
