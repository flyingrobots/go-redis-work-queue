## In deployments/scripts/setup-monitoring.sh around lines 146 to 218, the script

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039058

- [review_comment] 2025-09-18T15:56:34Z by coderabbitai[bot] (deployments/scripts/setup-monitoring.sh:218)

```text
In deployments/scripts/setup-monitoring.sh around lines 146 to 218, the script
checks for the secret name "alertmanager-main" but creates
"alertmanager-rbac-config", so detection and creation refer to different
resources; make them consistent by changing the kubectl get secret check to look
for "alertmanager-rbac-config" (and update any log messages if needed), or
alternatively change the created secret name to "alertmanager-main" so both
lines reference the same secret; modify only the secret name in the detection or
creation block to match the other and keep log messages aligned.
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
