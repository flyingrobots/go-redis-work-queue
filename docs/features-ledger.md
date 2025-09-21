---
noteID: 5713fc44-8949-4e4a-a051-8fa5790fa2ee
---
# Features Ledger

This is the canonical, grouped snapshot of features — shipped, in‑progress, and planned — including progress, tasks, tests, and remarks. TUI and other feature tasks live here (not in AGENTS.md).

## Progress

<!-- progress:begin -->
```text
█████████████████████▓░░░░░░░░░░░░░░░░░░ 55%
---------|---------|---------|---------|
        MVP      Alpha     Beta  v1.0.0 
```
<!-- progress:end -->

Weighted by feature size. Updated by `python3 scripts/update_progress.py`.

## Status Model

- Planned → In Progress → MVP → Alpha → Beta → V1
- We use stage names directly; “Shipped” is implied by MVP/Alpha/V1.

Definitions
- MVP: minimal viable; usable for demos/tests; rough edges allowed
- Alpha: feature complete; internal‑ready; known limitations; needs hardening
- Beta: feature complete; externally usable; not yet battle‑tested (needs soak/perf/chaos/coverage)
- V1: production‑ready; strong tests/docs; battle‑tested

Weighting method: For feature‑specific modules, w = 1 + log10(Go LOC + 10) / 3; minimum w = 0.5 if no resolvable code path. Overall = Σ(p_i·w_i)/Σ(w_i).

Emoji status mapping
- 📋 Planned
- ⏳ In Progress
- 🚼 MVP
- 🅰️ Alpha
- 🅱️ Beta
- ✅ V1 (Shipped)

Update via script
- Run `python3 scripts/update_progress.py` after editing table rows or Code links. The script updates the bars here and in README.md.

---

### Core & Platform
<!-- group-progress:core-platform:begin -->
```text
█████████████████████████████████▓░░░░░░ 84%
---------|---------|---------|---------|
        MVP      Alpha     Beta  v1.0.0 
weight=8.58 features=9 kloc=17.4
```
<!-- group-progress:core-platform:end -->

| Emoji | Feature                                               | Area          | Spec                                      | Code                                                                                                | KLoC (approx) | Status | Progress %       | Bar          | Current State                                                                        | Todo (Tasks)                                                     | Tests                          | Remarks                                        |
| ----- | ----------------------------------------------------- | ------------- | ----------------------------------------- | --------------------------------------------------------------------------------------------------- | ------------- | ------ | ---------------- | ------------ | ------------------------------------------------------------------------------------ | ---------------------------------------------------------------- | ------------------------------ | ---------------------------------------------- |
|🅰️ | [Core Job Queue](../README.md) | Core/Runtime | — | [internal/queue](../internal/queue), [worker](../internal/worker), [producer](../internal/producer) | 0.8 | Alpha | 80% (conf: high) | `████████░░` | Stable enqueue/consume; retries + metrics present. Conf high from breadth and usage. | Retry/backoff polish; graceful shutdown semantics; perf passes. | Unit + some integration; good. | Foundation is solid. |
|🅱️ | Admin API v1 (HTTP) | Platform/API | — | [internal/admin-api](../internal/admin-api) | 5.4 | Beta | 90% (conf: high) | `█████████░` | Endpoints + middleware + OpenAPI shipped. | TUI switchover for Stats; expand e2e; gRPC decision; soak/chaos; share port-forward helper across deployment scripts; add policy-as-code checks for manifest security/secret rules. | Unit + integration; good. | Productionize defaults; audit destructive ops. |
|🅰️ | Storage Backends | Core/Storage | — | [internal/storage-backends](../internal/storage-backends) | 5.9 | Alpha | 75% (conf: med) | ███████░░░ | Adapters + tests; conformance pending. | Complete adapter matrix; conformance; migration docs. | Unit + integration; fair. | Track compat matrix. |
|🅱️ | RBAC & Tokens | Security | — | [internal/rbac-and-tokens](../internal/rbac-and-tokens) | 3.1 | Beta | 85% (conf: high) | █████████░ | Manager + middleware; hardened. | Expand scopes; e2e coverage; audit trails; soak/rotation tests. | Unit + middleware; good. | Security foundation. |
|🅱️ | Observability Core | Observability | — | [internal/obs](../internal/obs) | 1.4 | Beta | 85% (conf: high) | █████████░ | Logger/metrics/tracing wiring. | Dashboards; error budgets; SLO dashboards; alert tuning. | Unit present. | Solid base. |
|🅱️ | Reaper | Maintenance | — | [internal/reaper](../internal/reaper) | 0.1 | Beta | 90% (conf: high) | █████████░ | TTL/cleanup working. | Tune policies; monitoring; long-run soak. | Unit present. | Keep safe defaults. |

