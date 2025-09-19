## In deployments/scripts/test-staging-deployment.sh around lines 278 to 287,

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033256

- [review_comment] 2025-09-18T15:55:20Z by coderabbitai[bot] (deployments/scripts/test-staging-deployment.sh:287)

```text
In deployments/scripts/test-staging-deployment.sh around lines 278 to 287,
remove the blind sleep and instead poll the TCP socket (or HTTP endpoint) until
it becomes available, honoring a TIMEOUT environment variable to avoid hangs;
implement a loop that repeatedly attempts to connect (e.g., with curl -sSf or a
simple /dev/tcp check or nc) with short sleeps between tries and aborts with a
non-zero exit if the timeout is reached, then proceed to the health endpoint
test only after the socket/HTTP check succeeds.
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
