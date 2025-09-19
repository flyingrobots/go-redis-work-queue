## In deployments/docker/rbac-configs/roles.yaml around lines 107 to 116, the RBAC

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066723

- [review_comment] 2025-09-18T16:02:30Z by coderabbitai[bot] (deployments/docker/rbac-configs/roles.yaml:116)

```text
In deployments/docker/rbac-configs/roles.yaml around lines 107 to 116, the RBAC
role assignment comments are vague about who enforces rules, when they are
applied, and the precedence; add a clear documentation block named
role_assignment_rules immediately above or beside the existing domain rules that
states: Enforced by: RBAC Token Service during token issuance; Precedence:
explicit user assignment > group membership > domain rules > default_role;
Evaluation: token email domain matched against rules in order. Ensure the block
is concise, YAML-comment friendly (or a YAML mapping if you prefer
machine-readable docs), and placed where maintainers reading the file will see
it before the domain entries.
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
