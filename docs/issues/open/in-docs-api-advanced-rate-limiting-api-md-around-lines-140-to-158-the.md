## In docs/api/advanced-rate-limiting-api.md around lines 140 to 158, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912558

- [review_comment] 2025-09-18T12:12:36Z by coderabbitai[bot] (docs/api/advanced-rate-limiting-api.md:158)

```text
In docs/api/advanced-rate-limiting-api.md around lines 140 to 158, the
MinGuaranteedShare description and guardrail are contradictory; adopt Model B
(weight‑proportional minima): rename or clarify the field as a total
MinGuaranteedBudget (0.0–1.0) that is distributed per priority as minimum_i =
(weight_i / Σweights) * MinGuaranteedBudget, update the guardrail to require
MinGuaranteedBudget ≤ 1.0, document the per‑priority calculation, and keep the
existing renormalisation behavior and warning log semantics (cap values at 1.0
and clamp negatives to 0) so operators know what will happen if the budget
exceeds capacity.
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
