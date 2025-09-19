## In deployments/admin-api/k8s-deployment.yaml around lines 152–167, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360038867

- [review_comment] 2025-09-18T15:56:32Z by coderabbitai[bot] (deployments/admin-api/k8s-deployment.yaml:167)

```text
In deployments/admin-api/k8s-deployment.yaml around lines 152–167, the
annotation nginx.ingress.kubernetes.io/rate-limit is non‑standard and will be
ignored; replace it with a documented ingress‑nginx annotation such as
nginx.ingress.kubernetes.io/limit-rps: "100" (or
nginx.ingress.kubernetes.io/limit-rpm: "6000" if you prefer per‑minute limits)
and remove any non‑standard keys like
nginx.ingress.kubernetes.io/rate-limit-window (or map it to an appropriate
documented setting if needed), then redeploy and validate that the controller
enforces the configured limits; apply the same replacement for the other files
referenced in the review.
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
