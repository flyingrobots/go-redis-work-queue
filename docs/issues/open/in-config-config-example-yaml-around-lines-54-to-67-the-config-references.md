## In config/config.example.yaml around lines 54 to 67, the config references

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856184

- [review_comment] 2025-09-16T23:20:21Z by coderabbitai[bot] (config/config.example.yaml:67)

```text
In config/config.example.yaml around lines 54 to 67, the config references
"{tenant}" in key_pattern/hash_key_pattern but never defines how "tenant" is
derived or configured; update the example and comments to explicitly define
"tenant" (e.g., per-application tenant ID, header-derived value, or environment
variable), show the exact configuration option name used to set it (or how to
derive it from request headers/metadata), and clarify its format/constraints;
also add a short note pointing to the relevant docs page (insert documentation
link placeholder) and add a docs cross-reference (README or operator guide) that
explains tenant resolution, recommended defaults, and examples for single-tenant
vs multi-tenant usage.
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