|Emoji | Feature | Area | Spec | Code | KLoC (approx) | Status | Progress % | Bar | Current State | Todo (Tasks) | Tests | Remarks |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
|🅰️ | Breaker | Core/Runtime | — | [internal/breaker](../internal/breaker) | 0.2 | Alpha | 75% (conf: med) | ███████░░░ | Circuit breaker in place with unit tests. | Integrate metrics; document use; tune defaults. | Unit good. | Keep simple + safe. |
|🅱️ | Config | Core/Runtime | — | [internal/config](../internal/config) | 0.3 | Beta | 85% (conf: high) | █████████░ | Config loader stable. | Extend validation; env overrides docs; backwards compat policy. | Unit present. | Foundation module. |
|🅱️ | Redis Client | Core/Runtime | — | [internal/redisclient](../internal/redisclient) | 0.0 | Beta | 90% (conf: high) | █████████░ | Thin wrapper around go-redis v9. | Connection tests; pool tuning; resilience docs. | None | Unified to v9. |

### TUI & UX
<!-- group-progress:tui-ux:begin -->
```text
█████████████████▓░░░░░░░░░░░░░░░░░░░░░░ 43%
---------|---------|---------|---------|
        MVP      Alpha     Beta  v1.0.0 
weight=14.46 features=13 kloc=39.3
```
<!-- group-progress:tui-ux:end -->

| Emoji | Feature                                                                 | Area          | Spec                                               | Code                                                                        | KLoC (approx) | Status      | Progress %       | Bar        | Current State                                  | Todo (Tasks)                                                                                 | Tests                       | Remarks                   |
| ----- | ----------------------------------------------------------------------- | ------------- | -------------------------------------------------- | --------------------------------------------------------------------------- | ------------- | ----------- | ---------------- | ---------- | ---------------------------------------------- | -------------------------------------------------------------------------------------------- | --------------------------- | ------------------------- |
|⏳ | [TUI Shell (Tabs/Layout)](../docs/TUI/README.md) | UX/TUI | [Spec](../docs/TUI/README.md) | [internal/tui](../internal/tui) | 2.7 | In Progress | 70% (conf: high) | ███████░░ | Tabs, charts expand, tiny‑term fixes, and table polish (striping/glyph/thresholds) done. | Wire Admin API; persist UI state; help overlay polish; adjustable panel split. | Manual + some unit; fair. | Incremental polish. |
|⏳ | DLQ Remediation UI | Ops/TUI | — | [internal/dlq-remediation-ui](../internal/dlq-remediation-ui) | 2.9 | In Progress | 55% (conf: med) | █████░░░░░ | API + TUI model exist; paging/filters pending. | Server‑side paging/filters; TUI list/peek; RBAC/audit hooks. | Unit present; needs e2e. | Prioritize perf. |
|⏳ | [Workers View (TUI)](../docs/TUI/README.md) | UX/TUI | [Spec](../docs/TUI/README.md) | [internal/tui](../internal/tui) | 2.7 | In Progress | 35% (conf: med) | ███░░░░░░░ | Placeholder; no live list yet. | Use Admin workers endpoint; sort/filter; heartbeat display. | None; add UI tests. | Needs workers API wiring. |
|⏳ | [Settings View (TUI)](../docs/TUI/README.md) | UX/TUI | [Spec](../docs/TUI/README.md) | [internal/tui](../internal/tui) | 2.7 | In Progress | 40% (conf: med) | ████░░░░░░ | Minimal snapshot. | Theme toggle; config path; copy/open shortcuts. | None; add snapshot tests. | Quick win. |
|⏳ | Right‑click Context Menus | UX/TUI | — | [internal/right-click-context-menus](../internal/right-click-context-menus) | 2.3 | In Progress | 50% (conf: med) | █████░░░░░ | Menus/zones exist; focus wiring pending. | Connect to table rows; actions; tests; double‑click peek; header sort. | Unit present; needs UI/e2e. | Pair with bubblezone. |
|📋 | Bubblezone Hitboxes | UX/TUI | — | [internal/right-click-context-menus](../internal/right-click-context-menus) | 2.3 | Planned | 10% (conf: med) | █░░░░░░░░░ | Not started; design known. | Integrate bubblezone; zone mapping for tabs/rows/splitter. | None. | Enables precise mouse UX. |
|⏳ | JSON Payload Studio | UX/TUI | — | [internal/json-payload-studio](../internal/json-payload-studio) | 4.4 | In Progress | 40% (conf: med) | ████░░░░░░ | Core handlers; not in TUI. | TUI editor; schemas/templates; enqueue path. | Unit present. | UX heavy. |
|⏳ | Calendar View | UX/TUI | — | [internal/calendar-view](../internal/calendar-view) | 5.0 | In Progress | 45% (conf: med) | ████░░░░░░ | Routes/UI; auth/multi‑queue TODOs. | Add auth context; filters; paging. | Unit + TODOs. | Verify perf. |
|🚼 | Theme Playground | UX/TUI | — | [internal/theme-playground](../internal/theme-playground) | 5.3 | MVP | 70% (conf: high) | ███████░░░ | Persistence + tests shipped. | Centralize styles; Settings toggle; accessible palettes. | Unit + integration; good. | Accessibility focus. |
|📋 | Terminal Voice Commands | UX/CLI | — | — | 0.0 | Planned | 0% (conf: low) | ░░░░░░░░░░ | Archived 2025-09-20; terminal voice module removed from repo. | Re-evaluate approach if feature resurfaces. | — | Optional, flashy. |
|⏳ | Plugin Panel System | Extensibility | — | [internal/plugin-panel-system](../internal/plugin-panel-system) | 3.7 | In Progress | 50% (conf: med) | █████░░░░░ | Lifecycle + permissions. | Sandbox; TUI registry; SDK docs. | Unit present. | Watch plugin trust. |
|⏳ | Visual DAG Builder | UX/Flow | — | [internal/visual-dag-builder](../internal/visual-dag-builder) | 4.0 | In Progress | 40% (conf: med) | ████░░░░░░ | Orchestrator/types; not wired. | Backend validation; DAG execution; TUI builder. | Unit partial. | Longer‑term. |
|⏳ | Collaborative Session | UX/TUI | — | [internal/collaborative-session](../internal/collaborative-session) | 1.4 | In Progress | 25% (conf: low) | ██░░░░░░░░ | Early scaffolding only. | Define protocol/permissions; host/guest; TUI controls. | None | Nice-to-have. |

