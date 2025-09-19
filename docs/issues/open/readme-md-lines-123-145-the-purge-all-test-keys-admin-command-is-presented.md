## README.md lines 123-145: the "Purge all (test keys)" admin command is presented

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569846

- [review_comment] 2025-09-16T03:16:21Z by coderabbitai[bot] (README.md:145)

```text
README.md lines 123-145: the "Purge all (test keys)" admin command is presented
without sufficient gating; update the README to clearly label this command as
"DEV ONLY" (or similar), and document that it requires an explicit additional
safety flag (e.g., --dev or --force-dev) and the --yes confirmation to run;
include a prominent one-line warning explaining it is destructive and should not
be used in production and show the exact safe invocation (e.g., with both --yes
and the extra dev-only flag) so readers know how to run it deliberately.
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
