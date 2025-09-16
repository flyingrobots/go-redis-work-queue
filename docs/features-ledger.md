# Features Ledger

This is the canonical, grouped snapshot of features — shipped, in‑progress, and planned — including progress, tasks, tests, and remarks. TUI and other feature tasks live here (not in AGENTS.md).

## Progress

<!-- progress:begin -->
```text
██████████████████████▓░░░░░░░░░░░░░░░░░ 55%
------------|-------------|------------|
           MVP          Alpha    v1.0.0 
```
<!-- progress:end -->

Weighted by feature size. Updated by `python3 scripts/update_progress.py`.

## Status Model

- Planned → In Progress → MVP → Alpha → V1
- We use stage names directly; “Shipped” is implied by MVP/Alpha/V1.

Weighting method: For feature‑specific modules, w = 1 + log10(Go LOC + 10) / 3; minimum w = 0.5 if no resolvable code path. Overall = Σ(p_i·w_i)/Σ(w_i).

Update via script
- Run `python3 scripts/update_progress.py` after editing table rows or Code links. The script updates the bars here and in README.md.

---

### Core & Platform
<!-- group-progress:core-platform:begin -->
```text
█████████████████████████████████▓░░░░░░ 84%
------------|-------------|------------|
           MVP          Alpha    v1.0.0 
weight=6.34 features=6 kloc=15.6
```
<!-- group-progress:core-platform:end -->

|Emoji | Feature | Area | Spec | Code | KLoC (approx) | Status | Progress % | Bar | Current State | Todo (Tasks) | Tests | Remarks |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
|🧱 | [Core Job Queue](../README.md) | Core/Runtime | — | [internal/queue](../internal/queue), [worker](../internal/worker), [producer](../internal/producer) | 0.8 | Alpha | 80% (conf: high) | ████████░░ | Stable enqueue/consume; retries + metrics present. Conf high from breadth and usage. | Retry/backoff polish; graceful shutdown semantics; perf passes. | Unit + some integration; good. | Foundation is solid. |
|🔌 | [Admin API v1 (HTTP)](../docs/ideas/admin-api.md) | Platform/API | [Spec](../docs/ideas/admin-api.md) | [internal/admin-api](../internal/admin-api) | 5.2 | V1 | 90% (conf: high) | █████████░ | Endpoints + middleware + OpenAPI shipped. | TUI switchover for Stats; expand e2e; gRPC decision. | Unit + integration; good. | Productionize defaults; audit destructive ops. |
|🗄️ | [Storage Backends](../docs/ideas/storage-backends.md) | Core/Storage | [Spec](../docs/ideas/storage-backends.md) | [internal/storage-backends](../internal/storage-backends) | 5.3 | Alpha | 75% (conf: med) | ███████░░░ | Adapters + tests; conformance pending. | Complete adapter matrix; conformance; migration docs. | Unit + integration; fair. | Track compat matrix. |
|🛡️ | [RBAC & Tokens](../docs/ideas/rbac-and-tokens.md) | Security | [Spec](../docs/ideas/rbac-and-tokens.md) | [internal/rbac-and-tokens](../internal/rbac-and-tokens) | 3.1 | V1 | 85% (conf: high) | █████████░ | Manager + middleware; hardened. | Expand scopes; e2e coverage; audit trails. | Unit + middleware; good. | Security foundation. |
|🔭 | Observability Core | Observability | — | [internal/obs](../internal/obs) | 1.0 | V1 | 85% (conf: high) | █████████░ | Logger/metrics/tracing wiring. | Dashboards; error budgets. | Unit present. | Solid base. |
|🧹 | Reaper | Maintenance | — | [internal/reaper](../internal/reaper) | 0.1 | V1 | 90% (conf: high) | █████████░ | TTL/cleanup working. | Tune policies; monitoring. | Unit present. | Keep safe defaults. |

