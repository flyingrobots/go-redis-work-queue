## In deployments/README.md around lines 146-148, the docs mention rate limits but

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572123

- [review_comment] 2025-09-16T03:17:31Z by coderabbitai[bot] (deployments/README.md:167)

```text
In deployments/README.md around lines 146-148, the docs mention rate limits but
omit that /metrics must not be internet-facing; update the notes to state
explicitly that the metrics endpoint must be exposed only via a ClusterIP-only
Service (no Ingress/LoadBalancer) and protected with a NetworkPolicy restricting
access to Prometheus scrape targets, and document that scraping should be done
via a ServiceMonitor or Prometheus scrape config targeting the ClusterIP service
only.
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
