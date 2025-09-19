## In deployments/scripts/deploy-rbac-staging.sh around lines 40-47 (and also apply

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066843

- [review_comment] 2025-09-18T16:02:31Z by coderabbitai[bot] (deployments/scripts/deploy-rbac-staging.sh:47)

```text
In deployments/scripts/deploy-rbac-staging.sh around lines 40-47 (and also apply
same fix to 116-123 and 126-132), you build a Docker image locally but never
push or load it so a remote cluster will get ImagePullBackOff; either push the
built image to a registry and ensure rbac-token-service-deployment.yaml
references the exact same image name/tag (use $IMAGE_NAME consistently and push
tags like :staging), or if targeting a local cluster (kind/minikube) replace the
push step with loading the image into the cluster via kind load docker-image
"$IMAGE_NAME" or minikube image load "$IMAGE_NAME"; pick one approach and make
the script perform the corresponding push or load and keep the deployment
YAML/image variable consistent.
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
