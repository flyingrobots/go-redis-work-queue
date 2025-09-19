## In deployments/docker/admin-api.env.example lines 1-4, replace the toy token

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033178

- [review_comment] 2025-09-18T15:55:19Z by coderabbitai[bot] (deployments/docker/admin-api.env.example:4)

```text
In deployments/docker/admin-api.env.example lines 1-4, replace the toy token
placeholders with clear secure-generation guidance and aligned names: require
tokens to be >=32 bytes entropy (provide examples for base64 and hex), mandate
role-prefixed keys (e.g. rq_admin_..., rq_read_), include example generation
commands (openssl rand -base64 32 and openssl rand -hex 32), instruct rotatation
and secure storage, and update placeholders to non-trivial examples; ensure the
env var names (API_TOKEN_1/API_TOKEN_2) and the secret key names
(api-token-1/api-token-2) remain consistent with
deployments/kubernetes/admin-api-deployment.yaml and deployments/README.md.
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
