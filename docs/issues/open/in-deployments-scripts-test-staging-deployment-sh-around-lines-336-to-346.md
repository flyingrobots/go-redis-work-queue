## In deployments/scripts/test-staging-deployment.sh around lines 336 to 346,

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033270

- [review_comment] 2025-09-18T15:55:20Z by coderabbitai[bot] (deployments/scripts/test-staging-deployment.sh:346)

```text
In deployments/scripts/test-staging-deployment.sh around lines 336 to 346,
replace the blind "sleep 5" before fetching the bootstrap token with the same
readiness loop used in the RBAC tests: poll kubectl (with a timeout and short
sleep interval) until the rbac-secrets secret (or its admin-bootstrap-token
field) exists and is readable, then proceed to read and base64-decode the token;
ensure the loop exits with a clear error if the secret never appears within the
timeout so the script fails fast instead of waiting blindly.
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
