## In deployments/scripts/deploy-rbac-staging.sh around line 211, the unquoted

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578102

- [review_comment] 2025-09-16T03:19:59Z by coderabbitai[bot] (deployments/scripts/deploy-rbac-staging.sh:196)

```text
In deployments/scripts/deploy-rbac-staging.sh around line 211, the unquoted
variable in the kill invocation can cause word-splitting or glob expansion
issues; update the command to quote the variable (e.g., kill
"$PORT_FORWARD_PID") and optionally guard against empty values (e.g., test for
non-empty before kill) to satisfy ShellCheck and avoid runtime failures.
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
