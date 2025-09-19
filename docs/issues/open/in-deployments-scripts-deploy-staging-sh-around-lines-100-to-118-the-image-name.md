## In deployments/scripts/deploy-staging.sh around lines 100 to 118, the IMAGE_NAME

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066877

- [review_comment] 2025-09-18T16:02:31Z by coderabbitai[bot] (deployments/scripts/deploy-staging.sh:118)

```text
In deployments/scripts/deploy-staging.sh around lines 100 to 118, the IMAGE_NAME
is built without a required registry namespace which can yield
docker.io/<app>:tag (the implicit "library" namespace) and cause push failures;
validate that DOCKER_NAMESPACE is set (fail fast if missing), construct a
fully-qualified IMAGE_NAME combining REGISTRY (trim trailing slash) +
DOCKER_NAMESPACE + APP_NAME + IMAGE_TAG, and use that canonical IMAGE_NAME
everywhere (build, push, and later kubectl set image) rather than embedding
"$REGISTRY/$APP_NAME:$IMAGE_TAG" inline; also audit
deployments/admin-api/deploy.sh and CI workflows (.github/workflows/*) to
standardize on REGISTRY + DOCKER_NAMESPACE + APP_NAME and add the same
validation where applicable.
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
