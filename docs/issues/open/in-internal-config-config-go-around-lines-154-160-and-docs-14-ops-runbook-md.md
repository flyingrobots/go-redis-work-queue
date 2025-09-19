## In internal/config/config.go (around lines 154–160) and docs/14_ops_runbook.md

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856247

- [review_comment] 2025-09-16T23:20:23Z by coderabbitai[bot] (docs/14_ops_runbook.md:48)

```text
In internal/config/config.go (around lines 154–160) and docs/14_ops_runbook.md
(lines 44–48): the docs claim double-underscore maps to nested keys but
config.go currently uses strings.NewReplacer(".", "_") with v.AutomaticEnv(),
which only maps dots to single underscores so CIRCUIT_BREAKER__COOLDOWN_PERIOD
will not resolve. Fix by either (A) code change: update the env key replacer to
translate double-underscores back to dots (and optionally also handle
single-underscore mapping) so ENV keys like CIRCUIT_BREAKER__COOLDOWN_PERIOD map
to circuit_breaker.cooldown_period, or (B) docs change: remove the
double-underscore example and explicitly document the actual mapping (e.g.,
WORKER_COUNT → worker.count, REDIS_ADDR → redis.addr) and parsing rules for
booleans/durations; apply one of these fixes and update tests/docs accordingly.
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
