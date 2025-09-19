## In docs/13_release_versioning.md around lines 27 to 32, the instructions

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814715

- [review_comment] 2025-09-16T22:46:57Z by coderabbitai[bot] (docs/13_release_versioning.md:32)

```text
In docs/13_release_versioning.md around lines 27 to 32, the instructions
currently create annotated tags and push them but should prefer signed tags,
avoid lightweight tags, and ensure exactly one ref is pushed; update the example
to use git tag -s for signed tags (with a descriptive "release: vX.Y.Z[-pre]"
message), add a commented annotated-tag fallback for CI environments that cannot
sign, and ensure git push references the exact tag name (push one ref) so the
documentation shows using signed tags by default and the annotated fallback as a
comment.
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
