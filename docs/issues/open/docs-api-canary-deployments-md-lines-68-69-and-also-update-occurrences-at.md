## docs/api/canary-deployments.md lines 68-69 (and also update occurrences at

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912737

- [review_comment] 2025-09-18T12:12:38Z by coderabbitai[bot] (docs/api/canary-deployments.md:71)

```text
docs/api/canary-deployments.md lines 68-69 (and also update occurrences at
366-371, 373-378, 779-785, 560-566, 594-595): the spec currently declares ULIDs
(uppercase, 26 chars) but examples use a prefixed UUID ("canary_..."), causing a
contradiction; pick one format and make amendments: either (A) adopt plain
uppercase 26-char ULIDs everywhere — remove the "canary_" prefix from all
example IDs and ensure any descriptive text and regex examples reflect 26
uppercase ULID characters, or (B) keep the "canary_" prefix — update the spec
text to state "prefix + ULID" and adjust any regex/validation examples to accept
the literal prefix followed by a 26-char uppercase ULID; apply the chosen change
consistently to the listed line ranges and any related ID examples in the
document.
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
