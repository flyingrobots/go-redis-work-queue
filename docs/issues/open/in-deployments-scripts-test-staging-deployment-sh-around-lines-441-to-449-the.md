## In deployments/scripts/test-staging-deployment.sh around lines 441 to 449, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061179

- [review_comment] 2025-09-18T16:01:16Z by coderabbitai[bot] (deployments/scripts/test-staging-deployment.sh:449)

```text
In deployments/scripts/test-staging-deployment.sh around lines 441 to 449, the
monitoring namespace is hardcoded to "monitoring"; change this to a parameter by
adding a -m/--monitoring-ns option to the script's argument parsing (matching
the style used in setup-monitoring.sh), introduce a local variable (e.g.,
monitoring_ns) set from that option with a default of "monitoring", update the
help/usage text, and ensure all checks and add_test_result calls use that
variable so existing behavior stays the same when the flag is not provided.
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
