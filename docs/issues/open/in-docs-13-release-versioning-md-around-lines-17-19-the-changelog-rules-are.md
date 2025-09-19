## In docs/13_release_versioning.md around lines 17–19, the changelog rules are

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814707

- [review_comment] 2025-09-16T22:46:56Z by coderabbitai[bot] (docs/13_release_versioning.md:19)

```text
In docs/13_release_versioning.md around lines 17–19, the changelog rules are
incomplete: update this section to explicitly require Conventional Commits
v1.0.0 semantics (type(scope)!: description / type(scope): description), mandate
allowed types (feat, fix, docs, ci, refactor, test, chore, perf, build), require
optional scope, optional body, and FOOTER/BREAKING CHANGE format for breaking
changes; add mapping rules for generating standardized CHANGELOG.md sections
(Features, Fixes, Docs, CI, Refactor, Tests, Chore, Performance, Build) and
rules for incrementing semver based on types/BREAKING CHANGE, include a short
example commit and breaking-change example, and add enforcement notes to run
commitlint and CI hook to validate commits and produce machine‑readable
changelog output for release tooling.
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
