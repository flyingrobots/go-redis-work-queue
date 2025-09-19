## .github/workflows/goreleaser.yml around lines 39 to 45: the workflow uses

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353724341

- [review_comment] 2025-09-16T21:42:54Z by coderabbitai[bot] (.github/workflows/goreleaser.yml:45)

```text
.github/workflows/goreleaser.yml around lines 39 to 45: the workflow uses
goreleaser action but doesn’t grant OIDC permission for keyless signing or
provenance/SBOM emission; add repository permissions including "id-token: write"
(and any other required permissions for writing artifacts/provenance if
applicable) in the workflow YAML, and ensure your .goreleaser.yaml has
SBOM/provenance and signing enabled if you intend to produce/sign
provenance/SBOMs; if you aren’t signing/emitting provenance, no change to
permissions is required.
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
