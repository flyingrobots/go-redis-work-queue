## In deployments/admin-api/k8s-deployment.yaml around lines 41 to 44, the inline

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360072299

- [review_comment] 2025-09-18T16:03:44Z by coderabbitai[bot] (deployments/admin-api/k8s-deployment.yaml:44)

```text
In deployments/admin-api/k8s-deployment.yaml around lines 41 to 44, the inline
comments after the empty secret values do not have two spaces before the “# …”
which fails the linter; update each line so there are exactly two spaces between
the value and the inline comment (e.g., change '"" #' to '""  #' for both
jwt-secret and redis-password) and save the file.
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