### Reliability & Ops
<!-- group-progress:reliability-ops:begin -->
```text
████████████████████▓░░░░░░░░░░░░░░░░░░░ 51%
---------|---------|---------|---------|
        MVP      Alpha     Beta  v1.0.0 
weight=16.83 features=14 kloc=60.2
```
<!-- group-progress:reliability-ops:end -->

|Emoji | Feature | Area | Spec | Code | KLoC (approx) | Status | Progress % | Bar | Current State | Todo (Tasks) | Tests | Remarks |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
|⏳ | DLQ Remediation Pipeline | Reliability | — | [internal/dlq-remediation-pipeline](../internal/dlq-remediation-pipeline) | 4.7 | In Progress | 45% (conf: med) | ████░░░░░░ | Pipeline scaffolding; classifiers/rules TBD. | Rules engine; rate‑limited requeue; safety bounds. | Light unit; needs scenario tests. | Integrate with DLQ UI. |
|🚼 | Exactly‑once Patterns | Reliability | — | [internal/exactly_once](../internal/exactly_once), [internal/exactly-once-patterns](../internal/exactly-once-patterns) | 7.9 | MVP | 70% (conf: high) | ███████░░░ | Idempotency/outbox ready; some TODOs. | Finalize metrics; publisher adapters; docs. | Unit + integration; good. | Strong differentiator. |
|⏳ | Advanced Rate Limiting | Throughput | — | [internal/advanced-rate-limiting](../internal/advanced-rate-limiting) | 1.6 | In Progress | 55% (conf: high) | █████░░░░░ | Lua token bucket + fairness done. | Admin API runtime updates; TUI widget; producer/worker hooks. | Unit + integration; good. | High leverage; wire into SDKs. |
|⏳ | Producer Backpressure | SDKs | — | [internal/producer-backpressure](../internal/producer-backpressure) | 3.3 | In Progress | 40% (conf: med) | ████░░░░░░ | Signals present; not linked to RL. | Integrate with rate limiter; client SDK examples. | Unit present. | Needs producer docs. |
|⏳ | Policy Simulator | Ops/Safety | — | [internal/policy-simulator](../internal/policy-simulator) | 4.7 | In Progress | 45% (conf: med) | ████░░░░░░ | Core present; retrieval/rollback TODO. | Preview/apply/rollback endpoints; persist scenarios. | Unit present. | Pair with Admin API. |
|⏳ | Worker Fleet Controls | Ops | — | [internal/worker-fleet-controls](../internal/worker-fleet-controls) | 3.1 | In Progress | 45% (conf: med) | ████░░░░░░ | Control scaffolding; safety checks TBD. | Pause/drain/resume + RBAC; per‑node metrics; TUI controls. | Unit present. | Add safety gates. |
|⏳ | Long‑term Archives | Ops/Data | — | [internal/long-term-archives](../internal/long-term-archives) | 4.2 | In Progress | 45% (conf: med) | ████░░░░░░ | Archival hooks; adapters TBD. | S3/ClickHouse adapters; retention; export path. | Unit partial. | Define retention/SLO. |
|⏳ | Event Hooks | Integrations | — | [internal/event-hooks](../internal/event-hooks) | 3.6 | In Progress | 50% (conf: med) | █████░░░░░ | Plumbing exists; config/signing TODO. | Configurable base URL; HMAC signatures; retries; Admin mgmt. | Unit present. | Security first. |

