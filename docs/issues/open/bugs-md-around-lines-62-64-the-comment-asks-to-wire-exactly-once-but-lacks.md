## BUGS.md around lines 62-64: the comment asks to "wire exactly-once" but lacks

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912408

- [review_comment] 2025-09-18T12:12:35Z by coderabbitai[bot] (BUGS.md:124)

```text
BUGS.md around lines 62-64: the comment asks to "wire exactly-once" but lacks
concrete handler integration points and a clear contract; update the doc to
explicitly spell out the handler contract and example steps: 1) derive an
idempotency key from the task, 2) call IdempotencyManager.Begin(ctx, key,
WithTTL) and short-circuit (return no-op) when it reports a duplicate or error,
3) ensure IdempotencyManager.End(ctx, key) is deferred/always-run after Begin
succeeds, 4) perform side-effects only after Begin succeeds and wrap any emitted
events in the outbox transaction/Enqueue so they are persisted atomically with
the handler, 5) on success mark the idempotency entry done, and 6) record
metrics idem_hits, idem_misses and outbox_flush_failures; replace the loose
suggestion with this precise step-by-step handler contract and note where to
increment metrics.
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
