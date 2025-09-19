## docs/api/anomaly-radar-slo-budget.md around lines 223 to 241: the /status

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039186

- [review_comment] 2025-09-18T15:56:37Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:241)

```text
docs/api/anomaly-radar-slo-budget.md around lines 223 to 241: the /status
payload currently exposes "config" at the top-level of "slo_budget" while
/config uses a nested shape under "slo" -> "thresholds", which will break
clients; either remove "config" from /status or, preferably, change the /status
example to match /config by nesting those fields under "slo": { "thresholds": {
... } } (preserve the same keys and values), update any surrounding text to
reference slo.thresholds instead of slo_budget.config, and ensure timestamp and
other fields remain at the same levels.
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
