## In docs/api/capacity-planning-api.md around lines 479 to 494, the code

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912898

- [review_comment] 2025-09-18T12:12:39Z by coderabbitai[bot] (docs/api/capacity-planning-api.md:494)

```text
In docs/api/capacity-planning-api.md around lines 479 to 494, the code
auto-applies scaling whenever Plan.Confidence >= config.ConfidenceThreshold
without checking whether the proposed plan will keep SLOs met; add an SLO gate
before auto-apply by computing/consulting a predicted SLO compliance check
(e.g., predictedSLOCompliant(response.Plan) or using
response.Plan.PredictedSLOCompliance) and only call applyScalingPlan when both
confidence >= threshold AND predicted SLO compliance is true; also log a clear
message when auto-apply is skipped due to SLO risk and surface the predicted SLO
metrics in the log for operators to inspect.
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
