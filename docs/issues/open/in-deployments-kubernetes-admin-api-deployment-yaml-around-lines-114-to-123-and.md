## In deployments/kubernetes/admin-api-deployment.yaml around lines 114 to 123 (and

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033200

- [review_comment] 2025-09-18T15:55:19Z by coderabbitai[bot] (deployments/kubernetes/admin-api-deployment.yaml:123)

```text
In deployments/kubernetes/admin-api-deployment.yaml around lines 114 to 123 (and
also lines 125-126), currently API tokens are injected as environment variables
which leaks secrets; replace those env entries with a projected/secret volume:
define a volume that sources the admin-api-secrets secret, mount it into the
container at /var/run/secrets/admin-api, remove the API_TOKEN_1/API_TOKEN_2 env
entries, and update the app startup/config to read tokens/passwords from the
files /var/run/secrets/admin-api/api-token-1 and
/var/run/secrets/admin-api/api-token-2 instead of from environment variables.
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