|Emoji | Feature | Area | Spec | Code | KLoC (approx) | Status | Progress % | Bar | Current State | Todo (Tasks) | Tests | Remarks |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
|⏳ | Job Budgeting | Reliability | — | [internal/job-budgeting](../internal/job-budgeting) | 4.4 | In Progress | 45% (conf: med) | ████░░░░░░ | Budget manager, cost model; limited UI. | Enforcement hooks; Admin API; notifications. | Unit present. | Wire to TUI. |
|⏳ | Smart Payload Dedup | Reliability | — | [internal/smart-payload-deduplication](../internal/smart-payload-deduplication) | 4.3 | In Progress | 50% (conf: med) | █████░░░░░ | Compression/dedup logic; TODOs on dict build. | Dict training; stats; enqueue integration. | Unit present. | Useful cost saver. |
|🅰️ | Smart Retry Strategies | Reliability | — | [internal/smart-retry-strategies](../internal/smart-retry-strategies) | 5.0 | Alpha | 75% (conf: high) | ███████░░░ | Strategies + tests; metrics TODO. | Prometheus metrics; TUI selector. | Unit/integration good. | Solid baseline. |
|⏳ | Automatic Capacity Planning | Planning | — | [internal/automatic-capacity-planning](../internal/automatic-capacity-planning) | 5.1 | In Progress | 55% (conf: med) | █████░░░░░ | Planner + simulator; needs hooks. | Expose Admin API; scheduling; tests. | Unit/integration fair. | Pair with forecasting. |
|⏳ | Chaos Harness | Ops/Safety | — | [internal/chaos-harness](../internal/chaos-harness) | 2.4 | In Progress | 45% (conf: med) | ████░░░░░░ | Fault injection scaffolding. | Profiles; RBAC; kill switch; dashboards. | Light unit. | Guardrails required. |
|⏳ | Canary Deployments | Ops | — | [internal/canary-deployments](../internal/canary-deployments) | 5.9 | In Progress | 50% (conf: med) | █████░░░░░ | Canary logic present; guardrails TBD. | Rollback/abort endpoints; audit logging. | Minimal tests. | Add e2e. |

### Scale & Multi‑Cluster
<!-- group-progress:scale-multi-cluster:begin -->
```text
███████████████████▓░░░░░░░░░░░░░░░░░░░░ 48%
---------|---------|---------|---------|
        MVP      Alpha     Beta  v1.0.0 
weight=4.52 features=4 kloc=11.6
```
<!-- group-progress:scale-multi-cluster:end -->

|Emoji | Feature | Area | Spec | Code | KLoC (approx) | Status | Progress % | Bar | Current State | Todo (Tasks) | Tests | Remarks |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
|⏳ | Multi‑cluster Control | Scale | — | [internal/multi-cluster-control](../internal/multi-cluster-control) | 4.1 | In Progress | 60% (conf: med) | ██████░░░░ | Manager/handlers + tests; UI pending. | TUI tabs; Admin fan‑out actions; compare/replicate ops. | Many tests; good. | Solid engine; wire UX. |
|⏳ | Kubernetes Operator | Platform | — | [internal/kubernetes-operator](../internal/kubernetes-operator) | 3.9 | In Progress | 55% (conf: med) | █████░░░░░ | Controllers/webhooks; examples/tests. | CRDs; reconcile backoff; e2e on kind. | Unit + integration; fair. | Mind CRD validation. |
|⏳ | Multi‑tenant Isolation | Security | — | [internal/multi-tenant-isolation](../internal/multi-tenant-isolation) | 2.8 | In Progress | 40% (conf: med) | ████░░░░░░ | Handlers with RBAC TODOs. | Enforce quotas/keys; authz middleware; tests. | Unit present. | Needs policy decisions. |
|⏳ | Tenant | Security | — | [internal/tenant](../internal/tenant) | 0.8 | In Progress | 35% (conf: low) | ███░░░░░░░ | Early scaffolding. | Define tenant model; integrate with RBAC/multi-tenant. | Unit minimal. | Tie into isolation. |

