## In append_metadata.py around lines 57 to 113, remove the hard-coded

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856153

- [review_comment] 2025-09-16T23:20:20Z by coderabbitai[bot] (append_metadata.py:138)

```text
In append_metadata.py around lines 57 to 113, remove the hard-coded
infrastructure_nodes and instead import the canonical infrastructure list and
normalization helpers from dependency_analysis; build node_map keyed by the
normalized name (e.g., use normalize_name(name)) while storing the original
display name in the node dict, normalize every dependency name before doing
lookups so feature→feature edges are not dropped due to kebab/snake differences,
and replace any manual path concatenation with os.path.join(ideas_dir, ...) when
resolving spec paths.
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
