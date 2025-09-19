## In deployments/scripts/deploy-rbac-staging.sh around lines 13 to 35, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360038970

- [review_comment] 2025-09-18T15:56:33Z by coderabbitai[bot] (deployments/scripts/deploy-rbac-staging.sh:35)

```text
In deployments/scripts/deploy-rbac-staging.sh around lines 13 to 35, the
prerequisite check function currently verifies kubectl and docker but misses
validating required tools openssl and curl; add checks similar to the existing
ones: test command -v openssl and command -v curl, emit an error message and
exit 1 if either is missing, and include them before the cluster connect check
so the script fails fast with clear guidance to install the missing utilities.
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
