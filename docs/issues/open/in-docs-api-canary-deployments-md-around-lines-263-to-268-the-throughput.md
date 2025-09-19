## In docs/api/canary-deployments.md around lines 263 to 268, the throughput

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912789

- [review_comment] 2025-09-18T12:12:38Z by coderabbitai[bot] (docs/api/canary-deployments.md:270)

```text
In docs/api/canary-deployments.md around lines 263 to 268, the throughput
message mixes English and math by saying "Throughput decrease: -5.1%"; update
the message generation so the sign and wording match: either format as
"Throughput decrease: 5.1%" (remove the negative sign when using the word
"decrease") or change the label to a neutral term like "Throughput change:
-5.1%" (keep the negative sign). Modify the template or formatting logic
accordingly so negative values drop the minus when using "decrease" or retain
the minus when using "change".
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
