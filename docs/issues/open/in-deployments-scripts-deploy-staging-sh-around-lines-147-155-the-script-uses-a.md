## In deployments/scripts/deploy-staging.sh around lines 147-155, the script uses a

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066887

- [review_comment] 2025-09-18T16:02:32Z by coderabbitai[bot] (deployments/scripts/deploy-staging.sh:155)

```text
In deployments/scripts/deploy-staging.sh around lines 147-155, the script uses a
direct if [ $? -ne 0 ] conditional to check rollout status; replace this with an
explicit capture of the command's exit code immediately after running the
rollout/status command (e.g., run kubectl rollout status ... --timeout=... and
store its exit code in a variable), then test that variable (if [ "$status" -ne
0 ]) to decide logging, calling rollback, and exiting; ensure the error log
includes the captured exit code or command output for clarity and use that same
code in exit so the caller sees the actual failure code.
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