### TUI & UX
<!-- group-progress:tui-ux:begin -->
```text
███████████████████▓░░░░░░░░░░░░░░░░░░░░ 48%
------------|-------------|------------|
           MVP          Alpha    v1.0.0 
weight=14.12 features=12 kloc=43.0
```
<!-- group-progress:tui-ux:end -->

|Emoji | Feature | Area | Spec | Code | KLoC (approx) | Status | Progress % | Bar | Current State | Todo (Tasks) | Tests | Remarks |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
|🖥️ | [TUI Shell (Tabs/Layout)](../docs/TUI/README.md) | UX/TUI | [Spec](../docs/TUI/README.md) | [internal/tui](../internal/tui) | 2.5 | In Progress | 65% (conf: high) | ██████░░░░ | Tabs, charts expand, tiny‑term fixes done. | Wire Admin API; persist UI state; help overlay polish; table polish; adjustable panel split. | Manual + some unit; fair. | Incremental polish. |
|🧟 | [DLQ Remediation UI](../docs/ideas/dlq-remediation-ui.md) | Ops/TUI | [Spec](../docs/ideas/dlq-remediation-ui.md) | [internal/dlq-remediation-ui](../internal/dlq-remediation-ui) | 2.9 | In Progress | 55% (conf: med) | █████░░░░░ | API + TUI model exist; paging/filters pending. | Server‑side paging/filters; TUI list/peek; RBAC/audit hooks. | Unit present; needs e2e. | Prioritize perf. |
|👷 | [Workers View (TUI)](../docs/TUI/README.md) | UX/TUI | [Spec](../docs/TUI/README.md) | [internal/tui](../internal/tui) | 2.5 | In Progress | 35% (conf: med) | ███░░░░░░░ | Placeholder; no live list yet. | Use Admin workers endpoint; sort/filter; heartbeat display. | None; add UI tests. | Needs workers API wiring. |
|⚙️ | [Settings View (TUI)](../docs/TUI/README.md) | UX/TUI | [Spec](../docs/TUI/README.md) | [internal/tui](../internal/tui) | 2.5 | In Progress | 40% (conf: med) | ████░░░░░░ | Minimal snapshot. | Theme toggle; config path; copy/open shortcuts. | None; add snapshot tests. | Quick win. |
|🖱️ | [Right‑click Context Menus](../docs/ideas/right-click-context-menus.md) | UX/TUI | [Spec](../docs/ideas/right-click-context-menus.md) | [internal/right-click-context-menus](../internal/right-click-context-menus) | 2.3 | In Progress | 50% (conf: med) | █████░░░░░ | Menus/zones exist; focus wiring pending. | Connect to table rows; actions; tests; double‑click peek; header sort. | Unit present; needs UI/e2e. | Pair with bubblezone. |
|🎯 | Bubblezone Hitboxes | UX/TUI | — | [internal/right-click-context-menus](../internal/right-click-context-menus) | 2.3 | Planned | 10% (conf: med) | █░░░░░░░░░ | Not started; design known. | Integrate bubblezone; zone mapping for tabs/rows/splitter. | None. | Enables precise mouse UX. |
|🧾 | [JSON Payload Studio](../docs/ideas/json-payload-studio.md) | UX/TUI | [Spec](../docs/ideas/json-payload-studio.md) | [internal/json-payload-studio](../internal/json-payload-studio) | 4.0 | In Progress | 40% (conf: med) | ████░░░░░░ | Core handlers; not in TUI. | TUI editor; schemas/templates; enqueue path. | Unit present. | UX heavy. |
|📅 | [Calendar View](../docs/ideas/calendar-view.md) | UX/TUI | [Spec](../docs/ideas/calendar-view.md) | [internal/calendar-view](../internal/calendar-view) | 5.0 | In Progress | 45% (conf: med) | ████░░░░░░ | Routes/UI; auth/multi‑queue TODOs. | Add auth context; filters; paging. | Unit + TODOs. | Verify perf. |
|🎨 | [Theme Playground](../docs/ideas/theme-playground.md) | UX/TUI | [Spec](../docs/ideas/theme-playground.md) | [internal/theme-playground](../internal/theme-playground) | 5.3 | MVP | 70% (conf: high) | ███████░░░ | Persistence + tests shipped. | Centralize styles; Settings toggle; accessible palettes. | Unit + integration; good. | Accessibility focus. |
|🎙️ | [Terminal Voice Commands](../docs/ideas/terminal-voice-commands.md) | UX/CLI | [Spec](../docs/ideas/terminal-voice-commands.md) | [internal/terminal-voice-commands](../internal/terminal-voice-commands) | 5.8 | MVP | 70% (conf: high) | ███████░░░ | Core + tests done. | Privacy/offline; tutorial; TUI affordances. | Rich unit; good. | Optional, flashy. |
|🔌 | [Plugin Panel System](../docs/ideas/plugin-panel-system.md) | Extensibility | [Spec](../docs/ideas/plugin-panel-system.md) | [internal/plugin-panel-system](../internal/plugin-panel-system) | 3.7 | In Progress | 50% (conf: med) | █████░░░░░ | Lifecycle + permissions. | Sandbox; TUI registry; SDK docs. | Unit present. | Watch plugin trust. |
|🗺️ | [Visual DAG Builder](../docs/ideas/visual-dag-builder.md) | UX/Flow | [Spec](../docs/ideas/visual-dag-builder.md) | [internal/visual-dag-builder](../internal/visual-dag-builder) | 4.0 | In Progress | 40% (conf: med) | ████░░░░░░ | Orchestrator/types; not wired. | Backend validation; DAG execution; TUI builder. | Unit partial. | Longer‑term. |

