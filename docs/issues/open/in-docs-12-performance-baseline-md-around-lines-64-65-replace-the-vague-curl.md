## In docs/12_performance_baseline.md around lines 64-65, replace the vague "curl

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360067080

- [review_comment] 2025-09-18T16:02:34Z by coderabbitai[bot] (docs/12_performance_baseline.md:65)

```text
In docs/12_performance_baseline.md around lines 64-65, replace the vague "curl
/metrics" note with a concrete one-liner that sets METRICS_URL to the binary's
default metrics address and shows a copy-pasteable curl that saves metrics to a
timestamped file; to do this, inspect the binary's flag parsing to determine the
actual default for --metrics-addr and use that host:port in the METRICS_URL
default (replace 9091 if the code's default is different), then add the two
lines: METRICS_URL=${METRICS_URL:-http://<actual-default>/metrics}  # set to
your --metrics-addr and curl -fsSL "$METRICS_URL" | tee "metrics_$(date
+%s).prom".
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
