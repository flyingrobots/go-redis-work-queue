## In deployments/kubernetes/rbac-token-service-deployment.yaml around lines 367 to

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360038937

- [review_comment] 2025-09-18T15:56:32Z by coderabbitai[bot] (deployments/kubernetes/rbac-token-service-deployment.yaml:371)

```text
In deployments/kubernetes/rbac-token-service-deployment.yaml around lines 367 to
371, the annotations using nginx.ingress.kubernetes.io/rate-limit* are invalid;
replace them with the real NGINX ingress annotations
nginx.ingress.kubernetes.io/limit-rps and
nginx.ingress.kubernetes.io/limit-burst and set values equivalent to the
previous intent (e.g., for 60 requests per minute use
nginx.ingress.kubernetes.io/limit-rps: "1" and
nginx.ingress.kubernetes.io/limit-burst: "60"), keeping the other annotations
(rewrite-target, ssl-redirect, cert-manager) unchanged.
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
