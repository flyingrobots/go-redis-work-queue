## In docs/api/advanced-rate-limiting-api.md around lines 365 to 375, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912613

- [review_comment] 2025-09-18T12:12:37Z by coderabbitai[bot] (docs/api/advanced-rate-limiting-api.md:375)

```text
In docs/api/advanced-rate-limiting-api.md around lines 365 to 375, the
"Configure TTLs" best-practice is too vague; replace or augment bullet 5 with a
concrete TTL rule-of-thumb: add a new bullet 5 stating "KeyTTL >=
max(2×RefillInterval, 2×BurstSize/RatePerSecond) and never set below 2× the
longest expected idle gap (to avoid bucket evaporation and cold-start spikes)".
Keep numbering of subsequent items, ensure formatting matches existing bullets,
and keep the language concise.
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
