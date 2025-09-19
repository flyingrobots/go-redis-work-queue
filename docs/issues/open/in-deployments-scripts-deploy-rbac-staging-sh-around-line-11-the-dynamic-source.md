## In deployments/scripts/deploy-rbac-staging.sh around line 11, the dynamic source

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066829

- [review_comment] 2025-09-18T16:02:31Z by coderabbitai[bot] (deployments/scripts/deploy-rbac-staging.sh:11)

```text
In deployments/scripts/deploy-rbac-staging.sh around line 11, the dynamic source
of "${SCRIPT_DIR}/lib/logging.sh" triggers ShellCheck SC1091; add a ShellCheck
source hint comment immediately above the source line to point to the actual
path of the file in the repo (for example: # shellcheck
source=deployments/scripts/lib/logging.sh) so ShellCheck can resolve it, then
keep the existing source "${SCRIPT_DIR}/lib/logging.sh" unchanged.
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
