## In deployments/scripts/setup-monitoring.sh around lines 225-226, the kill call

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683107

- [review_comment] 2025-09-16T14:20:39Z by coderabbitai[bot] (deployments/scripts/setup-monitoring.sh:226)

```text
In deployments/scripts/setup-monitoring.sh around lines 225-226, the kill call
uses an unquoted variable which can break if the PID contains spaces or is
empty; change it to quote the variable (use kill "$port_forward_pid" 2>/dev/null
|| true) so the PID is passed safely and to avoid word-splitting or globbing.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | No | No | - | - |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Pending review. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:1148
>
> **Alternatives Considered**
> Not documented.
>
> **Lesson(s) Learned**
> None recorded.
