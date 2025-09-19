## In docs/14_ops_runbook.md around lines 24 to 30, the local build example

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856245

- [review_comment] 2025-09-16T23:20:23Z by coderabbitai[bot] (docs/14_ops_runbook.md:30)

```text
In docs/14_ops_runbook.md around lines 24 to 30, the local build example
incorrectly uses --push (should use --load for local images) and the GO_VERSION
build-arg is inconsistent with the repository Dockerfiles (root Dockerfile line
3 uses FROM golang:1.23 and does not consume ARG GO_VERSION while go.mod and
other deployment Dockerfiles target Go 1.25); change the example flag from
--push to --load and reconcile the GO_VERSION mismatch by either adding ARG
GO_VERSION to the root Dockerfile and using it in the FROM (e.g., FROM
golang:${GO_VERSION}) so the build-arg takes effect, or update the
docs/build-arg value to the actual base image version used (1.23) so they match.
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
