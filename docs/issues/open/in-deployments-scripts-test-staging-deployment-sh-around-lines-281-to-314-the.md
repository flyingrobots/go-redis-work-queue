## In deployments/scripts/test-staging-deployment.sh around lines 281 to 314, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360067037

- [review_comment] 2025-09-18T16:02:33Z by coderabbitai[bot] (deployments/scripts/test-staging-deployment.sh:314)

```text
In deployments/scripts/test-staging-deployment.sh around lines 281 to 314, the
timeout command uses a hard-coded 30 seconds in two places; replace those
literals with the configured TIMEOUT variable (e.g. timeout "$TIMEOUT" ...) so
the script honors the configured timeout consistently for both health and
metrics checks, ensuring proper quoting of the variable and keeping the same
bash -c loop structure.
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
