## In docs/13_release_versioning.md around line 23, the phrase "Ensure CI green;

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814709

- [review_comment] 2025-09-16T22:46:57Z by coderabbitai[bot] (docs/13_release_versioning.md:23)

```text
In docs/13_release_versioning.md around line 23, the phrase "Ensure CI green;
govulncheck passes; tests (unit/race/e2e) pass." is vague; replace it with an
explicit list of CI gates to enforce (e.g., status: build passes, lint/format
checks, govulncheck report no findings, unit tests, race-detector tests,
integration/e2e tests, and required approvals). Update the line to enumerate
those checks and any minimum thresholds (e.g., coverage or no vuln findings) and
reference the exact CI job names used in the pipeline so reviewers know which
gates must be green.
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
