## In docs/api/anomaly-radar-slo-budget.md around lines 1–15, add a "Versioning &

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583584

- [review_comment] 2025-09-16T03:22:08Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:15)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 1–15, add a "Versioning &
Deprecation" section that declares supported API versions (e.g., v1), the
compatibility guarantees (minor/patch compatibility, no silent breaking
changes), the breaking-change policy (how breaking changes are evaluated and
approved), the deprecation timeline (minimum 90 days notice before removal), the
changelog/release process (where changes are recorded and how releases are
communicated), and concise migration guidance for clients (examples of typical
migration steps and links to relevant types like SLOConfig, BurnRateThresholds,
AnomalyThresholds, Alert, MetricSnapshot); also add the same section to the
central API docs file (docs/api/_index.md or the repository’s central API docs
entry) so the policy is discoverable project-wide, and ensure any references to
routes (internal/anomaly-radar-slo-budget/handlers.go RegisterRoutes) and types
are linked or cross-referenced for implementer guidance.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | No | No | - | - |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Pending review. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:598
>
> **Alternatives Considered**
> Not documented.
>
> **Lesson(s) Learned**
> None recorded.
