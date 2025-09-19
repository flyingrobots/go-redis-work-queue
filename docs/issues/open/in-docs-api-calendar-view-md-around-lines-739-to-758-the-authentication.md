## In docs/api/calendar-view.md around lines 739 to 758, the authentication

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033336

- [review_comment] 2025-09-18T15:55:21Z by coderabbitai[bot] (docs/api/calendar-view.md:758)

```text
In docs/api/calendar-view.md around lines 739 to 758, the authentication
instructions mix an explicit X-User-ID header with JWT-based identity; clarify
that identity is derived from validated JWT claims and either remove the
X-User-ID example or mark it as internal/testing-only and explicitly ignored
when a valid Authorization: Bearer <jwt> is provided; if you choose to support
both, document that X-User-ID must be validated against the JWT sub/claims on
the server and only accepted when it matches after strict verification.
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
