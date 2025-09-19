## In deployments/scripts/setup-monitoring.sh around lines 211 to 215, the secret

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039081

- [review_comment] 2025-09-18T15:56:35Z by coderabbitai[bot] (deployments/scripts/setup-monitoring.sh:215)

```text
In deployments/scripts/setup-monitoring.sh around lines 211 to 215, the secret
is created using --from-literal which leaks the config into process arguments;
instead feed the config via stdin or --from-file reading from /dev/stdin.
Replace the --from-literal usage with a method that pipes the
$alertmanager_config into kubectl (for example using a here-doc or printf |
kubectl and --from-file=alertmanager.yml=/dev/stdin or by constructing the
secret YAML on stdin) so the secret contents do not appear in the process list
or shell args; ensure quoting/encoding is preserved when piping and then
continue to dry-run -o yaml | kubectl apply -f - as before.
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
