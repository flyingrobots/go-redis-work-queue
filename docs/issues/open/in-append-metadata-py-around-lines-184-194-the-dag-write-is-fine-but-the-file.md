## In append_metadata.py around lines 184-194, the DAG write is fine but the file

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360060975

- [review_comment] 2025-09-18T16:01:14Z by coderabbitai[bot] (append_metadata.py:194)

```text
In append_metadata.py around lines 184-194, the DAG write is fine but the file
has critical issues to fix elsewhere: remove the static infrastructure_nodes
list defined at/near line 87 and instead import infrastructure data from
dependency_analysis; replace the brittle front-matter check at/near line 67 (if
content.endswith("---")) with a proper YAML/front-matter parser that
extracts/loads the front-matter block (e.g., locate the leading/trailing '---'
and yaml.safe_load the slice) so edge cases are handled; stop hardcoding spec
paths at/near line 155 by constructing them with os.path.join(ideas_dir, "docs",
"ideas", f"{feature_name}.md") or the correct platform-safe path for your repo
layout; and add normalized imports at the top of the file: from
dependency_analysis import get_normalized_feature_map, normalize_name,
infrastructure (and any other required symbols) so the earlier removals use
those functions/values.
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
