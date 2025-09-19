## In deployments/kubernetes/rbac-token-service-deployment.yaml around lines

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066805

- [review_comment] 2025-09-18T16:02:31Z by coderabbitai[bot] (deployments/kubernetes/rbac-token-service-deployment.yaml:151)

```text
In deployments/kubernetes/rbac-token-service-deployment.yaml around lines
146–151 the CORS allowed_origins are hardcoded to staging/prod domains; replace
this with a parameterized source by loading origins from an environment variable
or ConfigMap (e.g., ORIGINS_CSV) or via your Helm/template values, have the app
parse the CSV into the allowed_origins array at startup, provide a sensible
default/fallback and document how to set the env/config, and ensure
allowed_methods and other CORS fields remain populated from the same
configurable source.
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
