## .github/workflows/changelog.yml around lines 28-29: the command uses unquoted

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567123

- [review_comment] 2025-09-16T03:15:00Z by coderabbitai[bot] (.github/workflows/changelog.yml:29)

```text
.github/workflows/changelog.yml around lines 28-29: the command uses unquoted
command substitution $(go env GOPATH) which can break due to word-splitting
(SC2046); change it to quote the substitution so the path is treated as a single
word (e.g. use "$(go env GOPATH)/bin/git-chglog" in the run line) and keep the
existing fallback (|| echo ...) as-is so the workflow behaves the same on
failure.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Confirmed the workflow already wraps the substitution in quotes and preserved fallback behaviour. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Confirmed the workflow already wraps the substitution in quotes and preserved fallback behaviour. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:168
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## .github/workflows/ci.yml around lines 38 to 45: the workflow immediately starts

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567136

- [review_comment] 2025-09-16T03:15:01Z by coderabbitai[bot] (.github/workflows/ci.yml:45)

```text
.github/workflows/ci.yml around lines 38 to 45: the workflow immediately starts
5 E2E runs against localhost:6379 without verifying Redis is ready, causing
flaky failures; add a readiness gate prior to the for-loop that waits for Redis
to respond (e.g., loop using redis-cli ping until it returns PONG or timeout, or
use a small curl/openssl TCP probe) and fail fast if timeout is reached; place
the wait logic just before the for i in {1..5} loop so tests only start after
Redis health is confirmed.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | CI workflow now waits for Redis before the E2E loop, matching the requested gating. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> CI workflow now waits for Redis before the E2E loop, matching the requested gating. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:241
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In .github/workflows/goreleaser.yml around lines 25 to 27, the echo redirections

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567141

- [review_comment] 2025-09-16T03:15:01Z by coderabbitai[bot] (.github/workflows/goreleaser.yml:27)

```text
In .github/workflows/goreleaser.yml around lines 25 to 27, the echo redirections
use an unquoted $GITHUB_ENV which can break on filenames with spaces or special
chars; update the two lines so the redirection target is quoted (use >>
"$GITHUB_ENV") and keep the echoed strings quoted as-is to ensure safe, portable
assignment into the env file.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Updated workflow writes to `"$GITHUB_ENV"`, eliminating the unquoted redirection. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Updated workflow writes to `"$GITHUB_ENV"`, eliminating the unquoted redirection. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:278
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## .github/workflows/markdownlint.yml around line 6: the branch array uses spaces

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567155

- [review_comment] 2025-09-16T03:15:01Z by coderabbitai[bot] (.github/workflows/markdownlint.yml:6)

```text
.github/workflows/markdownlint.yml around line 6: the branch array uses spaces
inside the brackets ("[ main ]") which trips YAML/style linters; change it to
use no inner spacing ("[main]") so the array is formatted as a compact literal
and passes linting.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Branch array rewritten as `[main]`, satisfying markdownlint and YAML format guidance. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Branch array rewritten as `[main]`, satisfying markdownlint and YAML format guidance. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:351
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## .github/workflows/markdownlint.yml lines 12-21: the workflow lacks

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567164

- [review_comment] 2025-09-16T03:15:02Z by coderabbitai[bot] (.github/workflows/markdownlint.yml:21)

```text
.github/workflows/markdownlint.yml lines 12-21: the workflow lacks
least-privilege permissions and concurrency control; add a permissions block
granting only what the job needs (e.g., permissions: contents: read) at the
workflow level and add a concurrency key to cancel duplicate runs (e.g., group
using workflow/ref or workflow/run id with cancel-in-progress: true) so runners
aren’t wasted and attack surface is reduced.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Workflow now ships with least-privilege permissions and concurrency guard as requested. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Workflow now ships with least-privilege permissions and concurrency guard as requested. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:385
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## .github/workflows/markdownlint.yml lines 12–16: the workflow uses mutable tags

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567169

- [review_comment] 2025-09-16T03:15:02Z by coderabbitai[bot] (.github/workflows/markdownlint.yml:16)

```text
.github/workflows/markdownlint.yml lines 12–16: the workflow uses mutable tags
for actions; update the two uses entries to pinned commit SHAs as suggested —
change actions/checkout@v4 to
actions/checkout@08eba0b27e820071cde6df949e0beb9ba4906955 (keep with:
fetch-depth: 0) and change DavidAnson/markdownlint-cli2-action@v17 to
DavidAnson/markdownlint-cli2-action@db43aef879112c3119a410d69f66701e0d530809 so
both actions are pinned to a specific commit SHA.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Workflow now pins `actions/checkout` and `markdownlint-cli2-action` to commit SHAs. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Workflow now pins `actions/checkout` and `markdownlint-cli2-action` to commit SHAs. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:421
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## .goreleaser.yaml around lines 15 to 20: the archives block currently produces

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567181

- [review_comment] 2025-09-16T03:15:02Z by coderabbitai[bot] (.goreleaser.yaml:20)

```text
.goreleaser.yaml around lines 15 to 20: the archives block currently produces
tar.gz for all OSes (including Windows); change it to keep tar.gz for
non-Windows and add a format_overrides entry that sets format: zip for
goos/windows so Windows builds produce zip archives. Update the archives stanza
to include format_overrides with a selector for goos: windows -> format: zip
(and ensure name_template remains appropriate for both formats).
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Added Windows-specific format override so GoReleaser emits `.zip` archives. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added Windows-specific format override so GoReleaser emits `.zip` archives. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:458
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## .goreleaser.yaml around lines 38 to 41: the current owner/name fields use bare

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567191

- [review_comment] 2025-09-16T03:15:02Z by coderabbitai[bot] (.goreleaser.yaml:41)

```text
.goreleaser.yaml around lines 38 to 41: the current owner/name fields use bare
templates which fail YAML linting and are unnecessary for releasing to the same
repository; either remove the two lines entirely so GoReleaser infers the repo,
or if you must keep them wrap the template values in quotes (e.g. owner: "{{
.Env.GITHUB_REPOSITORY_OWNER }}" and name: "{{ .Env.GITHUB_REPOSITORY_NAME }}")
to produce valid YAML.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Wrapped templated owner/name values in quotes to satisfy YAML tooling. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Wrapped templated owner/name values in quotes to satisfy YAML tooling. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:494
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## .goreleaser.yaml around lines 43 to 45: there is an extra trailing blank line

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567194

- [review_comment] 2025-09-16T03:15:02Z by coderabbitai[bot] (.goreleaser.yaml:45)

```text
.goreleaser.yaml around lines 43 to 45: there is an extra trailing blank line
after the snapshot.name_template block; remove the blank line so the file ends
immediately after the "name_template: SNAPSHOT-{{ .ShortCommit }}" line to
satisfy yamllint and CI.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Removed stray blank line so file ends directly after the snapshot template. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Removed stray blank line so file ends directly after the snapshot template. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:530
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In .markdownlint.yaml at line 4 the config disables MD013 repo-wide; remove the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567196

- [review_comment] 2025-09-16T03:15:03Z by coderabbitai[bot] (.markdownlint.yaml:4)

```text
In .markdownlint.yaml at line 4 the config disables MD013 repo-wide; remove the
global "MD013: false" entry and instead apply targeted exceptions: delete that
line from the repo config and either add per-file overrides in the markdownlint
config for specific file globs, or keep MD013 enabled and use inline markers
(<!-- markdownlint-disable MD013 --> / <!-- markdownlint-enable MD013 -->),
per-line disables, or file front-matter to silence the rule only where long
lines are intentional.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Removed global MD013 disable and documented per-file opt-out approach. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Removed global MD013 disable and documented per-file opt-out approach. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:564
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## .vscode/extensions.json lines 1-6: The workspace recommendations only include

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567204

- [review_comment] 2025-09-16T03:15:03Z by coderabbitai[bot] (.vscode/extensions.json:6)

```text
.vscode/extensions.json lines 1-6: The workspace recommendations only include
the Go extension but this repo uses YAML and Docker files; update the
recommendations array to add "redhat.vscode-yaml" and
"ms-azuretools.vscode-docker" so VS Code suggests installing YAML and Docker
extensions. Keep existing entries, avoid duplicates, and leave
unwantedRecommendations untouched.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Added YAML and Docker extensions to VS Code workspace recommendations. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added YAML and Docker extensions to VS Code workspace recommendations. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:601
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In .vscode/settings.json around lines 9 to 13, the workspace is not enabling

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567208

- [review_comment] 2025-09-16T03:15:03Z by coderabbitai[bot] (.vscode/settings.json:13)

```text
In .vscode/settings.json around lines 9 to 13, the workspace is not enabling
staticcheck or key gopls analyses and also contains a trailing comma that can
break JSON; update the gopls settings to set "staticcheck": true and enable
analyses such as "nilness", "shadow", "unusedparams", and "unusedwrite" (as
appropriate for your project), and remove the trailing comma after the last
property so the JSON remains valid.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Enabled staticcheck plus key gopls analyses and tidied JSON. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Enabled staticcheck plus key gopls analyses and tidied JSON. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:637
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In CHANGELOG.md around lines 17 to 30, the current entry is a freeform

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567211

- [review_comment] 2025-09-16T03:15:03Z by coderabbitai[bot] (CHANGELOG.md:30)