### Observability & Analytics
<!-- group-progress:observability-analytics:begin -->
```text
██████████████████████▓░░░░░░░░░░░░░░░░░ 57%
---------|---------|---------|---------|
        MVP      Alpha     Beta  v1.0.0 
weight=10.49 features=9 kloc=28.7
```
<!-- group-progress:observability-analytics:end -->

|Emoji | Feature | Area | Spec | Code | KLoC (approx) | Status | Progress % | Bar | Current State | Todo (Tasks) | Tests | Remarks |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
|🅰️ | Distributed Tracing Integration | Observability | — | [internal/distributed-tracing-integration](../internal/distributed-tracing-integration) | 3.0 | Alpha | 85% (conf: high) | █████████░ | OTEL propagation + trace URLs done. | Link from TUI; config docs. | Unit + integration; good. | Low risk polish. |
|⏳ | Trace Drill‑down + Log Tail | Observability | — | [internal/trace-drilldown-log-tail](../internal/trace-drilldown-log-tail) | 3.9 | In Progress | 50% (conf: med) | █████░░░░░ | Trace links ok; log tail TBD. | Tail with filters; privacy; TUI links. | Unit partial. | Watch PII. |
|⏳ | Anomaly Radar + SLO Budget | Observability | — | [internal/anomaly-radar-slo-budget](../internal/anomaly-radar-slo-budget) | 3.2 | In Progress | 60% (conf: med) | ██████░░░░ | Scope-aware handlers, pagination cursors, public pkg wrapper, and OpenAPI/docs alignment landed; SLO maths and widget still pending. | Add SLO budget calc + TUI widget. | Unit + handler tests cover cursors; docs/spec CI validates. | Needs calibration + UI wiring. |
|⏳ | Forecasting | Planning | — | [internal/forecasting](../internal/forecasting) | 2.7 | In Progress | 40% (conf: med) | ████░░░░░░ | Stubs exist. | Baseline models; eval harness; TUI preview. | Unit partial. | Keep simple first. |
|⏳ | Queue Snapshot Testing | QA | — | [internal/queue-snapshot-testing](../internal/queue-snapshot-testing) | 2.4 | In Progress | 50% (conf: med) | █████░░░░░ | Framework + snapshots. | Broaden differ; golden tests; docs. | Unit; fair. | Useful for regressions. |
|⏳ | Patterned Load Generator | Testing | — | [internal/patterned-load-generator](../internal/patterned-load-generator) | 2.1 | In Progress | 45% (conf: med) | ████░░░░░░ | Handlers + generator; guardrails missing. | Add sine/burst/ramp; cancel/stop; profiles; TUI overlay. | Unit present; needs e2e. | Add caps; confirmations. |
|🅰️ | Bench (Basic) | Testing | — | [internal/admin](../internal/admin), [internal/tui](../internal/tui) | 3.4 | Alpha | 60% (conf: med) | ██████░░░░ | Running; progress UI present; baseline delta pending. | Baseline from initial completed list; cancel; ETA/throughput; guardrails. | Manual + some unit. | Guardrails for high rates. |
|⏳ | Job Genealogy Navigator | Analytics | — | [internal/job-genealogy-navigator](../internal/job-genealogy-navigator) | 3.6 | In Progress | 40% (conf: med) | ████░░░░░░ | Types + traversal; integration TBD. | Admin API; TUI drill‑down; pagination. | Unit present. | Pair with tracing. |
|🅰️ | Time‑Travel Debugger | Debugging | — | [internal/time-travel-debugger](../internal/time-travel-debugger) | 4.2 | Alpha | 80% (conf: high) | ████████░░ | Capture/replay + simple TUI implemented. | Selective replay; export/import; docs. | Unit rich. | Powerful debugging. |
