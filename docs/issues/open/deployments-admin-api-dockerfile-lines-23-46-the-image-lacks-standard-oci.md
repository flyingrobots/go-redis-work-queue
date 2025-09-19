## deployments/admin-api/Dockerfile lines 23-46: the image lacks standard OCI

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360044829

- [review_comment] 2025-09-18T15:57:46Z by coderabbitai[bot] (deployments/admin-api/Dockerfile:46)

```text
deployments/admin-api/Dockerfile lines 23-46: the image lacks standard OCI
metadata labels; add a LABEL instruction near the top of the Dockerfile
(immediately after the FROM) that sets common OCI labels such as
org.opencontainers.image.title, org.opencontainers.image.description,
org.opencontainers.image.version, org.opencontainers.image.revision (commit
SHA), org.opencontainers.image.created (build timestamp),
org.opencontainers.image.authors, org.opencontainers.image.licenses,
org.opencontainers.image.url/source and org.opencontainers.image.vendor;
implement values via build-time ARGs with sensible defaults (e.g., VERSION,
VCS_REF, BUILD_DATE, MAINTAINER) so CI can inject real values, and keep existing
functionality unchanged.
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
