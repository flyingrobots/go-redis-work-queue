# Design–Implementation Drift Report

This report compares each feature spec in `docs/ideas/` against the current implementation under `internal/` and related code. It highlights alignment, gaps, and concrete next steps.

## Executive Summary

- Overall alignment score: 60.1/100 (Medium)
- Overall drift: 39.9%
- Scope: 37 feature specs reviewed; corresponding `internal/<feature>` modules scanned (handlers, logic, tests, and TODOs), plus TUI and Admin API integration points where relevant.

Key observations
- Foundation is strong: Admin API, tracing, exactly-once, storage backends, theme playground, terminal voice, and time-travel debugger are comparatively well aligned (low drift ≤25%).
- Productization gaps recur: runtime configuration endpoints, RBAC tie‑in, pagination at scale, and TUI wiring are the most common sources of drift.
- Several modules are substantial but not yet integrated into TUI flows or Admin API surfaces (e.g., rate limiting, DLQ UI depth, right‑click menus, patterned load).

## Scoring Model

- We compute a feature Alignment Score (0–100) from code presence, tests, and obvious TODOs; Drift % = 100 − Alignment.
- Categories (by Alignment Score):
  - Critical: 0–39 (very high drift)
  - High: 40–59
  - Medium: 60–79
  - Low: 80–89
  - Excellent: 90–100 (very low drift)

Heuristic inputs used for this pass: presence of `internal/<feature>` module, number of `.go` files and tests, explicit TODOs/WIP markers, basic route/handler availability, and TUI/Admin integration indicators found via grep. This is a fast quantitative baseline; targeted deep‑dives recommended for “High” and “Critical”.

## Critical Issues (Immediate Attention)

- Collaborative Session (Drift 80%): Minimal code, no tests, not integrated.
- Advanced Rate Limiting (Drift 60%): Strong core limiter present, but runtime Admin API controls and TUI widgets missing; producer/worker integration incomplete.
- Patterned Load Generator (Drift 60%): Handlers exist; patterns/scheduler/guardrails and TUI overlay not wired end‑to‑end.

## Detailed Findings by Category

- Low Drift (≤25%): admin-api, distributed-tracing-integration, storage-backends, time-travel-debugger, rbac-and-tokens, smart-retry-strategies, terminal-voice-commands, theme-playground
- Medium Drift (30–45%): exactly-once-patterns, multi-cluster-control, visual-dag-builder, automatic-capacity-planning, kubernetes-operator, producer-backpressure, event-hooks, job-genealogy-navigator, calendar-view, anomaly-radar-slo-budget, canary-deployments, chaos-harness, forecasting, json-payload-studio, job-budgeting, dlq-remediation-pipeline, dlq-remediation-ui, trace-drilldown-log-tail, worker-fleet-controls, right-click-context-menus, smart-payload-deduplication
- High–Critical (≥55%): multi-tenant-isolation, queue-snapshot-testing, advanced-rate-limiting, patterned-load-generator, collaborative-session

## Recommendations (Cross‑Cutting)

- Close the loop to TUI: For DLQ UI, rate limiting, patterned load, context menus — wire handlers into `internal/tui` and persist state where needed.
- Expose runtime controls via Admin API: Add update endpoints for limits, toggles, and policies; document OpenAPI and add tests.
- Pagination and guardrails: Ensure DLQ, list endpoints, and patterned load enforce bounds and use server‑side pagination.
- RBAC integration everywhere: Require tokens for sensitive ops; confirm deny‑by‑default semantics; add audit logs to destructive flows.
- Tests and docs: Add handler and e2e tests for newly exposed endpoints; update README/TUI docs with new keybindings and flows.

---

## Feature Drift Table

