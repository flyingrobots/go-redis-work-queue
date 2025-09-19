## In .goreleaser.yaml around lines 8 to 13, the build configuration lacks

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353792367

- [review_comment] 2025-09-16T22:30:41Z by coderabbitai[bot] (.goreleaser.yaml:13)

```text
In .goreleaser.yaml around lines 8 to 13, the build configuration lacks
reproducibility flags; add the -trimpath flag to the Go ldflags and enable
mod_timestamp (set to a fixed value like 0) in the goreleaser build
configuration so that file paths are trimmed from binaries and timestamps are
stamped consistently across builds; update the ldflags entry to include
-trimpath and add the mod_timestamp setting at the appropriate builds/archives
level.
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