```text
In CHANGELOG.md around lines 17 to 30, the current entry is a freeform
brain-dump and must be converted to "Keep a Changelog" style: split the content
into explicit sections such as Added, Changed, Fixed (and optionally
Removed/Deprecated) and move each bullet under the appropriate section, convert
informal bullets into concise changelog-style lines, and append PR references
(e.g. " (#123)") for each item — leave a placeholder for PR numbers to be filled
in once merged and add a short header with the release version and date; ensure
the final format matches other entries in the file and includes a
[request_verification] note to confirm PR numbers after merge.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Reworked Unreleased section to follow Keep a Changelog formatting with placeholders for PR refs. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Reworked Unreleased section to follow Keep a Changelog formatting with placeholders for PR refs. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:673
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In create_review_tasks.py around lines 22-24 (and also update the similar logic

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567213

- [review_comment] 2025-09-16T03:15:03Z by coderabbitai[bot] (create_review_tasks.py:24)

```text
In create_review_tasks.py around lines 22-24 (and also update the similar logic
at lines 27-35), the code lexically sorts task identifiers which misorders items
like "T10" vs "T9"; change the sort to extract the numeric portion and sort by
that numeric value instead. Implement a sort key that parses the integer from
each task string (e.g., regex or split to grab trailing digits), fallback to 0
if no number present, then use sorted(completed_tasks, key=that_numeric_key) and
slice [:12]; apply the same numeric-key sorting to the other block at lines
27-35.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Numeric sort helper ensures tasks order naturally (T9 before T10). |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Numeric sort helper ensures tasks order naturally (T9 before T10). Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:712
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In demos/responsive-tui.tape around lines 20 to 27 (and similarly at 81-88 and

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567216

- [review_comment] 2025-09-16T03:15:04Z by coderabbitai[bot] (demos/responsive-tui.tape:27)

```text
In demos/responsive-tui.tape around lines 20 to 27 (and similarly at 81-88 and
138-146), the script sets COLUMNS to simulate mobile layout but this is cosmetic
for most tools; update the demo to either display an explicit on-screen note
that setting COLUMNS is only cosmetic, or trigger the TUI's real layout switch
API/flag if available so the UI truly reflows; also reduce sleep/delay durations
used around these sections to speed CI. Ensure the note is visible before the
COLUMNS change (or replace the COLUMNS step with a proper layout switch), and
apply the same change to the other line ranges mentioned.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Demo now calls out COLUMNS overrides as cosmetic and trims long sleeps. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Demo now calls out COLUMNS overrides as cosmetic and trims long sleeps. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:750
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In dependency_analysis.py around lines 233 to 243, the infrastructure dict is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567217

- [review_comment] 2025-09-16T03:15:04Z by coderabbitai[bot] (dependency_analysis.py:243)

```text
In dependency_analysis.py around lines 233 to 243, the infrastructure dict is
missing many referenced components which causes the validator to fail and DAG
edges to be dropped; add missing infra stubs such as "plugin_runtime",
"event_sourcing", "idempotency_keys", "controller_runtime", "k8s_api" and any
other referenced components in the codebase (e.g., "service_mesh",
"policy_engine", "sidecar_injector", "ci_cd", "secrets_manager") as minimal
string descriptions or promote them to feature objects consistent with the
surrounding structure so the validator recognizes them and edges in the DAG
remain intact.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Added missing infrastructure stubs so dependency validation no longer drops edges. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added missing infrastructure stubs so dependency validation no longer drops edges. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:788
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deploy/docker-compose.yml around lines 19 to 22 the service volume mounts

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567222

- [review_comment] 2025-09-16T03:15:04Z by coderabbitai[bot] (deploy/docker-compose.yml:22)

```text
In deploy/docker-compose.yml around lines 19 to 22 the service volume mounts
reference ./data but the repository has deploy/deploy/data (deploy/data is
missing), so either move the directory deploy/deploy/data → deploy/data to match
the current ./data mount (preferred) or update the compose file mounts for
app-all and app-producer to point to the existing path (e.g.,
./deploy/data:/data or an absolute path). Ensure the chosen fix is applied
consistently for all services and update any related documentation or .gitignore
entries if paths change.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Verified compose now mounts `./deploy/data` after relocating the data directory. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Verified compose now mounts `./deploy/data` after relocating the data directory. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:827
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deploy/grafana/dashboards/work-queue.json around lines 6 to 9, the PromQL

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567230

- [review_comment] 2025-09-16T03:15:04Z by coderabbitai[bot] (deploy/grafana/dashboards/work-queue.json:9)

```text
In deploy/grafana/dashboards/work-queue.json around lines 6 to 9, the PromQL
uses incorrect aggregation syntax "sum(...) by (le)"; change it to use the "sum
by (le) (...)" form and wrap the rate() call inside that aggregation so the
histogram_quantile receives a properly aggregated timeseries (i.e., compute rate
on the bucket metric over 5m, then apply sum by (le) around that result before
calling histogram_quantile).
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Updated query now aggregates with `sum by (le)` outside the `rate()` call. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Updated query now aggregates with `sum by (le)` outside the `rate()` call. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:865
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deploy/grafana/dashboards/work-queue.json around lines 20-24 (and likewise

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567239

- [review_comment] 2025-09-16T03:15:04Z by coderabbitai[bot] (deploy/grafana/dashboards/work-queue.json:24)

```text
In deploy/grafana/dashboards/work-queue.json around lines 20-24 (and likewise
adjust panels at lines 31-35), the stat panels lack units, thresholds and
field/state mappings; update the "Circuit Breaker State" stat to include
field/value mappings (0 -> "Closed (OK)" with green, 1 -> "Open (Alert)" with
red, 2 -> "Half-Open (Warn)" with amber), add a unit (e.g., "state" or "none"),
and set a reduce/threshold/color scheme so the numerical values render as
labeled colored states; for the "active workers" stat set unit to "none", add
thresholds so >0 is green and =0 is red, and ensure both panels use appropriate
fieldConfig -> mappings and thresholds entries so the UI shows readable labels
and colors at a glance.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Circuit breaker and worker stat panels now include mappings, thresholds, and units. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Circuit breaker and worker stat panels now include mappings, thresholds, and units. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:901
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deploy/grafana/dashboards/work-queue.json around lines 26 to 29, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567245

- [review_comment] 2025-09-16T03:15:04Z by coderabbitai[bot] (deploy/grafana/dashboards/work-queue.json:29)

```text
In deploy/grafana/dashboards/work-queue.json around lines 26 to 29, the
Prometheus target uses a raw metric "queue_length" which emits one time series
per replica; change the query to aggregate across replicas per logical queue
(for example use sum by (queue)(queue_length) or sum without grouping then
group_by the queue label) so the panel shows one series per queue rather than
one per instance, and update the legend/labeling to include the queue label for
clarity.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Queue length panel now queries `sum by (queue) (queue_length)`, collapsing per-instance series. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Queue length panel now queries `sum by (queue) (queue_length)`, collapsing per-instance series. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:941
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## deployments/admin-api/k8s-redis.yaml lines 1-62: the manifest lacks any

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567261

- [review_comment] 2025-09-16T03:15:05Z by coderabbitai[bot] (deployments/admin-api/k8s-redis.yaml:62)

```text
deployments/admin-api/k8s-redis.yaml lines 1-62: the manifest lacks any
pod/container security hardening; add a podSecurityContext and container
securityContext plus a locked ServiceAccount: set
spec.template.spec.serviceAccountName to a restricted SA (create a dedicated
redis SA with minimal RBAC), add spec.template.spec.securityContext with
runAsNonRoot: true and fsGroup (e.g., 1000) so /data is writable, and in the
container securityContext set runAsUser (non-root UID),
allowPrivilegeEscalation: false, capabilities: drop: ["ALL"],
readOnlyRootFilesystem: false (since Redis needs /data) and seccompProfile: {
type: RuntimeDefault }; ensure the SA and any PSP/PSP-replacement (PodSecurity
admission) policies permit these constraints.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Manifest now provisions a dedicated ServiceAccount and applies pod/container security contexts. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Manifest now provisions a dedicated ServiceAccount and applies pod/container security contexts. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:978
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/admin-api/k8s-redis.yaml around lines 62 to 74, the Service

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567278

- [review_comment] 2025-09-16T03:15:05Z by coderabbitai[bot] (deployments/admin-api/k8s-redis.yaml:74)

```text
In deployments/admin-api/k8s-redis.yaml around lines 62 to 74, the Service
manifest is missing a trailing newline and references the namespace
"redis-work-queue" without creating it; add a newline at EOF to satisfy the
linter and either add a Namespace manifest (apiVersion: v1, kind: Namespace,
metadata.name: redis-work-queue) in this file or a separate manifest applied
with the Service, or alternatively add a short comment/README note stating that
"redis-work-queue" must be created beforehand.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Namespace resource added alongside Service, satisfying manifest lint. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Namespace resource added alongside Service, satisfying manifest lint. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:1021
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/01_product_roadmap.md around lines 34 to 39, the roadmap dates list RC

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567287

- [review_comment] 2025-09-16T03:15:05Z by coderabbitai[bot] (docs/01_product_roadmap.md:39)

```text
In docs/01_product_roadmap.md around lines 34 to 39, the roadmap dates list RC
and GA without mapping which CI checks/gates must pass; update the roadmap to
enumerate the required checks and link them to the Release Plan and CI jobs.
Specifically, add a short checklist under the RC/GA entries naming the exact CI
jobs that must be green (unit, integration, e2e-with-redis, security-scan,
performance-smoke, deploy-preview), reference the canonical release-plan
file/path and the CI workflow file(s) (e.g., .github/workflows/release.yml), and
note any gating rules (e.g., “all listed jobs must pass and branch protection
enabled”) so the roadmap dates align with the Release Plan and CI gates.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | RC/GA bullets now enumerate gating CI jobs and reference release artifacts. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> RC/GA bullets now enumerate gating CI jobs and reference release artifacts. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:1059
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/01_product_roadmap.md around lines 48–52, the "Dependencies" bullets are

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567299

- [review_comment] 2025-09-16T03:15:05Z by coderabbitai[bot] (docs/01_product_roadmap.md:52)

```text
In docs/01_product_roadmap.md around lines 48–52, the "Dependencies" bullets are
vague; update each dependency to name the responsible owner and the concrete
artifact(s) that satisfy it (e.g., PR number, spec/doc link, or task ID).
Replace the three bullets with explicit entries such as "Tracing propagation —
owner: @alice — depends on finalized Job struct (PR #123) and processor API doc
(docs/processor-api.md#version-2)", "Reaper improvements — owner: @bob — depends
on reliable heartbeat semantics (RFC #45 / task JIRA-678)", and "Performance
tuning — owner: @carol — depends on priority dequeue semantics (PR #130) and
metrics completeness (metrics/README or dashboards PR #140)"; keep formatting
consistent with the file and add links or references where available.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Dependencies list now names owners and the exact upstream artifacts. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Dependencies list now names owners and the exact upstream artifacts. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:1098
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/03_milestones.md around lines 6 to 8, the milestone entries lack

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567307

- [review_comment] 2025-09-16T03:15:05Z by coderabbitai[bot] (docs/03_milestones.md:8)

```text
In docs/03_milestones.md around lines 6 to 8, the milestone entries lack
assigned owners/DRIs; add an owner for each milestone (name, role, contact) and
a backup/secondary DRI, and include a one-line responsibility statement per
owner. Update the milestones list or table to add an "Owner / DRI" column or a
subheading under each milestone with the owner's name, role, email/Slack handle,
and their specific accountability (e.g., delivery lead, QA lead), and ensure
dependencies or decision gates note who is responsible for sign-off.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Every milestone now lists owner, backup DRI, and responsibility statement. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Every milestone now lists owner, backup DRI, and responsibility statement. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:28
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/04_sprint_plans.md around lines 5 to 8, replace the ambiguous term

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567313

- [review_comment] 2025-09-16T03:15:05Z by coderabbitai[bot] (docs/04_sprint_plans.md:8)

```text
In docs/04_sprint_plans.md around lines 5 to 8, replace the ambiguous term
"bi-weekly sprints" with the explicit phrase "two-week sprints" (and any other
occurrences in this file) so the plan unambiguously states sprint length; update
the sentence to read something like "Four two-week sprints lead to v1.0.0." and
verify surrounding text remains grammatically correct.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 10 | Terminology now states "two-week sprints" explicitly. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Terminology now states "two-week sprints" explicitly. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:65
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/06_technical_spec.md around lines 113–117, the reaper section is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567317

- [review_comment] 2025-09-16T03:15:06Z by coderabbitai[bot] (docs/06_technical_spec.md:117)

```text
In docs/06_technical_spec.md around lines 113–117, the reaper section is
hand-wavy: stop inferring original queue from payload and instead persist an
origin_queue field in job metadata; limit SCAN cost by using SCAN/SSCAN with
COUNT plus a per-invocation time budget and randomized jitter between scans to
avoid thundering-herd effects; and perform RPOP/LPUSH re-queue operations inside
a Redis Lua script (EVAL) so list mutation and heartbeat checks are atomic and
consistent. Ensure the spec describes how the reaper reads heartbeat keys, skips
live workers, uses the persisted origin_queue to determine destination list,
enforces a page/time limit per run, and sleeps with jitter between pages.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Spec now details origin queue persistence, bounded SCAN windows with jitter, and Lua-based requeue steps. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Spec now details origin queue persistence, bounded SCAN windows with jitter, and Lua-based requeue steps. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:100
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/13_release_versioning.md around lines 21 to 25, the release checklist

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567321

- [review_comment] 2025-09-16T03:15:06Z by coderabbitai[bot] (docs/13_release_versioning.md:25)

```text
In docs/13_release_versioning.md around lines 21 to 25, the release checklist
lacks supply-chain verification gates; add a new checklist item requiring an
SBOM, provenance (SLSA/OIDC) attestation, and signed artifacts (cosign) before
release. Update the numbered list to include a clear line such as “4) Ensure
supply-chain artifacts present: SBOM generated, build provenance/SLSA
attestations (e.g. OIDC) available, and release artifacts signed (e.g. cosign)”;
optionally add brief links or references to existing internal tooling or
standards used for SBOM/provenance/signing and ensure CI blocks release when
these artifacts are missing.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Release checklist now includes an explicit supply-chain verification step. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Release checklist now includes an explicit supply-chain verification step. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:139
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/13_release_versioning.md around lines 26 to 31, the current instructions

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567323

- [review_comment] 2025-09-16T03:15:06Z by coderabbitai[bot] (docs/13_release_versioning.md:31)

```text
In docs/13_release_versioning.md around lines 26 to 31, the current instructions
use a lightweight tag and recommend git push --tags which can push all local
tags; change to recommend creating an annotated or signed tag and pushing only
that single ref. Update the steps to show using git tag -a (or -s) with a
message, then git push origin <tag-name>, replacing the generic --tags flow so
only the new release tag is published.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Release guide now instructs creating annotated tags and pushing the single ref. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Release guide now instructs creating annotated tags and pushing the single ref. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:178
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/YOU ARE WORKER 6.rb around lines 3 to 4, the README uses vendor-specific

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567336

- [review_comment] 2025-09-16T03:15:06Z by coderabbitai[bot] (docs/YOU ARE WORKER 6.rb:4)

```text
In docs/YOU ARE WORKER 6.rb around lines 3 to 4, the README uses vendor-specific
phrasing ("Claude worker"); replace that with a neutral role description (e.g.,
"a worker in the SLAPS task execution system" or "task worker") so the text
reads project-neutral: update the sentence to remove the vendor name and ensure
it still communicates that the role claims and executes tasks for the
go-redis-work-queue project.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Role description now uses vendor-neutral wording and adds safety guidance. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Role description now uses vendor-neutral wording and adds safety guidance. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:251
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/YOU ARE WORKER 6.rb around lines 6 to 13, the workflow step that tells

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567349

- [review_comment] 2025-09-16T03:15:06Z by coderabbitai[bot] (docs/YOU ARE WORKER 6.rb:13)

```text
In docs/YOU ARE WORKER 6.rb around lines 6 to 13, the workflow step that tells
workers to "Claim a task by moving it to slaps-coordination/claude-001/" lacks
any caution about mv being non-atomic across filesystems and about race
conditions that can corrupt the queue; add a brief safety note instructing
maintainers to (1) prefer an atomic rename on the same filesystem or implement
an explicit locking/claim protocol (e.g., create a unique temporary/claim file
or use link+unlink or filesystem locks) to avoid collisions, (2) detect and
handle failures from mv/rename and retry/abort safely, and (3) fall back to an
explicit copy+atomic-rename pattern with checks when crossing filesystems to
prevent duplicate processing.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Added move safety guidance outlining atomic renames, copy+rename fallbacks, and error checks. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added move safety guidance outlining atomic renames, copy+rename fallbacks, and error checks. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:287
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/YOU ARE WORKER 6.rb around lines 21 to 26, the example shell commands

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567354

- [review_comment] 2025-09-16T03:15:06Z by coderabbitai[bot] (docs/YOU ARE WORKER 6.rb:26)

```text
In docs/YOU ARE WORKER 6.rb around lines 21 to 26, the example shell commands
use paths with spaces unquoted and a plain mv that can clobber files; update the
examples to quote all filesystem paths (e.g.
"slaps-coordination/open-tasks/P1.T001.json") and add the -n flag to mv (mv -n
"source" "dest/") to prevent overwriting; ensure any path that could be
interpreted as an option is protected by quoting or by using -- where
appropriate.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Documentation now demonstrates safe quoted `mv -n` usage. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Documentation now demonstrates safe quoted `mv -n` usage. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:327
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In auto_commit.sh around line 1, the script lacks Bash strict mode which can

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569555

- [review_comment] 2025-09-16T03:16:13Z by coderabbitai[bot] (auto_commit.sh:1)

```text
In auto_commit.sh around line 1, the script lacks Bash strict mode which can
hide failures in the infinite git loop; add a strict-mode invocation immediately
after the shebang and set a safe IFS: enable errexit, errtrace, nounset and
pipefail (i.e. set -Eeuo pipefail) and set IFS to newline+tab so word-splitting
is safe; ensure this is the first executable code so the script exits on errors
and treats unset variables as failures.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Script now enables `set -Eeuo pipefail` and sets IFS safely right after the shebang. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Script now enables `set -Eeuo pipefail` and sets IFS safely right after the shebang. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:364
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In auto_commit.sh around lines 4-6 (and similarly lines 45-47) the script starts

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569560

- [review_comment] 2025-09-16T03:16:14Z by coderabbitai[bot] (auto_commit.sh:6)

```text
In auto_commit.sh around lines 4-6 (and similarly lines 45-47) the script starts
an unconditional infinite loop which can DOS CI; add a guard using an
environment variable (e.g., MAX_ITERATIONS) with a sensible default and an
iteration counter that increments each loop and breaks when the max is reached,
and add signal handling: install a trap for SIGINT and SIGTERM that sets a flag
(or exits) so the loop can terminate cleanly; ensure the loop checks the flag
before each iteration and after sleep, and prefer a configurable SLEEP_SECONDS
(default 300) so cadence remains adjustable.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Loop now honours MAX_ITERATIONS, sleep interval overrides, and signal handlers. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Loop now honours MAX_ITERATIONS, sleep interval overrides, and signal handlers. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:400
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In auto_commit.sh around lines 30 to 41, the script currently parses git push

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569563

- [review_comment] 2025-09-16T03:16:14Z by coderabbitai[bot] (auto_commit.sh:41)

```text
In auto_commit.sh around lines 30 to 41, the script currently parses git push
output with grep which is brittle; instead call git rev-parse --abbrev-ref HEAD
to get the current branch, run git push origin <current-branch> (or git push
--set-upstream origin <current-branch> if upstream is not set) and check git’s
exit status ($?) to determine success; remove the grep pipeline and use the push
command’s exit code to log success or failure, and when upstream is unset detect
that (e.g., by checking git rev-parse --symbolic-full-name @{u} or examining
push exit code) and set upstream only when needed.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Push logic now checks git exit codes directly and handles upstream detection without grepping output. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Push logic now checks git exit codes directly and handles upstream detection without grepping output. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:438
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In config/config.example.yaml around line 2, the default Redis address is set to

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569570

- [review_comment] 2025-09-16T03:16:14Z by coderabbitai[bot] (config/config.example.yaml:2)

```text
In config/config.example.yaml around line 2, the default Redis address is set to
"redis:6379" which breaks local development; change the default to
"localhost:6379" and add a one-line comment explaining that Docker Compose uses
the service name (e.g., redis:6379) and show an example override instruction
(mention using a compose-specific config or environment variable to set addr to
redis:6379 for containerized runs). Ensure the example indicates how to override
the value (compose config or env) and keep the default local-friendly value in
the example file.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Default now targets localhost with a comment noting Docker overrides. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Default now targets localhost with a comment noting Docker overrides. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:476
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In config/config.example.yaml around lines 50 to 66, the idempotency settings

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569573

- [review_comment] 2025-09-16T03:16:14Z by coderabbitai[bot] (config/config.example.yaml:66)

```text
In config/config.example.yaml around lines 50 to 66, the idempotency settings
lack precise semantics; update the example with comments that (1) define exactly
what constitutes the idempotency key (e.g., which headers and/or payload hash
are used and the format that fills {queue},{tenant},{key}), (2) describe
collision behavior (what happens when two requests produce the same key — e.g.,
first-wins, overwrite, or atomic check-and-set) and how retries interact with
stored results, (3) clarify TTL semantics versus cleanup_interval (TTL is how
long a record is valid; cleanup is periodic garbage collection and may not
immediately remove expired keys), and (4) note implications for storage options
(e.g., Redis hashes vs keys, compression effects) so users can configure
default_ttl, cleanup_interval, max_retries and retry_delay correctly.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Added inline comments covering key format, collision behaviour, and TTL vs cleanup semantics. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added inline comments covering key format, collision behaviour, and TTL vs cleanup semantics. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:514
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In config/config.example.yaml around lines 67 to 80, the outbox section is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569576

- [review_comment] 2025-09-16T03:16:14Z by coderabbitai[bot] (config/config.example.yaml:80)

```text
In config/config.example.yaml around lines 67 to 80, the outbox section is
disabled and only notes "Requires database setup" without any example DB
configuration or usage guidance; add a commented example DSN and required
schema/table names (e.g., outbox table name, columns used), an example polling
query or SQL snippet, and recommended config knobs (connection string, table,
schema, retention/purge query) or a link to the docs page that defines the
outbox setup so users can enable and configure it easily.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Outbox config now includes sample DSN, table name, and polling query hints. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Outbox config now includes sample DSN, table name, and polling query hints. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:555
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In create_review_tasks.py around lines 1-4 and also update lines 102-112,

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569580

- [review_comment] 2025-09-16T03:16:14Z by coderabbitai[bot] (create_review_tasks.py:4)

```text
In create_review_tasks.py around lines 1-4 and also update lines 102-112,
convert the script into a CLI tool by wrapping execution in if __name__ ==
'__main__' and adding argparse flags --limit (int) and --dir (path) to control
output quantity and directory; add an optional --timestamp flag (ISO8601 or
epoch) that, when provided, is used instead of datetime.now() so outputs are
deterministic for CI/tests; refactor functions that currently call
datetime.now() or use globals to accept a timestamp parameter (defaulting to
now) and ensure the main entry parses flags, injects the parsed timestamp into
those functions, and writes outputs to the specified dir with behavior unchanged
when flags are omitted.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Script now exposes argparse flags for limit, dirs, and timestamp with a guarded main entry. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Script now exposes argparse flags for limit, dirs, and timestamp with a guarded main entry. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:592
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In create_review_tasks.py around lines 9 to 11, the code calls

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569586

- [review_comment] 2025-09-16T03:16:14Z by coderabbitai[bot] (create_review_tasks.py:11)

```text
In create_review_tasks.py around lines 9 to 11, the code calls
os.listdir(completed_dir) which will crash if completed_dir does not exist; add
a guard to check for directory existence (os.path.isdir or os.path.exists)
before listing, and either create the directory (os.makedirs(completed_dir,
exist_ok=True)) or skip processing when it’s missing; then proceed to iterate
over files only when the directory exists to avoid the crash.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Duplicate filter now lowercases filenames before checking. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Duplicate filter now lowercases filenames before checking. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:958
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In create_review_tasks.py around lines 30-31, the code constructs timestamps

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569589

- [review_comment] 2025-09-16T03:16:14Z by coderabbitai[bot] (create_review_tasks.py:31)

```text
In create_review_tasks.py around lines 30-31, the code constructs timestamps
with datetime.now().isoformat()+"Z" which produces a naive local time mislabeled
as UTC; change this to use an aware UTC timestamp by calling
datetime.now(timezone.utc).isoformat() (and add/import timezone from datetime if
missing) so the produced ISO string reflects real UTC (or, if you must keep the
trailing "Z", use datetime.now(timezone.utc).isoformat().replace("+00:00","Z")).
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Timestamps now derive from `datetime.now(timezone.utc)` and render with `Z`. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Timestamps now derive from `datetime.now(timezone.utc)` and render with `Z`. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:668
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In demos/lipgloss-transformation.tape around lines 136-142 (and similarly for

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569595

- [review_comment] 2025-09-16T03:16:15Z by coderabbitai[bot] (demos/lipgloss-transformation.tape:142)

```text
In demos/lipgloss-transformation.tape around lines 136-142 (and similarly for
147-162) you are scripting "Type" commands that will fail if the referenced
binary doesn’t exist; add a clear guard or disclaimer: check for the binary file
(e.g., test -x ./bin/queue-tui) before running the Type sequence and skip or
print an explanatory echo if missing, or replace the sequence with a
commented/echoed mock note indicating this is a scripted demo and not executed;
ensure the demo prints a clear message when skipped so users know why the
interactive steps weren’t run.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Tape now runs `if [ -x ./bin/queue-tui ]` before invoking the binary. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Tape now runs `if [ -x ./bin/queue-tui ]` before invoking the binary. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:704
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In demos/lipgloss-transformation.tape around lines 271 to 276, the script calls

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569601

- [review_comment] 2025-09-16T03:16:15Z by coderabbitai[bot] (demos/lipgloss-transformation.tape:276)

```text
In demos/lipgloss-transformation.tape around lines 271 to 276, the script calls
figlet directly which will break on hosts without figlet; wrap the figlet
invocation in a guard that checks for the figlet binary (e.g., `command -v
figlet >/dev/null`) and, if missing, output a sensible fallback (plain "Redis
TUI" or a simple ASCII alternative) so the tape continues; apply the same
guarded pattern used elsewhere in the repo for the other figlet invocation.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | figlet call now wrapped in `command -v` test with plain-text fallback. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> figlet call now wrapped in `command -v` test with plain-text fallback. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:742
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In dependency_analysis.py around lines 7 to 231, feature keys use kebab-case

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569619

- [review_comment] 2025-09-16T03:16:15Z by coderabbitai[bot] (dependency_analysis.py:231)

```text
In dependency_analysis.py around lines 7 to 231, feature keys use kebab-case
(e.g., "admin-api") while dependency lists use snake_case (e.g., "admin_api"),
causing resolution failures; normalize names to one canonical form (pick either
kebab-case or snake_case) at import by mapping all feature keys and all
dependency entries through the same normalizer (replace - with _ or vice versa)
before building the graph, correct known typos (e.g., change
"distributed_tracing" to "distributed-tracing-integration" or its normalized
equivalent), and add a validation pass that checks each dependency resolves to a
defined feature and logs or raises on unresolved refs so downstream DAGs fail
fast.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Added normalization helpers, alias mapping, infrastructure stubs, and a validation routine. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added normalization helpers, alias mapping, infrastructure stubs, and a validation routine. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:814
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/05_architecture.md around line 7, the architecture text omits the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569626

- [review_comment] 2025-09-16T03:16:16Z by coderabbitai[bot] (docs/05_architecture.md:7)

```text
In docs/05_architecture.md around line 7, the architecture text omits the
exactly-once/idempotency component; update the paragraph to add an
“Exactly-once” component and describe its flows: record/check idempotency key at
job intake, guard worker processing with idempotency lookup, persist
side-effects to an outbox before acknowledging completion, and publish outbox
entries to external systems; also mention how the reaper and circuit breaker
interact with idempotency (do not double-process keys) and that observability
should include metrics/traces for idempotency/outbox operations.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Component list and data flows now call out the idempotency/outbox layer. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Component list and data flows now call out the idempotency/outbox layer. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:854
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/06_technical_spec.md around lines 124-129 and also 131-134, the metrics

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569642

- [review_comment] 2025-09-16T03:16:16Z by coderabbitai[bot] (docs/06_technical_spec.md:129)

```text
In docs/06_technical_spec.md around lines 124-129 and also 131-134, the metrics
and logging section lacks explicit label names, cardinality bounds, unit
verification, and stable log key rules; update the doc to (1) list every metric
label schema (e.g., queue_length{queue}) and state a max cardinality or allowed
value set for each label, (2) verify and declare that the histogram metric uses
seconds (or rename suffix) so `_seconds` matches actual units, (3) state exact,
enforced log key names (trace_id, span_id, job_id, queue, worker_id) and forbid
logging secrets or PII, and (4) add a short note about how to enforce these
constraints in code/review (e.g., validation rules or linter/checklist).
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Spec now enumerates metric labels, units, and canonical log keys with guidance. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Spec now enumerates metric labels, units, and canonical log keys with guidance. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:892
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/09_requirements.md around lines 43-49, the acceptance criteria are

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569663

- [review_comment] 2025-09-16T03:16:16Z by coderabbitai[bot] (docs/09_requirements.md:49)

```text
In docs/09_requirements.md around lines 43-49, the acceptance criteria are
currently vague; update them to (1) enumerate the exact metric names and
expected types exposed at /metrics (e.g., request_count: counter,
request_duration_seconds: histogram, job_queue_length: gauge,
worker_registered_total: gauge) so tests can verify them, (2) specify that
/readyz must return healthy only if a successful Redis PING is received and at
least one worker is registered (describe the exact probe: call Redis PING and
check worker registration API/state), and (3) require admin destructive commands
to prompt for confirmation interactively and also accept a --yes flag for
non-interactive runs; additionally require automated unit/integration tests that
assert metric names/types, the /readyz behavior under Redis failure and
no-worker conditions, and admin command behavior with and without --yes.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Acceptance criteria now spell out metric names, readiness semantics, and admin safeguards. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Acceptance criteria now spell out metric names, readiness semantics, and admin safeguards. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:931
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/12_performance_baseline.md around lines 26 to 28, the example starts a

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569678

- [review_comment] 2025-09-16T03:16:17Z by coderabbitai[bot] (docs/12_performance_baseline.md:28)

```text
In docs/12_performance_baseline.md around lines 26 to 28, the example starts a
Redis container without pinning a specific minor version and lacks teardown
steps; update the run command to use a fixed minor Redis tag (e.g.,
7.2.x-alpine) and add explicit cleanup commands (stop/remove or force remove) so
reviewers can reproduce results reliably and avoid leftover containers.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Baseline now pins Redis to 7.2.4-alpine and documents cleanup. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Baseline now pins Redis to 7.2.4-alpine and documents cleanup. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:973
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/12_performance_baseline.md around lines 31 to 33, the README tells users

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569692

- [review_comment] 2025-09-16T03:16:17Z by coderabbitai[bot] (docs/12_performance_baseline.md:33)

```text
In docs/12_performance_baseline.md around lines 31 to 33, the README tells users
to run Redis locally but the example/default config uses redis:6379 and
therefore mismatches; update the doc to explicitly instruct readers to set
redis.addr="localhost:6379" for this baseline (or update the example config to
use localhost:6379) so the instructions and config agree—prefer adding a
one-line note beneath the config block that says: "Note: set
redis.addr=\"localhost:6379\" if running Redis locally for this baseline."
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Documentation now highlights the localhost address needed for the baseline setup. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Documentation now highlights the localhost address needed for the baseline setup. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:1008
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/14_ops_runbook.md around lines 21 to 26, replace the single unpinned

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569713

- [review_comment] 2025-09-16T03:16:18Z by coderabbitai[bot] (docs/14_ops_runbook.md:26)

```text
In docs/14_ops_runbook.md around lines 21 to 26, replace the single unpinned
"docker build -t job-queue-system:local ." entry with a pinned, reproducible
multi-arch build and an alternative compose build flow: update the docs to show
a buildx command that specifies platforms (e.g., linux/amd64,linux/arm64), pins
base image via build args or explicit tags, enables inline cache
(BUILDKIT_INLINE_CACHE=1), and uses --pull (and --push if publishing) so images
are reproducible across architectures; also add a separate example showing how
to build the same image via docker compose build for parity with later compose
notes.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Ops runbook now documents buildx multi-arch builds and compose parity commands. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Ops runbook now documents buildx multi-arch builds and compose parity commands. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:1045
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/14_ops_runbook.md around lines 32 to 35, the env-var mapping description

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569730

- [review_comment] 2025-09-16T03:16:18Z by coderabbitai[bot] (docs/14_ops_runbook.md:35)

```text
In docs/14_ops_runbook.md around lines 32 to 35, the env-var mapping description
is too vague; replace the hand-wavy “upper snake case replaces dots” with
explicit examples and parsing notes: add 1–2 concrete mappings (e.g.,
WORKER_COUNT -> worker.count and REDIS_ADDR -> redis.addr), show the
transformation rule (dots -> underscores, keys uppercased), and note how
booleans (true/false/1/0) and durations/times are parsed (e.g., "30s" ->
duration) and any required quoting; update the list to include these exact
mappings and parsing expectations so operators know how to set env vars
unambiguously.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Added explicit mapping examples and parsing notes for environment overrides. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added explicit mapping examples and parsing notes for environment overrides. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:1084
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/14_ops_runbook.md around lines 38 to 42, the guidance currently

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569756

- [review_comment] 2025-09-16T03:16:19Z by coderabbitai[bot] (docs/14_ops_runbook.md:42)

```text
In docs/14_ops_runbook.md around lines 38 to 42, the guidance currently
documents /healthz, /readyz, /metrics but doesn't instruct how to restrict
access; update this section to recommend binding health/metrics endpoints to
localhost or a dedicated admin interface, or expose them on a separate
port/interface, and add explicit protection guidance: enforce network
policies/firewall rules to restrict access, require authentication/authorization
(mTLS, bearer tokens or HTTP basic+IP allowlist) for admin/metrics endpoints,
and note Prometheus should scrape via a securely proxied or authenticated
endpoint; keep the examples concise and state to avoid exposing these endpoints
on public listeners.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Ops runbook now calls out binding metrics/health endpoints to controlled interfaces and restricting access. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Ops runbook now calls out binding metrics/health endpoints to controlled interfaces and restricting access. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:1123
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/14_ops_runbook.md around lines 51 to 75, the purge-dlq example and

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569759

- [review_comment] 2025-09-16T03:16:19Z by coderabbitai[bot] (docs/14_ops_runbook.md:75)

```text
In docs/14_ops_runbook.md around lines 51 to 75, the purge-dlq example and
surrounding admin CLI docs lack a dry-run example and an explicit RBAC note;
update the purge-dlq command example to include a --dry-run (and keep --yes
separate) showing safe preview usage, and add a short sentence noting that purge
operations require admin RBAC (e.g., only users/roles with purge/delete
permissions may execute) and recommend running dry-run first before --yes; keep
the other examples unchanged.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Runbook now documents `--dry-run` usage and mentions RBAC requirements for purge operations. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Runbook now documents `--dry-run` usage and mentions RBAC requirements for purge operations. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:28
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/dlq-remediation-pipeline.md around lines 121 to 137, the sample

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569774

- [review_comment] 2025-09-16T03:16:19Z by coderabbitai[bot] (docs/api/dlq-remediation-pipeline.md:137)

```text
In docs/api/dlq-remediation-pipeline.md around lines 121 to 137, the sample
response mixes integer and float types and embeds unit labels only in keys (ms),
which causes churn for clients; standardize metric types and lock units by (1)
making counts strictly integers (jobs_processed, jobs_matched, actions_executed,
actions_successful, actions_failed, rate_limit_hits, circuit_breaker_trips), (2)
expressing timing metrics as numbers in milliseconds as integers
(classification_time_ms, action_time_ms, end_to_end_time_ms) or explicitly state
they are floats if sub-millisecond precision is required, (3) keeping ratios/hit
rates as floats between 0 and 1 (cache_hit_rate), (4) update the JSON example to
use consistent types and values accordingly, and (5) add a short typed schema
section immediately after the example listing each field name, its JSON type,
and the unit (e.g., "classification_time_ms: integer (ms)") so clients have an
authoritative contract.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Example response now uses integer millisecond values with a schema table. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Example response now uses integer millisecond values with a schema table. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:65
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/dlq-remediation-pipeline.md around lines 541–606, the audit example

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569781

- [review_comment] 2025-09-16T03:16:20Z by coderabbitai[bot] (docs/api/dlq-remediation-pipeline.md:606)

```text
In docs/api/dlq-remediation-pipeline.md around lines 541–606, the audit example
exposes full payloads and redaction behavior is not documented; update the text
and example to state that audit payloads are redacted by default, configurable
via an audit_redaction setting, and show a redacted response example. Specify a
minimal canonical list of always-masked fields (e.g., ssn, email, phone,
full_name, address, credit_card, payment_token, auth_token, password) and note
that nested payload keys matching patterns are masked; replace the
before_state/after_state content in the JSON example with redacted placeholders
(e.g., "<redacted>") and add a short note pointing to the config section that
explains how to change redaction level and add custom fields.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Audit example now shows redacted payloads with guidance on configurable masks. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Audit example now shows redacted payloads with guidance on configurable masks. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:147
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/dlq-remediation-pipeline.md around lines 860 to 876, the WebSocket

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569789

- [review_comment] 2025-09-16T03:16:20Z by coderabbitai[bot] (docs/api/dlq-remediation-pipeline.md:876)

```text
In docs/api/dlq-remediation-pipeline.md around lines 860 to 876, the WebSocket
events section lacks details about authentication and backpressure/heartbeat
handling; update the doc to require a bearer/token query param or Authorization
header for the /ws/dlq-remediation/events endpoint and show the token format and
renewal behavior, specify heartbeat/ping semantics (client must respond to
server pings and send an application-level heartbeat every N seconds, include
ping/pong timeouts and reconnect guidance), and define a
slow-consumer/backpressure policy (per-connection send buffer limits,
server-side queue thresholds, and the chosen strategy: drop oldest messages vs.
close connection with a close code and reason), plus recommend monitoring
metrics and recommended defaults (buffer size, ping interval, timeout) so
implementers can prevent memory blowup.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | WebSocket documentation now covers auth, ping/pong expectations, slow-consumer limits, and metrics. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> WebSocket documentation now covers auth, ping/pong expectations, slow-consumer limits, and metrics. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:187
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/dlq-remediation-ui.md around line 9, the doc currently states the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569797

- [review_comment] 2025-09-16T03:16:20Z by coderabbitai[bot] (docs/api/dlq-remediation-ui.md:9)

```text
In docs/api/dlq-remediation-ui.md around line 9, the doc currently states the
API has no auth; change the implementation notes to require authentication, CSRF
for browser clients, and RBAC with default-deny: add a mandatory auth middleware
(JWT/OAuth session) on all DLQ remediation endpoints, enforce CSRF validation on
state-changing requests originating from browsers, implement role checks (e.g.,
require "dlq_admin" or specific capability to purge/modify DLQs) and return 403
by default for unauthorized users, and document the required roles, token scope,
and recommended audit logging for all purge operations.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Authentication section now mandates RBAC, CSRF, and bearer/session tokens. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Authentication section now mandates RBAC, CSRF, and bearer/session tokens. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:229
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/dlq-remediation-ui.md around lines 231 to 241, the API currently

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569804

- [review_comment] 2025-09-16T03:16:20Z by coderabbitai[bot] (docs/api/dlq-remediation-ui.md:241)

```text
In docs/api/dlq-remediation-ui.md around lines 231 to 241, the API currently
uses a query param confirm=true which is not secure; replace this with a signed
confirmation token mechanism, change the endpoint to require a JSON POST body
that contains an explicit boolean dry_run flag (must be provided) and the signed
confirmation token, and add a mandatory change_reason string field that will be
validated and persisted to logs; update request validation to reject
query-string confirmation, validate and verify the token signature/expiry,
enforce dry_run presence before executing real changes, and ensure the reason is
recorded in audit logs for every remediation action.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Purge-all now takes a POST body with `confirm`, structured filters, and requires an `Idempotency-Key` header. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Purge-all now takes a POST body with `confirm`, structured filters, and requires an `Idempotency-Key` header. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:188
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/dlq-remediation-ui.md around lines 387 to 394, the documented rate

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569811

- [review_comment] 2025-09-16T03:16:20Z by coderabbitai[bot] (docs/api/dlq-remediation-ui.md:394)

```text
In docs/api/dlq-remediation-ui.md around lines 387 to 394, the documented rate
limits are arbitrary and not tied to roles/endpoints; update the section to
specify rate limits per endpoint and per role, differentiating read (list/get)
vs destructive (purge/requeue) operations. Change the list to explicitly state
limits per endpoint group (e.g., list, bulk, individual) and add stricter,
per-principal limits for destructive endpoints (purge/requeue) with
rate/permission mappings (e.g., admin/service account vs regular user), and
document how limits are enforced (per-IP, per-token) and any burst/penalty
behavior.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Rate limit section now lists per-role/per-endpoint quotas and enforcement details. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Rate limit section now lists per-role/per-endpoint quotas and enforcement details. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:306
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/SLAPS/FINAL-POSTMORTEM.md around lines 8-9 (and also lines 324-336), the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569820

- [review_comment] 2025-09-16T03:16:20Z by coderabbitai[bot] (docs/SLAPS/FINAL-POSTMORTEM.md:9)

```text
In docs/SLAPS/FINAL-POSTMORTEM.md around lines 8-9 (and also lines 324-336), the
header shows "Success Rate: 100% (74 completed, 14 remaining)" but the body
computes 74/88 = 84% — make these consistent by either (A) replacing the header
with a clear completion metric like "Completion: 84% (74/88 completed)" or (B)
keeping "Success Rate: 100%" and clarifying it refers to attempted tasks (e.g.,
"Success Rate: 100% (74 succeeded of 74 attempted); 14 tasks not yet
attempted"), then update the related occurrences at lines 324-336 to match the
chosen wording and ensure the numeric labels (completed/attempted/remaining) are
accurate.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Postmortem now reports “Completion: 84% (74/88 tasks finished)” with consistent follow-up bullets. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Postmortem now reports “Completion: 84% (74/88 tasks finished)” with consistent follow-up bullets. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:345
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In README.md around lines 3 to 6, the intro mentions features without a clear,

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569828

- [review_comment] 2025-09-16T03:16:21Z by coderabbitai[bot] (README.md:6)

```text
In README.md around lines 3 to 6, the intro mentions features without a clear,
maintainable reference; add a link to a living feature matrix (e.g., a
FEATURES.md or a table in the repo/docs site) indicating
supported/experimental/deprecated status. Update the intro to include a short
parenthetical or sentence like "See the feature matrix: <relative-link>" and
create or point to that living document in the repo, ensuring it is kept
up-to-date and clearly lists feature statuses and notes.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | README now links to `docs/features-ledger.md` for the live feature matrix. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> README now links to `docs/features-ledger.md` for the live feature matrix. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:384
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In append_metadata.py around line 11, the script uses a hardcoded absolute path

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350571937

- [review_comment] 2025-09-16T03:17:27Z by coderabbitai[bot] (append_metadata.py:11)

```text
In append_metadata.py around line 11, the script uses a hardcoded absolute path
(/Users/james/...), which breaks on other machines and CI; replace it with a
configurable value: accept an --ideas-dir CLI argument (via argparse) and/or
read an IDEAS_DIR environment variable, falling back to a sensible relative
default (e.g., ./docs/ideas) and expanding user (~) with os.path.expanduser;
validate the path exists and fail with a clear error if not.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Script now takes `--ideas-dir` (env override) instead of a hardcoded path. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Script now takes `--ideas-dir` (env override) instead of a hardcoded path. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:568
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In auto_commit.sh around lines 8 to 10, the current use of "ls ... | wc -l | tr

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350571968

- [review_comment] 2025-09-16T03:17:27Z by coderabbitai[bot] (auto_commit.sh:10)

```text
In auto_commit.sh around lines 8 to 10, the current use of "ls ... | wc -l | tr
-d ' '" is brittle for filenames and empty directories; replace each count with
a robust find invocation that restricts depth and type. For example, use find
slaps-coordination/open-tasks -mindepth 1 -maxdepth 1 -type f | wc -l (and
similarly for finished-tasks and help-me) to accurately count files regardless
of names or locales; ensure any non-existent directory errors are suppressed
(e.g., redirect find's stderr to /dev/null) if needed.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Replaced ls/wc pipeline with a `find`-based helper that tolerates missing dirs. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Replaced ls/wc pipeline with a `find`-based helper that tolerates missing dirs. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:604
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In auto_commit.sh around lines 16–28, the git commit call currently silences

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350571987

- [review_comment] 2025-09-16T03:17:27Z by coderabbitai[bot] (auto_commit.sh:28)

```text
In auto_commit.sh around lines 16–28, the git commit call currently silences
errors with 2>/dev/null and embeds a third‑party marketing line plus a
Co-Authored-By footer; remove the stderr redirection so failures surface, and
remove the marketing/Co‑Authored‑By text from the default commit message (or
make those footers opt‑in via an environment variable like INCLUDE_GENERATOR),
then ensure the script checks git commit's exit code and exits non‑zero (or
prints the git error) on failure so CI/users see the problem.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Commit step now surfaces errors and uses a neutral message without co-author footers. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Commit step now surfaces errors and uses a neutral message without co-author footers. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:641
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In claude_worker.py around lines 90-145 the worker currently blocks on

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572059

- [review_comment] 2025-09-16T03:17:28Z by coderabbitai[bot] (claude_worker.py:145)

```text
In claude_worker.py around lines 90-145 the worker currently blocks on
interactive input() calls to mark tasks done/help/failed, which turns the module
into a manual tool; replace the interactive loop with programmatic task status
handling: remove input() calls and instead call a task execution method (e.g.,
self.execute_task(task_data)) that returns a status enum/tuple (status,
summary_or_error), set timestamps/actor from that result, serialize and move the
file to finished/help/failed directories accordingly, and implement retries,
timeout handling, and logging for failures; alternatively, if the intention is
truly manual, rename the file/class to manual_task_tracker.py and update
docs/imports. Ensure no user prompts remain and keep file moves and JSON writes
identical to current behavior.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Worker now delegates to a pluggable executor returning TaskStatus instead of prompting via input(). |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Worker now delegates to a pluggable executor returning TaskStatus instead of prompting via input(). Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:755
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In claude_worker.py around lines 157-159 the bare "except: pass" silently

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572062

- [review_comment] 2025-09-16T03:17:28Z by coderabbitai[bot] (claude_worker.py:159)

```text
In claude_worker.py around lines 157-159 the bare "except: pass" silently
swallows all exceptions; replace it by catching and handling only the expected
exceptions (e.g., json.JSONDecodeError, OSError/IOError, ValueError) and log the
exception with context via the module logger, return False for
handled/non-critical errors, and re-raise truly critical exceptions
(KeyboardInterrupt, SystemExit, MemoryError) so they propagate; ensure logs
include the exception message and stack trace (logger.exception) to aid
debugging.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Replaced bare except with targeted exception handling and failure logging. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Replaced bare except with targeted exception handling and failure logging. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:796
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/admin-api/monitoring.yaml lines 1-66: this ConfigMap duplicates

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572067

- [review_comment] 2025-09-16T03:17:29Z by coderabbitai[bot] (deployments/admin-api/monitoring.yaml:66)

```text
In deployments/admin-api/monitoring.yaml lines 1-66: this ConfigMap duplicates
alerts already managed in deployments/kubernetes/monitoring.yaml and conflicts
with the monitoring operator (ConfigMap-based rules vs PrometheusRule). Fix by
either deleting this file entirely, or converting its contents into a
PrometheusRule (and ServiceMonitor if needed) CRD placed in the same namespace
and using the same labels/owner/namespace conventions as the existing monitoring
manifests under deployments/kubernetes/monitoring.yaml so the operator picks it
up; ensure you do not keep both ConfigMap and PrometheusRule definitions for the
same alerts to avoid duplication.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Converted alert ConfigMap into a PrometheusRule resource. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Converted alert ConfigMap into a PrometheusRule resource. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:834
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/admin-api/monitoring.yaml around line 5 (and also line 71), the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572071

- [review_comment] 2025-09-16T03:17:29Z by coderabbitai[bot] (deployments/admin-api/monitoring.yaml:5)

```text
In deployments/admin-api/monitoring.yaml around line 5 (and also line 71), the
namespace is set to "redis-work-queue" which conflicts with the expected
"work-queue"; update the namespace value(s) at those lines to the canonical
"work-queue" so all manifests/dashboards use the same namespace and avoid 404s.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Namespaces now match the `work-queue` convention in both the rule and dashboard manifests. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Namespaces now match the `work-queue` convention in both the rule and dashboard manifests. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:873
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/admin-api/monitoring.yaml around lines 58-65, the alert divides

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572072

- [review_comment] 2025-09-16T03:17:29Z by coderabbitai[bot] (deployments/admin-api/monitoring.yaml:65)

```text
In deployments/admin-api/monitoring.yaml around lines 58-65, the alert divides
by container_spec_memory_limit_bytes without matching labels which creates
cardinality/vector-matching issues; change the denominator to the
kube-state-metrics memory limit metric (e.g.
kube_pod_container_resource_limits_bytes{resource="memory"} or the equivalent
kube_pod_container_resource_limits{resource="memory", unit="byte"}) and perform
an explicit vector match so the usage and limit align (for example use
on(namespace,pod,container) or include identical pod/container selectors), or
remove the alert until you can implement correct label-matched limits.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Alert now joins usage with kube-state-metrics limits and guards against zero values. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Alert now joins usage with kube-state-metrics limits and guards against zero values. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:907
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/admin-api/monitoring.yaml around lines 82 to 99 (and also lines

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572075

- [review_comment] 2025-09-16T03:17:29Z by coderabbitai[bot] (deployments/admin-api/monitoring.yaml:99)

```text
In deployments/admin-api/monitoring.yaml around lines 82 to 99 (and also lines
~118-125), PromQL label matchers use single quotes (e.g. {job='admin-api'})
which is invalid; update every PromQL target in this file to use double quotes
for label values (e.g. {job="admin-api"}), including status regexes and any
other label matchers, and search/replace all occurrences across the file so all
targets use double-quoted label values.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | JSON dashboard targets now escape double quotes for all PromQL label matchers. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> JSON dashboard targets now escape double quotes for all PromQL label matchers. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:946
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/admin-api/monitoring.yaml around line 128, the file currently

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572077

- [review_comment] 2025-09-16T03:17:29Z by coderabbitai[bot] (deployments/admin-api/monitoring.yaml:128)

```text
In deployments/admin-api/monitoring.yaml around line 128, the file currently
lacks a trailing newline at EOF; open the file and add a single newline
character at the end (ensure the file ends with a single newline), then save so
the file ends with a proper newline.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Verified the manifest ends with a single newline to satisfy linters. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Verified the manifest ends with a single newline to satisfy linters. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:982
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/kubernetes/monitoring.yaml around lines 1 to 17, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572079

- [review_comment] 2025-09-16T03:17:29Z by coderabbitai[bot] (deployments/kubernetes/monitoring.yaml:17)

```text
In deployments/kubernetes/monitoring.yaml around lines 1 to 17, the
ServiceMonitor resource is using the wrong apiVersion; replace "apiVersion: v1"
with "apiVersion: monitoring.coreos.com/v1" so the ServiceMonitor CRD is
recognized, then validate the manifest (kubectl apply --dry-run=client or
kubectl apply) and ensure the Prometheus Operator CRDs are installed in the
cluster before applying.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | ServiceMonitor now declares `monitoring.coreos.com/v1` to match the CRD. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> ServiceMonitor now declares `monitoring.coreos.com/v1` to match the CRD. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:1016
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/kubernetes/monitoring.yaml around lines 72-84, the alert

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572084

- [review_comment] 2025-09-16T03:17:29Z by coderabbitai[bot] (deployments/kubernetes/monitoring.yaml:84)

```text
In deployments/kubernetes/monitoring.yaml around lines 72-84, the alert
expression divides container_memory_usage_bytes by
container_spec_memory_limit_bytes with mismatched labels and can divide by zero;
replace it to use the kube-state-metrics limits metric (e.g.
kube_pod_container_resource_limits_bytes or kube_pod_container_resource_limits)
and perform a proper vector match by namespace/pod/container (or use
on(namespace,pod,container) group_left if needed) and guard against zero limits
by filtering the limit metric to > 0 (or applying clamp_min(limit,1)) before
division so the resulting ratio is valid and safe for comparison.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Alert now divides usage by a namespace-scoped kube-state memory limit with clamp_min safeguards. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Alert now divides usage by a namespace-scoped kube-state memory limit with clamp_min safeguards. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:1052
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/kubernetes/monitoring.yaml around lines 96-107, the rule

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572088

- [review_comment] 2025-09-16T03:17:30Z by coderabbitai[bot] (deployments/kubernetes/monitoring.yaml:107)

```text
In deployments/kubernetes/monitoring.yaml around lines 96-107, the rule
currently uses a boolean expr so $value becomes 0/1 and the annotation shows
garbage; change the rule to (1) filter by namespace, (2) keep a boolean expr to
fire the alert:
(certmanager_certificate_expiration_timestamp_seconds{name="admin-api-tls",namespace="your-namespace"}
- time()) < 7*24*3600, and (3) update the annotation to display the
time-to-expiry by evaluating the time-left expression, e.g. use {{
humanizeDuration
(certmanager_certificate_expiration_timestamp_seconds{name="admin-api-tls",namespace="your-namespace"}
- time()) }} so the annotation shows remaining time instead of 0/1.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Alert now scopes to `namespace="work-queue"` and surfaces the remaining TTL via `humanizeDuration`. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Alert now scopes to `namespace="work-queue"` and surfaces the remaining TTL via `humanizeDuration`. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:1091
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/kubernetes/monitoring.yaml around lines 109 to 119, the alert

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572094

- [review_comment] 2025-09-16T03:17:30Z by coderabbitai[bot] (deployments/kubernetes/monitoring.yaml:119)

```text
In deployments/kubernetes/monitoring.yaml around lines 109 to 119, the alert
uses the metric http_requests_rate_limited_total which conflicts with other
files using rate_limit_exceeded_total; pick one canonical metric name (e.g.,
rate_limit_exceeded_total) and update this alert to use that name, then search
the repo for the other occurrences and replace them to the chosen name so all
rules/dashboards/queries are consistent; ensure the metric label set and any
recording rules or exporters that emit the metric are updated to match the new
canonical name and run a quick lint/test of Prometheus rules to verify no
references remain to the old name.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Alert and dashboard now reference the canonical `rate_limit_exceeded_total` metric. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Alert and dashboard now reference the canonical `rate_limit_exceeded_total` metric. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:1131
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/kubernetes/monitoring.yaml around line 220, the file is missing a

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572101

- [review_comment] 2025-09-16T03:17:30Z by coderabbitai[bot] (deployments/kubernetes/monitoring.yaml:220)

```text
In deployments/kubernetes/monitoring.yaml around line 220, the file is missing a
trailing newline at EOF; add a single newline character at the end of the file
so the final line terminator is present and the file ends with a newline.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 6 | Verified the manifest now ends with a single newline to satisfy linters. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Verified the manifest now ends with a single newline to satisfy linters. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:28
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/README.md around lines 39-44 (and also update occurrences at

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572107

- [review_comment] 2025-09-16T03:17:30Z by coderabbitai[bot] (deployments/README.md:44)

```text
In deployments/README.md around lines 39-44 (and also update occurrences at
165-170), the docs inconsistently reference /health, /healthz and /readyz;
choose the Kubernetes conventions and make them consistent: use /healthz for
liveness and /readyz for readiness across the entire document, update the curl
examples and any runbook references accordingly, and verify no other places
still reference /health (or swap meanings) so probes and examples match the
chosen endpoints.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | README and deployment probes now consistently use `/healthz` and `/readyz`. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> README and deployment probes now consistently use `/healthz` and `/readyz`. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:61
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/README.md around lines 80 to 88, the README claims a rate limit

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572110

- [review_comment] 2025-09-16T03:17:30Z by coderabbitai[bot] (deployments/README.md:88)

```text
In deployments/README.md around lines 80 to 88, the README claims a rate limit
of “100 rpm, burst 20” but no environment variables or flags are documented to
control those values; add explicit env config entries for the rate limiter
(e.g., RATE_LIMIT_RPM, RATE_LIMIT_BURST, optionally RATE_LIMIT_WINDOW_SECONDS
and RATE_LIMIT_ENABLED) with clear descriptions and sensible defaults (100, 20,
60, true), specify expected types (integer/boolean), and note that the
application should read these envs to configure the limiter; update the table to
include these variables, their descriptions and defaults so users can actually
tune the rate limiting behavior.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Rate limiter environment variables and defaults are now documented. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Rate limiter environment variables and defaults are now documented. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:98
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/README.md around lines 93 to 100, the current example uses

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572115

- [review_comment] 2025-09-16T03:17:30Z by coderabbitai[bot] (deployments/README.md:100)

```text
In deployments/README.md around lines 93 to 100, the current example uses
kubectl create secret ... --from-literal with real tokens which leaks secrets to
shell history and CI logs; update the docs to instruct creating secrets from
files or stdin (store tokens in files with restrictive permissions or pass via
pipe/stdin), or generate a YAML manifest with kubectl --dry-run=client -o yaml
and apply that manifest, and emphasise not to paste secrets directly into
shell/CI logs and to use environment variables or secret managers in CI.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | All admin API manifests and deployment defaults now point to the `work-queue` namespace. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> All admin API manifests and deployment defaults now point to the `work-queue` namespace. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:174
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/README.md around lines 95–100 the secret is created with -n

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572119

- [review_comment] 2025-09-16T03:17:30Z by coderabbitai[bot] (deployments/README.md:100)

```text
In deployments/README.md around lines 95–100 the secret is created with -n
work-queue but other admin-api manifests use -n redis-work-queue, causing
namespace drift; pick the cluster's canonical namespace (confirm with kubectl
get namespaces or team docs) and make it consistent across all affected files:
either change this README secret to -n redis-work-queue, or update every
manifest/script that uses redis-work-queue to work-queue. Specifically update
deployments/README.md; deployments/admin-api/* (k8s-deployment.yaml,
k8s-redis.yaml, monitoring.yaml, deploy.sh); deployments/kubernetes/*
(admin-api-deployment.yaml, rbac-token-service-deployment.yaml,
monitoring.yaml); deployments/scripts/* (setup-monitoring.sh,
test-staging-deployment.sh); and update docs/monitoring to reflect the chosen
namespace. Ensure all kubectl commands, YAML metadata.namespace fields, and Helm
values (if any) match the chosen namespace and run a dry validation (kubectl
apply --server-dry-run or kubeval) to confirm no mismatches remain.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | All admin API manifests and deployment defaults now point to the `work-queue` namespace. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> All admin API manifests and deployment defaults now point to the `work-queue` namespace. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:174
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/00_assessment.md around lines 20–21, the doc currently pins "go-redis

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572129

- [review_comment] 2025-09-16T03:17:31Z by coderabbitai[bot] (docs/00_assessment.md:21)

```text
In docs/00_assessment.md around lines 20–21, the doc currently pins "go-redis
v8" — update wording to "go-redis v9 (v9+)" and add a concise migration
checklist that lists the breaking changes to account for (pipelines not
thread-safe; changed timeout/cancel behavior; Pipeline.Close/WithContext
removed; option renames MaxConnAge→ConnMaxLifetime and
IdleTimeout→ConnMaxIdleTime; connection reaper removal in favor of MaxIdleConns;
redis.Z type change from *Z→Z; reworked hooks API including DialHook; RESP3
behavior differences). Also append a short upgrade plan (steps: upgrade
dependency in lockfile, run tests, audit pipeline usages and option names,
update hooks and types, run performance/RESP3 smoke tests) plus a rollback
justification (pin v8 in lockfile and revert dependency if issues) or explicit
justification for staying on v8. Ensure the text stays brief, actionable, and
includes a checklist and upgrade/rollback plan.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Assessment now highlights go-redis v9 and documents migration/rollback steps. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Assessment now highlights go-redis v9 and documents migration/rollback steps. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:255
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/07_test_plan.md around lines 27 to 29, replace the vague "Chaos (where

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572137

- [review_comment] 2025-09-16T03:17:31Z by coderabbitai[bot] (docs/07_test_plan.md:29)

```text
In docs/07_test_plan.md around lines 27 to 29, replace the vague "Chaos (where
feasible in CI)" note with a deterministic list of fault-injection scenarios and
explicit pass/fail criteria: enumerate concrete failures (e.g., Redis SIGSTOP
for 30s, introduce 200ms p95 latency using tc netem on Redis port, inject 5% TCP
connection resets via iptables or tc loss), provide exact commands or CI steps
to run each injection (so they can be replayed in CI or locally), define how
long each injection should run and the sequence/timing, and state clear
pass/fail criteria for each (e.g., service stays healthy, no data loss, retries
succeed within X seconds, error rate < Y%) so reviewers and CI can
deterministically validate outcomes.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Chaos section now lists explicit failures, commands, durations, and pass/fail criteria. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Chaos section now lists explicit failures, commands, durations, and pass/fail criteria. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:298
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/07_test_plan.md around lines 41 to 45, the benchmark notes lack

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572146

- [review_comment] 2025-09-16T03:17:31Z by coderabbitai[bot] (docs/07_test_plan.md:45)

```text
In docs/07_test_plan.md around lines 41 to 45, the benchmark notes lack
reproducibility details; update the test plan to pin the GH runner size (exact
VM type/VM image), specify the exact Go version used, set and document
GOMAXPROCS (and recommend exporting it in the runner), fix and export the RNG
seed used by the synthetic job producer, document CPU governor/settings used
during runs, and ensure each run prints the commit SHA (and any build
tags/flags) and exports the seed so results can be reproduced; include these
fields in the Reporting section so CPU/memory/Redis/queue metrics are captured
alongside runner size, GOMAXPROCS, seed, CPU governor, and Go version.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Benchmark section now specifies runner type, Go version, GOMAXPROCS, seed, and reporting fields. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Benchmark section now specifies runner type, Go version, GOMAXPROCS, seed, and reporting fields. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:338
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/anomaly-radar-slo-budget.md around lines 124 to 176, the duration

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572177

- [review_comment] 2025-09-16T03:17:32Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:176)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 124 to 176, the duration
field "window" is ambiguous as currently shown ("720h0m0s"); update the docs to
explicitly state that durations are encoded as Go time.Duration strings and list
accepted formats (e.g., "72h", "720h0m0s", "30m", "1h30m", "1500ms"), include a
short note about parsing behavior (supports negative values and sub-second units
like "1500ms" or "1.5s"), and add one or two alternate example values in the
JSON response to demonstrate valid variants so downstream clients know how to
parse them.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Added explicit RFC3339/UTC notes under every response example. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added explicit RFC3339/UTC notes under every response example. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:808
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/anomaly-radar-slo-budget.md around lines 186 to 217, the config

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572224

- [review_comment] 2025-09-16T03:17:33Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:217)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 186 to 217, the config
endpoint documentation is missing required Authentication/Authorization details
for this admin surface; update the endpoint docs to include an
"Authentication/Authorization" block that states the endpoint requires admin
scope and specifies the auth scheme (Bearer JWT) consistent with the Admin API
docs, include required header (Authorization: Bearer <token>), required
roles/scopes (e.g., "admin" or specific scope name used by the Admin API), and a
short example note on denied responses (401/403) so readers know auth is
mandatory.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Endpoint doc now calls out Bearer auth requirements and 401/403 behaviours. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Endpoint doc now calls out Bearer auth requirements and 401/403 behaviours. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:415
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/canary-deployments.md around lines 15 to 19, the authentication

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572234

- [review_comment] 2025-09-16T03:17:33Z by coderabbitai[bot] (docs/api/canary-deployments.md:19)

```text
In docs/api/canary-deployments.md around lines 15 to 19, the authentication
headers listed are inconsistent with the rest of the repo (Bearer + X-API-Key).
Update the section to standardize on Authorization: Bearer <token> as the
primary method, replace X-API-Key with the repository-preferred alias
X-API-Token (if an alias is needed), and add one sentence clarifying semantics
and precedence (Authorization header takes precedence over X-API-Token when both
are present). Ensure the examples and header names in this block are changed
accordingly.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Auth section now aligns with repo conventions (Bearer primary, `X-API-Token` alias). |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Auth section now aligns with repo conventions (Bearer primary, `X-API-Token` alias). Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:454
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/canary-deployments.md around lines 303–345, the JSON mixes units

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572245

- [review_comment] 2025-09-16T03:17:33Z by coderabbitai[bot] (docs/api/canary-deployments.md:345)

```text
In docs/api/canary-deployments.md around lines 303–345, the JSON mixes units
(error_rate shown as 0.96 while success_rate is 99.04) — pick a single
convention and make fields explicit; update the schema to use percentage fields
(rename error_rate → error_percent and success_rate → success_percent) and set
their values to (error_count/job_count)*100 and (success_count/job_count)*100
respectively (rounded consistently, e.g., two decimal places), and add a short
comment/inline note in the docs that these fields are percentages to avoid
future confusion.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Metrics examples now expose `error_percent`/`success_percent` and clarify 0–100 scaling. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Metrics examples now expose `error_percent`/`success_percent` and clarify 0–100 scaling. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:493
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/canary-deployments.md around lines 556 to 592, the Deployment Object

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572255

- [review_comment] 2025-09-16T03:17:34Z by coderabbitai[bot] (docs/api/canary-deployments.md:592)

```text
In docs/api/canary-deployments.md around lines 556 to 592, the Deployment Object
"id" field is underspecified; update the schema and prose to explicitly state
the expected ID format (e.g., "uuid" v4 or "ULID"), allowed characters and
length, give a concrete example, and indicate that the server validates IDs and
will return HTTP 400 on invalid IDs; add a short note on the validation rules
(regex or exact format), whether IDs are case-sensitive, and link to the API
error response shape for 400 validation errors.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Deployment object now shows a ULID example and documents validation regex + 400 behaviour. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Deployment object now shows a ULID example and documents validation regex + 400 behaviour. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:532
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/capacity-planning-api.md around lines 311 to 318, the import uses a

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572269

- [review_comment] 2025-09-16T03:17:34Z by coderabbitai[bot] (docs/api/capacity-planning-api.md:318)

```text
In docs/api/capacity-planning-api.md around lines 311 to 318, the import uses a
hardcoded placeholder module path "github.com/yourorg/..." which will mislead
users; update the import to either the repository's actual Go module path or
replace it with a neutral placeholder comment (e.g., // replace with your module
path) and show an example like module/path/to/automatic-capacity-planning so
readers know to substitute their own module path; ensure the docs clearly state
to replace the placeholder with the user's real module path.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Import example now references the canonical repo path with guidance to swap in local module names. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Import example now references the canonical repo path with guidance to swap in local module names. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:607
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/exactly-once-admin.md around lines 299 to 321, replace the literal

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572287

- [review_comment] 2025-09-16T03:17:34Z by coderabbitai[bot] (docs/api/exactly-once-admin.md:321)

```text
In docs/api/exactly-once-admin.md around lines 299 to 321, replace the literal
"YOUR_TOKEN" used in the curl examples with a clearly identifiable placeholder
(e.g. <YOUR_API_TOKEN_HERE> or ${API_TOKEN}) and add a brief note in the
examples that this must be substituted with a real token or read from an
environment variable; update all four curl snippets to use the chosen
placeholder consistently so readers cannot copy-paste a literal string by
mistake.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Curl examples now require `${API_TOKEN}` placeholders with guidance to source real tokens securely. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Curl examples now require `${API_TOKEN}` placeholders with guidance to source real tokens securely. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:645
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/PRD.md around lines 162-168 the current metric definition

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572294

- [review_comment] 2025-09-16T03:17:34Z by coderabbitai[bot] (docs/PRD.md:168)

```text
In docs/PRD.md around lines 162-168 the current metric definition
queue_length{queue=...} risks unbounded label cardinality; update the spec to
require limiting/sanitizing queue labels by adding a new config entry (e.g.
metrics.allowed_queues: list of permitted queue names) and a fallback strategy
(hash or bucket remaining queues into a small fixed set or use a generic "other"
label), describe how the exporter/collector will emit metrics (only allowed
queue labels appear verbatim; disallowed names are hashed/bucketed or
aggregated), and include default config values and expected cardinality behavior
in the doc.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | PRD now documents `metrics.allowed_queues`, fallback hashing, and default cardinality controls. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> PRD now documents `metrics.allowed_queues`, fallback hashing, and default cardinality controls. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:682
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## `

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350574887

- [review_comment] 2025-09-16T03:18:41Z by coderabbitai[bot] (.claude/agents/feature-enhancer.md:12)

```text
`
In .claude/agents/feature-enhancer.md around lines 8 to 12 (and similar issues
at 20-36, 44-74, 80-126, 190-212, 252-269), the markdown has lint failures
(incorrect heading spacing, missing blank lines around fenced code blocks,
inconsistent fenced code ticks, trailing newline missing). Run make mdlint-fix
and commit the result; if fixing manually, ensure headings have a blank line
above, add a blank line before and after every
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Markdown headings now include required spacing and fenced blocks are separated properly. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Markdown headings now include required spacing and fenced blocks are separated properly. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:722
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In AGENTS.md around lines 183 to 193, replace the fake placeholder link

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350574894

- [review_comment] 2025-09-16T03:18:41Z by coderabbitai[bot] (AGENTS.md:193)

```text
In AGENTS.md around lines 183 to 193, replace the fake placeholder link
"https://fake.com" and any "(to PR)" text with the actual PR or issue URLs (or
remove them entirely if no real links exist), and update the nonstandard
admonition syntax "[!info]-" to the proper GitHub-flavored blockquote admonition
format such as "> [!NOTE]" or "> [!WARNING]" (without the trailing hyphen),
ensuring links use valid markdown link syntax and no placeholder text remains.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Replaced placeholder links and converted admonitions to GitHub-style syntax. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Replaced placeholder links and converted admonitions to GitHub-style syntax. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:759
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In cmd/tui/main.go around line 31, the FlagSet is created with flag.ExitOnError

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350574900

- [review_comment] 2025-09-16T03:18:42Z by coderabbitai[bot] (cmd/tui/main.go:31)

```text
In cmd/tui/main.go around line 31, the FlagSet is created with flag.ExitOnError
which calls os.Exit on parse errors and bypasses your function's error handling;
change the creation to use flag.ContinueOnError so Parse returns an error
instead, then call fs.Parse(os.Args[1:]) and handle/return that error
consistently with the rest of the function (optionally set fs.SetOutput to
discard or a logger if you want to suppress default printed usage).
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Flag parsing now uses `ContinueOnError` with explicit error handling. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Flag parsing now uses `ContinueOnError` with explicit error handling. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:795
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In cmd/tui/main.go around line 43, the error returned by fs.Parse(os.Args[1:])

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350574908

- [review_comment] 2025-09-16T03:18:42Z by coderabbitai[bot] (cmd/tui/main.go:43)

```text
In cmd/tui/main.go around line 43, the error returned by fs.Parse(os.Args[1:])
is being discarded; instead capture the error, check if it's non-nil, write the
error message to stderr (and/or call fs.Usage()), and exit with a non-zero
status (e.g., os.Exit(2>) or return the error) so parsing failures are properly
reported and the process terminates.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Parse errors now propagate and abort the program with a clear message. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Parse errors now propagate and abort the program with a clear message. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:831
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In cmd/tui/main.go around lines 64-66, the code pings Redis and merely logs on

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350574913

- [review_comment] 2025-09-16T03:18:42Z by coderabbitai[bot] (cmd/tui/main.go:66)

```text
In cmd/tui/main.go around lines 64-66, the code pings Redis and merely logs on
error but continues running with a broken connection; change this to fail fast:
on Ping error write the error to stderr (or process logger) and exit the process
with a non-zero status (or return the error from main) so the TUI does not
proceed with an invalid Redis client; alternatively implement a small
retry/backoff loop before exiting if transient errors are expected, but do not
allow execution to continue when Ping ultimately fails.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Redis ping now exits on failure and CLI flags map into config/TUI options. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Redis ping now exits on failure and CLI flags map into config/TUI options. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:866
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In cmd/tui/main.go around line 68, the TODO leaves many CLI flags un-wired

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350574918

- [review_comment] 2025-09-16T03:18:42Z by coderabbitai[bot] (cmd/tui/main.go:68)

```text
In cmd/tui/main.go around line 68, the TODO leaves many CLI flags un-wired
(redisURL, cluster, namespace, readOnly, metricsAddr, theme, fps) so the TUI
starts without required runtime options; fix by reading these flags from the
root/Cobra command or shared config struct and passing them into the TUI
initializer: add parameters on the TUI options/ctor for
redisURL/cluster/namespace (used to construct the redis/backend client),
readOnly (toggle input/editing and backend write operations), metricsAddr
(start/forward metrics collection if non-empty), and theme/fps (apply to
renderer/refresh loop); ensure the flags are defined and defaulted at command
setup and validate/convert types before constructing the TUI, then remove the
TODO.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | TUI receives structured options (cluster, namespace, read-only, theme, fps, metrics addr) instead of ignoring flags. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TUI receives structured options (cluster, namespace, read-only, theme, fps, metrics addr) instead of ignoring flags. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:903
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/admin-api/deploy.sh around lines 1 to 10, the script uses weak

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350574922

- [review_comment] 2025-09-16T03:18:42Z by coderabbitai[bot] (deployments/admin-api/deploy.sh:10)

```text
In deployments/admin-api/deploy.sh around lines 1 to 10, the script uses weak
bash settings and unquoted variable expansions; update the shebang area to
enable strict mode by replacing the current set -e with set -Eeuo pipefail, and
ensure all variable usages in the script are quoted (e.g., use "$NAMESPACE",
"$APP_NAME", "${ENVIRONMENT}", "${VERSION}" wherever referenced) to prevent
word-splitting and undefined-variable errors; also ensure positional defaults
remain as shown but are safely assigned (retain ENVIRONMENT="${1:-staging}" and
VERSION="${2:-latest}").
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Script now runs with `set -Eeuo pipefail`, uses a safe IFS, and quotes variables. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Script now runs with `set -Eeuo pipefail`, uses a safe IFS, and quotes variables. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:944
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/admin-api/deploy.sh around lines 31 to 43, the docker tag/push

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350574938

- [review_comment] 2025-09-16T03:18:42Z by coderabbitai[bot] (deployments/admin-api/deploy.sh:43)

```text
In deployments/admin-api/deploy.sh around lines 31 to 43, the docker tag/push
steps use unquoted variable expansions and assume a good registry value which
can break with spaces, empty vars, or missing namespace; update to quote all
variable expansions (e.g. "${APP_NAME}" "${VERSION}" "${REGISTRY_URL}" ),
validate or default DOCKER_REGISTRY and DOCKER_NAMESPACE explicitly before use
(fail fast if missing), construct a fully qualified repository string like
"${REGISTRY_URL%/}/${DOCKER_NAMESPACE:-your-namespace}/${APP_NAME}:${VERSION}"
to avoid double slashes or implicit Docker Hub quirks, and use that quoted
repository value for both docker tag and docker push; also consider requiring
docker login before push (or check LOGIN env) and exit with an error if push
fails.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Image tagging/pushing now validates registry inputs and uses a normalized repository string. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Image tagging/pushing now validates registry inputs and uses a normalized repository string. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:982
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/admin-api/deploy.sh around lines 49 to 61, the script creates the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350574945

- [review_comment] 2025-09-16T03:18:43Z by coderabbitai[bot] (deployments/admin-api/deploy.sh:61)

```text
In deployments/admin-api/deploy.sh around lines 49 to 61, the script creates the
${NAMESPACE} but then applies manifests without specifying that namespace;
update the kubectl apply invocations to target the created namespace by adding
-n ${NAMESPACE} (or --namespace=${NAMESPACE}) to the redis and admin-api apply
commands so both kubectl apply -f deployments/admin-api/k8s-redis.yaml and
kubectl apply -f deployments/admin-api/k8s-deployment.yaml run against the
intended namespace; keep the existing kubectl wait which already uses -n
${NAMESPACE}.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | `kubectl apply` now targets the intended namespace for Redis and Admin API manifests. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> `kubectl apply` now targets the intended namespace for Redis and Admin API manifests. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:1023
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/admin-api/deploy.sh around lines 73-85 (and also apply same

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350574957

- [review_comment] 2025-09-16T03:18:43Z by coderabbitai[bot] (deployments/admin-api/deploy.sh:85)

```text
In deployments/admin-api/deploy.sh around lines 73-85 (and also apply same
change to lines 113-116), the script currently starts a kubectl port-forward in
the "local" branch without ensuring it is cleaned up and also runs port-forward
in Docker "local" mode; update the script so it only performs kubectl
port-forward when running against a k8s cluster (not when ENVIRONMENT indicates
Docker/local-Docker), capture the background PID immediately, set a trap on EXIT
(and SIGINT/SIGTERM) to kill the port-forward if it was started, and ensure any
sleep/wait uses a small readiness check rather than a blind sleep; apply the
identical guard+trap pattern to the other location (lines 113-116).
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Port-forwarding now runs only during k8s smoke tests and is cleaned up via trap. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Port-forwarding now runs only during k8s smoke tests and is cleaned up via trap. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:1061
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/admin-api/deploy.sh around lines 170-177, replace the deprecated

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350574963

- [review_comment] 2025-09-16T03:18:43Z by coderabbitai[bot] (deployments/admin-api/deploy.sh:177)

```text
In deployments/admin-api/deploy.sh around lines 170-177, replace the deprecated
docker-compose usage with the modern `docker compose` subcommand and add a
pre-check that verifies `docker compose` is available: implement a small check
(e.g., run `docker compose version` or `command -v docker` + `docker compose
version` and test its exit code) before the local deploy branch and if the check
fails print a clear error and exit non-zero; then change the local case to call
`docker compose -f deployments/admin-api/docker-compose.yaml up -d` (or bail if
unavailable).
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Local deploy now requires `docker compose` and invokes the modern subcommand. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Local deploy now requires `docker compose` and invokes the modern subcommand. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:1100
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/admin-api/docker-compose.yaml around line 28, the JWT_SECRET is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350574973

- [review_comment] 2025-09-16T03:18:43Z by coderabbitai[bot] (deployments/admin-api/docker-compose.yaml:28)

```text
In deployments/admin-api/docker-compose.yaml around line 28, the JWT_SECRET is
set to a weak default which is insecure; remove the hardcoded fallback so the
environment variable must be provided (e.g. change JWT_SECRET=${JWT_SECRET} with
no default), add a tiny validation service or init container that checks
process.env.JWT_SECRET and exits non‑zero if missing, and update the admin-api
service to depend_on that validation service so the stack fails fast when
JWT_SECRET is not supplied; ensure local/dev docs or .env.example instructs how
to provide a strong secret.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Compose stack now enforces `JWT_SECRET` via an env-check container and no longer ships a weak default. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Compose stack now enforces `JWT_SECRET` via an env-check container and no longer ships a weak default. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:1138
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/docker/docker-compose.yaml around lines 36-38 (and also update

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350574996

- [review_comment] 2025-09-16T03:18:43Z by coderabbitai[bot] (deployments/docker/docker-compose.yaml:38)

```text
In deployments/docker/docker-compose.yaml around lines 36-38 (and also update
similar occurrences at 45-49 and 74-77), the RBAC mount is ./rbac-configs ->
/app/rbac but the env vars point to /app/configs/*.yaml and API_TOKEN_* values
are inlined; change the RBAC env vars to point to the actual mounted path
(/app/rbac/roles.yaml and /app/rbac/resources.yaml) and remove any hard-coded
API_TOKEN_* from the compose file, instead reference an env_file (admin-api.env)
or Docker secrets; update service definitions to load admin-api.env (ensure that
file is .gitignored) or declare docker secrets and replace API_TOKEN_* entries
with secret references.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Verified the compose stack now reads RBAC configs from the mounted path and loads API tokens via env_file. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Verified the compose stack now reads RBAC configs from the mounted path and loads API tokens via env_file. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:66
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/docker/docker-compose.yaml around lines 121 to 123, the file is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575000

- [review_comment] 2025-09-16T03:18:44Z by coderabbitai[bot] (deployments/docker/docker-compose.yaml:123)

```text
In deployments/docker/docker-compose.yaml around lines 121 to 123, the file is
missing a trailing newline at EOF and the review flagged potential
incompatibility with depends_on health condition; add a single newline character
at the end of the file, and if you rely on depends_on: condition:
service_healthy ensure each dependent service defines a proper healthcheck and
verify behaviour against your Compose implementation (run docker compose
version) — if incompatibility exists, either remove the condition and gate
startup in the service entrypoint or migrate to a Compose version that supports
health conditions.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Compose v2.39.2 confirmed to honour service_healthy gating and the manifest now carries the required newline. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Compose v2.39.2 confirmed to honour service_healthy gating and the manifest now carries the required newline. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:105
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/docker/Dockerfile.admin-api around lines 52-53, the HEALTHCHECK

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575007

- [review_comment] 2025-09-16T03:18:44Z by coderabbitai[bot] (deployments/docker/Dockerfile.admin-api:53)

```text
In deployments/docker/Dockerfile.admin-api around lines 52-53, the HEALTHCHECK
uses wget which is inconsistent with other Dockerfiles; replace the wget
invocation with an equivalent curl-based healthcheck command (use curl options
to fail on non-2xx, be silent, set a timeout, and return non-zero on failure)
and ensure the Dockerfile installs curl in the runtime image (add curl to the
runtime package list or use a base image that includes curl) so the healthcheck
works and matches project convention.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Runtime layer now installs curl and the healthcheck aligns with other images. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Runtime layer now installs curl and the healthcheck aligns with other images. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:144
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/docker/Dockerfile.rbac-token-service around lines 39-40, the COPY

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575011

- [review_comment] 2025-09-16T03:18:44Z by coderabbitai[bot] (deployments/docker/Dockerfile.rbac-token-service:40)

```text
In deployments/docker/Dockerfile.rbac-token-service around lines 39-40, the COPY
commands currently suppress failures with "2>/dev/null || true", which hides
missing config files; decide whether these configs are required or optional and
remove the error-suppression accordingly: if required, delete the redirections
and "|| true" so the build fails loudly when files are missing and update build
artifacts to include the files; if optional, remove the suppressions and instead
implement an explicit presence check (or adjust the application to handle
missing configs) and use a conditional copy strategy in the Docker build
pipeline (e.g., stage the files only when present) so missing files are handled
intentionally rather than silently ignored.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Build now fails loudly if RBAC config bundle is missing. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Build now fails loudly if RBAC config bundle is missing. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:181
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/docker/rbac-configs/resources.yaml around line 231, the file is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575018

- [review_comment] 2025-09-16T03:18:44Z by coderabbitai[bot] (deployments/docker/rbac-configs/resources.yaml:231)

```text
In deployments/docker/rbac-configs/resources.yaml around line 231, the file is
missing a trailing newline; add a single newline character at the end of the
file so the last line ("GET /api/v1/admin/audit": ["admin:audit"]) is terminated
by a newline to satisfy linters and POSIX file conventions.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Endpoint map now mirrors the Admin API contract and includes metrics/health routes. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Endpoint map now mirrors the Admin API contract and includes metrics/health routes. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:142
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/docker/rbac-configs/roles.yaml around lines 95 to 102, add a

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575024

- [review_comment] 2025-09-16T03:18:44Z by coderabbitai[bot] (deployments/docker/rbac-configs/roles.yaml:102)

```text
In deployments/docker/rbac-configs/roles.yaml around lines 95 to 102, add a
terminating newline at EOF to satisfy YAML linters, and update the file to
explicitly document how role_assignment_rules are enforced and their precedence:
state which component (token service, API gateway, or other) applies these
rules, describe whether matching is based on token claims, user attributes, or
group memberships, and clarify precedence order between domain-based rules and
explicit user/group mappings (e.g., user/group overrides domain rules or vice
versa). Ensure the documentation snippet is concise, placed near the
role_assignment_rules section, and includes an example showing resolution order
so readers know which mapping wins.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Comments now explain how domain rules interact with explicit mappings and newline is restored. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Comments now explain how domain rules interact with explicit mappings and newline is restored. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:255
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/docker/rbac-configs/token-service.yaml around line 28, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575032

- [review_comment] 2025-09-16T03:18:44Z by coderabbitai[bot] (deployments/docker/rbac-configs/token-service.yaml:28)

```text
In deployments/docker/rbac-configs/token-service.yaml around line 28, the
encryption_key currently falls back to the insecure default
"default-dev-key-change-in-production"; remove that hardcoded default so the
value must come from RBAC_KEY_ENCRYPTION_KEY (e.g. use the env var without a
default), and add startup validation that requires RBAC_KEY_ENCRYPTION_KEY (and
REDIS_PASSWORD) and fails fast on missing vars; implement a small
startup_validation section that lists required_env_vars and fail_fast: true so
the service refuses to start when those secrets are not provided.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Config now insists on external secrets and bails if they are missing. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Config now insists on external secrets and bails if they are missing. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:295
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/docker/rbac-configs/token-service.yaml around lines 72-75 the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575036

- [review_comment] 2025-09-16T03:18:44Z by coderabbitai[bot] (deployments/docker/rbac-configs/token-service.yaml:75)

```text
In deployments/docker/rbac-configs/token-service.yaml around lines 72-75 the
CORS allowed_origins contains hardcoded staging.example.com and example.com
domains which are environment-specific; replace these hardcoded entries with a
generated list sourced from an environment variable (e.g., ALLOWED_ORIGINS) or a
secret/config map so the manifest is templated at deploy time; update the
manifest/template to read a comma-separated or YAML/JSON list from the env var
(or ConfigMap) and document the required env var format so different
environments can supply their own allowed origins.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | CORS origins now come from env vars, eliminating hardcoded staging/prod hosts. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> CORS origins now come from env vars, eliminating hardcoded staging/prod hosts. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:333
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/docker/rbac-configs/token-service.yaml around line 114, the file

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575042

- [review_comment] 2025-09-16T03:18:45Z by coderabbitai[bot] (deployments/docker/rbac-configs/token-service.yaml:114)

```text
In deployments/docker/rbac-configs/token-service.yaml around line 114, the file
is missing a trailing newline; add a single newline character at the end of the
file so the final line ("content_security_policy: \"default-src 'self'\"") is
terminated with a proper line ending and save the file to ensure POSIX-compliant
EOF newline is present.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 10 | File now ends with a proper newline as required by linters. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> File now ends with a proper newline as required by linters. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:371
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/README-RBAC-Deployment.md around lines 19 to 36, replace the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575049

- [review_comment] 2025-09-16T03:18:45Z by coderabbitai[bot] (deployments/README-RBAC-Deployment.md:36)

```text
In deployments/README-RBAC-Deployment.md around lines 19 to 36, replace the
ASCII art block with a precise architecture section: remove the diagram and
instead add clear, textual and/or tool-compatible diagram content that documents
network topology (node/pod/service relationships and ingress/egress paths),
security boundaries (namespaces, network policies, RBAC roles and token flows),
data flow diagrams (request/response paths between Admin API, RBAC Token
Service, Redis and monitoring stack), exact port configurations (service and
container ports, protocols), and explicit service dependencies and startup
order; if a visual diagram is required, include a PlantUML or Mermaid diagram
snippet and link to any external draw.io/diagrams file so the doc is
machine-readable and appropriate for production use.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | ASCII art replaced with a detailed, tool-friendly architecture walkthrough. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> ASCII art replaced with a detailed, tool-friendly architecture walkthrough. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:406
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/README-RBAC-Deployment.md around lines 138-146, the env var table

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575055

- [review_comment] 2025-09-16T03:18:45Z by coderabbitai[bot] (deployments/README-RBAC-Deployment.md:146)

```text
In deployments/README-RBAC-Deployment.md around lines 138-146, the env var table
exposes secret defaults (Redis password, RBAC_SIGNING_KEY, RBAC_ENCRYPTION_KEY);
remove any hint of actual values and replace the Default column for these
entries with a neutral placeholder such as "REDACTED — provide via secret" or
"must be set from a secret", and add a short note (one line) on how to reference
them (e.g., via Kubernetes Secret, Docker secret, or env var injection) without
printing values; do not include actual keys or passwords anywhere in the README.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Table now points deployers to secrets instead of leaking representative values. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Table now points deployers to secrets instead of leaking representative values. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:447
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/README-RBAC-Deployment.md around lines 257 to 266, the "STOP at

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575064

- [review_comment] 2025-09-16T03:18:45Z by coderabbitai[bot] (deployments/README-RBAC-Deployment.md:266)

```text
In deployments/README-RBAC-Deployment.md around lines 257 to 266, the "STOP at
staging" note leaves the production deployment section incomplete; replace it
with a full production deployment procedure (or remove the section if not
intended) by adding concrete, ordered steps covering: blue‑green (or canary)
deployment workflow and tooling, detailed rollback and emergency rollback
procedures, database migration plan with safe migration strategy and backups,
production-specific security hardening and approvals, load balancing and scaling
configuration, load and smoke testing acceptance criteria, monitoring/alerting
runbook and SLOs, disaster recovery and backup verification steps, and required
pre/post‑deployment checklist and signoff process so the file contains an
actionable, auditable production deployment guide.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Production section now contains a full, auditable runbook. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Production section now contains a full, auditable runbook. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:484
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/advanced-rate-limiting-api.md around lines 140–155, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575072

- [review_comment] 2025-09-16T03:18:45Z by coderabbitai[bot] (docs/api/advanced-rate-limiting-api.md:155)

```text
In docs/api/advanced-rate-limiting-api.md around lines 140–155, the
FairnessConfig docs lack a guardrail explaining that cumulative minimum shares
can exceed 100%; update the docs to state the constraint "Sum(Weights) *
MinGuaranteedShare must be < 1.0" and add a clear fallback behavior: when the
constraint is violated the implementation will renormalize and cap minimum
guarantees (proportionally scale down per-priority MinGuaranteedShare so the
total < 1.0 and ensure no negative/greater-than-1 values), and mention any
rounding or epsilon used; keep the wording concise and add an example or single
sentence showing renormalization occurs when exceeded.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Docs now spell out the MinGuaranteedShare guardrail and fallback behaviour. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Docs now spell out the MinGuaranteedShare guardrail and fallback behaviour. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:525
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/anomaly-radar-slo-budget.md around lines 36 to 38, the Go import

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575092

- [review_comment] 2025-09-16T03:18:46Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:38)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 36 to 38, the Go import
uses the path
"github.com/flyingrobots/go-redis-work-queue/internal/anomaly-radar-slo-budget"
but the code references the package as anomalyradarslobudget which will not
compile; update the import to use an explicit alias (e.g. anomalyradarslobudget
"github.com/flyingrobots/go-redis-work-queue/internal/anomaly-radar-slo-budget")
and ensure the target package file declares package anomalyradarslobudget
(rename the package identifier if it currently contains hyphens or a different
name).
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Added a public re-export at `pkg/anomaly-radar-slo-budget` and updated docs to reference the stable path. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added a public re-export at `pkg/anomaly-radar-slo-budget` and updated docs to reference the stable path. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:637
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/calendar-view.md around lines 39 to 45, the CalendarView struct uses

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575102

- [review_comment] 2025-09-16T03:18:46Z by coderabbitai[bot] (docs/api/calendar-view.md:45)

```text
In docs/api/calendar-view.md around lines 39 to 45, the CalendarView struct uses
*time.Location for the timezone which does not JSON-marshal as a simple string;
change the Timezone field to a string (e.g., Timezone string `json:"timezone"` )
and document/expect an IANA timezone name (or offset) in requests/responses,
then update the Go example below to use Timezone: "UTC" (or another string)
instead of Timezone: "UTC" as a *time.Location value so clients receive a plain
string.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Calendar view now expects string timezones, matching JSON clients. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Calendar view now expects string timezones, matching JSON clients. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:643
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/calendar-view.md around lines 193 to 241, the bulk reschedule

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575110

- [review_comment] 2025-09-16T03:18:46Z by coderabbitai[bot] (docs/api/calendar-view.md:241)

```text
In docs/api/calendar-view.md around lines 193 to 241, the bulk reschedule
endpoint lacks guardrails and transactional behavior details; update the docs to
state a configurable max items per bulk request (default 100), describe
server-side rate limit expectations, and require clients to paginate or batch
larger payloads; specify that the response returns an overall transaction id and
a per-item status array (success/failure with error reasons), and document how
partial failures are handled (best-effort rollback when feasible, otherwise
leave successful items committed and mark failed ones with retry guidance) so
clients can implement safe retry/backoff.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Bulk reschedule docs now define limits, transaction IDs, and partial failure semantics. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Bulk reschedule docs now define limits, transaction IDs, and partial failure semantics. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:680
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/calendar-view.md around lines 575 to 582, the table exposes numeric

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575116

- [review_comment] 2025-09-16T03:18:46Z by coderabbitai[bot] (docs/api/calendar-view.md:582)

```text
In docs/api/calendar-view.md around lines 575 to 582, the table exposes numeric
enum values in public JSON; change the public representation to use string
status values instead of numbers and document the allowed string set (scheduled,
running, completed, failed, canceled). Update the table and any JSON examples to
show "status":"scheduled"|"running"|"completed"|"failed"|"canceled" and add a
note that the server may map these strings to internal integer constants. Ensure
examples and descriptions consistently use the string names and remove raw
numeric examples.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Calendar view docs now use string statuses/view types to match JSON payloads. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Calendar view docs now use string statuses/view types to match JSON payloads. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:719
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/distributed-tracing-integration.md around lines 41 to 50, the sample

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575123

- [review_comment] 2025-09-16T03:18:46Z by coderabbitai[bot] (docs/api/distributed-tracing-integration.md:50)

```text
In docs/api/distributed-tracing-integration.md around lines 41 to 50, the sample
YAML sets insecure: true which encourages unsafe production usage; change the
example to show insecure: false by default and demonstrate TLS first (e.g., use
otlp+https or a TLS endpoint URL and set insecure: false), and update the text
to flip the recommendation order (TLS first, then opt-in insecure for local
dev). Additionally add a short note stating the system should fail fast when the
configured endpoint appears to require TLS (e.g., uses https or known TLS port)
but insecure: true is set, instructing implementers to validate endpoint
scheme/port and return a clear error if that mismatch is detected.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Default tracing config now prefers TLS and documents the safety check. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Default tracing config now prefers TLS and documents the safety check. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:757
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/distributed-tracing-integration.md around lines 348–359, the example

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575137

- [review_comment] 2025-09-16T03:18:47Z by coderabbitai[bot] (docs/api/distributed-tracing-integration.md:359)

```text
In docs/api/distributed-tracing-integration.md around lines 348–359, the example
shell commands use macOS-only tools (pbcopy, open); update the doc to either
detect and show cross-platform equivalents or explicitly flag the OS: for
copying, mention pbcopy (macOS), xclip/xsel (Linux) and clip or PowerShell
alternatives (Windows); for opening URLs, mention open (macOS), xdg-open (Linux)
and start or PowerShell Invoke-Item (Windows); add a short note explaining how
to detect OS in a script (or that users should pick the command for their OS) so
readers on Linux/Windows can follow the steps.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Added cross-platform clipboard/URL commands so instructions work beyond macOS. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added cross-platform clipboard/URL commands so instructions work beyond macOS. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:796
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## docs/SLAPS/worker-reflections/claude-001-reflection.md lines 39-43: the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575201

- [review_comment] 2025-09-16T03:18:49Z by coderabbitai[bot] (docs/SLAPS/worker-reflections/claude-001-reflection.md:43)

```text
docs/SLAPS/worker-reflections/claude-001-reflection.md lines 39-43: the
“Metrics-Disabled Testing” section is too vague — replace the hand‑waving with
an exact description of the toggle and patterns used to avoid Prometheus
duplicate-collector panics: state the config key or env var name (e.g.,
METRICS_ENABLED=false), describe the code guard used (e.g., wrap MustRegister
calls with a check or use a test-only conditional registry), and list the
concrete alternatives tried (per-test Registry via prometheus.NewRegistry(),
using promtest helpers, or using registerOnce/guarded register functions).
Include the exact snippet locations where tests should set the switch (e.g.,
test setup/teardown) and a short note on preferred approach for the repo.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Reflection now documents the exact metrics toggle, registry strategy, and alternatives tested. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Reflection now documents the exact metrics toggle, registry strategy, and alternatives tested. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:954
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/admin-api/deploy.sh around lines 123 to 137, the ServiceMonitor

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350577939

- [review_comment] 2025-09-16T03:19:56Z by coderabbitai[bot] (deployments/admin-api/deploy.sh:137)

```text
In deployments/admin-api/deploy.sh around lines 123 to 137, the ServiceMonitor
manifest uses the wrong apiVersion and assumes a named service port that may not
exist; change apiVersion from v1 to monitoring.coreos.com/v1 for the
ServiceMonitor CRD, and update your k8s Service definition (k8s-deployment.yaml)
to expose the port with name: http (e.g., name: http, port: <port-number>,
targetPort: <container-port>) so the ServiceMonitor endpoint port: http matches
a named port on the Service.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | ServiceMonitor now targets the correct CRD apiVersion. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> ServiceMonitor now targets the correct CRD apiVersion. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:1030
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/admin-api/k8s-deployment.yaml around lines 42-43 the jwt-secret

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350577957

- [review_comment] 2025-09-16T03:19:56Z by coderabbitai[bot] (deployments/admin-api/k8s-deployment.yaml:43)

```text
In deployments/admin-api/k8s-deployment.yaml around lines 42-43 the jwt-secret
is set to a weak hardcoded value ("change-me-in-production"); replace this by
generating a cryptographically secure secret (e.g., 32+ random bytes,
base64/hex-encoded) and remove the hardcoded value from the manifest, store the
secret in a Kubernetes Secret (or external secret manager like Vault/Secrets
Manager) and reference it via secretKeyRef/envFrom in the Deployment; ensure the
secret is not checked into source control, grant minimal RBAC to access it, and
document rotation/management procedures.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Secret manifest no longer ships a hard-coded JWT secret. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Secret manifest no longer ships a hard-coded JWT secret. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:1067
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/admin-api/k8s-deployment.yaml around line 65, the container image

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350577959

- [review_comment] 2025-09-16T03:19:56Z by coderabbitai[bot] (deployments/admin-api/k8s-deployment.yaml:65)

```text
In deployments/admin-api/k8s-deployment.yaml around line 65, the container image
is pinned to the non-deterministic tag "redis-work-queue/admin-api:latest";
replace this with a reproducible identifier (semantic version tag like
redis-work-queue/admin-api:vX.Y.Z or the image digest
redis-work-queue/admin-api@sha256:...) produced by your CI build, update the
manifest to use that tag/digest, and ensure your release pipeline publishes and
updates manifests automatically (or documents the manual step) so Kubernetes
deployments reference an immutable image for production.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Deployment manifest now references a tagged image instead of `latest`. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Deployment manifest now references a tagged image instead of `latest`. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:1105
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/admin-api/k8s-deployment.yaml around line 197, the file is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350577964

- [review_comment] 2025-09-16T03:19:56Z by coderabbitai[bot] (deployments/admin-api/k8s-deployment.yaml:197)

```text
In deployments/admin-api/k8s-deployment.yaml around line 197, the file is
missing a trailing newline at EOF; fix by adding a single newline character at
the end of the file so the file ends with a blank line, then commit;
additionally, add a CI linting rule (e.g., via eol-last in linters or a
YAML/file-format check) to enforce trailing newlines for all YAML files.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | YAML now ends with a newline and a lint script guards against regressions. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> YAML now ends with a newline and a lint script guards against regressions. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:1143
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/docker/docker-compose.yaml around lines 50 to 53, the healthcheck

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350577973

- [review_comment] 2025-09-16T03:19:57Z by coderabbitai[bot] (deployments/docker/docker-compose.yaml:53)

```text
In deployments/docker/docker-compose.yaml around lines 50 to 53, the healthcheck
uses wget which may not exist in minimal base images; replace it with a portable
curl invocation or run the check via CMD-SHELL so the shell builtin or installed
tools can be used. Update the test to use something like: use curl --fail
--silent --show-error http://localhost:8080/health (or wrap the existing wget in
a CMD-SHELL call) and keep interval/timeout/retries unchanged to ensure the
container healthcheck works across base images.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Healthcheck now uses curl with fail-fast flags, avoiding wget dependency. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Healthcheck now uses curl with fail-fast flags, avoiding wget dependency. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:28
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/docker/docker-compose.yaml around lines 86–106 (Prometheus) and

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350577980

- [review_comment] 2025-09-16T03:19:57Z by coderabbitai[bot] (deployments/docker/docker-compose.yaml:106)

```text
In deployments/docker/docker-compose.yaml around lines 86–106 (Prometheus) and
98–113 (Grafana), stop using :latest — pin Prometheus and Grafana images to
specific immutable versions (replace prom/prometheus:latest and
grafana/grafana:latest with chosen version tags), add a restart policy line
(restart: unless-stopped) to both service blocks, and resolve the missing
dashboards path referenced for Grafana by either adding the dashboards JSON
files under deployments/docker/grafana/dashboards or updating the Grafana
volumes entry to point to the correct existing dashboards directory.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Prometheus/Grafana now use pinned images, restart policies, and real provisioning files. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Prometheus/Grafana now use pinned images, restart policies, and real provisioning files. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:65
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/docker/rbac-configs/resources.yaml around lines 91 to 205 the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350577986

- [review_comment] 2025-09-16T03:19:57Z by coderabbitai[bot] (deployments/docker/rbac-configs/resources.yaml:205)

```text
In deployments/docker/rbac-configs/resources.yaml around lines 91 to 205 the
actions catalog is missing actions referenced by roles (metrics:read,
health:read) and contains an ambiguous queues:list entry; add explicit action
definitions for "metrics:read" and "health:read" mirroring the pattern
(description, risk_level, audit_required) used elsewhere, and resolve the
queues:list ambiguity by either removing "queues:list" from the catalog and all
roles or by explicitly mapping "queues:list" to the same internal capability as
"queues:read" (documenting it in the description) so semantics are not
duplicated or ambiguous.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Action catalog now covers metrics/health and documents the queues:list alias. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Action catalog now covers metrics/health and documents the queues:list alias. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:103
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## deployments/docker/rbac-configs/resources.yaml around lines 205–231: the DLQ

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350577993

- [review_comment] 2025-09-16T03:19:57Z by coderabbitai[bot] (deployments/docker/rbac-configs/resources.yaml:231)

```text
deployments/docker/rbac-configs/resources.yaml around lines 205–231: the DLQ
endpoints and verbs are mismatched with the Admin API; change the mapping key
"POST /api/v1/dlq/retry" to "POST /api/v1/dlq/requeue" and change the mapping
key "DELETE /api/v1/dlq" to "POST /api/v1/dlq/purge" while keeping "DELETE
/api/v1/queues/dlq": ["dlq:purge"] as-is; after making these edits, verify the
final endpoint→permission mappings against internal/admin-api/server.go (GET
/api/v1/dlq, POST /api/v1/dlq/requeue, POST /api/v1/dlq/purge, GET
/api/v1/workers) and internal/rbac-and-tokens/config.go, add monitoring
endpoints if missing (GET /metrics → metrics:read and GET /healthz →
health:read), and update related roles/tests/docs (test/e2e, test/integration,
docs/*, deployments/*) to reflect the canonical routes.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Endpoint map now mirrors the Admin API contract and includes metrics/health routes. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Endpoint map now mirrors the Admin API contract and includes metrics/health routes. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:142
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/docker/rbac-configs/roles.yaml around lines 19 to 23, the roles

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578007

- [review_comment] 2025-09-16T03:19:57Z by coderabbitai[bot] (deployments/docker/rbac-configs/roles.yaml:23)

```text
In deployments/docker/rbac-configs/roles.yaml around lines 19 to 23, the roles
file uses undefined actions and wildcards and uses ad-hoc condition flags that
don't match resources.yaml; update operator.permissions to replace "queues:list"
with "queues:read"; replace admin.permissions wildcard entries "admin:*" and
"system:*" with the explicit actions present in resources.yaml (e.g.,
"admin:users","admin:tokens","admin:audit","admin:system"); for
monitoring.permissions either add missing actions "metrics:read" and
"health:read" to deployments/docker/rbac-configs/resources.yaml or map them to
existing actions such as "stats:read"/"stats:export"; for emergency.permissions
replace "admin:all" and "emergency:*" with explicit actions (for example
admin:users, admin:tokens, admin:audit, admin:system, queues:delete, dlq:purge)
or implement true wildcard semantics in the enforcer (don’t leave magic); and
change emergency.resource_constraints.conditions to reference schedules defined
in resources.yaml (e.g., schedule: "emergency_only" or "after_hours") rather
than ad-hoc flags; finally run the RBAC validation to ensure every permission in
roles.yaml exists under actions in resources.yaml.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Roles now reference defined actions and list emergency powers explicitly. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Roles now reference defined actions and list emergency powers explicitly. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:183
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/kubernetes/admin-api-deployment.yaml around line 98, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578020

- [review_comment] 2025-09-16T03:19:57Z by coderabbitai[bot] (deployments/kubernetes/admin-api-deployment.yaml:98)

```text
In deployments/kubernetes/admin-api-deployment.yaml around line 98, the
container image is pinned to the non-deterministic tag
"work-queue/admin-api:latest"; replace it with a specific immutable tag
(semantic version like work-queue/admin-api:v1.2.3 or an image digest like
work-queue/admin-api@sha256:<digest>) so deployments are reproducible and not
affected by upstream image updates—update the manifest to point to the chosen
version/digest and ensure your CI/CD publishes and references that exact
tag/digest.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Deployment now references the tagged admin-api image. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Deployment now references the tagged admin-api image. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:229
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/kubernetes/admin-api-deployment.yaml around lines 190-193 the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578025

- [review_comment] 2025-09-16T03:19:58Z by coderabbitai[bot] (deployments/kubernetes/admin-api-deployment.yaml:193)

```text
In deployments/kubernetes/admin-api-deployment.yaml around lines 190-193 the
RoleRule grants get/list/watch on all configmaps and secrets which is too broad;
replace the blanket resource access with explicit resourceNames for each secret
and configmap the admin API actually needs (e.g. add resourceNames:
["<specific-secret-name>","<specific-configmap-name>"]) and remove wide-scoped
entries, or split into separate rules per resource type with only the minimal
verbs required; ensure the role is also scoped to the correct namespace and
update any references to match the new names.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | RBAC now grants get access only to the needed configmap/secret. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> RBAC now grants get access only to the needed configmap/secret. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:267
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/kubernetes/rbac-monitoring.yaml around lines 1 to 18, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578037

- [review_comment] 2025-09-16T03:19:58Z by coderabbitai[bot] (deployments/kubernetes/rbac-monitoring.yaml:18)

```text
In deployments/kubernetes/rbac-monitoring.yaml around lines 1 to 18, the
manifest uses a core API version for a ServiceMonitor (which is part of the
Prometheus Operator CRDs) and will fail; change the apiVersion to
monitoring.coreos.com/v1, keep kind: ServiceMonitor, ensure the ServiceMonitor
CRD is installed (Prometheus Operator) in the cluster and that the namespace
exists, and confirm the selector/labels match the target Service so Prometheus
can discover it.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | ServiceMonitor manifest now references monitoring.coreos.com/v1. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> ServiceMonitor manifest now references monitoring.coreos.com/v1. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:343
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/kubernetes/rbac-monitoring.yaml around line 43, the runbook_url

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578051

- [review_comment] 2025-09-16T03:19:58Z by coderabbitai[bot] (deployments/kubernetes/rbac-monitoring.yaml:43)

```text
In deployments/kubernetes/rbac-monitoring.yaml around line 43, the runbook_url
is pointing to a dummy/non-existent wiki; replace that value with the correct,
accessible runbook URL for the RBAC service (or remove the runbook_url field if
no runbook exists) so on-call engineers have a valid link. Ensure the new URL
points to the canonical incident runbook (or the team's runbook index) and
verify accessibility before committing.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Runbook link now points to the documented ops guide. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Runbook link now points to the documented ops guide. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:380
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/kubernetes/rbac-monitoring.yaml around lines 354 and 362, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578062

- [review_comment] 2025-09-16T03:19:58Z by coderabbitai[bot] (deployments/kubernetes/rbac-monitoring.yaml:354)

```text
In deployments/kubernetes/rbac-monitoring.yaml around lines 354 and 362, the
Slack webhook entries are set to the placeholder 'YOUR_SLACK_WEBHOOK_URL' which
will cause Alertmanager to fail or silently drop alerts; replace those
placeholders with the real Slack webhook URLs (or remove the Slack receiver
blocks entirely if you don't want Slack notifications), and instead of
hardcoding secrets inline ensure you reference the webhook via a Kubernetes
Secret (create a secret containing the webhook and use secretKeyRef or envFrom
in the Alertmanager config/Secret) so the actual webhook is not stored in plain
text in the manifest.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Slack receivers now read webhooks from mounted secrets instead of inline placeholders. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Slack receivers now read webhooks from mounted secrets instead of inline placeholders. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:416
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/kubernetes/rbac-token-service-deployment.yaml around lines

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578074

- [review_comment] 2025-09-16T03:19:59Z by coderabbitai[bot] (deployments/kubernetes/rbac-token-service-deployment.yaml:204)

```text
In deployments/kubernetes/rbac-token-service-deployment.yaml around lines
203-204, the container image is pinned to the "latest" tag with imagePullPolicy:
Always which causes non-deterministic deployments and rollback issues; replace
"work-queue/rbac-token-service:latest" with a semantic version tag (e.g. v1.2.3)
and set imagePullPolicy to IfNotPresent (or keep Always only if you must always
pull unique digests), update your release/CI pipeline to stamp the manifest with
the specific image tag (or image digest) during build/release so every
deployment references an immutable image and next time you need to roll back you
can reference the previous version.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Token service deployment now points to a tagged image and avoids Always pulls. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Token service deployment now points to a tagged image and avoids Always pulls. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:494
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/kubernetes/rbac-token-service-deployment.yaml around lines 301 to

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578076

- [review_comment] 2025-09-16T03:19:59Z by coderabbitai[bot] (deployments/kubernetes/rbac-token-service-deployment.yaml:307)

```text
In deployments/kubernetes/rbac-token-service-deployment.yaml around lines 301 to
307, the Role grants "watch" on configmaps and secrets which is excessive for a
token service; remove "watch" from the verbs for those resources (leaving "get"
and "list") unless you actually implement runtime reload behavior that relies on
watches — if you do need watches, document the exact code path and limit the
Role to the smallest namespace/specific resource names possible and add a
comment explaining why "watch" is required.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Removed unnecessary watch verb from configmap/secret access. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Removed unnecessary watch verb from configmap/secret access. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:533
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/scripts/deploy-rbac-staging.sh around lines 17-19 (and similarly

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578092

- [review_comment] 2025-09-16T03:19:59Z by coderabbitai[bot] (deployments/scripts/deploy-rbac-staging.sh:19)

```text
In deployments/scripts/deploy-rbac-staging.sh around lines 17-19 (and similarly
21-23, 25-27) the log functions are duplicated; extract them to a single shared
shell utility (e.g., deployments/scripts/utils/logging.sh), put color constants
and log(), info(), warn(), error() helpers there, make the file idempotent if
sourced, and ensure it sets no side-effects; then replace the copied function
blocks in each deployment script with a single source ../utils/logging.sh (use
the correct relative path), verify scripts still execute (make the utility
readable/executable as needed) and run CI/lint to confirm no path or naming
regressions.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Logging helpers now live in `lib/logging.sh` and the script sources them. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Logging helpers now live in `lib/logging.sh` and the script sources them. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:608
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/scripts/deploy-rbac-staging.sh around line 155, the SERVICE_IP

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578097

- [review_comment] 2025-09-16T03:19:59Z by coderabbitai[bot] (deployments/scripts/deploy-rbac-staging.sh:155)

```text
In deployments/scripts/deploy-rbac-staging.sh around line 155, the SERVICE_IP
variable is assigned but never used; remove the dead assignment line entirely
(or if intended to be used, reference SERVICE_IP where needed) so there is no
unused variable left in the script and re-run ShellCheck to ensure no other
issues.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Removed unused SERVICE_IP assignment before shellcheck flagged it. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Removed unused SERVICE_IP assignment before shellcheck flagged it. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:647
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/scripts/deploy-staging.sh around line 71, the docker build

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578104

- [review_comment] 2025-09-16T03:19:59Z by coderabbitai[bot] (deployments/scripts/deploy-staging.sh:71)

```text
In deployments/scripts/deploy-staging.sh around line 71, the docker build
invocation uses an unquoted variable ($IMAGE_NAME) which will break if it
contains spaces or special characters; update the command to quote the variable
(e.g. -t "$IMAGE_NAME") and review other shell variables in the script to ensure
all are quoted similarly to avoid word-splitting and globbing issues.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Docker build now quotes the image tag, preventing shell word-splitting. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Docker build now quotes the image tag, preventing shell word-splitting. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:717
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/scripts/deploy-staging.sh around lines 73, 85, 99, and 122, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578110

- [review_comment] 2025-09-16T03:20:00Z by coderabbitai[bot] (deployments/scripts/deploy-staging.sh:73)

```text
In deployments/scripts/deploy-staging.sh around lines 73, 85, 99, and 122, the
script uses archaic checks like "if [ $? -ne 0 ]; then" after commands; replace
each pattern by testing the command directly (e.g., change "some_command; if [
$? -ne 0 ]; then" to "if ! some_command; then" or invert as appropriate for
success checks), making sure any command substitutions, pipes, or redirects are
preserved and that the conditional logic (success vs failure) remains the same
for each occurrence.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Replaced `$?` checks with direct command conditions for build/push. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Replaced `$?` checks with direct command conditions for build/push. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:752
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/scripts/deploy-staging.sh around line 83 (and also at lines 112

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578112

- [review_comment] 2025-09-16T03:20:00Z by coderabbitai[bot] (deployments/scripts/deploy-staging.sh:83)

```text
In deployments/scripts/deploy-staging.sh around line 83 (and also at lines 112
and 192), several unquoted shell variables like $IMAGE_NAME are being used;
update each occurrence to use quoted parameter expansion (e.g., "$IMAGE_NAME")
so values with spaces or special characters are handled correctly, and review
surrounding commands to ensure quotes are applied consistently for all variable
expansions and any necessary escaping is preserved.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Quoted kubectl arguments and hardened port-forward cleanup throughout the script. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Quoted kubectl arguments and hardened port-forward cleanup throughout the script. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:789
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/scripts/deploy-staging.sh around line 225, the conditional uses

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578116

- [review_comment] 2025-09-16T03:20:00Z by coderabbitai[bot] (deployments/scripts/deploy-staging.sh:225)

```text
In deployments/scripts/deploy-staging.sh around line 225, the conditional uses
the deprecated negated -z test ("if [ ! -z \"$INGRESS_HOST\" ];"); replace it
with the canonical non-empty test using -n (i.e., "if [ -n \"$INGRESS_HOST\"
];") to match the project's style and avoid deprecated negation; ensure the
variable remains quoted to preserve whitespace and exit status semantics.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Condition now uses `[[ -n "$INGRESS_HOST" ]]`, matching suggested style. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Condition now uses `[[ -n "$INGRESS_HOST" ]]`, matching suggested style. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:825
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/scripts/health-check-rbac.sh around line 238, the current command

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578120

- [review_comment] 2025-09-16T03:20:00Z by coderabbitai[bot] (deployments/scripts/health-check-rbac.sh:238)

```text
In deployments/scripts/health-check-rbac.sh around line 238, the current command
uses grep | wc -l to count "error" occurrences; replace that pipeline with grep
-i -c to count matches directly and keep the existing stderr suppression and
options, i.e., invoke kubectl logs ... 2>/dev/null | grep -i -c error and assign
the output to error_count to follow ShellCheck best practices and be more
efficient.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Log error counting now uses `grep -i -c`, eliminating the extra pipeline. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Log error counting now uses `grep -i -c`, eliminating the extra pipeline. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:860
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/scripts/setup-monitoring.sh around lines 17 to 31 the logging

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578130

- [review_comment] 2025-09-16T03:20:00Z by coderabbitai[bot] (deployments/scripts/setup-monitoring.sh:31)

```text
In deployments/scripts/setup-monitoring.sh around lines 17 to 31 the logging
functions (log, info, warn, error) are duplicated; extract these functions into
a new deployments/scripts/lib/logging.sh and replace the duplicate block here
with a single source statement (e.g. source "$(dirname "$0")/lib/logging.sh");
do the same in the other scripts that currently copy these functions, remove the
duplicates, and ensure lib/logging.sh is idempotent (guard against
double-sourcing or redefinition by checking a sentinel variable) and uses the
same color variables so all scripts keep identical behavior.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Monitoring script sources the shared logging helpers instead of duplicating them. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Monitoring script sources the shared logging helpers instead of duplicating them. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:896
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/scripts/setup-monitoring.sh around lines 132 to 180, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578133

- [review_comment] 2025-09-16T03:20:00Z by coderabbitai[bot] (deployments/scripts/setup-monitoring.sh:180)

```text
In deployments/scripts/setup-monitoring.sh around lines 132 to 180, the
Alertmanager YAML is using hardcoded placeholder values (e.g., localhost:587 and
admin@company.com) which will not work in production; update the script to
source SMTP host/port, from address, and alert recipient addresses from
environment variables or Kubernetes Secrets (fallback to no-default or clearly
fail), replace the hardcoded values in the generated alertmanager.yml with those
variables, and add a brief comment or README note next to the creation step
documenting that real values or secrets must be provided before deployment.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Alertmanager config now sources SMTP + recipients from env vars. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Alertmanager config now sources SMTP + recipients from env vars. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:934
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/scripts/test-staging-deployment.sh around lines 63-68 (and

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578141

- [review_comment] 2025-09-16T03:20:00Z by coderabbitai[bot] (deployments/scripts/test-staging-deployment.sh:68)

```text
In deployments/scripts/test-staging-deployment.sh around lines 63-68 (and
similarly at 71-76 and 79-84), add a prerequisite check for the jq binary using
the same pattern as the kubectl check: use command -v jq &> /dev/null to detect
presence, call add_test_result "jq-available" "PASS" "jq is available" when
found, otherwise call add_test_result "jq-available" "FAIL" "jq not found" and
set all_passed=false; keep messaging consistent with the existing kubectl check
and ensure the new check appears before any code that uses jq.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Prerequisite checks now verify jq before the script proceeds. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Prerequisite checks now verify jq before the script proceeds. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:972
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/scripts/test-staging-deployment.sh around lines 331-333, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578148

- [review_comment] 2025-09-16T03:20:01Z by coderabbitai[bot] (deployments/scripts/test-staging-deployment.sh:333)

```text
In deployments/scripts/test-staging-deployment.sh around lines 331-333, the
base64 decode pipeline can fail silently and return an empty/invalid
bootstrap_token; update the retrieval to enable strict pipe error detection
(e.g., set -o pipefail for the block or run commands sequentially), check the
exit status of kubectl and base64, validate the token is non-empty and valid,
and on failure emit a clear error to stderr and exit with a non-zero status so
the script does not proceed with an invalid token.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Bootstrap token retrieval now handles pipeline failures safely. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Bootstrap token retrieval now handles pipeline failures safely. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:1009
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/scripts/test-staging-deployment.sh around lines 451-453 (and

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578153

- [review_comment] 2025-09-16T03:20:01Z by coderabbitai[bot] (deployments/scripts/test-staging-deployment.sh:453)

```text
In deployments/scripts/test-staging-deployment.sh around lines 451-453 (and
similarly lines 461-463), the script currently compares volume mount paths
against hardcoded values like "/app/audit" and "/app/keys"; change this to
reference configurable variables or values derived from the deployment manifest:
introduce variables at the top of the script (e.g., AUDIT_MOUNT_PATH and
KEYS_MOUNT_PATH) or query the deployment spec (kubectl get deployment -o
jsonpath=...) to obtain the expected mount paths, then replace the hardcoded
string literals in the comparisons with those variables so the checks follow the
manifest/config rather than fixed paths.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Volume mount checks now read expected paths from the deployment spec. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Volume mount checks now read expected paths from the deployment spec. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:1046
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/FEATURE_ENHANCEMENT_AGENT_PROMPT.md around lines 32 to 35, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578161

- [review_comment] 2025-09-16T03:20:01Z by coderabbitai[bot] (docs/FEATURE_ENHANCEMENT_AGENT_PROMPT.md:35)

```text
In docs/FEATURE_ENHANCEMENT_AGENT_PROMPT.md around lines 32 to 35, the
admonition uses a non-standard collapsible syntax (`> [!note]- **🗣️ CLAUDE'S
THOUGHTS 💭**`) which may not render across Markdown processors; replace it with
standard Markdown (e.g., a heading or blockquote) or explicitly document the
required Markdown extension, and update the template to use a portable pattern
such as a level-3 heading "### 🗣️ CLAUDE'S THOUGHTS" or a standard blockquote
starting with "> **🗣️ CLAUDE'S THOUGHTS:**" so rendering is consistent.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Template now uses a standard heading + blockquote for Claude's thoughts. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Template now uses a standard heading + blockquote for Claude's thoughts. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:1085
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/FEATURE_ENHANCEMENT_AGENT_PROMPT.md around lines 156–180 the color

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578168

- [review_comment] 2025-09-16T03:20:01Z by coderabbitai[bot] (docs/FEATURE_ENHANCEMENT_AGENT_PROMPT.md:180)

```text
In docs/FEATURE_ENHANCEMENT_AGENT_PROMPT.md around lines 156–180 the color
palette is embedded as an XML-style comment block which is not machine-friendly
or easily reusable; extract the palette into a proper config (e.g.,
docs/colors.yml or docs/colors.json) listing named roles and hex values, replace
the XML comment in the markdown with either a compact markdown table pointing to
the new config file or a brief reference link, and ensure each entry includes a
clear role key (background/panel/border/text/button) and hex value so downstream
code/docs can import or reference it.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Palette now lives in `docs/colors.yml` with a summary table in the template. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Palette now lives in `docs/colors.yml` with a summary table in the template. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:1122
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In Makefile around line 13, remove the unnecessary GO111MODULE=on prefix from

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578185

- [review_comment] 2025-09-16T03:20:01Z by coderabbitai[bot] (Makefile:13)

```text
In Makefile around line 13, remove the unnecessary GO111MODULE=on prefix from
the go build command and ensure the linker variable used in LDFLAGS exists:
either add a package-level variable in package main (e.g. in cmd/$(APP)/main.go
declare var version string) so -X main.version=$(VERSION) will succeed, or
change the -X flag to point to the actual package and symbol that holds the
version; then run the build to confirm the "Missing main.version symbol for -X"
error is resolved.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Dropped the unused GO111MODULE env and confirmed the linker flag resolves against the existing `main.version` symbol. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Dropped the unused GO111MODULE env and confirmed the linker flag resolves against the existing `main.version` symbol. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:64
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In Makefile around lines 16 to 18, the TUI build target uses a different go

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578194

- [review_comment] 2025-09-16T03:20:01Z by coderabbitai[bot] (Makefile:18)

```text
In Makefile around lines 16 to 18, the TUI build target uses a different go
build invocation than the other build targets; make it consistent by using the
same variables and flags as the other builds: ensure the target creates the bin
directory if needed and invoke go build with the shared GOFLAGS and LDFLAGS
(e.g. use GO111MODULE=on go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o bin/tui
./cmd/tui) so the TUI is built with the same environment and linker flags as the
rest of the project.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Unified the TUI build recipe with the primary binary and added the missing version symbol so LDFLAGS apply cleanly. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Unified the TUI build recipe with the primary binary and added the missing version symbol so LDFLAGS apply cleanly. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:101
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In .githooks/pre-commit around lines 7 to 15, the hook assumes python3 exists

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350580859

- [review_comment] 2025-09-16T03:21:08Z by coderabbitai[bot] (.githooks/pre-commit:15)

```text
In .githooks/pre-commit around lines 7 to 15, the hook assumes python3 exists
and will fail if it's absent; update it to detect an available interpreter at
runtime by checking for "python3" first, then "python" (using quoted paths), and
if neither exists either exit with a clear error message or skip the progress
update gracefully; use the detected interpreter variable when invoking
scripts/update_progress.py and ensure git add only runs when the update
succeeded.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Hook now discovers python3/python dynamically and exits early if neither exists. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Hook now discovers python3/python dynamically and exits early if neither exists. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:138
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In .github/workflows/update-progress.yml lines 1-6, YAMLLint flagged the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350580871

- [review_comment] 2025-09-16T03:21:08Z by coderabbitai[bot] (.github/workflows/update-progress.yml:6)

```text
In .github/workflows/update-progress.yml lines 1-6, YAMLLint flagged the
reserved key and bracket spacing: wrap the reserved key "on" in quotes (e.g.
"on":) and tighten the branch list brackets by removing inner spaces and quoting
the value (e.g. branches: ['main']) so the YAML parses cleanly.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Quoted the reserved `on` key and normalized the branches list to satisfy YAML linters. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Quoted the reserved `on` key and normalized the branches list to satisfy YAML linters. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:175
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In .github/workflows/update-progress.yml around lines 8 to 10, the workflow

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350580880

- [review_comment] 2025-09-16T03:21:08Z by coderabbitai[bot] (.github/workflows/update-progress.yml:10)

```text
In .github/workflows/update-progress.yml around lines 8 to 10, the workflow
lacks concurrency control which allows overlapping runs on rapid pushes; add a
top-level concurrency block (e.g., concurrency: { group: 'update-progress-${{
github.ref }}', cancel-in-progress: true }) to serialize runs per branch/ref and
cancel any in-progress run when a new one starts; place this block at the same
indentation level as permissions and jobs.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Added a concurrency block so only one update-progress run executes per ref. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added a concurrency block so only one update-progress run executes per ref. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:209
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## .github/workflows/update-progress.yml lines 31-41: the workflow uses unguarded

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350580886

- [review_comment] 2025-09-16T03:21:08Z by coderabbitai[bot] (.github/workflows/update-progress.yml:41)

```text
.github/workflows/update-progress.yml lines 31-41: the workflow uses unguarded
git add and unquoted $GITHUB_OUTPUT redirections; update the script to first
check for the presence of the files (or use a safe add that won’t fail) before
running git add, and ensure all instances of >> $GITHUB_OUTPUT are changed to >>
"$GITHUB_OUTPUT" (quote the variable) so the redirection target is not
word-split; keep behavior the same otherwise (commit only when changes exist).
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Hardened the commit step to quote `$GITHUB_OUTPUT` and only add files that exist. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Hardened the commit step to quote `$GITHUB_OUTPUT` and only add files that exist. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:245
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## .github/workflows/update-progress.yml around lines 43-47 contains an extra

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350580891

- [review_comment] 2025-09-16T03:21:08Z by coderabbitai[bot] (.github/workflows/update-progress.yml:47)

```text
.github/workflows/update-progress.yml around lines 43-47 contains an extra
trailing blank line after the git push step; remove the empty line after the
final "git push" run block so the file ends immediately after the command (no
extra newline line), keeping YAML lint happy.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Simplified the push step to a single-line run so no stray blank line remains. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Simplified the push step to a single-line run so no stray blank line remains. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:281
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In cmd/job-queue-system/main.go around lines 85 to 92, the code always starts

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350580900

- [review_comment] 2025-09-16T03:21:09Z by coderabbitai[bot] (cmd/job-queue-system/main.go:92)

```text
In cmd/job-queue-system/main.go around lines 85 to 92, the code always starts
the observability HTTP server (metrics/healthz/readyz) even for role=admin; this
can cause unnecessary port conflicts for one-shot admin commands. Modify the
logic to only call obs.StartHTTPServer and defer its Shutdown when the
configured role is not "admin" (e.g., if cfg.Role != "admin" or a bool IsAdmin
check is false); otherwise skip starting the HTTP server entirely. Ensure
readyCheck remains defined for non-admin paths and that no server is started or
shutdown deferred when running as admin.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Admin executions now skip starting the HTTP metrics server to avoid port conflicts. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Admin executions now skip starting the HTTP metrics server to avoid port conflicts. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:315
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In cmd/job-queue-system/main.go around lines 112–114, the background metrics

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350580914

- [review_comment] 2025-09-16T03:21:09Z by coderabbitai[bot] (cmd/job-queue-system/main.go:114)

```text
In cmd/job-queue-system/main.go around lines 112–114, the background metrics
updater is started unconditionally for CLI/admin paths; guard the
obs.StartQueueLengthUpdater call so it is skipped when running the admin/CLI
mode. Change the code to check an existing config flag (e.g. cfg.Admin or
cfg.IsAdmin) or add a new flag like cfg.DisableBackgroundUpdaters/DisableMetrics
and only call obs.StartQueueLengthUpdater when that flag is false; ensure the
updater is not launched for admin CLI invocations.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Admin runs no longer spawn the background queue length updater. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Admin runs no longer spawn the background queue length updater. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:353
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In cmd/job-queue-system/main.go around lines 142 to 149, the admin handling is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350580918

- [review_comment] 2025-09-16T03:21:09Z by coderabbitai[bot] (cmd/job-queue-system/main.go:149)

```text
In cmd/job-queue-system/main.go around lines 142 to 149, the admin handling is
nested inside the switch's default branch which obscures flow and risks
accidental execution when new roles are added; refactor by branching early or
adding an explicit case for "admin" instead of placing it in default: move the
admin check before the switch or add a dedicated case "admin" that calls
runAdmin(...) and returns, and keep the default branch solely to log/exit on
unknown roles so control flow is clear and future role additions won't
accidentally run admin logic.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Refactored the role switch to give admin its own case and leave default for true unknowns. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Refactored the role switch to give admin its own case and leave default for true unknowns. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:390
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In cmd/job-queue-system/main.go around lines 187-188, the purge-all branch

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350580928

- [review_comment] 2025-09-16T03:21:09Z by coderabbitai[bot] (cmd/job-queue-system/main.go:188)

```text
In cmd/job-queue-system/main.go around lines 187-188, the purge-all branch
prints a human-readable string with fmt.Printf which is inconsistent with other
admin commands that emit machine-readable JSON; replace the fmt.Printf call with
code that outputs a JSON object (e.g. {"purged": n}) to stdout using the
encoding/json package (or fmt.Fprintf with a properly escaped JSON string) and
ensure it ends with a newline and returns the same exit path as other admin
outputs.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Purge-all now emits JSON matching other admin outputs. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Purge-all now emits JSON matching other admin outputs. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:428
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/kubernetes/rbac-monitoring.yaml around line 35 (and also review

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350580943

- [review_comment] 2025-09-16T03:21:09Z by coderabbitai[bot] (deployments/kubernetes/rbac-monitoring.yaml:35)

```text
In deployments/kubernetes/rbac-monitoring.yaml around line 35 (and also review
lines 78, 88, 99, 109, 120, 130, 140, 151, 161, 172, 182), the Prometheus alert
expressions reference non-existent rbac_* metrics and job="rbac-token-service";
either register the missing rbac_* metrics in the RBAC token service (create and
MustRegister prometheus metrics, instrument code, and expose them on /metrics)
or update the YAML alerts to reference actual metric names we already export
(e.g., jobs_*, backpressure_*, exactly_once_*) and/or ensure Prometheus scrape
configs include a target with job="rbac-token-service" so
up{job="rbac-token-service"} is valid; choose one approach and make consistent
changes across all listed lines.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Updated the RBAC monitoring stack to use existing http/app metrics and dropped unreachable token counters. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Updated the RBAC monitoring stack to use existing http/app metrics and dropped unreachable token counters. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:465
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In Dockerfile around line 3, the build stage uses golang:1.23 while CI/docs

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350580952

- [review_comment] 2025-09-16T03:21:09Z by coderabbitai[bot] (Dockerfile:3)

```text
In Dockerfile around line 3, the build stage uses golang:1.23 while CI/docs
require Go 1.25; update the builder image tag to golang:1.25 to align
toolchains. Edit the FROM line to use golang:1.25 (or the exact patch variant
your CI/doc specify) so the compile environment matches CI/docs, then rebuild
and run CI to verify no compatibility issues.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 10 | Bumped the builder image to Go 1.25 to match CI/tooling requirements. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Bumped the builder image to Go 1.25 to match CI/tooling requirements. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:505
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/15_promotion_checklists.md around lines 21–33, the promotion checklist

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350580966

- [review_comment] 2025-09-16T03:21:09Z by coderabbitai[bot] (docs/15_promotion_checklists.md:33)

```text
In docs/15_promotion_checklists.md around lines 21–33, the promotion checklist
table is missing OWNER and EXIT CRITERIA columns which prevents enforceable
gates; add two new columns "OWNER" and "EXIT CRITERIA" to the table header and
for each row populate a responsible owner (GitHub handle or team) and a clear,
testable exit criterion (e.g., "E2E pass 5× in CI", ">=80% coverage on admin &
obs", "alert rules merged + helm docs") and update any status cells accordingly
so CI can validate these criteria; also add a short note under the table
describing how CI will read these columns (OWNER format and exact phrasing
required for automated checks).
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Added Owner/Exit Criteria columns across promotion tables and documented CI expectations. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added Owner/Exit Criteria columns across promotion tables and documented CI expectations. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:540
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/admin-api.md around lines 41 to 44, the configuration shows a single

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350580975

- [review_comment] 2025-09-16T03:21:10Z by coderabbitai[bot] (docs/api/admin-api.md:44)

```text
In docs/api/admin-api.md around lines 41 to 44, the configuration shows a single
confirmation_phrase while the endpoints require different phrases (e.g.,
CONFIRM_DELETE_ALL) causing inconsistency; choose one approach and make docs and
code consistent: either document separate keys (e.g., dlq_confirmation_phrase
and purge_all_confirmation_phrase) and update the sample config and README to
list both keys, or change the endpoints to validate against the single
configured confirmation_phrase and update any endpoint docs/samples to reference
that single key; apply the chosen change across the docs and codebase so the
config keys and endpoint expectations match.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Split destructive confirmations into DLQ/purge-all phrases and documented the fallback. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Split destructive confirmations into DLQ/purge-all phrases and documented the fallback. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:579
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/admin-api.md around lines 106 to 132, the queue parameter

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350581001

- [review_comment] 2025-09-16T03:21:10Z by coderabbitai[bot] (docs/api/admin-api.md:132)

```text
In docs/api/admin-api.md around lines 106 to 132, the queue parameter
description is ambiguous about alias-to-Redis-key mappings; update the docs to
explicitly list accepted aliases (high, low, completed, dead_letter) and show
the exact Redis key each alias resolves to (or state that a full Redis key may
be provided), and reference the configuration fields that control those mappings
by name (worker.queues.* for priority queues, completed_list for completed jobs,
dead_letter_list for dead-letter queue). Mention accepted value formats (alias
or full Redis key) and provide a short example mapping table or inline examples
referencing the config keys so readers know where to change the mappings.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Clarified queue alias mapping with a table tied to the worker config fields. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Clarified queue alias mapping with a table tied to the worker config fields. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:618
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/admin-api.md around lines 260 to 268, the CORS guidance currently

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350581008

- [review_comment] 2025-09-16T03:21:10Z by coderabbitai[bot] (docs/api/admin-api.md:268)

```text
In docs/api/admin-api.md around lines 260 to 268, the CORS guidance currently
implies using cors_allow_origins: ["*"]; update the text to recommend an empty
list as the safe default and explicitly warn that using "*" is dangerous when
require_auth: true. Replace the current bullet with instructions to set
cors_allow_origins to an explicit, environment-specific list of allowed origins
(or leave empty to block cross-origin requests), add a short note discouraging
"*" for authenticated endpoints, and include a brief recommendation to use
specific subdomains or environment variables for allowed origins and to test
CORS in staging before production.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Updated CORS guidance to promote explicit allow-lists and warn against `"*"` with auth. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Updated CORS guidance to promote explicit allow-lists and warn against `"*"` with auth. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:657
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## docs/api/event-hooks.md around lines 110 to 124: the HMAC signature currently

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350581023

- [review_comment] 2025-09-16T03:21:10Z by coderabbitai[bot] (docs/api/event-hooks.md:124)

```text
docs/api/event-hooks.md around lines 110 to 124: the HMAC signature currently
only covers the body which allows replay attacks because the listed
X-Webhook-Timestamp is not bound to the signature; update the docs to require
that the signature is computed over a canonical string that includes the
timestamp (e.g., timestamp + "." + body) and that receivers verify the timestamp
is within a configurable freshness window (e.g., ±N seconds) before accepting
the signature, and document that the server must reject deliveries with
missing/old timestamps or mismatched signatures.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Documented the timestamp-bound HMAC string and freshness window. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Documented the timestamp-bound HMAC string and freshness window. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:696
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/event-hooks.md around lines 246 to 264, the docs currently describe

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350581042

- [review_comment] 2025-09-16T03:21:11Z by coderabbitai[bot] (docs/api/event-hooks.md:264)

```text
In docs/api/event-hooks.md around lines 246 to 264, the docs currently describe
rate limiting and retry policy but do not document idempotency or replay
semantics for DLH replays; add a short subsection stating that replays may
duplicate deliveries and receivers must treat deliveries as potentially
repeated, require an idempotency header (e.g., X-Webhook-Delivery) containing a
unique delivery ID, and optionally include a replay indicator (e.g.,
X-Webhook-Replay: true). Explain receiver behavior: persist the delivery ID with
a configurable TTL, deduplicate by returning a successful 2xx response for
already-processed IDs, use the idempotency key to make non-idempotent operations
safe (skip or noop on duplicate IDs), and document recommended TTL and retention
guidance for the idempotency cache.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Documented replay semantics, idempotency key usage, and the replay header. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Documented replay semantics, idempotency key usage, and the replay header. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:734
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 41-44, the docs show an

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350581052

- [review_comment] 2025-09-16T03:21:11Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:44)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 41-44, the docs show an
incorrect go test command using a filename; replace it with package-aware
commands using -run filters: from repo root use "go test -v ./... -run
'^TestHMACSigner_'" or, when inside the package directory, "cd path/to/package
&& go test -v -run '^TestHMACSigner_'", and update the fenced bash block
accordingly so tests run reliably and with dependencies resolved.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 10 | Updated signature test instructions to use a package-aware `go test -run` pattern. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Updated signature test instructions to use a package-aware `go test -run` pattern. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:775
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 61 to 65 the example test

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350581060

- [review_comment] 2025-09-16T03:21:11Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:65)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 61 to 65 the example test
command uses a filename which is incorrect when run from the repository root;
update the docs to use package paths or a patterned run flag instead: replace
the current `go test -v ./event_filter_test.go` example with a command that runs
tests by package or name, e.g. `go test -v ./... -run '^TestEventFilter_'`, or
alternatively instruct readers to change into the directory containing the test
and run `go test -v` — ensure the doc shows one clear correct command and
removes the filename-based invocation.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 10 | Swapped the event filter docs to use `go test -run` instead of a filename. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Swapped the event filter docs to use `go test -run` instead of a filename. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:811
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 86-90, the docs currently

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350581065

- [review_comment] 2025-09-16T03:21:11Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:90)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 86-90, the docs currently
instruct running a single test file; update the example to run the integration
test by pattern instead of a filename. Replace the existing command with one
that runs the package tests using the -run flag to match the TestWebhookHarness_
tests (e.g., cd test/integration && go test -v -run '^TestWebhookHarness_'), and
ensure the fenced code block language remains bash.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 10 | Integration harness docs now run tests via `-run` within the package. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Integration harness docs now run tests via `-run` within the package. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:849
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 112 to 116, the example

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350581070

- [review_comment] 2025-09-16T03:21:11Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:116)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 112 to 116, the example
command and fenced block are incorrect; replace the existing bash block that
runs `go test -v ./nats_transport_test.go` with a bash fenced block that runs
`cd test/integration && go test -v -run '^TestNATSTransport_'` so the docs use
the correct go test invocation to run the NATS transport tests by name pattern
and ensure the code fence language is "bash".
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 10 | Updated NATS transport instructions to use the package-pattern invocation. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Updated NATS transport instructions to use the package-pattern invocation. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:885
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 136 to 139, the example test

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350581075

- [review_comment] 2025-09-16T03:21:11Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:139)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 136 to 139, the example test
command and fenced block are incorrect for running only DLH tests; replace the
existing bash fenced block (which runs the specific file) with a bash fenced
block that executes the Go test runner with the -run '^TestDLH_' pattern (i.e.,
change the command to: cd test/integration && go test -v -run '^TestDLH_') so
the documentation shows running only DLH tests via the -run flag.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 10 | DLH replay guidance now uses the `-run '^TestDLH_'` pattern. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> DLH replay guidance now uses the `-run '^TestDLH_'` pattern. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:921
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## `

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350581084

- [review_comment] 2025-09-16T03:21:11Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:186)

```text
`
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 182 to 186, the security tests
section currently shows an incorrect go test command that passes a file path;
replace the snippet so it runs the specific test pattern instead (use go test -v
-run '^TestSignatureService_') and ensure the fenced bash block remains intact
and formatted as
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 10 | Security test doc now points to the package pattern instead of a single file. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Security test doc now points to the package pattern instead of a single file. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:957
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 216-224, the docs incorrectly

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350581089

- [review_comment] 2025-09-16T03:21:11Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:224)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 216-224, the docs incorrectly
suggest using file globbing (go test ./*.go) which fails in multi-package repos
and misuses coverage tools; update the commands to run tests across all packages
(replace ./*.go with ./...), use go test -v ./... and go test -v
-coverprofile=coverage.out ./... and then run go tool cover -func=coverage.out
(or -html=coverage.out -o coverage.html) to generate coverage reports so the
instructions work correctly in multi-package projects.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 10 | Coverage instructions now run across packages and include `go tool cover -func`. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Coverage instructions now run across packages and include `go tool cover -func`. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:993
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 228-235, the docs use a

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350581096

- [review_comment] 2025-09-16T03:21:11Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:235)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 228-235, the docs use a
working-directory change and file glob (cd test/integration + go test -v ./*.go)
which doesn't scope tests by package; update the examples to use explicit
package paths instead — replace that block with a single command using the
package path: "go test -v ./test/integration", and similarly update the
"Security Tests Only" example to use the package path (e.g. "go test -v
./test/security_test.go" or "go test -v ./test/integration -run Security"
depending on intended scope) so tests are run by package path rather than
relying on cd and file globs.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 10 | Integration & security examples now rely on explicit package paths. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Integration & security examples now rely on explicit package paths. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:1030
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 239 to 244, the benchmark

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350581101

- [review_comment] 2025-09-16T03:21:12Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:244)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 239 to 244, the benchmark
examples run tests unintentionally and target specific files; update the
commands to filter out tests with -run '^$' and run across packages with ./...
and use a proper benchmark regex for specific benchmarks (e.g.,
-bench='^BenchmarkName$') so benchmarks run only and across packages instead of
executing tests or limiting to ./*.go.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 10 | Benchmark docs now isolate benches with `-run '^$'` and module-wide paths. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Benchmark docs now isolate benches with `-run '^$'` and module-wide paths. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:1069
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In create_postmortem_tasks.py around lines 107 to 108, the dependencies array is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583501

- [review_comment] 2025-09-16T03:22:07Z by coderabbitai[bot] (create_postmortem_tasks.py:108)

```text
In create_postmortem_tasks.py around lines 107 to 108, the dependencies array is
hard-coded with ten POSTMORTEM IDs which is brittle; replace the static list
with a dynamic generation that builds the dependency list from the source of
truth (e.g., the tasks/workers list or a count) so additions/removals stay in
sync — for example, derive task IDs from the tasks collection or generate using
a formatted range like "POSTMORTEM.{:03d}".format(i) and assign that resulting
list to the "dependencies" key.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Coordinator task now derives dependencies from the generated worker tasks instead of a handwritten list. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Coordinator task now derives dependencies from the generated worker tasks instead of a handwritten list. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:1105
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In create_postmortem_tasks.py around lines 114 to 117, the two os.makedirs calls

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583510

- [review_comment] 2025-09-16T03:22:07Z by coderabbitai[bot] (create_postmortem_tasks.py:117)

```text
In create_postmortem_tasks.py around lines 114 to 117, the two os.makedirs calls
use exist_ok=True but lack error handling; wrap each directory creation in a
try/except that catches OSError, detect and handle permission/readonly errors
(e.g. errno.EACCES, errno.EROFS) and any other OS-level failures, and on error
either log a clear message with the path and errno/details and exit with a
non-zero status or re-raise a new exception with contextual information so the
script fails loudly rather than silently.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Added explicit error handling so directory creation fails fast with actionable context. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added explicit error handling so directory creation fails fast with actionable context. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:28
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In demos/lipgloss-transformation.tape lines 1–278, the demo invokes a

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583515

- [review_comment] 2025-09-16T03:22:07Z by coderabbitai[bot] (demos/lipgloss-transformation.tape:278)

```text
In demos/lipgloss-transformation.tape lines 1–278, the demo invokes a
nonexistent ./bin/queue-tui --enhanced; update the tape to run the actual
entrypoint instead (replace the "./bin/queue-tui --enhanced" line with either a
build+run sequence "go build -o bin/queue-tui ./cmd/tui && ./bin/queue-tui" or a
direct run "go run ./cmd/tui" and remove the unsupported "--enhanced" flag), or
alternatively implement and document a matching --enhanced CLI flag in
cmd/tui/main.go that toggles the enhanced view; pick one approach and make the
corresponding change so the demo invocation runs successfully.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Updated the VHS tape to invoke the actual TUI binary and removed the unsupported flag. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Updated the VHS tape to invoke the actual TUI binary and removed the unsupported flag. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:65
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deploy/deploy/data/test.txt lines 1-1 the test data is incorrectly placed

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583522

- [review_comment] 2025-09-16T03:22:07Z by coderabbitai[bot] (deploy/deploy/data/test.txt:1)

```text
In deploy/deploy/data/test.txt lines 1-1 the test data is incorrectly placed
under deploy/deploy; move the file into a Go-style testdata directory such as
testdata/producer/input.ndjson, update any source references, tests, and
CI/deploy manifests to point to the new path, and perform the move as a tracked
rename in git so history is preserved.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Replaced the placeholder text with deterministic NDJSON representing real jobs. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Replaced the placeholder text with deterministic NDJSON representing real jobs. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:138
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deploy/deploy/data/test.txt around lines 1 to 1, the file contains only the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583522

- [review_comment] 2025-09-16T03:22:07Z by coderabbitai[bot] (deploy/deploy/data/test.txt:1)

```text
In deploy/deploy/data/test.txt around lines 1 to 1, the file contains only the
useless line "test file for producer"; replace it with a deterministic NDJSON
fixture representing a real job payload (one JSON object per line) used by the
producer tests. Construct a minimal, valid payload including required fields
(e.g., id, type, payload data, timestamps, and any flags the consumer expects),
ensure values are deterministic (static IDs/timestamps), and save as NDJSON so
each test run consumes identical input for reproducible assertions.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Replaced the placeholder text with deterministic NDJSON representing real jobs. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Replaced the placeholder text with deterministic NDJSON representing real jobs. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:138
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/admin-api/Dockerfile around lines 16 to 18, the go build

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583532

- [review_comment] 2025-09-16T03:22:07Z by coderabbitai[bot] (deployments/admin-api/Dockerfile:18)

```text
In deployments/admin-api/Dockerfile around lines 16 to 18, the go build
invocation doesn't strip debug symbols, set a version variable, or trim paths
for reproducible, smaller binaries; change the build step to accept a VERSION
build-arg (ARG VERSION), add -trimpath and -ldflags '-s -w -X
main.version=${VERSION}' to the go build command (keeping CGO_ENABLED=0
GOOS=linux), so the produced admin-api binary is stripped, has embedded version
metadata, and uses reproducible paths.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Docker build now trims paths, strips symbols, and honors a VERSION build arg. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Docker build now trims paths, strips symbols, and honors a VERSION build arg. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:223
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/admin-api/Dockerfile around lines 20 to 23, the image currently

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583546

- [review_comment] 2025-09-16T03:22:08Z by coderabbitai[bot] (deployments/admin-api/Dockerfile:23)

```text
In deployments/admin-api/Dockerfile around lines 20 to 23, the image currently
uses unpinned alpine:latest and runs as root; change to a specific, pinned
Alpine version (e.g., alpine:3.18 or a project-approved tag) and create a
non-root user/group before switching to it: install packages as root, create a
dedicated user and group, create and set ownership of any workdir/home, drop
privileges with USER <username>, and ensure file permissions are set so the
container does not run as root at runtime.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Runtime image now pins Alpine, drops root, and avoids baking configs. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Runtime image now pins Alpine, drops root, and avoids baking configs. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:260
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/admin-api/Dockerfile around lines 26 to 31, the Dockerfile

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583553

- [review_comment] 2025-09-16T03:22:08Z by coderabbitai[bot] (deployments/admin-api/Dockerfile:31)

```text
In deployments/admin-api/Dockerfile around lines 26 to 31, the Dockerfile
currently copies an environment-specific config into the image (COPY
--from=builder /app/configs/admin-api.yaml ./configs/); remove that COPY so
environment-specific configs are not baked into the image and instead rely on
runtime mounting (volume/ConfigMap/Secret). Update the Dockerfile to stop
copying configs, ensure the image expects configs at a runtime path (e.g.,
./configs/) and add a brief comment indicating configs must be mounted at
runtime via your deployment manifests.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 10 | Fixed the table of contents slug so it matches the section heading. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Fixed the table of contents slug so it matches the section heading. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:297
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/docker/Dockerfile.admin-api around lines 20-21, the Go build

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583558

- [review_comment] 2025-09-16T03:22:08Z by coderabbitai[bot] (deployments/docker/Dockerfile.admin-api:21)

```text
In deployments/docker/Dockerfile.admin-api around lines 20-21, the Go build
command should strip debug info, embed a version, and make builds reproducible;
change the build to add -trimpath and -ldflags (e.g. -ldflags "-s -w -buildid=
-X main.version=${VERSION}") to bake in version and remove symbol tables, then
run strip on the resulting binary (or use 'go build' with -ldflags "-s -w" and
follow with 'strip admin-api') so the image is smaller and builds are
reproducible; ensure VERSION is provided via build-arg or environment.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Documented the baseline environment and introduced the new bench payload flag. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Documented the baseline environment and introduced the new bench payload flag. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:335
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/docker/Dockerfile.admin-api around lines 38-40, the COPY line

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583572

- [review_comment] 2025-09-16T03:22:08Z by coderabbitai[bot] (deployments/docker/Dockerfile.admin-api:40)

```text
In deployments/docker/Dockerfile.admin-api around lines 38-40, the COPY line
assumes /app/configs/admin-api.yaml exists in the builder stage which makes the
build fail when it’s absent; fix by guaranteeing the file always exists in the
builder stage (add a step in the builder stage to mkdir -p /app/configs and
create a default admin-api.yaml if missing, or copy a repository default config
into that path) so the final-stage COPY is deterministic and never errors.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Switched the documentation to the interface-based collector example. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Switched the documentation to the interface-based collector example. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:393
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/03_milestones.md around lines 11-13 (and similarly at lines 46-51), the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583575

- [review_comment] 2025-09-16T03:22:08Z by coderabbitai[bot] (docs/03_milestones.md:13)

```text
In docs/03_milestones.md around lines 11-13 (and similarly at lines 46-51), the
table-of-contents uses the incorrect anchor `#gono-go-decision-gates`; update
those links to the correct slug `#go-no-go-decision-gates` (and search the file
for any other occurrences of `gono-go` to replace) so the ToC links point to the
actual section anchor.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Added explicit `BurnRateThresholds` type with documented fields. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added explicit `BurnRateThresholds` type with documented fields. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:492
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/anomaly-radar-slo-budget.md around lines 36-38 (and also apply the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583587

- [review_comment] 2025-09-16T03:22:09Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:38)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 36-38 (and also apply the
same change at 52-75 and 475-483), the Go import uses the long package path
"github.com/flyingrobots/go-redis-work-queue/internal/anomaly-radar-slo-budget";
update the examples to alias this import to a short, readable identifier (e.g.,
ars or slo) and update all references in the examples accordingly so the code is
concise and consistent across the noted ranges.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Added a public re-export at `pkg/anomaly-radar-slo-budget` and updated docs to reference the stable path. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added a public re-export at `pkg/anomaly-radar-slo-budget` and updated docs to reference the stable path. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:637
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/anomaly-radar-slo-budget.md around lines 36 to 38, the docs instruct

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583587

- [review_comment] 2025-09-16T03:22:09Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:38)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 36 to 38, the docs instruct
users to import an internal package path which is not accessible to consumers;
update the documentation to point to the public exported package path (the
module's published import path for the anomaly-radar-slo-budget package), ensure
the package is exported (move/rename from internal if necessary or add a public
wrapper package), and replace the internal import line with the correct public
import path that consumers can use.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Added a public re-export at `pkg/anomaly-radar-slo-budget` and updated docs to reference the stable path. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added a public re-export at `pkg/anomaly-radar-slo-budget` and updated docs to reference the stable path. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:637
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/anomaly-radar-slo-budget.md around lines 124-176 (and apply same

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583610

- [review_comment] 2025-09-16T03:22:09Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:176)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 124-176 (and apply same
changes to ranges 179-218, 220-238, 240-270, 272-296, 298-327, 329-350,
352-369): the API endpoint docs lack an auth model, standard error schema, and
explicit Content-Type and status codes; add a short "Authentication &
Authorization" subsection stating the auth scheme (e.g., Bearer token) and
required RBAC roles for the endpoint, add a "Errors" subsection with the
standard JSON error shape (fields: error.code, error.message, error.details,
request_id) and list applicable HTTP error responses (401, 403, 422, 429 where
relevant) and finally specify response Content-Type (application/json) for
request and response examples so every endpoint doc includes auth, error schema,
relevant status codes, and Content-Type.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Added explicit RFC3339/UTC notes under every response example. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added explicit RFC3339/UTC notes under every response example. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:808
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## docs/api/anomaly-radar-slo-budget.md around lines 128 to 176: the response

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583616

- [review_comment] 2025-09-16T03:22:09Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:176)

```text
docs/api/anomaly-radar-slo-budget.md around lines 128 to 176: the response
example lacks an explicit timezone policy for timestamps; add a short sentence
directly beneath each response example block stating that all timestamps are
formatted as RFC3339 in UTC and include the trailing "Z" (e.g., "All timestamps
are RFC3339 in UTC and use the trailing 'Z'"), ensuring the note appears under
every response example in the file.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Added explicit RFC3339/UTC notes under every response example. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added explicit RFC3339/UTC notes under every response example. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:808
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/anomaly-radar-slo-budget.md around lines 220 to 238, the PUT

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583622

- [review_comment] 2025-09-16T03:22:09Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:238)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 220 to 238, the PUT
/api/v1/anomaly-radar/config section lacks semantics for validation, unknown
fields, partial updates and immutability; update the doc to state whether PUT
performs a full replace or a deep PATCH-like merge (choose one and describe
behavior for nested objects), enumerate validation rules (e.g.,
availability_target must be between 0 and 1 inclusive, latency_threshold_ms
positive integer, threshold rates between 0 and 1, types), state that invalid
input returns HTTP 422 with a JSON body listing field errors, explain how
unknown fields are handled (reject with 400 or ignore with warning), and call
out any fields that are immutable at runtime and require a restart to change.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Documented PUT as a full replace with explicit validation, 400/422 behaviours, and immutability notes. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Documented PUT as a full replace with explicit validation, 400/422 behaviours, and immutability notes. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:880
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/anomaly-radar-slo-budget.md around lines 246-249, the query param

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583631

- [review_comment] 2025-09-16T03:22:10Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:249)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 246-249, the query param
`max_samples` is underspecified and can lead to huge responses; set and document
a sensible default (e.g., default=1000) and a hard upper bound (e.g.,
max=10_000), add an optional pagination token query param (e.g., `next_cursor`)
and show the paginated response structure including `metrics`, `count`, and
`next_cursor` (opaque token) so callers know how to request subsequent pages;
also mention that `max_samples` cannot exceed the hard limit and that server
will return `count` and `next_cursor=null` when no more data exists.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Documented pagination defaults, limits, and `next_cursor` contract alongside the updated handlers. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Documented pagination defaults, limits, and `next_cursor` contract alongside the updated handlers. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:920
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/anomaly-radar-slo-budget.md around lines 250-266 (and similarly for

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583635

- [review_comment] 2025-09-16T03:22:10Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:266)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 250-266 (and similarly for
lines 339-350), the example metrics response omits p90 while the percentiles
endpoint includes it; update the historical metrics payload to include a
p90_latency_ms field for each metric entry (matching the same format and units
as p50/p95/p99) so both endpoints use the same percentile set, and ensure the
surrounding documentation text reflects that the metrics include
p50/p90/p95/p99.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Added `p90_latency_ms` to the metrics response and updated the descriptive text. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added `p90_latency_ms` to the metrics response and updated the descriptive text. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:958
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/anomaly-radar-slo-budget.md around lines 371 to 374, the health

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583645

- [review_comment] 2025-09-16T03:22:10Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:374)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 371 to 374, the health
endpoint HTTP status codes are too limited; add entries for 500, 429, and 206
with brief conditions: 500 Internal Server Error for unexpected/internal
failures, 429 Too Many Requests when collectors or clients are being throttled,
and 206 Partial Content when the endpoint returns partial or degraded data
(include a short parenthetical or one-line condition for each). Ensure
formatting matches the existing bullet list and keep descriptions concise.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Documented additional status codes (500/429/206) for the health endpoint. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Documented additional status codes (500/429/206) for the health endpoint. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:1012
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/anomaly-radar-slo-budget.md around lines 375 to 383, the Start/Stop

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583651

- [review_comment] 2025-09-16T03:22:10Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:383)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 375 to 383, the Start/Stop
section lacks idempotency and authorization details; update the docs to
explicitly define semantics for repeated calls (POST /api/v1/anomaly-radar/start
returns 200 OK with a message "already started" if radar is running, otherwise
202 Accepted or 200 OK when started; POST /api/v1/anomaly-radar/stop returns 200
OK with "already stopped" if not running, otherwise 202/200 when stopping), list
required authentication/authorization (e.g., requires bearer token with role
"slo_admin" or permission "anomaly_radar:manage"), document possible status
codes and example request/response bodies for both success, idempotent-no-op,
and unauthorized (401/403) cases, and add note about concurrency handling
(server guarantees idempotent behavior and returns current state) so callers
know double start/stop are safe.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Documented auth requirements, idempotent responses, and concurrency guarantees for start/stop. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Documented auth requirements, idempotent responses, and concurrency guarantees for start/stop. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:1049
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/anomaly-radar-slo-budget.md around lines 435 to 441, the "Batch

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583660

- [review_comment] 2025-09-16T03:22:10Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:441)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 435 to 441, the "Batch
Operations: Use batch endpoints for efficient data retrieval" bullet references
endpoints that are not documented; remove this bullet or add a new "Batch
endpoints" section detailing the routes and request/response shapes. If
removing, delete bullet 4 and renumber/adjust wording to keep the list coherent;
if adding, create a new subsection immediately after the Performance list that
documents each batch route (path, method), expected request payload, response
schema, and example use-cases so the docs are accurate and not misleading.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Removed the stray batch-operations bullet until matching endpoints exist. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Removed the stray batch-operations bullet until matching endpoints exist. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:1091
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/anomaly-radar-slo-budget.md around lines 488 to 503, the Prometheus

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583663

- [review_comment] 2025-09-16T03:22:10Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:503)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 488 to 503, the Prometheus
exporter incorrectly calls Inc() during scrape which mutates metrics on each
scrape; instead compute counts and set gauges idempotently. Change the loop to
tally active alert counts by severity, call
alertCountVec.WithLabelValues(sev).Set(count) for each severity, and ensure any
previously-exposed severity labels not present are either Set(0) or removed
(e.g., DeleteLabelValues) so the exporter is fully idempotent.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Updated the exporter example to tally counts and use `Set` for idempotent scrapes. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Updated the exporter example to tally counts and use `Set` for idempotent scrapes. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:1129
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/anomaly-radar-slo-budget.md around lines 519 to 521, the README

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583669

- [review_comment] 2025-09-16T03:22:10Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:521)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 519 to 521, the README
references a /debug endpoint without specifying its contract; either fully
document the endpoint or remove the snippet. If keeping it, add a clear spec:
HTTP method, full path, required auth/headers, request body or query params with
types and validation rules, example request (curl) and example successful and
error responses with status codes and JSON schema; if removing it, delete the
curl snippet and any other mentions of /debug in this doc to avoid misleading
users.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Removed the undocumented `/debug` callout to keep the guide aligned with supported endpoints. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Removed the undocumented `/debug` callout to keep the guide aligned with supported endpoints. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:31
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/chaos-harness.md around lines 259 to 283, the example imports an

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583678

- [review_comment] 2025-09-16T03:22:10Z by coderabbitai[bot] (docs/api/chaos-harness.md:283)

```text
In docs/api/chaos-harness.md around lines 259 to 283, the example imports an
internal package which cannot be imported outside the module; change the import
to the public package path (for example
github.com/flyingrobots/go-redis-work-queue/pkg/chaosharness) and update the
example to use that package name, and if the chaosharness code currently lives
under internal/ move or re-export it under pkg/chaosharness (or otherwise make
it publicly importable) so the example compiles for external users.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Added a `pkg/chaos-harness` re-export and updated the samples to import it via an explicit alias. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added a `pkg/chaos-harness` re-export and updated the samples to import it via an explicit alias. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:70
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/dlq-remediation-ui.md around lines 7 to 10, the documentation

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583687

- [review_comment] 2025-09-16T03:22:11Z by coderabbitai[bot] (docs/api/dlq-remediation-ui.md:10)

```text
In docs/api/dlq-remediation-ui.md around lines 7 to 10, the documentation
currently states the API is unauthenticated; update the docs and implementation
guidance to require authentication and role-based access control for any
endpoints that can requeue or purge jobs. Replace the “no authentication” note
with explicit instructions that anonymous access is banned, list the required
authentication mechanism (e.g., JWT bearer tokens or mTLS) and required RBAC
roles/permissions (e.g., admin:dlq:manage), and add an example request header
and a short note to enforce auth in production deployments and tests. Ensure the
docs also call out auditing/logging for destructive actions and recommend
least-privilege role assignment.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Authentication section now requires bearer tokens/CSRF, documents RBAC scopes, and shows a concrete header example with auditing guidance. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Authentication section now requires bearer tokens/CSRF, documents RBAC scopes, and shows a concrete header example with auditing guidance. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:108
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/dlq-remediation-ui.md around lines 25 to 39, the docs currently

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583693

- [review_comment] 2025-09-16T03:22:11Z by coderabbitai[bot] (docs/api/dlq-remediation-ui.md:39)

```text
In docs/api/dlq-remediation-ui.md around lines 25 to 39, the docs currently
state "max 1000" and mention "rate limits" ambiguously; update the text to
explicitly state that the page_size maximum of 1000 is enforced server-side and
that any rate limits are enforced server-side as well (include where applicable:
page_size and any API rate limiting behavior), e.g., change descriptive cells to
assert server-side enforcement and add a short note clarifying that requests
exceeding page_size or rate limits will be rejected with appropriate HTTP error
codes and messages.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Clarified that `page_size` and rate limits are enforced server-side and documented the rejection behaviour. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Clarified that `page_size` and rate limits are enforced server-side and documented the rejection behaviour. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:149
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/dlq-remediation-ui.md around lines 223-241, the purge-all endpoint

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583698

- [review_comment] 2025-09-16T03:22:11Z by coderabbitai[bot] (docs/api/dlq-remediation-ui.md:241)

```text
In docs/api/dlq-remediation-ui.md around lines 223-241, the purge-all endpoint
currently uses a dangerous confirm=true query parameter; change the docs to
require a POST body with a boolean "confirm" and a "filter" object
(queue,type,error_pattern,start_time,end_time,min_attempts,max_attempts) instead
of query confirm, and document that callers MUST supply an Idempotency-Key
header and a privileged scope/permission to invoke this operation; add
requirements to validate the request body schema, enforce the privileged scope
check on the server, persist the Idempotency-Key with the operation result to
make the purge idempotent and return the stored result (or a 409/appropriate
response) for duplicate keys, and update the endpoint example and params table
to show the request body and required Idempotency-Key header instead of a
confirm query param.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Purge-all now takes a POST body with `confirm`, structured filters, and requires an `Idempotency-Key` header. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Purge-all now takes a POST body with `confirm`, structured filters, and requires an `Idempotency-Key` header. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:188
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/dlq-remediation-ui.md around lines 247 to 257, the JSON example

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583704

- [review_comment] 2025-09-16T03:22:11Z by coderabbitai[bot] (docs/api/dlq-remediation-ui.md:257)

```text
In docs/api/dlq-remediation-ui.md around lines 247 to 257, the JSON example
contains an invalid string "..." inside the successful array; replace it with
valid JSON by either listing only the real example entries (e.g. two sample IDs)
or by truncating the array (e.g. show the first two entries and remove the
ellipsis entirely), keeping the rest of the response fields unchanged so the
example is valid JSON.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 6 | Cleaned the purge-all response example so it is valid JSON without ellipses. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Cleaned the purge-all response example so it is valid JSON without ellipses. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:231
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/dlq-remediation-ui.md around lines 299 to 307, the HTTP status codes

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583715

- [review_comment] 2025-09-16T03:22:11Z by coderabbitai[bot] (docs/api/dlq-remediation-ui.md:307)

```text
In docs/api/dlq-remediation-ui.md around lines 299 to 307, the HTTP status codes
table omits authentication/authorization responses; add rows for 401
Unauthorized and 403 Forbidden with concise descriptions (e.g., "401 -
Unauthorized: Authentication required or invalid credentials" and "403 -
Forbidden: Authenticated but insufficient permissions") and ensure the security
section references these codes where relevant so auth failures are documented
alongside other status codes.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 6 | Added 401/403/429 rows to the status table so auth and throttling failures are documented. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added 401/403/429 rows to the status table so auth and throttling failures are documented. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:268
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 246 to 262, the listed

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583735

- [review_comment] 2025-09-16T03:22:12Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:262)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 246 to 262, the listed
performance numbers are presented without context and are misleading; update the
section to replace the single numeric estimates with concrete measurement
details: for each metric state the test environment (hardware, OS, network—e.g.,
localhost vs remote), payload sizes, concurrency level, measurement
tool/version, sample size and duration, and the statistical results (p50/p95/p99
plus mean and standard deviation) and link to raw test scripts/logs; remove or
qualify any numbers that cannot be reproduced and, where appropriate, note
whether the metric was measured under unit, integration, or load-testing
conditions so readers can reproduce the results.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Replaced bare timing bullets with reproducible benchmark tables (environment, workload, p50/p95/p99) and linked commands. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Replaced bare timing bullets with reproducible benchmark tables (environment, workload, p50/p95/p99) and linked commands. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:306
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In .github/workflows/markdownlint.yml around line 3 the reserved YAML key on is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679306

- [review_comment] 2025-09-16T14:19:23Z by coderabbitai[bot] (.github/workflows/markdownlint.yml:3)

```text
In .github/workflows/markdownlint.yml around line 3 the reserved YAML key on is
unquoted and triggers yamllint; either quote the key by changing on: to "on":
(or 'on':) or add a yamllint disable directive for that line (e.g. the
appropriate yamllint disable-line comment) so the linter is silenced while
preserving the existing workflow semantics.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 10 | Quoted the reserved `on` key so yamllint stops complaining. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Quoted the reserved `on` key so yamllint stops complaining. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:347
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## .github/workflows/update-progress.yml lines 16-24: the workflow currently

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679317

- [review_comment] 2025-09-16T14:19:23Z by coderabbitai[bot] (.github/workflows/update-progress.yml:24)

```text
.github/workflows/update-progress.yml lines 16-24: the workflow currently
references actions with mutable tags (actions/checkout@v4 and
actions/setup-python@v5); replace those two uses: entries with the recommended
immutable commit SHAs provided
(actions/checkout@08eba0b27e820071cde6df949e0beb9ba4906955 and
actions/setup-python@a26af69be951a213d495a4c3e4e4022e16d87065) so the workflow
is pinned to fixed commits while preserving existing with: options (fetch-depth
and python-version).
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Pinned checkout/setup-python to immutable SHAs to avoid surprise upgrades. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Pinned checkout/setup-python to immutable SHAs to avoid surprise upgrades. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:383
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In append_metadata.py around lines 30 to 33, format_list currently returns "-

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679359

- [review_comment] 2025-09-16T14:19:24Z by coderabbitai[bot] (append_metadata.py:33)

```text
In append_metadata.py around lines 30 to 33, format_list currently returns "-
[]" for empty lists which emits YAML that is a list with a string "[]"; change
format_list to emit a proper YAML empty list (i.e., the key followed by "[]" on
the same line without a dash) when items is empty, and when items exist produce
one "- item" per line with the prefix preserved; then update
generate_yaml_metadata to inline empty collections for soft/enables/provides on
the same line (key: []) rather than emitting a dash line, and apply this same
empty-list handling to soft, enables, and provides.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Empty dependency lists now render as `field: []` while populated lists keep bullet formatting. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Empty dependency lists now render as `field: []` while populated lists keep bullet formatting. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:493
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In append_metadata.py around lines 35 to 59 (and also apply same pattern to

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679374

- [review_comment] 2025-09-16T14:19:24Z by coderabbitai[bot] (append_metadata.py:59)

```text
In append_metadata.py around lines 35 to 59 (and also apply same pattern to
lines 125-131), the script currently mutates files on import; wrap the
procedural code in a main() function and add an if __name__ == "__main__":
main() guard so nothing runs on import, surround file-processing logic with
try/except to log or print errors and continue safely, and ensure the target
DAG/ideas directory exists before writing (os.makedirs(ideas_dir,
exist_ok=True)) to avoid missing-directory errors; update the referenced lines
accordingly to call main and handle exceptions.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | File updates now run under an explicit `main()` with defensive directory creation and error logging. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> File updates now run under an explicit `main()` with defensive directory creation and error logging. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:532
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In BUGS.md around lines 12 to 25, the heartbeat snippet is unsafe: it ignores

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679383

- [review_comment] 2025-09-16T14:19:24Z by coderabbitai[bot] (BUGS.md:25)

```text
In BUGS.md around lines 12 to 25, the heartbeat snippet is unsafe: it ignores
errors, can lose the key TTL, and doesn’t create the heartbeat atomically. Fix
by (1) creating the hb key atomically at start using Redis SET with NX and PX
(or PEXPIRE immediately after a successful SET), (2) renewing using SET with XX
and PX (or PEXPIRE) so you only extend an existing key, (3) adding randomized
jitter to the renewal interval (e.g., renew at TTL/3 ± jitter) to avoid
thundering herds, (4) checking and logging Redis command errors and handling
transient failures with a small retry/backoff or escalate and exit the worker if
renewals repeatedly fail, and (5) ensuring the ticker is stopped and cancel() is
called before performing final LREM/DEL so the background renew goroutine exits
cleanly.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Heartbeat guidance now uses atomic SET NX/XX with jittered renewals and error handling. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Heartbeat guidance now uses atomic SET NX/XX with jittered renewals and error handling. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:571
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In BUGS.md around lines 32-38, the current code scans Redis with pattern

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679399

- [review_comment] 2025-09-16T14:19:25Z by coderabbitai[bot] (BUGS.md:38)

```text
In BUGS.md around lines 32-38, the current code scans Redis with pattern
"jobqueue:*" which floods Redis Cluster across slots; change the reaper to
either (A) maintain and iterate a registry (e.g., a SET of active worker/queue
IDs) and use SMEMBERS to fetch per-worker keys, or (B) use Redis hash-tags so
all jobqueue keys for a worker live in the same slot (e.g.,
jobqueue:{workerID}:processing) and then SCAN only that single slot per worker;
implement one of these approaches and update the description and example in the
doc to show scanning per-worker keys or iterating the SET instead of a global
SCAN "jobqueue:*".
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Guidance now directs operators to maintain a worker registry and hash-tagged processing keys instead of global scans. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Guidance now directs operators to maintain a worker registry and hash-tagged processing keys instead of global scans. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:613
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In BUGS.md around lines 45–53 the scheduler mover uses ZRANGEBYSCORE + ZREM +

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679403

- [review_comment] 2025-09-16T14:19:25Z by coderabbitai[bot] (BUGS.md:53)

```text
In BUGS.md around lines 45–53 the scheduler mover uses ZRANGEBYSCORE + ZREM +
LPUSH in a pipeline which is racy and can double‑enqueue under concurrency;
replace this with an atomic operation: either implement a loop using ZPOPMIN to
pop due members and LPUSH them (ensuring you process only members with score <=
now) or, preferably, write a single Lua script that atomically pops all due
members from the sorted set and pushes them to the list in one server-side
operation; update the code to call the Lua script (or ZPOPMIN loop) and remove
the non‑atomic pipeline approach so deliveries cannot be duplicated.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Replaced the non-atomic pipeline advice with a ZPOPMIN/Lua-based atomic mover. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Replaced the non-atomic pipeline advice with a ZPOPMIN/Lua-based atomic mover. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:653
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In BUGS.md around lines 55–61, the advice to write NDJSON ledger files to local

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679408

- [review_comment] 2025-09-16T14:19:25Z by coderabbitai[bot] (BUGS.md:61)

```text
In BUGS.md around lines 55–61, the advice to write NDJSON ledger files to local
disk risks data loss and PII leakage; change the guidance to require routing
ledger events to a durable sink (S3/Blob store, Kafka, or equivalent) or, if
local append-only files are used, mandate atomic appenders with daily rotation,
size caps, gzip/archive rotation, and documented retention; require payload
redaction/scrubbing of PII before writing (or write only event metadata and
references), implement backpressure and retry/fallback logic on IO failures so
handlers are not blocked, and add documentation for retention/rotation policy
and monitoring/alerts.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Ledger guidance now points to durable sinks (or guarded local appenders) with rotation, retention, and PII scrubbing. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Ledger guidance now points to durable sinks (or guarded local appenders) with rotation, retention, and PII scrubbing. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:692
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In claude_worker.py around lines 29 to 34, only self.my_dir is created but other

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679417

- [review_comment] 2025-09-16T14:19:25Z by coderabbitai[bot] (claude_worker.py:34)

```text
In claude_worker.py around lines 29 to 34, only self.my_dir is created but other
directories (self.finished_dir, self.failed_dir, self.help_dir, and
self.open_tasks_dir) are written to later and may not exist; update the
initialization to mkdir(parents=True, exist_ok=True) for each of those directory
attributes as well (create self.finished_dir, self.failed_dir, self.help_dir,
and self.open_tasks_dir up front) so all required directories exist before any
file operations.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Worker initialisation now ensures open/finished/failed/help directories exist before use. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Worker initialisation now ensures open/finished/failed/help directories exist before use. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:732
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In cmd/job-queue-system/main.go around lines 159-161 (and similarly at 169-171

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679439

- [review_comment] 2025-09-16T14:19:25Z by coderabbitai[bot] (cmd/job-queue-system/main.go:161)

```text
In cmd/job-queue-system/main.go around lines 159-161 (and similarly at 169-171
and 193-201) you’re ignoring json.Marshal errors and printing possibly corrupt
output; replace json.MarshalIndent usage with a json.Encoder writing to
os.Stdout (use encoder.SetIndent("", "  ") for pretty output) and call
encoder.Encode(value), check and handle the returned error (log to stderr and
return/exit with non-zero status or propagate the error) instead of discarding
it; apply the same pattern to every JSON output in runAdmin.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Swapped `json.MarshalIndent` for a shared `json.Encoder` helper that logs encode failures. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Swapped `json.MarshalIndent` for a shared `json.Encoder` helper that logs encode failures. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:770
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In create_postmortem_tasks.py around lines 1-3, the code appends "Z" to a naive

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679449

- [review_comment] 2025-09-16T14:19:25Z by coderabbitai[bot] (create_postmortem_tasks.py:3)

```text
In create_postmortem_tasks.py around lines 1-3, the code appends "Z" to a naive
local datetime which is incorrect; instead import and use a timezone-aware UTC
datetime (e.g., add "from datetime import timezone" and call
datetime.now(timezone.utc)) and emit an ISO8601 UTC timestamp (convert to ISO
format and normalize to Z or use isoformat with UTC) so timestamps are real Zulu
time rather than local time with a trailing "Z".
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Timestamps now use `datetime.now(timezone.utc).isoformat().replace("+00:00", "Z")`. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Timestamps now use `datetime.now(timezone.utc).isoformat().replace("+00:00", "Z")`. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:808
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In create_postmortem_tasks.py around lines 15 to 16, the code appends "Z" to

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679454

- [review_comment] 2025-09-16T14:19:26Z by coderabbitai[bot] (create_postmortem_tasks.py:16)

```text
In create_postmortem_tasks.py around lines 15 to 16, the code appends "Z" to
datetime.now() which produces a fake UTC timestamp; update both timestamp sites
to produce real UTC-aware ISO timestamps by using
datetime.now(timezone.utc).isoformat() (or
datetime.utcnow().replace(tzinfo=timezone.utc).isoformat()), import timezone
from datetime, and remove the manual "Z" concatenation so timestamps are
properly UTC-aware.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Same UTC fix applied to the worker task loop. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Same UTC fix applied to the worker task loop. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:845
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In create_postmortem_tasks.py around lines 27 to 39, the criteria strings still

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679460

- [review_comment] 2025-09-16T14:19:26Z by coderabbitai[bot] (create_postmortem_tasks.py:39)

```text
In create_postmortem_tasks.py around lines 27 to 39, the criteria strings still
contain raw placeholders "{i}" and "{worker}" which won't be interpolated;
replace those literal placeholders with properly formatted/interpolated strings
(e.g., use f-strings or .format with the loop variables i and worker) so the
criteria read the actual worker index and name, and ensure any literal braces
intended to stay are escaped; update the list entries to build their strings
dynamically using the current i/worker variables.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 6 | Criteria now use f-strings, so the rendered instructions include the correct worker ID/name. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Criteria now use f-strings, so the rendered instructions include the correct worker ID/name. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:883
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In create_postmortem_tasks.py around lines 69-70, the timestamp is created with

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679467

- [review_comment] 2025-09-16T14:19:26Z by coderabbitai[bot] (create_postmortem_tasks.py:70)

```text
In create_postmortem_tasks.py around lines 69-70, the timestamp is created with
datetime.now(), which uses local time; change it to produce a UTC timestamp
(e.g., use datetime.utcnow().isoformat() + "Z" or
datetime.now(timezone.utc).isoformat()) and apply the same UTC-fix to the
coordinator task elsewhere in this file so all created_at values are uniform and
in UTC.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Coordinator task timestamps now use the UTC helper as well. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Coordinator task timestamps now use the UTC helper as well. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:921
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In create_review_tasks.py around lines 10-11, the check for "duplicate" is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679474

- [review_comment] 2025-09-16T14:19:26Z by coderabbitai[bot] (create_review_tasks.py:11)

```text
In create_review_tasks.py around lines 10-11, the check for "duplicate" is
case-sensitive so files like "Duplicate..." slip through; change the condition
to test against a lowercased filename (e.g., use 'duplicate' not in f.lower())
while keeping the .json check (you can call f.lower() for the duplicate check
only) so task_id = f[:-5] remains unchanged.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Duplicate filter now lowercases filenames before checking. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Duplicate filter now lowercases filenames before checking. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:958
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In create_review_tasks.py around lines 14 to 21, the code uses a bare except

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679482

- [review_comment] 2025-09-16T14:19:26Z by coderabbitai[bot] (create_review_tasks.py:21)

```text
In create_review_tasks.py around lines 14 to 21, the code uses a bare except
which hides real errors; replace it by catching the specific exceptions that can
occur (e.g., IndexError and ValueError) when splitting/parsing task_id and,
instead of a silent pass, either continue the loop or log the parse failure; for
any truly unexpected exception re-raise or log and raise so real bugs aren’t
masked.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Invalid filenames now log a skip message instead of being silently swallowed. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Invalid filenames now log a skip message instead of being silently swallowed. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:994
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In demos/responsive-tui.tape around lines 72-73 (and also at 129-130, 214-215,

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679493

- [review_comment] 2025-09-16T14:19:26Z by coderabbitai[bot] (demos/responsive-tui.tape:73)

```text
In demos/responsive-tui.tape around lines 72-73 (and also at 129-130, 214-215,
307-308, 365), the script uses "Sleep 3s" which wastes CI minutes; remove these
Sleep commands and instead either remove the pause entirely or replace with a
deterministic check/wait-for-condition (e.g., wait for expected output or
prompt) so the test proceeds immediately when ready; update the surrounding
steps to rely on explicit assertions or readiness checks rather than fixed
sleeps.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Removed all `Sleep` directives from the tape so tests run without artificial waits. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Removed all `Sleep` directives from the tape so tests run without artificial waits. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:1031
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In dependency_analysis.py around lines 23–166, there’s a naming inconsistency:

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679502

- [review_comment] 2025-09-16T14:19:27Z by coderabbitai[bot] (dependency_analysis.py:44)

```text
In dependency_analysis.py around lines 23–166, there’s a naming inconsistency:
keys were defined with hyphens (e.g., "distributed-tracing-integration") but
code still references snake_case names ("distributed_tracing",
"storage_backend"); update all occurrences listed (distributed_tracing at lines
23, 41, 89, 107, 166 and storage_backend at line 112) to use the canonical
hyphenated keys, or alternatively add an explicit alias mapping near the top
that maps "distributed_tracing" -> "distributed-tracing-integration" and
"storage_backend" -> "storage-backends" and then use the canonical names
throughout; after making the edits run the provided rg command to confirm zero
unresolved nodes.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Dependencies now reference `distributed-tracing-integration`/`storage-backends`; alias map handles legacy names. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Dependencies now reference `distributed-tracing-integration`/`storage-backends`; alias map handles legacy names. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:1069
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/admin-api/k8s-redis.yaml around lines 16 to 52, the Pod and

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679524

- [review_comment] 2025-09-16T14:19:27Z by coderabbitai[bot] (deployments/admin-api/k8s-redis.yaml:52)

```text
In deployments/admin-api/k8s-redis.yaml around lines 16 to 52, the Pod and
container lack required security hardening: add automountServiceAccountToken:
false at the pod spec level; add pod-level securityContext with seccompProfile:
{ type: RuntimeDefault }, runAsNonRoot: true and fsGroup: 10001; for the redis
container set imagePullPolicy: Always and pin the image to a digest
(redis:7-alpine@sha256:...); add container securityContext with runAsUser:
10001, runAsGroup: 10001, allowPrivilegeEscalation: false, capabilities: { drop:
["ALL"] } and readOnlyRootFilesystem: true; add an emptyDir mount at /tmp and
keep the redis-data PVC mounted but ensure ownership/permissions allow write by
UID 10001 (use fsGroup and init or securityContext mechanisms); and replace the
readinessProbe exec with a shell command like sh -c 'redis-cli -h 127.0.0.1 -p
6379 ping' for a robust probe.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Added the requested hardening (SA token off, RuntimeDefault, pinned digest, read-only root, /tmp emptyDir, shell probes). |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added the requested hardening (SA token off, RuntimeDefault, pinned digest, read-only root, /tmp emptyDir, shell probes). Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:1147
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/docker/Dockerfile.rbac-token-service around lines 25-27 (and also

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679556

- [review_comment] 2025-09-16T14:19:27Z by coderabbitai[bot] (deployments/docker/Dockerfile.rbac-token-service:27)

```text
In deployments/docker/Dockerfile.rbac-token-service around lines 25-27 (and also
update the similar install at lines 53-54), the image currently installs
ca-certificates, tzdata and wget but the project standard prefers curl for
healthchecks; update the RUN apk --no-cache add ... commands to include curl so
curl is available at runtime for healthcheck scripts and make the same change
for the duplicate install at lines 53-54.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Verified runtime image now ships curl and the healthcheck uses it. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Verified runtime image now ships curl and the healthcheck uses it. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:31
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/00_assessment.md around line 3, the "Last updated: 2025-09-12" header is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679567

- [review_comment] 2025-09-16T14:19:27Z by coderabbitai[bot] (docs/00_assessment.md:3)

```text
In docs/00_assessment.md around line 3, the "Last updated: 2025-09-12" header is
stale for this 2025-09-16 PR; update the timestamp to the current PR date (e.g.,
"Last updated: 2025-09-16") or remove the line entirely if you prefer not to
track last-updated metadata.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Replaced the stale timestamp convention with an evergreen status line per maintainer guidance. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Replaced the stale timestamp convention with an evergreen status line per maintainer guidance. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:68
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/02_release_plan.md around lines 6–7, the release plan text needs

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679574

- [review_comment] 2025-09-16T14:19:28Z by coderabbitai[bot] (docs/02_release_plan.md:7)

```text
In docs/02_release_plan.md around lines 6–7, the release plan text needs
explicit freeze windows, rollout/rollback and go/no‑go gates: add a 48–72h code
freeze before each milestone date (2025-09-26, 2025-10-10, 2025-10-24,
2025-11-07), list the required sign‑off owners for each gate and a short
rollback plan for each release, and add “no High/Critical CVEs” plus
“govulncheck clean” as hard pre‑release gates; update the milestone entries to
include these freeze window dates, the sign‑off owner field, and a one‑line
rollback procedure for each.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Rebuilt the release plan with explicit freezes, owners, rollback notes, and security gates. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Rebuilt the release plan with explicit freezes, owners, rollback notes, and security gates. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:103
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/10_risk_register.md around line 3, the "Last updated" timestamp is stale

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679581

- [review_comment] 2025-09-16T14:19:28Z by coderabbitai[bot] (docs/10_risk_register.md:3)

```text
In docs/10_risk_register.md around line 3, the "Last updated" timestamp is stale
(2025-09-12); update that line to "Last updated: 2025-09-16" and save the file,
and ensure the updated commit includes this change (or implement a quick check
to update the date when editing this file in future).
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Converted the stale timestamp to the new status wording called out by the maintainer. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Converted the stale timestamp to the new status wording called out by the maintainer. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:142
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/admin-api.md around lines 356 to 359, the docs mention a “minimum

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679589

- [review_comment] 2025-09-16T14:19:28Z by coderabbitai[bot] (docs/api/admin-api.md:359)

```text
In docs/api/admin-api.md around lines 356 to 359, the docs mention a “minimum
reason length” but do not define it; update the text to state the exact minimum
(e.g., "Minimum reason length: 10 characters") and the validation behavior: if
the reason is missing or shorter than the minimum the API responds with HTTP 400
and an error body containing code "REASON_REQUIRED" (include a brief example
response schema or one-line note showing HTTP 400 + code REASON_REQUIRED).
Ensure the wording is precise and replaces the vague phrase with the concrete
minimum and the stated HTTP error/code behavior.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Documented the exact minimum reason length and failure envelope in the purge docs. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Documented the exact minimum reason length and failure envelope in the purge docs. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:177
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/canary-deployments.md around lines 7 to 11 (and also apply the same

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679606

- [review_comment] 2025-09-16T14:19:28Z by coderabbitai[bot] (docs/api/canary-deployments.md:11)

```text
In docs/api/canary-deployments.md around lines 7 to 11 (and also apply the same
change to lines 753-760), the README currently shows insecure HTTP URLs by
default; change the examples to use HTTPS (and WSS for WebSocket examples) as
the default secure scheme, and move any http/ws examples into a clearly labeled
"Local development" note that indicates using http://localhost only for local
testing; update the Base URL block to use https://... and add a short local-dev
snippet showing the http://localhost URL with an explicit note.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Switched all public examples to HTTPS/WSS and tucked loopback guidance into a local-dev appendix. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Switched all public examples to HTTPS/WSS and tucked loopback guidance into a local-dev appendix. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:254
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/exactly-once-admin.md around lines 25-33 (and also apply the same

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679626

- [review_comment] 2025-09-16T14:19:29Z by coderabbitai[bot] (docs/api/exactly-once-admin.md:33)

```text
In docs/api/exactly-once-admin.md around lines 25-33 (and also apply the same
change to lines 56-61), the deduplication field "hit_rate" is ambiguous; rename
the field to "hit_percent" and update its value semantics to be a percentage
(e.g., 2.28 means 2.28%), then update the dedup stats JSON example accordingly
and edit the "Fields" documentation block to reflect the new name and explicitly
state that hit_percent is a percentage value (not a fraction) with its units.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Renamed `hit_rate` to `hit_percent` and clarified units everywhere it appears. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Renamed `hit_rate` to `hit_percent` and clarified units everywhere it appears. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:333
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/SLAPS/coordinator-observations.md around lines 114-116 (and also apply

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679641

- [review_comment] 2025-09-16T14:19:29Z by coderabbitai[bot] (docs/SLAPS/coordinator-observations.md:116)

```text
In docs/SLAPS/coordinator-observations.md around lines 114-116 (and also apply
the same fix at 235-242), the text shows “19 tasks completed successfully” while
elsewhere it shows “74 completed,” causing confusion; update the copy to
explicitly annotate that “19” refers to an early snapshot or intermediate
checkpoint and “74” is the final total (or reconcile to a single consistent
number), e.g., add a parenthetical or an extra sentence clarifying the
timeline/source of each number so readers understand they are different
snapshots rather than inconsistent data.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Added context tying the 19-task snapshot to the later 74-task total. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added context tying the 19-task snapshot to the later 74-task total. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:370
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/SLAPS/coordinator-observations.md around lines 121 to 130, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679654

- [review_comment] 2025-09-16T14:19:29Z by coderabbitai[bot] (docs/SLAPS/coordinator-observations.md:130)

```text
In docs/SLAPS/coordinator-observations.md around lines 121 to 130, the
timestamps are missing timezone/offset information; update every timestamp to a
full ISO‑8601 format including date and timezone offset (e.g.
2025-09-16T12:10:00-07:00 or use Z for UTC) so Start Time, End Time and any
other time entries explicitly include timezone/offset; keep human-friendly
labels but ensure machine-parseable ISO strings are used consistently throughout
the file.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Converted the timetable snippet to ISO-8601 with timezone offsets. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Converted the timetable snippet to ISO-8601 with timezone offsets. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:409
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## docs/SLAPS/coordinator-observations.md around lines 249-251: the document

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679656

- [review_comment] 2025-09-16T14:19:29Z by coderabbitai[bot] (docs/SLAPS/coordinator-observations.md:251)

```text
docs/SLAPS/coordinator-observations.md around lines 249-251: the document
contains a profane user quote ("fuck lol thats amazing") which is inappropriate
for a public repo; remove or sanitize the phrase by either replacing it with a
neutral placeholder quote (e.g., "User: 'That's amazing'") or redact the
offending word (e.g., "User: '[redacted] lol that's amazing'"), and ensure the
surrounding sentence still reads naturally and retains the original intent.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Replaced the profane quote with a redacted version to keep the repo family-friendly. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Replaced the profane quote with a redacted version to keep the repo family-friendly. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:447
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/YOU ARE WORKER 6.rb around lines 1 to 5, the file contains non-Ruby

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679659

- [review_comment] 2025-09-16T14:19:29Z by coderabbitai[bot] (docs/YOU ARE WORKER 6.rb:5)

```text
In docs/YOU ARE WORKER 6.rb around lines 1 to 5, the file contains non-Ruby
documentation which breaks RuboCop; either rename the file to a documentation
extension (e.g., docs/YOU ARE WORKER 6.md) or keep the .rb extension but wrap
the entire content in Ruby-safe comment or code fences (e.g., block comments or
a heredoc) so RuboCop ignores it; pick one approach, update the filename or wrap
the content accordingly, and adjust any references in the repo to the new path
if renamed.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Renamed the worker guide to `.md` so lint and tooling stop treating it as Ruby. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Renamed the worker guide to `.md` so lint and tooling stop treating it as Ruby. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:484
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 314 to 320, the examples use

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679672

- [review_comment] 2025-09-16T14:19:29Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:320)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 314 to 320, the examples use
file globs and improper flag placement; replace them with package-wide
invocations and a proper -run regex: use go test -v ./... -args -debug to enable
verbose logging with the debug arg, and use go test -v ./... -run
'^TestSpecificTest$' to run a single test (anchored regex) so tests target
packages correctly and flags are applied as intended.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Updated the go test snippets to use package-aware invocations with anchored -run. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Updated the go test snippets to use package-aware invocations with anchored -run. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:522
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In AGENTS.md around lines 10 to 41, the table of contents uses Obsidian-style

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683010

- [review_comment] 2025-09-16T14:20:36Z by coderabbitai[bot] (AGENTS.md:41)

```text
In AGENTS.md around lines 10 to 41, the table of contents uses Obsidian-style
wiki links ([[...]]), which won’t render correctly in standard Markdown; convert
each wiki link to a standard Markdown anchor link using the format
[Text](#anchor) where the anchor is the lowercased, hyphenated heading text
(remove extra hashes and punctuation), and update nested entries accordingly so
all TOC links point to the correct section IDs.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Converted the TOC to GitHub-flavoured Markdown anchors. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Converted the TOC to GitHub-flavoured Markdown anchors. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:559
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/admin-api/k8s-deployment.yaml around lines 62 to 116, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683020

- [review_comment] 2025-09-16T14:20:37Z by coderabbitai[bot] (deployments/admin-api/k8s-deployment.yaml:116)

```text
In deployments/admin-api/k8s-deployment.yaml around lines 62 to 116, the
container/pod lacks basic security hardening; add a pod-level and
container-level securityContext to prevent privilege escalation and enforce
least privilege: set spec.template.spec.automountServiceAccountToken: false,
podSecurityContext.runAsNonRoot: true and runAsUser (e.g., 1000) and fsGroup if
needed; in the container securityContext set allowPrivilegeEscalation: false,
readOnlyRootFilesystem: true, runAsNonRoot: true (or match pod runAsUser), and
drop Linux capabilities (capabilities.drop: ["ALL"]); also enable seccomp by
setting securityContext.seccompProfile.type: RuntimeDefault (or annotation for
older k8s) to enforce a default seccomp profile. Ensure values are consistent
between pod and container contexts and do not conflict with existing volume
mounts.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Hardened the admin API deployment with pod/container security contexts and RuntimeDefault seccomp. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Hardened the admin API deployment with pod/container security contexts and RuntimeDefault seccomp. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:596
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/admin-api/k8s-deployment.yaml around lines 117 to 159, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683027

- [review_comment] 2025-09-16T14:20:37Z by coderabbitai[bot] (deployments/admin-api/k8s-deployment.yaml:159)

```text
In deployments/admin-api/k8s-deployment.yaml around lines 117 to 159, the
ingress is fine but you must ensure the Deployment's container probe paths
differ: update the Deployment spec (the container's probes) so
readinessProbe.path is /ready or /readyz and the livenessProbe stays on a
different path (e.g., /health or /live); apply this change in the Deployment
manifest (set readinessProbe.path to /ready or /readyz, keep livenessProbe on
the existing health endpoint) so the readiness check does not use the same
endpoint as liveness.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Ensured readiness probes target `/readyz` while liveness stays on `/healthz`. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Ensured readiness probes target `/readyz` while liveness stays on `/healthz`. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:639
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/docker/docker-compose.yaml around lines 105 to 109 the compose

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683036

- [review_comment] 2025-09-16T14:20:37Z by coderabbitai[bot] (deployments/docker/docker-compose.yaml:109)

```text
In deployments/docker/docker-compose.yaml around lines 105 to 109 the compose
mounts refer to ./grafana/dashboards and ./grafana/datasources but the
repository stores these under deploy/grafana; update the volume paths to point
to ../../deploy/grafana/dashboards and ../../deploy/grafana/datasources (or
alternatively create the missing local ./grafana/datasources directory and add
the required datasource files), ensuring the compose file uses the correct
relative paths to the existing deploy/grafana directories.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Corrected Grafana volume mounts to reference the actual deploy directory. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Corrected Grafana volume mounts to reference the actual deploy directory. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:678
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/docker/rbac-configs/resources.yaml around lines 91-104 (and also

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683042

- [review_comment] 2025-09-16T14:20:37Z by coderabbitai[bot] (deployments/docker/rbac-configs/resources.yaml:104)

```text
In deployments/docker/rbac-configs/resources.yaml around lines 91-104 (and also
check 146-167 and 167-204), the actions catalog is missing the referenced
monitoring/health actions and contains an undefined queues:list alias; add
explicit action entries for "metrics:read" and "health:read" (with description,
risk_level and audit_required) and either add a clear explicit "queues:list"
action entry if you want it aliased to an existing queue read/list permission or
remove any references to queues:list from roles; ensure all roles reference only
actions that are declared in this file.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 6 | Confirmed the monitoring and health actions exist and clarified their descriptions. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Confirmed the monitoring and health actions exist and clarified their descriptions. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:716
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/docker/rbac-configs/token-service.yaml around lines 21 to 24, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683049

- [review_comment] 2025-09-16T14:20:37Z by coderabbitai[bot] (deployments/docker/rbac-configs/token-service.yaml:24)

```text
In deployments/docker/rbac-configs/token-service.yaml around lines 21 to 24, the
JWT config uses a symmetric HS256 key which is unsafe across multiple services;
replace it with an asymmetric algorithm (e.g., RS256 or EdDSA) or PASETO
v4.public, generate a signing private key and a separate public verification key
(use RSA 3072/4096 or Ed25519 as chosen), store the private key only in the
token-service secret and distribute the public key to all verifier services (or
publish via a JWKS endpoint), update the config to reflect algorithm and
key_size appropriate for the chosen algorithm, and add/update deployment docs to
describe the key rotation process (generate new keypair, deploy new public key
to verifiers, switch signer to new private key, and revoke old keys) so
sign/verify roles remain separated.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Switched the token service config to RS256, added key rotation notes, and introduced *_FILE env usage. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Switched the token service config to RS256, added key rotation notes, and introduced *_FILE env usage. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:755
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/kubernetes/admin-api-deployment.yaml around lines 90 to 100, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683057

- [review_comment] 2025-09-16T14:20:37Z by coderabbitai[bot] (deployments/kubernetes/admin-api-deployment.yaml:100)

```text
In deployments/kubernetes/admin-api-deployment.yaml around lines 90 to 100, the
pod-level securityContext exists but the container is missing baseline
hardening; add a container-level securityContext that sets
allowPrivilegeEscalation: false, privileged: false, drops all capabilities
(capabilities.drop: ["ALL"]), sets readOnlyRootFilesystem: true, and configures
a seccompProfile (type: RuntimeDefault) so the container cannot escalate
privileges, has no extra capabilities, uses a read-only root filesystem and
enforces seccomp; keep existing pod runAsNonRoot/runAsUser/fsGroup settings
intact.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Applied the requested security hardening and probe split to the Kubernetes Admin API deployment. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Applied the requested security hardening and probe split to the Kubernetes Admin API deployment. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:801
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/kubernetes/monitoring.yaml around lines 35 to 38, the alert

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683060

- [review_comment] 2025-09-16T14:20:38Z by coderabbitai[bot] (deployments/kubernetes/monitoring.yaml:38)

```text
In deployments/kubernetes/monitoring.yaml around lines 35 to 38, the alert
divides by sum(rate(http_requests_total{app="admin-api"}[5m])) which can be
zero; change the PromQL to guard the denominator (for example wrap the
denominator with clamp_min(..., 1) or otherwise ensure it’s >0 before dividing)
so the expression never performs a division by zero and the rule won’t flap when
traffic is 0.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Wrapped the error-rate denominator with clamp_min to avoid divide-by-zero flapping. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Wrapped the error-rate denominator with clamp_min to avoid divide-by-zero flapping. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:845
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/kubernetes/rbac-monitoring.yaml around lines 45–54, the expr

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683069

- [review_comment] 2025-09-16T14:20:38Z by coderabbitai[bot] (deployments/kubernetes/rbac-monitoring.yaml:54)

```text
In deployments/kubernetes/rbac-monitoring.yaml around lines 45–54, the expr
currently uses raw 5xx RPS instead of a ratio; replace it with a ratio of 5xx
requests to total requests over the same window (and aggregate across labels) —
for example use
sum(rate(http_requests_total{job="rbac-token-service",status=~"5.."}[5m])) /
sum(rate(http_requests_total{job="rbac-token-service"}[5m])) > 0.1 — keeping the
same for/labels/annotations.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Converted the RBAC token service alerts to ratios and proper histogram aggregation. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Converted the RBAC token service alerts to ratios and proper histogram aggregation. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:886
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/kubernetes/rbac-monitoring.yaml around lines 56 to 75, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683075

- [review_comment] 2025-09-16T14:20:38Z by coderabbitai[bot] (deployments/kubernetes/rbac-monitoring.yaml:75)

```text
In deployments/kubernetes/rbac-monitoring.yaml around lines 56 to 75, the
histogram_quantile calls are using raw bucket series instead of aggregated
buckets; replace the inner range vector with sum by (le) over the bucket
streams, e.g. histogram_quantile(0.95, sum by (le)
(rate(http_request_duration_seconds_bucket{job="rbac-token-service"}[5m]))), and
apply the same change for the second alert (the >1.0 and >5.0 thresholds remain
unchanged); ensure both alerts use the sum by (le) aggregation for correct
quantile calculation.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Hardened the RBAC token service deployment with file-based secrets, security contexts, and probe split. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Hardened the RBAC token service deployment with file-based secrets, security contexts, and probe split. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:928
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/kubernetes/rbac-token-service-deployment.yaml around lines

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683084

- [review_comment] 2025-09-16T14:20:38Z by coderabbitai[bot] (deployments/kubernetes/rbac-token-service-deployment.yaml:205)

```text
In deployments/kubernetes/rbac-token-service-deployment.yaml around lines
197–205 (and container block around lines ~231–238) the pod and container
security settings are insufficient: add under pod spec securityContext the
seccompProfile with type: RuntimeDefault and set automountServiceAccountToken:
false; in the rbac-token-service container add a container-level securityContext
with allowPrivilegeEscalation: false, readOnlyRootFilesystem: true, and
capabilities.drop: ["ALL"]; replace the image:
work-queue/rbac-token-service:latest with a pinned immutable image tag (e.g.,
:vX.Y.Z) and change imagePullPolicy to IfNotPresent for pinned images; after
updating, run Checkov/OPA Gatekeeper to verify CKV_K8S_* findings are resolved.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Added persistent forward tracking and cleanup traps to the staging deploy script. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added persistent forward tracking and cleanup traps to the staging deploy script. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:971
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/kubernetes/rbac-token-service-deployment.yaml around lines

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683090

- [review_comment] 2025-09-16T14:20:38Z by coderabbitai[bot] (deployments/kubernetes/rbac-token-service-deployment.yaml:229)

```text
In deployments/kubernetes/rbac-token-service-deployment.yaml around lines
209-229 (and also apply the same change to blocks at 255-262 and 263-271), stop
injecting sensitive secret values (redis-password, rbac-signing-key,
rbac-encryption-key) directly as env values; instead mount the existing Secret
as a volume and update the container spec to volumeMount those secret files,
then change the app to read the secret files from the mounted paths (or if the
app requires env paths, set non-sensitive env vars to the file paths only).
Remove the valueFrom secretKeyRef entries for the secret keys, add a volumes:
entry referencing the Secret name rbac-secrets, add corresponding volumeMounts
with a secure mountPath, and ensure RBAC_SIGNING_KEY, RBAC_ENCRYPTION_KEY and
REDIS_PASSWORD are no longer exposed directly in logs/environment but accessed
from the mounted files; repeat the same adjustments for the other noted blocks.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Replaced the blind sleep with an active poll for the RBAC health check port-forward. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Replaced the blind sleep with an active poll for the RBAC health check port-forward. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:1016
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/scripts/deploy-staging.sh around lines 161-166 (and also apply

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683092

- [review_comment] 2025-09-16T14:20:38Z by coderabbitai[bot] (deployments/scripts/deploy-staging.sh:166)

```text
In deployments/scripts/deploy-staging.sh around lines 161-166 (and also apply
same change to 203-204), the script uses unquoted PID variables and lacks a
cleanup trap; update to quote the PID variables and guard against empty values
(e.g. check [ -n "$PF_PID" ] before calling kill) to avoid globbing/empty-var
issues, and add a trap (e.g. trap cleanup EXIT) plus a small cleanup function
that safely kills quoted PIDs (using kill "$PF_PID" 2>/dev/null || true) to
ensure processes are terminated on exit.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Quoted the monitoring script’s kill command to handle empty or unusual PIDs. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Quoted the monitoring script’s kill command to handle empty or unusual PIDs. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:1063
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/scripts/health-check-rbac.sh around lines 173 to 191, replace the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683098

- [review_comment] 2025-09-16T14:20:39Z by coderabbitai[bot] (deployments/scripts/health-check-rbac.sh:191)

```text
In deployments/scripts/health-check-rbac.sh around lines 173 to 191, replace the
blind sleep used to wait for the port-forward with an active poll that verifies
the local port is accepting TCP connections: implement a loop that checks
localhost:8081 (using a portable method such as bash /dev/tcp/localhost/8081 or
nc if available) with the same $TIMEOUT, exiting success as soon as a connection
can be opened and logging a timeout error if the port never becomes ready; keep
the subsequent HTTP health/metrics checks unchanged and ensure the overall wait
honours the $TIMEOUT variable.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Added idempotency headers, integer duration fields, and dry-run semantics to the batch docs. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added idempotency headers, integer duration fields, and dry-run semantics to the batch docs. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:1105
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/dlq-remediation-pipeline.md around lines 676 to 687, the documented

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683129

- [review_comment] 2025-09-16T14:20:39Z by coderabbitai[bot] (docs/api/dlq-remediation-pipeline.md:687)

```text
In docs/api/dlq-remediation-pipeline.md around lines 676 to 687, the documented
error response only includes human text and status/timestamp; update the error
model to add a stable machine-readable "code" (string enum) and a "request_id"
(correlation id) field. Modify the example and API schema to include "code" with
a defined enum of values (e.g., rule_not_found, validation_error,
internal_error) and "request_id" as a UUID/string; document that every endpoint
must return a request_id header/value and include it in logs for tracing, and
update any related examples and schema definitions to reflect these new required
fields.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Standardized the DLQ pipeline error envelope with codes plus request IDs. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Standardized the DLQ pipeline error envelope with codes plus request IDs. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_011.md:31
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/dlq-remediation-pipeline.md around lines 701 to 712, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683132

- [review_comment] 2025-09-16T14:20:39Z by coderabbitai[bot] (docs/api/dlq-remediation-pipeline.md:712)

```text
In docs/api/dlq-remediation-pipeline.md around lines 701 to 712, the
rate-limiting section is missing Retry-After semantics, header units, and
scope/precedence of limits; add that 429 responses include a Retry-After header
(with seconds or an HTTP-date) and describe that X-RateLimit-Reset is a Unix
timestamp in seconds (epoch) indicating when the window resets; explicitly state
whether limits are enforced per-principal (API key/token) and/or per-IP and
which takes precedence when both apply (e.g., per-principal overrides per-IP or
both are enforced and the most restrictive applies), and clarify deterministic
behavior for clients (how remaining counters and reset are calculated and which
header to trust for backoff). Ensure the new text is concise and includes header
names, units, 429 use, and scope/precedence rules.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Clarified rate-limit headers, Retry-After semantics, and limit precedence. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Clarified rate-limit headers, Retry-After semantics, and limit precedence. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_011.md:75
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/dlq-remediation-pipeline.md around lines 713 to 731, the docs

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683135

- [review_comment] 2025-09-16T14:20:39Z by coderabbitai[bot] (docs/api/dlq-remediation-pipeline.md:731)

```text
In docs/api/dlq-remediation-pipeline.md around lines 713 to 731, the docs
currently only describe offset pagination which can misbehave under large lists
and concurrent writes; add a cursor-based pagination variant and document stable
sort keys: describe new query params (cursor — opaque token, limit with default
and max), explain that the backend must return a next_cursor (and prev_cursor
optionally) instead of has_next/has_prev, provide example request/response JSON
for the cursor flow, and call out that results must be ordered by one or more
stable sort keys (e.g., created_at ASC + id ASC as tie-breaker) so cursors
remain consistent under concurrent writes; also mention migration guidance
(support both offset and cursor endpoints or parameters) and recommend using
opaque base64 cursors that encode the last item sort-key values.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Documented cursor pagination alongside the legacy offset flow. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Documented cursor pagination alongside the legacy offset flow. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_011.md:121
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/api/dlq-remediation-ui.md around lines 432 to 436, the purge-all example

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683142

- [review_comment] 2025-09-16T14:20:39Z by coderabbitai[bot] (docs/api/dlq-remediation-ui.md:436)

```text
In docs/api/dlq-remediation-ui.md around lines 432 to 436, the purge-all example
uses an unsafe query-string API; replace it with the documented safe POST body
API: change the curl example to a bash code block that sends a POST to
/api/dlq/entries/purge-all with headers Content-Type: application/json,
Authorization: Bearer $TOKEN, and Idempotency-Key: $(uuidgen), and a JSON body
like {"confirm": true, "filter": {"queue":"test-queue"}} (use proper backtick
fencing for the bash block).
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Updated the purge-all example to the safe JSON POST form with idempotency. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Updated the purge-all example to the safe JSON POST form with idempotency. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_011.md:167
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## docs/SLAPS/worker-reflections/claude-008-reflection.md lines 1-16: add a YAML

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683152

- [review_comment] 2025-09-16T14:20:40Z by coderabbitai[bot] (docs/SLAPS/worker-reflections/claude-008-reflection.md:16)

```text
docs/SLAPS/worker-reflections/claude-008-reflection.md lines 1-16: add a YAML
front-matter block at the top containing date and worker_id (e.g., date:
YYYY-MM-DD and worker_id: claude-008) and update the headings to match the other
reflections (use consistent header levels and phrasing such as "SLAPS Reflection
— Worker 8 (claude-008)" and "Summary" / "Tasks" instead of the current informal
titles); ensure the front-matter is the very first content, remove or tighten
any verbose header wording to be parallel with other files, and keep the rest of
the content intact.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Brought the claude-008 reflection in line with the rest of the SLAPS logbook. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Brought the claude-008 reflection in line with the rest of the SLAPS logbook. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_011.md:209
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## .github/workflows/changelog.yml lines 20-29: workflow currently references

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353724315

- [review_comment] 2025-09-16T21:42:53Z by coderabbitai[bot] (.github/workflows/changelog.yml:29)

```text
.github/workflows/changelog.yml lines 20-29: workflow currently references
mutable action tags (actions/checkout@v4 and actions/setup-go@v5); replace those
with the corresponding pinned commit SHAs for each action (e.g.
actions/checkout@<full-commit-sha> and actions/setup-go@<full-commit-sha>) so
the workflow uses immutable references; locate the commit SHAs from the official
GitHub repos for each action and update the uses lines, and do the same for any
other third-party actions in this file to fully eliminate mutable tags.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Confirmed the workflow already wraps the substitution in quotes and preserved fallback behaviour. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Confirmed the workflow already wraps the substitution in quotes and preserved fallback behaviour. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:168
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## .github/workflows/markdownlint.yml lines 4-6: the workflow is currently

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353724344

- [review_comment] 2025-09-16T21:42:54Z by coderabbitai[bot] (.github/workflows/markdownlint.yml:6)

```text
.github/workflows/markdownlint.yml lines 4-6: the workflow is currently
triggered on every push and PR; add path filters so the job only runs for
Markdown, workflow, or config changes by adding a paths (or paths-ignore as
preferred) entry under both pull_request and push that includes patterns like
**/*.md, .github/**, and .github/workflows/** (and any other repo config globs
you want to include) so runners are only used when Markdown or workflow/config
files change.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Branch array rewritten as `[main]`, satisfying markdownlint and YAML format guidance. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Branch array rewritten as `[main]`, satisfying markdownlint and YAML format guidance. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:351
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## .github/workflows/markdownlint.yml lines 20-21: the checkout action uses

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353724350

- [review_comment] 2025-09-16T21:42:54Z by coderabbitai[bot] (.github/workflows/markdownlint.yml:21)

```text
.github/workflows/markdownlint.yml lines 20-21: the checkout action uses
fetch-depth: 0 which fetches full git history unnecessarily for a markdown
linter; change the checkout step to use fetch-depth: 1 to speed up the job
(replace 0 with 1 on that line).
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Workflow now ships with least-privilege permissions and concurrency guard as requested. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Workflow now ships with least-privilege permissions and concurrency guard as requested. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:385
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/YOU ARE WORKER 6.rb around lines 3 to 4, the sentence "You are a worker

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814728

- [review_comment] 2025-09-16T22:46:57Z by coderabbitai[bot] (docs/YOU ARE WORKER 6.rb:4)

```text
In docs/YOU ARE WORKER 6.rb around lines 3 to 4, the sentence "You are a worker
in the SLAPS task execution system. Your job is to claim and execute tasks for
the go-redis-work-queue project." was split across two lines with an extra
indent; fix it by joining into a single line without the stray indentation so
the sentence is one continuous line.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Role description now uses vendor-neutral wording and adds safety guidance. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Role description now uses vendor-neutral wording and adds safety guidance. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:251
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/admin-api/monitoring.yaml around lines 1-66, this file duplicates

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856215

- [review_comment] 2025-09-16T23:20:22Z by coderabbitai[bot] (deployments/admin-api/monitoring.yaml:66)

```text
In deployments/admin-api/monitoring.yaml around lines 1-66, this file duplicates
canonical admin-api alerts and is missing required metadata.labels/ownership and
uses the wrong namespace; either remove this duplicate file or align it to the
canonical manifest by changing namespace to work-queue, adding metadata.labels
for the PrometheusRule (prometheus: kube-prometheus, role: alert-rules, app:
admin-api, component: admin-api) and for any ConfigMap (app: admin-api,
grafana_dashboard: "1"), and ensure every alert rule includes labels component:
admin-api so the Operator/Grafana can discover them.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Converted alert ConfigMap into a PrometheusRule resource. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Converted alert ConfigMap into a PrometheusRule resource. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:834
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In docs/09_requirements.md around lines 47 to 49, the metrics list needs

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856235

- [review_comment] 2025-09-16T23:20:23Z by coderabbitai[bot] (docs/09_requirements.md:49)

```text
In docs/09_requirements.md around lines 47 to 49, the metrics list needs
explicit queue labeling and consistent Prometheus naming conventions: change the
queue_length entry to explicitly show the label (e.g.
queue_length{queue="<name>"}) to prevent accidental high cardinality, and ensure
all metric names use the correct suffixes and types (use _total for counters
like worker_registered_total and rate_limit_hits_total, _seconds for histograms
like job_processing_duration_seconds, and gauge for current values). Update the
text to explicitly state the required label and naming rules so automated tests
and implementers enforce queue label usage and consistent _total/_seconds
suffixes.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Acceptance criteria now spell out metric names, readiness semantics, and admin safeguards. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Acceptance criteria now spell out metric names, readiness semantics, and admin safeguards. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:931
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In cmd/job-queue-system/main.go around lines 100 to 114, the signal handling

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912487

- [review_comment] 2025-09-18T12:12:36Z by coderabbitai[bot] (cmd/job-queue-system/main.go:114)

```text
In cmd/job-queue-system/main.go around lines 100 to 114, the signal handling
goroutine never calls signal.Stop, leaking signal notifications; after receiving
the first signal and calling cancel(), call signal.Stop(sigCh) to unregister the
channel from the signal package, then proceed with the existing logic to either
handle a second signal or wait the timeout and return so the goroutine exits and
resources are released.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Admin runs no longer spawn the background queue length updater. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Admin runs no longer spawn the background queue length updater. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:353
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 232 to 235, add an additional

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358913071

- [review_comment] 2025-09-18T12:12:41Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:235)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 232 to 235, add an additional
example command to show how to run only the payload redactor tests; duplicate
the existing security-only example and append a new line with the go test
command using the '^TestPayloadRedactor_' run selector so contributors can
easily execute only redactor tests.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 10 | Integration & security examples now rely on explicit package paths. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Integration & security examples now rely on explicit package paths. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:1030
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In .gitignore around line 16, you added / .gocache/ but missed other common

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360032954

- [review_comment] 2025-09-18T15:55:17Z by coderabbitai[bot] (.gitignore:16)

```text
In .gitignore around line 16, you added / .gocache/ but missed other common
local/build artifacts; update the file to also ignore typical Go and editor/temp
files such as /bin/, /pkg/, /vendor/, *.exe, *.test, coverage.out, *.env and
.env.local, .DS_Store, and editor dirs like .idea/ and .vscode/ (or expand the
existing allowlist pattern), and ensure you do a quick sanity check that go.work
is not required by the repo before adding it to the ignore list.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Added `go.work.sum` to the ignore list alongside `go.work`. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added `go.work.sum` to the ignore list alongside `go.work`. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:422
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/docker/Dockerfile.admin-api around lines 52 to 53, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033189

- [review_comment] 2025-09-18T15:55:19Z by coderabbitai[bot] (deployments/docker/Dockerfile.admin-api:53)

```text
In deployments/docker/Dockerfile.admin-api around lines 52 to 53, the
HEALTHCHECK is targeting /health while other artifacts use /healthz; update the
HEALTHCHECK command to use the canonical /healthz path instead of /health (i.e.,
replace the URL to http://localhost:8080/healthz) so Dockerfile, Compose and K8s
probes are consistent and keep the existing curl flags and timing options
unchanged.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Runtime layer now installs curl and the healthcheck aligns with other images. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Runtime layer now installs curl and the healthcheck aligns with other images. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:144
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/admin-api/k8s-deployment.yaml around lines 71 to 116, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360038848

- [review_comment] 2025-09-18T15:56:31Z by coderabbitai[bot] (deployments/admin-api/k8s-deployment.yaml:116)

```text
In deployments/admin-api/k8s-deployment.yaml around lines 71 to 116, the
JWT_SECRET and REDIS_PASSWORD are exposed as environment variables; change to
mount the admin-api-secrets Secret as a volume and remove those secret ENV
entries, mounting the secret files (e.g. /etc/secrets/jwt-secret and
/etc/secrets/redis-password) into the container with appropriate readOnly
permissions and update volume/volumeMounts to reference the Secret; then update
the application startup logic to first read credentials from the mounted files
if present (fallback to env only if files missing) and ensure file permissions
and container user (runAsUser 1000) can read them.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Hardened the admin API deployment with pod/container security contexts and RuntimeDefault seccomp. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Hardened the admin API deployment with pod/container security contexts and RuntimeDefault seccomp. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:596
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In BUGS.md around lines 51 to 53, update the note about worker registry to

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360060998

- [review_comment] 2025-09-18T16:01:14Z by coderabbitai[bot] (BUGS.md:53)

```text
In BUGS.md around lines 51 to 53, update the note about worker registry to
document registry hygiene: specify that on startup or during periodic
maintenance the system must check each registered worker's heartbeat key and, if
the heartbeat key is missing or expired, remove that worker ID from the registry
(SREM) to prevent unbounded growth; avoid global SCAN/KEYS by iterating only the
registry set membership and checking each worker's heartbeat key, and clearly
state this pruning step and its frequency so implementers know to remove dead
workers when heartbeats are absent.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Replaced the non-atomic pipeline advice with a ZPOPMIN/Lua-based atomic mover. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Replaced the non-atomic pipeline advice with a ZPOPMIN/Lua-based atomic mover. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:653
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In demos/responsive-tui.tape around lines 19-27, the script sets COLUMNS

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061123

- [review_comment] 2025-09-18T16:01:15Z by coderabbitai[bot] (demos/responsive-tui.tape:27)

```text
In demos/responsive-tui.tape around lines 19-27, the script sets COLUMNS
cosmetically (export COLUMNS=35) which doesn't actually resize the terminal;
change that to drive a real terminal width before the section by using a
terminal resize command (for example: stty cols 35 or resize -s <rows> 35) so
the TUI hits its breakpoint; make the same replacement in the other demo blocks
noted (lines 75-83, 133-141, 219-227) so each section begins with an actual
terminal width change rather than only exporting COLUMNS.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Demo now calls out COLUMNS overrides as cosmetic and trims long sleeps. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Demo now calls out COLUMNS overrides as cosmetic and trims long sleeps. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:750
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/kubernetes/rbac-monitoring.yaml around lines 33 to 43, the alert

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066741

- [review_comment] 2025-09-18T16:02:30Z by coderabbitai[bot] (deployments/kubernetes/rbac-monitoring.yaml:43)

```text
In deployments/kubernetes/rbac-monitoring.yaml around lines 33 to 43, the alert
expression using up{app="rbac-token-service"} == 0 is brittle and can miss cases
where all targets are absent; replace the expr with an aggregate/absent-aware
expression such as using sum() and absent(), e.g. trigger when
absent(up{app="rbac-token-service"}) or when sum(up{app="rbac-token-service"})
== 0 so that total outages and aggregated down states are detected; keep the
same for/labels/annotations but update the expr to the combined absent OR sum
condition.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Runbook link now points to the documented ops guide. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Runbook link now points to the documented ops guide. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:380
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.


## In deployments/scripts/setup-monitoring.sh around lines 30-31 (and the similar

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066956

- [review_comment] 2025-09-18T16:02:32Z by coderabbitai[bot] (deployments/scripts/setup-monitoring.sh:31)

```text
In deployments/scripts/setup-monitoring.sh around lines 30-31 (and the similar
block at 194-209), the ALERTMANAGER_WEBHOOK_URL is defaulting to
http://localhost:9093/webhook which is unsafe/useless; change the logic to fail
fast when ALERTMANAGER_WEBHOOK_URL is unset or validate it and reject
localhost/loopback addresses. Specifically, remove the localhost default, check
if ALERTMANAGER_WEBHOOK_URL is non-empty, verify it looks like a sane URL
(http(s) scheme and host not localhost/127.0.0.1/::1), and exit with a clear
error if the check fails so the script requires a real webhook URL instead of
silently using localhost.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Monitoring script sources the shared logging helpers instead of duplicating them. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Monitoring script sources the shared logging helpers instead of duplicating them. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:896
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.