### Reliability & Ops
<!-- group-progress:reliability-ops:begin -->
```text
███████████████████▓░░░░░░░░░░░░░░░░░░░░ 50%
------------|-------------|------------|
           MVP          Alpha    v1.0.0 
weight=9.54 features=8 kloc=32.6
```
<!-- group-progress:reliability-ops:end -->

|Emoji | Feature | Area | Spec | Code | KLoC (approx) | Status | Progress % | Bar | Current State | Todo (Tasks) | Tests | Remarks |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
|🧰 | [DLQ Remediation Pipeline](../docs/ideas/dlq-remediation-pipeline.md) | Reliability | [Spec](../docs/ideas/dlq-remediation-pipeline.md) | [internal/dlq-remediation-pipeline](../internal/dlq-remediation-pipeline) | 4.7 | In Progress | 45% (conf: med) | ████░░░░░░ | Pipeline scaffolding; classifiers/rules TBD. | Rules engine; rate‑limited requeue; safety bounds. | Light unit; needs scenario tests. | Integrate with DLQ UI. |
|✅ | [Exactly‑once Patterns](../docs/ideas/exactly-once-patterns.md) | Reliability | [Spec](../docs/ideas/exactly-once-patterns.md) | [internal/exactly_once](../internal/exactly_once), [internal/exactly-once-patterns](../internal/exactly-once-patterns) | 7.4 | MVP | 70% (conf: high) | ███████░░░ | Idempotency/outbox ready; some TODOs. | Finalize metrics; publisher adapters; docs. | Unit + integration; good. | Strong differentiator. |
|⚖️ | [Advanced Rate Limiting](../docs/ideas/advanced-rate-limiting.md) | Throughput | [Spec](../docs/ideas/advanced-rate-limiting.md) | [internal/advanced-rate-limiting](../internal/advanced-rate-limiting) | 1.6 | In Progress | 55% (conf: high) | █████░░░░░ | Lua token bucket + fairness done. | Admin API runtime updates; TUI widget; producer/worker hooks. | Unit + integration; good. | High leverage; wire into SDKs. |
|🌊 | [Producer Backpressure](../docs/ideas/producer-backpressure.md) | SDKs | [Spec](../docs/ideas/producer-backpressure.md) | [internal/producer-backpressure](../internal/producer-backpressure) | 3.3 | In Progress | 40% (conf: med) | ████░░░░░░ | Signals present; not linked to RL. | Integrate with rate limiter; client SDK examples. | Unit present. | Needs producer docs. |
|🧮 | [Policy Simulator](../docs/ideas/policy-simulator.md) | Ops/Safety | [Spec](../docs/ideas/policy-simulator.md) | [internal/policy-simulator](../internal/policy-simulator) | 4.7 | In Progress | 45% (conf: med) | ████░░░░░░ | Core present; retrieval/rollback TODO. | Preview/apply/rollback endpoints; persist scenarios. | Unit present. | Pair with Admin API. |
|🛠️ | [Worker Fleet Controls](../docs/ideas/worker-fleet-controls.md) | Ops | [Spec](../docs/ideas/worker-fleet-controls.md) | [internal/worker-fleet-controls](../internal/worker-fleet-controls) | 3.1 | In Progress | 45% (conf: med) | ████░░░░░░ | Control scaffolding; safety checks TBD. | Pause/drain/resume + RBAC; per‑node metrics; TUI controls. | Unit present. | Add safety gates. |
|🗃️ | [Long‑term Archives](../docs/ideas/long-term-archives.md) | Ops/Data | [Spec](../docs/ideas/long-term-archives.md) | [internal/long-term-archives](../internal/long-term-archives) | 4.2 | In Progress | 45% (conf: med) | ████░░░░░░ | Archival hooks; adapters TBD. | S3/ClickHouse adapters; retention; export path. | Unit partial. | Define retention/SLO. |
|🔔 | [Event Hooks](../docs/ideas/event-hooks.md) | Integrations | [Spec](../docs/ideas/event-hooks.md) | [internal/event-hooks](../internal/event-hooks) | 3.6 | In Progress | 50% (conf: med) | █████░░░░░ | Plumbing exists; config/signing TODO. | Configurable base URL; HMAC signatures; retries; Admin mgmt. | Unit present. | Security first. |

