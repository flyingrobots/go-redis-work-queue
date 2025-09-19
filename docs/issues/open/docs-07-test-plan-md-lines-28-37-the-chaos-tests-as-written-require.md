## docs/07_test_plan.md lines 28–37: the chaos tests as written require

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061205

- [review_comment] 2025-09-18T16:01:16Z by coderabbitai[bot] (docs/07_test_plan.md:37)

```text
docs/07_test_plan.md lines 28–37: the chaos tests as written require
CAP_NET_ADMIN and will fail on GitHub-hosted runners; update the document to (1)
note the CAP_NET_ADMIN requirement and explicitly gate/skip these steps on
hosted runners, (2) provide alternatives and examples: run netem in a privileged
sidecar container with CAP_NET_ADMIN, run the tests on self-hosted runners that
grant the capability, or replace host-level injections with proxy-based tools
(toxiproxy/pumba) that work on hosted runners, and (3) add cleanup and detection
guidance so CI can detect capability absence and automatically skip these steps
while pointing users to the privileged-run instructions.
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