| Feature | Folder | Finished? | Drift % | Remarks | Recommendation |
| --- | --- | --- | ---:| --- | --- |
| admin-api | internal/admin-api | Yes | 20 | HTTP v1 endpoints (Stats, Keys, Peek, Purge DLQ/All, Bench) and middleware (auth, rate‑limit, audit, CORS, recovery) implemented; OpenAPI served; gRPC not present; TUI still mostly calls internal helpers directly. | Switch TUI Stats to Admin API; decide on gRPC scope (ship or de‑scope); expand CI integration tests; confirm deny‑by‑default tokens in all destructive routes. |
| advanced-rate-limiting | internal/advanced-rate-limiting | No | 60 | Redis Lua token bucket with priority fairness and tests in place; missing Admin API runtime updates, dry‑run preview endpoint, producer/worker integration, and TUI status widget. | Add Admin API CRUD for limits/weights; integrate with producer hints/worker throttling; surface TUI widget and metrics; document tuning. |
| anomaly-radar-slo-budget | internal/anomaly-radar-slo-budget | No | 45 | Handlers exist; metrics/alerts endpoints stubbed; thresholds/SLO budgets not clearly tuned; limited tests. | Define SLO config and thresholds; add Prometheus metrics; wire TUI “status/radar” widget; add calibration docs. |
| automatic-capacity-planning | internal/automatic-capacity-planning | No | 35 | Core structs and planner logic present; needs scheduler hooks and Admin API surface; limited integration tests. | Expose plan/apply/dry‑run via Admin API; add forecast inputs; e2e soak tests with patterned load. |
| calendar-view | internal/calendar-view | No | 35 | Routes and TUI helpers exist; TODOs for auth and multi‑queue filtering; pagination needs verification. | Add auth context, filters, and paging; connect to Admin API; add large‑list tests. |
| canary-deployments | internal/canary-deployments | No | 45 | Canary logic exists; limited production guardrails and rollback controls; minimal test coverage. | Add rollback/abort endpoints; audit logging; e2e with worker fleets. |
| chaos-harness | internal/chaos-harness | No | 45 | Fault injection scaffolding present; missing safety gates and observability glue. | Add scoped chaos profiles, RBAC, and kill‑switch; record effects to tracing/metrics. |
| collaborative-session | internal/collaborative-session | No | 80 | Minimal code; no tests; not wired to TUI. | Define protocol and permissions; add session host/guest modes; defer if out of scope this release. |
| distributed-tracing-integration | internal/distributed-tracing-integration | Yes | 20 | OTEL integration, trace propagation tests, and trace URL helpers implemented. | Link from TUI job views to external tracing UI; document configuration. |
| dlq-remediation-pipeline | internal/dlq-remediation-pipeline | No | 45 | Pipeline components present; classifier rules and auto‑retry policies limited; tests light. | Add rules engine, rate‑limited requeue, and safety bounds; expose via Admin API. |
| dlq-remediation-ui | internal/dlq-remediation-ui | No | 45 | API handlers (list/peek/requeue/purge) and TUI model exist; single test file; need pagination at scale and TUI polish; ensure Admin API parity. | Implement server‑side pagination and filters; expand tests; integrate with main TUI tab and keybindings; confirm RBAC/audit on destructive ops. |
| event-hooks | internal/event-hooks | No | 40 | Webhook plumbing in place; base URL TODO and health/status routes present. | Make base URL configurable; add signing/verification and retries; Admin API to manage subscriptions. |
| exactly-once-patterns | internal/exactly-once-patterns | No | 30 | Idempotency/outbox patterns implemented; some TODOs (hit‑rate calc, publishers). | Finalize metrics; add publisher adapters; document patterns and failure modes. |
| forecasting | internal/forecasting | No | 45 | Forecast stubs implemented; needs model selection and evaluation harness. | Provide baseline models (ARIMA/Prophet external or simple EMA); surface via Admin API and TUI. |
| job-budgeting | internal/job-budgeting | No | 45 | Budget manager, cost model, notifications present; limited tests and UI. | Add enforcement hooks and Admin API; TUI budget panel; alerting thresholds. |
| job-genealogy-navigator | internal/job-genealogy-navigator | No | 40 | Types and graph traversal present; non‑Go assets for views; integration unclear. | Expose via Admin API; TUI drill‑down; add pagination on lineage. |
| json-payload-studio | internal/json-payload-studio | No | 45 | Core handlers present; minimal tests; not fully plugged into TUI. | Add validation schemas, templates, and enqueue paths; TUI editor with previews. |
| kubernetes-operator | internal/kubernetes-operator | No | 35 | Operator scaffolding and tests exist; CRDs integration simulated. | Define CRDs; reconcile loops; e2e against kind; RBAC manifests. |
| long-term-archives | internal/long-term-archives | No | 45 | Archival hooks present; backends not fully pluggable; tests light. | Implement S3/ClickHouse adapters; retention/TTL policies; Admin API to export. |
| multi-cluster-control | internal/multi-cluster-control | No | 30 | Manager, errors, handlers, and many artifacts exist; compare/switch logic present; lots of non‑Go test assets. | Finalize e2e tests; wire to TUI tabs; Admin API for fan‑out actions with audit. |
| multi-tenant-isolation | internal/multi-tenant-isolation | No | 55 | Handlers with TODOs for RBAC validation; quotas/tenancy boundaries incomplete. | Enforce tenant authz middleware; define quotas and keys; add tests for isolation. |
| patterned-load-generator | internal/patterned-load-generator | No | 60 | Handlers exist; patterns/scheduling/guardrails and TUI overlay unproven; single test. | Implement sine/burst/ramp + stop/cancel; guardrails; profile save/load; chart overlay in TUI. |
| plugin-panel-system | internal/plugin-panel-system | No | 40 | Plugin lifecycle present; hot‑reload/sandboxing needs validation; non‑Go assets included. | Add permission model and sandbox; plugin SDK docs; TUI panel registry. |
| policy-simulator | internal/policy-simulator | No | 40 | Simulator with TODOs for retrieval/rollback; multiple tests present. | Wire Admin API preview/apply/rollback; persist scenarios; integrate with TUI. |
| producer-backpressure | internal/producer-backpressure | No | 35 | Backpressure scaffolding and tests; not hooked into rate limiter hints. | Integrate with advanced rate limiting; expose hints in client SDKs; metrics. |
| queue-snapshot-testing | internal/queue-snapshot-testing | No | 55 | Snapshot framework present with testdata; low code/test counts suggest early stage. | Expand differ coverage; add golden tests; docs for snapshot lifecycle. |
| rbac-and-tokens | internal/rbac-and-tokens | Yes | 25 | JWT manager, middleware, and handlers with tests; revoke and cache implemented. | Integrate with all sensitive Admin API routes; add per‑action scopes and audit trails. |
| right-click-context-menus | internal/right-click-context-menus | No | 50 | Zones/registry/menu implemented; TODO for focused item context; not wired into TUI tables. | Integrate with TUI table rows via bubblezone; add context actions and tests. |
| smart-payload-deduplication | internal/smart-payload-deduplication | No | 50 | Compression/dedup logic present with TODOs (dict build); limited integration. | Add dict training pipeline; expose dedup stats; integrate in enqueue path. |
| smart-retry-strategies | internal/smart-retry-strategies | Yes | 25 | Strategies and handlers implemented; metrics endpoint TODO; decent tests. | Implement Prometheus metrics; surface strategy selection in TUI; add docs. |
| storage-backends | internal/storage-backends | Yes | 20 | Multiple backends scaffolding present with tests; OpenAPI docs exist. | Complete adapter matrix; add conformance tests; document migration paths. |
| terminal-voice-commands | internal/terminal-voice-commands | Yes | 25 | Command mapping, handlers, and tests present; privacy/telemetry not addressed. | Add opt‑in, PII handling, and offline mode; short TUI tutorial. |
| theme-playground | internal/theme-playground | Yes | 25 | Theme system prototypes and tests; TUI integration partial. | Centralize styles; add theme toggle in Settings tab; docs for accessible palettes. |
| time-travel-debugger | internal/time-travel-debugger | Yes | 20 | Capture/replay and simple TUI implemented with tests. | Add selective replay controls; export/import; document guardrails. |
| trace-drilldown-log-tail | internal/trace-drilldown-log-tail | No | 45 | Trace ID plumbing present; log tail integration limited; tests partial. | Add tailing with filters; link TUI job to trace URL; privacy filtering. |
| visual-dag-builder | internal/visual-dag-builder | No | 30 | Orchestrator and types present; some .bak tests; not wired to enqueue pipeline. | Backend validation and DAG execution plan; TUI builder; Admin API to submit DAGs. |
| worker-fleet-controls | internal/worker-fleet-controls | No | 45 | Control handlers and audits exist; needs live graphs and safety checks. | Add pause/drain/resume with RBAC; per‑node metrics; TUI controls panel. |

---

## Methodology Notes

- Inputs: `docs/ideas/*.md` specs; `internal/<feature>` modules; grep for endpoints/routes, middleware, and TODOs; presence of tests and non‑Go assets; TUI and Admin API usage.
- Limitations: This pass uses heuristic signals and spot‑reads; it does not execute binaries or hit live services. High‑drift items should get a focused deep‑dive before scheduling.

## Next Steps

- Confirm priorities in AGENTS.md backlog; create tickets per “Recommendation” above.
- If desired, I can: (1) wire TUI to Admin API Stats, (2) add Admin API endpoints for rate‑limits, and (3) implement DLQ pagination + filters with tests.

