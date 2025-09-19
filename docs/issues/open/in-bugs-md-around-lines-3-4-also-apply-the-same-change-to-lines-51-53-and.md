## In BUGS.md around lines 3-4 (also apply the same change to lines 51-53 and

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360032989

- [review_comment] 2025-09-18T15:55:17Z by coderabbitai[bot] (BUGS.md:4)

```text
In BUGS.md around lines 3-4 (also apply the same change to lines 51-53 and
71-72), the intro contradicts itself by claiming the heartbeat/reaper will SCAN
jobqueue:worker:*:processing keys while elsewhere advising readers to avoid
using KEYS/SCAN; update the text to be consistent with the registry +
single-slot guidance by removing any endorsement of SCAN/KEYS and instead
describing that the heartbeat/reaper uses the registry and single-slot ownership
to locate processing keys (or explain that any reaper must be implemented via
the registry and safe single-slot semantics), and rephrase the listed items to
reflect that SCAN/KEYS are discouraged and not part of the recommended
implementation.
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
