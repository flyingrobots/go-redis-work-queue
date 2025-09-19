## In .github/workflows/goreleaser.yml around lines 28-33, the workflow logs into

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567149

- [review_comment] 2025-09-16T03:15:01Z by coderabbitai[bot] (.github/workflows/goreleaser.yml:37)

```text
In .github/workflows/goreleaser.yml around lines 28-33, the workflow logs into
GHCR but does not set up QEMU or Docker Buildx for multi-arch builds; add steps
before the login/build steps to (1) register QEMU emulators (use
actions/setup-qemu-action@v2) and (2) create/enable a buildx builder (use
docker/setup-buildx-action@v2), ensuring buildx is the active builder and
supports the target platforms; keep the login step but then invoke buildx-based
multi-platform build/push (or ensure goreleaser step uses buildx) so multi-arch
images are built correctly.
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
