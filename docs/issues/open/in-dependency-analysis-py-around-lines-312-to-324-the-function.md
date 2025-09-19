## In dependency_analysis.py around lines 312 to 324, the function

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033136

- [review_comment] 2025-09-18T15:55:19Z by coderabbitai[bot] (dependency_analysis.py:324)

```text
In dependency_analysis.py around lines 312 to 324, the function
get_normalized_feature_map() currently declares a return type of dict[str,
dict[str, list[str]]], but the payload includes original_name: str (a plain
string) causing a type mismatch; update the typing to accurately reflect the
shape (preferably define a TypedDict like FeatureNormalized { original_name:
str; hard: list[str]; soft: list[str]; enables: list[str]; provides: list[str] }
and change the function annotation to dict[str, FeatureNormalized]) and keep the
payload as-is (ensuring provides is a list[str]) or alternatively change
original_name to a single-element list if you prefer lists-only—apply one of
these fixes and update any imports/aliases accordingly.
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
