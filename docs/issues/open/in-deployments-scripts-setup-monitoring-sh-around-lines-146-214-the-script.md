## In deployments/scripts/setup-monitoring.sh around lines 146–214, the script

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066980

- [review_comment] 2025-09-18T16:02:33Z by coderabbitai[bot] (deployments/scripts/setup-monitoring.sh:214)

```text
In deployments/scripts/setup-monitoring.sh around lines 146–214, the script
checks for a secret named "alertmanager-main" but creates
"alertmanager-rbac-config", so Prometheus Operator will ignore the config;
change the creation/patch step to create or update the secret name the operator
expects (e.g., create/patch "alertmanager-main" in $MONITORING_NAMESPACE with
the generated alertmanager config) and ensure the secret key matches the
operator’s expected key (replace "alertmanager-rbac-config" with
"alertmanager-main" or vice‑versa consistently, using kubectl create secret
generic alertmanager-main --from-literal=alertmanager.yml="$alertmanager_config"
--dry-run=client -o yaml | kubectl apply -f - or perform a kubectl patch if you
need to update an existing secret).
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
