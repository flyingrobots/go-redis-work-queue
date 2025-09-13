# AGENTS Notes

Quick notes for working on this repo (Go Redis Work Queue) — things I’ve learned / want to remember when iterating fast.

- TUI stack and structure
  - Bubble Tea + Lip Gloss + Bubbles (`table`, `viewport`, `spinner`, `progress`) and a custom scrim overlay (no external overlay dep now).
  - Entry point: `cmd/tui/main.go` constructs `internal/tui` model with config + redis + zap logger.
  - Core TUI files: `internal/tui/{model,init,app,view,commands,overlays}.go`.
  - Tabs: `internal/tui/tabs.go` renders "Job Queue", "Workers", "Dead Letter", "Settings" with per-tab border colors + mouse switching.
  - Data polling: periodic `stats` + `keys` via `internal/admin` helpers; charts maintain short time series per queue alias.

- Overlays and input behavior
  - Confirmation modal and Help use a full-screen scrim overlay that centers content and dims background; resilient to any terminal size.
  - ESC priority:
    1) Close confirm modal if open
    2) Exit bench inputs if focused
    3) Clear active filter
    4) Otherwise toggle Help overlay

- Current tabs
  - Job Queue: existing dashboard (Queues table + Charts + Info). Filter (`f`/`/`), peek (`p`/enter), bench (`b` then enter), progress bar.
  - Workers: placeholder summary (heartbeats, processing lists); will grow to live workers view.
  - Dead Letter: placeholder summary (DLQ key + count) with future actions (peek/purge/requeue).
  - Settings: read-only snapshot of a few config values.

- Keybindings (important)
  - `q`/`ctrl+c`: quit (asks to confirm)
  - `esc`: help toggle, or exit modal/input as above
  - `tab`/`shift+tab`: move panel focus (within Job Queue tab)
  - `j/k`, mouse wheel: scroll
  - `p`/enter on a queue: peek
  - `b`: bench form (tab cycles inputs, enter runs)
  - `f` or `/`: filter queues (fuzzy); `esc` clears
  - `D` / `A`: confirm purge DLQ / purge ALL
  - Mouse: click tabs to switch, left-click (Job Queue) peeks selected

- Redis/admin plumbing
  - Uses `internal/admin` for `Stats`, `StatsKeys`, `Peek`, `Bench`, `Purge*`.
  - Completed progress for bench is polled from `cfg.Worker.CompletedList` (keep in mind large lists can be slow to LLen).

- Config + run
  - Config path flag: `--config config/config.yaml`, refresh via `--refresh`.
  - Build: `make build` or `go build -o bin/tui ./cmd/tui`; run `./bin/tui --config config/config.yaml`.

- Observability
  - Zap logger; metrics on `:9090/metrics`; liveness `/healthz`, readiness `/readyz`.

- Guardrails
  - Purge actions gated by confirm modal (DLQ / ALL). Don’t run in prod without care.
  - Bench can generate many jobs fast; prefer test env and lower rates.

- Near-term TODOs I’m targeting
  - Charts expand-on-click (toggle 2/3 vs 1/3), precise mouse hitboxes (bubblezone), table polish (striping, thresholds, selection glyph), enqueue actions (`e`/`E`), right-click peek.

## Working Tasklist (maintain and use this from now on)

Use this checklist to track work. Keep it prioritized, update statuses, and reference it in PRs/commits. Add new items as they surface; close them when done.