### Scale & Multi‑Cluster
<!-- group-progress:scale-multi-cluster:begin -->
```text
████████████████████▓░░░░░░░░░░░░░░░░░░░ 52%
------------|-------------|------------|
           MVP          Alpha    v1.0.0 
weight=3.53 features=3 kloc=10.2
```
<!-- group-progress:scale-multi-cluster:end -->

|Emoji | Feature | Area | Spec | Code | KLoC (approx) | Status | Progress % | Bar | Current State | Todo (Tasks) | Tests | Remarks |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
|🧭 | [Multi‑cluster Control](../docs/ideas/multi-cluster-control.md) | Scale | [Spec](../docs/ideas/multi-cluster-control.md) | [internal/multi-cluster-control](../internal/multi-cluster-control) | 3.6 | In Progress | 60% (conf: med) | ██████░░░░ | Manager/handlers + tests; UI pending. | TUI tabs; Admin fan‑out actions; compare/replicate ops. | Many tests; good. | Solid engine; wire UX. |
|☸️ | [Kubernetes Operator](../docs/ideas/kubernetes-operator.md) | Platform | [Spec](../docs/ideas/kubernetes-operator.md) | [internal/kubernetes-operator](../internal/kubernetes-operator) | 3.8 | In Progress | 55% (conf: med) | █████░░░░░ | Controllers/webhooks; examples/tests. | CRDs; reconcile backoff; e2e on kind. | Unit + integration; fair. | Mind CRD validation. |
|🏷️ | [Multi‑tenant Isolation](../docs/ideas/multi-tenant-isolation.md) | Security | [Spec](../docs/ideas/multi-tenant-isolation.md) | [internal/multi-tenant-isolation](../internal/multi-tenant-isolation) | 2.8 | In Progress | 40% (conf: med) | ████░░░░░░ | Handlers with RBAC TODOs. | Enforce quotas/keys; authz middleware; tests. | Unit present. | Needs policy decisions. |

