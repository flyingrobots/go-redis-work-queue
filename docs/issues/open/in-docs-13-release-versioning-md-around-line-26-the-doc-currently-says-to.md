## In docs/13_release_versioning.md around line 26, the doc currently says to

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814712

- [review_comment] 2025-09-16T22:46:57Z by coderabbitai[bot] (docs/13_release_versioning.md:26)

```text
In docs/13_release_versioning.md around line 26, the doc currently says to
"ensure" supply-chain artifacts but lacks concrete, blocking verification steps;
update the section to include explicit, copy-paste verification commands for (1)
cosign container signature verification bound to the tag and OIDC issuer, (2)
slsa-verifier provenance verification against the release.intoto.jsonl and the
repo+tag, and (3) SBOM emission via syft producing spdx-json, and instruct users
to replace placeholders (org/repo, TAG, registry/image@digest, provenance path,
artifacts) and to run these commands in a failing/CI-blocking mode (e.g., run in
a shell with errexit or check exit codes) so any verification failure causes the
release pipeline to stop.
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
