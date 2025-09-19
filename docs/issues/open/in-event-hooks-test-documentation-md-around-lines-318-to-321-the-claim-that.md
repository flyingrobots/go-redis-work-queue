## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 318 to 321, the claim that

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033449

- [review_comment] 2025-09-18T15:55:22Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:321)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 318 to 321, the claim that
"All tests are deterministic with controlled randomness" needs concrete details:
update the section to state that all randomized tests accept a TEST_SEED
environment variable and show example CI/test commands including the shuffle
flag (e.g., add text like "All randomized tests accept TEST_SEED; CI runs with
-shuffle=on" and an example command showing -shuffle=on), and ensure you mention
where to set TEST_SEED locally and in CI so readers can reproduce deterministic
runs.
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
