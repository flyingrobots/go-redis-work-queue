## In deployments/scripts/deploy-rbac-staging.sh around lines 36 to 47, the build

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360038975

- [review_comment] 2025-09-18T15:56:33Z by coderabbitai[bot] (deployments/scripts/deploy-rbac-staging.sh:47)

```text
In deployments/scripts/deploy-rbac-staging.sh around lines 36 to 47, the build
step creates and tags a local image work-queue/rbac-token-service:staging which
never gets pushed and does not match the manifest's pinned image; add
IMAGE_REPO, IMAGE_TAG and IMAGE variables near the top (after line 7) to match
the Deployment, update build_image to build with -t "$IMAGE", push the image to
the registry (docker push "$IMAGE"), and remove or stop using the local-only
staging tag so the deployed manifest pulls the exact pushed tag.
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
