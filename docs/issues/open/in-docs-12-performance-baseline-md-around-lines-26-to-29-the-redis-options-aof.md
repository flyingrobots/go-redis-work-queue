## In docs/12_performance_baseline.md around lines 26 to 29, the Redis options (AOF

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360072335

- [review_comment] 2025-09-18T16:03:44Z by coderabbitai[bot] (docs/12_performance_baseline.md:29)

```text
In docs/12_performance_baseline.md around lines 26 to 29, the Redis options (AOF
disabled, noeviction, tcp-keepalive=60) are asserted but not actually applied by
the documented docker run; update the docs to show a docker run (or
docker-compose) invocation that explicitly passes those Redis configuration
options to the container (or mounts a redis.conf) so they are enforced
reproducibly — specifically ensure AOF is disabled (appendonly no),
maxmemory-policy is set to noeviction, and tcp-keepalive is set to 60 in the
command or config file referenced by the doc.
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
