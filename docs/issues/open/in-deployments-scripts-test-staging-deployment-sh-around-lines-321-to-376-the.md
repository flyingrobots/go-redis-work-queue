## In deployments/scripts/test-staging-deployment.sh around lines 321 to 376, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360067046

- [review_comment] 2025-09-18T16:02:33Z by coderabbitai[bot] (deployments/scripts/test-staging-deployment.sh:376)

```text
In deployments/scripts/test-staging-deployment.sh around lines 321 to 376, the
RBAC tests start kubectl port-forward and then sleep 5 which can race and cause
test flakiness; replace the static sleep with a readiness poll that waits for
the local service to respond (e.g., loop up to a timeout calling curl -sS
--max-time 1 http://localhost:8081/health or the auth validate endpoint and
break when it returns 200/expected body), retrying every 1s and failing after a
configurable timeout; keep the port-forward in background, preserve
cleanup_port_forward/trap, and after the poll proceed with the bootstrap-token
retrieval and token-based tests (mark failure if readiness timeout occurs before
running tests).
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
