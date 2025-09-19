## In docs/api/capacity-planning-api.md around lines 631-640, the complexity table

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912932

- [review_comment] 2025-09-18T12:12:39Z by coderabbitai[bot] (docs/api/capacity-planning-api.md:640)

```text
In docs/api/capacity-planning-api.md around lines 631-640, the complexity table
is too vague: update the table and surrounding text to list practical caps and
numerical-stability guards (e.g., for M/M/c state that complexity is O(min(c,
C_MAX)) and document a configurable cap C_MAX and checks to avoid numerical
instability when c is large), clarify Holt-Winters complexity as O(n * k * it)
or O(n * s) by specifying per-iteration constants (k = number of seasonal
components or smoothing parameters and it = number of iterations) and any
early‑stop/regularization applied, and add brief notes for Simulation and
Pattern Extraction about applied caps or downsampling (safeguards like max
steps, max history, or sampling) so the table reflects real-world limits rather
than idealized O()s.
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
