## In deployments/scripts/test-staging-deployment.sh around lines 278 to 280,

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360067025

- [review_comment] 2025-09-18T16:02:33Z by coderabbitai[bot] (deployments/scripts/test-staging-deployment.sh:280)

```text
In deployments/scripts/test-staging-deployment.sh around lines 278 to 280,
remove the blind "sleep 5" and replace it with a socket polling loop that
repeatedly checks the forwarded local port until it accepts connections or a
configurable timeout is reached; implement the loop using a lightweight check
(e.g. nc -z, bash /dev/tcp/host/port, or timeout+curl) with short sleeps between
attempts, fail the script with a clear error if the port never becomes available
within the timeout, and only proceed to the HTTP checks once the socket is
confirmed open.
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
