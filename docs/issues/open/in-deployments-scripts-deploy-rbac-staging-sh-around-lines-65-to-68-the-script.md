## In deployments/scripts/deploy-rbac-staging.sh around lines 65 to 68, the script

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360038984

- [review_comment] 2025-09-18T15:56:33Z by coderabbitai[bot] (deployments/scripts/deploy-rbac-staging.sh:68)

```text
In deployments/scripts/deploy-rbac-staging.sh around lines 65 to 68, the script
currently uses "export RBAC_SIGNING_KEY RBAC_ENCRYPTION_KEY REDIS_PASSWORD
ADMIN_BOOTSTRAP_TOKEN" which unnecessarily places secrets into the environment
and risks leaking them; change these to plain shell variables (assign them
without export) so they remain in-script only (e.g., RBAC_SIGNING_KEY="..."
RBAC_ENCRYPTION_KEY="..." REDIS_PASSWORD="..." ADMIN_BOOTSTRAP_TOKEN="..."),
remove the export statement, and ensure no subsequent commands rely on these
variables being inherited by child processes; optionally unset the variables
before script exit for extra safety.
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