### Observability & Analytics
<!-- group-progress:observability-analytics:begin -->
```text
█████████████████████▓░░░░░░░░░░░░░░░░░░ 54%
------------|-------------|------------|
           MVP          Alpha    v1.0.0 
weight=8.06 features=7 kloc=20.2
```
<!-- group-progress:observability-analytics:end -->

|Emoji | Feature | Area | Spec | Code | KLoC (approx) | Status | Progress % | Bar | Current State | Todo (Tasks) | Tests | Remarks |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
|🔎 | [Distributed Tracing Integration](../docs/ideas/distributed-tracing-integration.md) | Observability | [Spec](../docs/ideas/distributed-tracing-integration.md) | [internal/distributed-tracing-integration](../internal/distributed-tracing-integration) | 3.0 | Alpha | 85% (conf: high) | █████████░ | OTEL propagation + trace URLs done. | Link from TUI; config docs. | Unit + integration; good. | Low risk polish. |
|🧵 | [Trace Drill‑down + Log Tail](../docs/ideas/trace-drilldown-log-tail.md) | Observability | [Spec](../docs/ideas/trace-drilldown-log-tail.md) | [internal/trace-drilldown-log-tail](../internal/trace-drilldown-log-tail) | 3.9 | In Progress | 50% (conf: med) | █████░░░░░ | Trace links ok; log tail TBD. | Tail with filters; privacy; TUI links. | Unit partial. | Watch PII. |
|📡 | [Anomaly Radar + SLO Budget](../docs/ideas/anomaly-radar-slo-budget.md) | Observability | [Spec](../docs/ideas/anomaly-radar-slo-budget.md) | [internal/anomaly-radar-slo-budget](../internal/anomaly-radar-slo-budget) | 2.8 | In Progress | 45% (conf: med) | ████░░░░░░ | Handlers/metrics skeleton. | Define SLO; thresholds; Prom metrics; widget. | Unit partial. | Needs calibration. |
|📊 | [Forecasting](../docs/ideas/forecasting.md) | Planning | [Spec](../docs/ideas/forecasting.md) | [internal/forecasting](../internal/forecasting) | 2.7 | In Progress | 40% (conf: med) | ████░░░░░░ | Stubs exist. | Baseline models; eval harness; TUI preview. | Unit partial. | Keep simple first. |
|🧪 | [Queue Snapshot Testing](../docs/ideas/queue-snapshot-testing.md) | QA | [Spec](../docs/ideas/queue-snapshot-testing.md) | [internal/queue-snapshot-testing](../internal/queue-snapshot-testing) | 2.4 | In Progress | 50% (conf: med) | █████░░░░░ | Framework + snapshots. | Broaden differ; golden tests; docs. | Unit; fair. | Useful for regressions. |
|📈 | [Patterned Load Generator](../docs/ideas/patterned-load-generator.md) | Testing | [Spec](../docs/ideas/patterned-load-generator.md) | [internal/patterned-load-generator](../internal/patterned-load-generator) | 2.1 | In Progress | 45% (conf: med) | ████░░░░░░ | Handlers + generator; guardrails missing. | Add sine/burst/ramp; cancel/stop; profiles; TUI overlay. | Unit present; needs e2e. | Add caps; confirmations. |
|🧪 | Bench (Basic) | Testing | — | [internal/admin](../internal/admin), [internal/tui](../internal/tui) | 3.3 | Alpha | 60% (conf: med) | ██████░░░░ | Running; progress UI present; baseline delta pending. | Baseline from initial completed list; cancel; ETA/throughput; guardrails. | Manual + some unit. | Guardrails for high rates. |
