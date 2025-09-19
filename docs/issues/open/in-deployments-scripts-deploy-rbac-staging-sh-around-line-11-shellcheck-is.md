## In deployments/scripts/deploy-rbac-staging.sh around line 11, ShellCheck is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360038953

- [review_comment] 2025-09-18T15:56:33Z by coderabbitai[bot] (deployments/scripts/deploy-rbac-staging.sh:11)

```text
In deployments/scripts/deploy-rbac-staging.sh around line 11, ShellCheck is
warning about sourcing a local library; add a ShellCheck source directive
immediately above the source line to point to the lib path (for example: "#
shellcheck source=lib/logging.sh") so ShellCheck knows where the sourced file
lives, then keep the existing source "${SCRIPT_DIR}/lib/logging.sh" line
unchanged.
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