- [ ] TUI: Charts expand-on-click (Charts 2/3 vs Queues 1/3; toggle back on Queues click)
- [ ] TUI: Integrate `bubblezone` for precise mouse hitboxes (tabs, table rows, future context menus)
- [ ] TUI: Table polish — colorized counts by thresholds (green/yellow/red), selection glyph, alternating row striping
- [ ] TUI: Enqueue actions — `e` enqueue 1 to selected; `E` prompt for count (inline in Info panel)
- [ ] TUI: Right-click on Queues — peek selected (later: context menu with actions)
- [ ] TUI: Keyboard shortcuts for tabs (`1`..`4` to switch)
- [ ] TUI: Persist UI state across sessions (active tab, focus panel, filter value)
- [ ] TUI: Improve tiny-terminal layout — stack panels vertically; clamp widths; hide charts if extremely narrow
- [ ] TUI: Adjustable panel split — keys (`[`/`]`) or drag on splitter (mouse) to change left/right ratio
- [ ] TUI: Bench UX — cancel with ESC; show ETA; live throughput; configurable payload size and jitter; concurrency knob
- [ ] TUI: Bench progress baseline — compute delta from initial `CompletedList` length (avoid overcount if list pre-populated)
- [ ] TUI: DLQ tab — list/paginate items; peek full payload; requeue selected; purge selected; search/filter
- [ ] TUI: Workers tab — list worker IDs, last heartbeat time, processing queue/job; sort and filter
- [ ] TUI: Settings tab — theme toggle; show config path; copy key values; open config file shortcut
- [ ] TUI: Theme system — centralize styles; dark/light + high-contrast palette via lipgloss adaptive colors
- [ ] TUI: Help overlay — expand with all shortcuts (tabs, enqueue, right-click), add mouse hints, link to README
- [ ] TUI: Mouse UX — double-click row to peek; click column header to sort if supported
- [ ] TUI: Non-blocking error toasts/status area for transient errors (top-right), with log tail in Info
- [ ] TUI: Unit tests for pure helpers (filtering, formatting, thresholds, clamp)
- [ ] Admin: Requeue-from-DLQ command with count/range support (exposed to TUI)
- [ ] Admin: Workers-list admin call (IDs, last heartbeat, active item) for Workers tab
- [ ] Metrics: Optional TUI runtime metrics (ticks, RPC latency) for debugging
- [ ] Docs: Update README TUI section with tabs, screenshots, and new keybindings
- [ ] Release: Add changelog entries for TUI tabbed layout and overlays

## WILD IDEAS — HAVE A BRAINSTORM

Capture ambitious, unconventional ideas. Some may be long-term or require new components; still worth recording for future exploration.

- TUI: Live log tail + trace drill-down — attach to worker logs, show correlated OpenTelemetry spans; press a job to open its trace waterfall.
- TUI: Visual DAG builder for multi-step workflows — drag-and-drop stages with dependencies, retries, and compensation actions; submit as a reusable pipeline.
- TUI: Anomaly radar — backlog growth, p95 latency spikes, failure-rate heatmap; SLO error budget meter with burn alerts.
- TUI: Interactive policy tuning — edit retry/backoff, rate limits, concurrency caps; preview impact with a simulator; apply with one keystroke.
- TUI: Patterned load generator — sine/burst/ramp traffic models; schedule runs; export reproducible profiles for CI.
- TUI: Multi-cluster control — tabs for multiple Redis endpoints; quick switch and side-by-side compare; propagate admin actions across clusters.
- TUI: Plugin panel system — drop-in panels (Go, WASM, or Lua) for custom org metrics, transforms, or actions; hot-reload safely.
- TUI: JSON payload studio — pretty-edit, validate, and enqueue; templates and snippets; schedule run-at/cron.
- TUI: Calendar view — visualize scheduled and recurring jobs; click to reschedule or pause a rule.
- TUI: Worker fleet controls — pause/resume/drain nodes; rolling restarts; live CPU/mem/net graphs per worker.
- TUI: Right-click context menus everywhere — requeue, purge, copy payload, copy Redis key, open trace, export sample.
- TUI: Collaborative session — multiplexed read-only share over SSH; presenter hands control with a key.
- TUI: Theme playground — high-contrast/accessible themes; auto-switch based on OS or time of day.
- Project: HTTP/gRPC admin API — first-class, versioned contract used by both TUI and a web UI; enable remote control and automation.
- Project: Kubernetes Operator — CRDs for queues/workers; reconcile deployments; autoscale by backlog and SLA targets; preemption policies.
- Project: Advanced rate limiting — token-bucket with priority fairness; global and per-tenant budgets; dynamic tuning via feedback signals.
- Project: Producer backpressure — SDK hints when queues are saturated; adaptive rate; circuit breaking by priority.
- Project: Multi-tenant isolation — quotas, per-tenant keys, encryption at rest (payload), audit logs, privacy scrubbing hooks.
- Project: DLQ remediation pipeline — automatic classifiers to cluster failures; rules to auto-retry, transform, or quarantine.
- Project: Storage backends — pluggable engines (Redis Streams, KeyDB/Dragonfly, Redis Cluster); optional Kafka outbox bridge.
- Project: Long-term archives — stream completed jobs to ClickHouse/S3; TTL retention; fast query for forensics.
- Project: Event hooks — webhooks or NATS for job state changes; Slack/PagerDuty notifications with deep links to TUI.
- Project: RBAC and tokens — signed admin commands, per-action permissions; audit trail UI.
- Project: Chaos harness — inject latency, drops, and Redis failovers; visualize recovery; automate soak/chaos scenarios.
- Project: Forecasting — simple ARIMA/Prophet on backlog/throughput; recommend scale-up/down and SLA adjustments.
- Project: Exactly-once patterns — idempotency keys, dedup sets, and transactional outbox patterns documented and optionally enforced.
