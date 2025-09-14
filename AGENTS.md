# AGENTS

- Quick notes for working on this repo (Go Redis Work Queue) 
- Things learned / want to remember when iterating fast
- Activity log
- Tasklist
- Ideas

<!-- Table of Contents -->
# Table of Contents
1. [[AGENTS# AGENTS|AGENTS]]
	1. [[AGENTS## Important Information|Important Information]]
		1. [[AGENTS### Sections You Must Actively Maintain|Sections You Must Actively Maintain]]
		2. [[AGENTS### Job Queue|Job Queue]]
		3. [[AGENTS### TUI App|TUI App]]
			1. [[AGENTS#### TUI stack and structure|TUI stack and structure]]
			2. [[AGENTS#### Overlays and input behavior|Overlays and input behavior]]
			3. [[AGENTS#### Current tabs|Current tabs]]
			4. [[AGENTS#### Keybindings (important)|Keybindings (important)]]
			5. [[AGENTS#### Redis/admin plumbing|Redis/admin plumbing]]
			6. [[AGENTS#### Config + run|Config + run]]
			7. [[AGENTS#### Observability|Observability]]
				1. [[AGENTS##### Guardrails|Guardrails]]
		4. [[AGENTS### Project Status|Project Status]]
		5. [[AGENTS### Notes|Notes]]
	2. [[AGENTS## Working Tasklist|Working Tasklist]]
		1. [[AGENTS### Prioritized Backlog|Prioritized Backlog]]
		2. [[AGENTS### Finished Log|Finished Log]]
	3. [[AGENTS## Daily Activity Logs|Daily Activity Logs]]
		1. [[AGENTS### 2025-09-13–Rewrote `AGENTS.md`|2025-09-13–Rewrote `AGENTS.md`]]
			1. [[AGENTS#### ##### 06:39 – Starting `AGENTS.md` Enhancements|##### 06:39 – Starting `AGENTS.md` Enhancements]]
				1. [[AGENTS##### 06:39 – Starting `AGENTS.md` Enhancements|06:39 – Starting `AGENTS.md` Enhancements]]
	4. [[AGENTS## APPENDIX B: WILD IDEAS — HAVE A BRAINSTORM|APPENDIX B: WILD IDEAS — HAVE A BRAINSTORM]]
		1. [[AGENTS### Codex's Top Picks|Codex's Top Picks]]
	5. [[AGENTS## Appendix C: Codex Ideas in Detail|Appendix C: Codex Ideas in Detail]]

<!-- End of TOC -->

---
## Important Information

### Sections You Must Actively Maintain

It is **CRITICAL** to keep the following sections of this document up-to-date as you work.

- What You Should Know
- Working Tasklist
- Daily Activity Logs
### Job Queue

This project is a legit job queue backed by Redis, implemented in Go. The aim is to build a robust, horizontally scalable job system, balancing powerful features against real-world pragmatism, and keep things easy to use and understand. 

See the [README.md](./README.md) for more information.

### TUI App

There's a fancy TUI for interacting with and monitoring the job system. The app's main view is a tabbed UX, where each tab is centered around various app domains. The user can use the keyboard and mouse to interact with, monitor, and debug the system.

#### TUI stack and structure

  - Bubble Tea + Lip Gloss + Bubbles (`table`, `viewport`, `spinner`, `progress`) and a custom scrim overlay (no external overlay dep now).
  - Entry point: `cmd/tui/main.go` constructs `internal/tui` model with config + redis + zap logger.
  - Core TUI files: `internal/tui/{model,init,app,view,commands,overlays}.go`.
  - Tabs: `internal/tui/tabs.go` renders "Job Queue", "Workers", "Dead Letter", "Settings" with per-tab border colors + mouse switching.
  - Data polling: periodic `stats` + `keys` via `internal/admin` helpers; charts maintain short time series per queue alias.

#### Overlays and input behavior

  - Confirmation modal and Help use a full-screen scrim overlay that centers content and dims background; resilient to any terminal size.
  - ESC priority:
    1) Close confirm modal if open
    2) Exit bench inputs if focused
    3) Clear active filter
    4) Otherwise toggle Help overlay
#### Current tabs

- Job Queue: existing dashboard (Queues table + Charts + Info). Filter (`f`/`/`), peek (`p`/enter), bench (`b` then enter), progress bar.
- Workers: placeholder summary (heartbeats, processing lists); will grow to live workers view.
- Dead Letter: placeholder summary (DLQ key + count) with future actions (peek/purge/requeue).
- Settings: read-only snapshot of a few config values.

#### Keybindings (important)

  - `q`/`ctrl+c`: quit (asks to confirm)
  - `esc`: help toggle, or exit modal/input as above
  - `tab`/`shift+tab`: move panel focus (within Job Queue tab)
  - `j/k`, mouse wheel: scroll
  - `p`/enter on a queue: peek
  - `b`: bench form (tab cycles inputs, enter runs)
  - `f` or `/`: filter queues (fuzzy); `esc` clears
  - `D` / `A`: confirm purge DLQ / purge ALL
  - Mouse: click tabs to switch, left-click (Job Queue) peeks selected

#### Redis/admin plumbing

  - Uses `internal/admin` for `Stats`, `StatsKeys`, `Peek`, `Bench`, `Purge*`.
  - Completed progress for bench is polled from `cfg.Worker.CompletedList` (keep in mind large lists can be slow to LLen).

#### Config + run

  - Config path flag: `--config config/config.yaml`, refresh via `--refresh`.
  - Build: `make build` or `go build -o bin/tui ./cmd/tui`; run `./bin/tui --config config/config.yaml`.

#### Observability

  - Zap logger; metrics on `:9090/metrics`; liveness `/healthz`, readiness `/readyz`.

##### Guardrails

  - Purge actions gated by confirm modal (DLQ / ALL). Don’t run in prod without care.
  - Bench can generate many jobs fast; prefer test env and lower rates.

### Project Status

- Alpha RC in PR
- TUI started

### Notes

Near-term TODOs I’m targeting:

- Charts expand-on-click (toggle 2/3 vs 1/3), precise mouse hitboxes (bubblezone), table polish (striping, thresholds, selection glyph), enqueue actions (`e`/`E`), right-click peek.

---

## Working Tasklist

(maintain and use this from now on)

Use this checklist to track work. Keep it prioritized, update statuses, and reference it in PRs/commits. Add new items as they surface; close them when done. This is your backlog.

### Prioritized Backlog

- [x] TUI: Charts expand-on-click (Charts 2/3 vs Queues 1/3; toggle back on Queues click)
- [ ] TUI: Integrate `bubblezone` for precise mouse hitboxes (tabs, table rows, future context menus)
- [ ] TUI: Table polish — colorized counts by thresholds (green/yellow/red), selection glyph, alternating row striping
- [ ] TUI: Enqueue actions — `e` enqueue 1 to selected; `E` prompt for count (inline in Info panel)
- [ ] TUI: Right-click on Queues — peek selected (later: context menu with actions)
- [x] TUI: Keyboard shortcuts for tabs (`1`..`4` to switch)
- [ ] TUI: Persist UI state across sessions (active tab, focus panel, filter value)
- [x] TUI: Improve tiny-terminal layout — stack panels vertically; clamp widths; hide charts if extremely narrow
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
- [x] TUI: Flexbox layout via `stickers` across tabs; cell-based sizing for contents
- [ ] Admin: Requeue-from-DLQ command with count/range support (exposed to TUI)
- [ ] Admin: Workers-list admin call (IDs, last heartbeat, active item) for Workers tab
- [ ] Metrics: Optional TUI runtime metrics (ticks, RPC latency) for debugging
- [x] Docs: Add TUI design README with SVG mockups
- [ ] Docs: Update README TUI section with tabs, screenshots, and new keybindings
- [ ] Release: Add changelog entries for TUI tabbed layout and overlays

### Finished Log
- [x] Rewrite `AGENTS.md` **2025-09-13 07:18** [Link to PR #123](https://fake.com)

- [x] TUI layout revamp (flexbox + animation) **2025-09-13**
  - Integrated `github.com/76creates/stickers/flexbox` for panel layout across all tabs
  - Applied borders at cell level; sized content to cell inner dims (fixed clipped corners/edges)
  - Added Harmonica spring animation for Charts expansion (animated 1:1 → 1:2)
  - Implemented narrow-width stacking (Queues → Charts → Info) to avoid overflow
  - Compact tab bar styling with visible borders; active tab colored; added 1–4 keybindings
  - Fixed Queues table bottom-border clipping by clamping table height from cell height
  - Rendered charts using cell width; resized Info viewport from cell dims
  - Added TUI design doc with SVG mockups under `docs/TUI/`

---
## Daily Activity Logs

(maintain and use this from now on)

Please keep this document up-to-date with records of what you've worked on as you're working. When you start a task, write down what you're about to do. When you finish something, log that you've finished it. If it was an item off the backlog (see below), check it off. Build up a commit graph of the day's activity and keep it up-to-date as you make commits. Use this not only to record activity, but to capture ideas, make notes/observations/insights, and jot down bugs you don't have time to deal with in the moment.

> [!info]- ### 2025-09-13–Rewrote `AGENTS.md`
> Today we rewrote `AGENTS.md` and now it's a very useful artifact that records past activity, captures current activity, plans for future activity, and helps both AI agents and humans remember what they're doing.
> 
> ```mermaid
> gitGraph 
> 	commit 
> 	commit 
> 	branch docs/example
> 	checkout docs/example 
> 	commit 
> 	commit 
> 	checkout main 
> 	merge docs/example id: "PR #213"
> ```
> ##### 06:39 – Starting `AGENTS.md` Enhancements
> 
> - Switched to branch `docs/example`
> - Refining the `AGENTS.md` file to be the one-top-shop/HUB of governance for this project, maintained by AI agents.
> - Organized top picks from yesterday's brainstorm 
> ##### 07:23 – Finished `AGENTS.md` Enhancements
> 
> - PR open at [URL](to PR)
> - Tests added: `/path/to/tests.whatever`
> ##### 13:42 – Bug Report
>   > [!warning]- **Bug: Infinite Loop in Foo.bar**
>   > Repro steps:
>   > etc…

---
> [!info]- ### 2025-09-13 – TUI Layout Revamp (Flexbox + Animation + Docs)
> Implemented a responsive, animated TUI layout and added design docs.
>
> Changes
> - Switched panel layout to stickers flexbox across all tabs
> - Borders are drawn at cell level; content sized to inner width/height
> - Charts expansion uses Harmonica spring animation (click right to expand, left to balance)
> - Narrow terminals stack panels vertically to prevent clipping
> - Tab bar restyled (compact, bordered); added numeric tab shortcuts (1–4)
> - Fixed queues bottom border by setting table height from cell height
> - Charts now render with precise cell width; Info viewport sized from cell dims
> - Wrote `docs/TUI/README.md` and color-coded SVG mockups for all screens
>
> Follow-ups
> - Bubblezone for precise mouse hitboxes (tabs/panels/rows)
> - Add `c` keyboard toggle for expand
> - Tune stacking threshold; optionally add min-widths per cell
>

---
## APPENDIX B: WILD IDEAS — HAVE A BRAINSTORM

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

### Codex's Top Picks

High‑leverage, high‑impact items to pursue first. Keep this table updated as priorities shift.

| Idea                                  | Why                                        | First Steps                                                                 | Remarks                                                      | Difficulty | Complexity  | Wow factor  | Leverage Factor |
| ------------------------------------- | ------------------------------------------ | --------------------------------------------------------------------------- | ------------------------------------------------------------ | ---------- | ----------- | ----------- | --------------- |
| HTTP/gRPC Admin API                   | Core enabler for TUI/web/automation/RBAC   | Define API (proto/OpenAPI); wrap existing admin funcs; add basic auth       | Version the API; unlocks Workers/DLQ features and remote ops | Medium     | Medium‑High | Medium      | High            |
| DLQ Remediation UI                    | Reduces incident toil; fast, visible value | List/paginate DLQ; peek; requeue/purge; add filters/search                  | Needs admin endpoints; great demo for reliability wins       | Medium     | Medium      | High        | High            |
| Trace Drill‑down + Log Tail           | Deep observability; faster RCA             | Ensure trace IDs; link to tracing UI; basic worker log tail with filters    | Start with external trace links; mind privacy/log volume     | Medium     | Medium      | High        | Medium          |
| Interactive Policy Tuning + Simulator | Prevents outages; safe “what‑if”           | Read‑only preview; simple backlog/throughput model; dry‑run apply; rollback | Requires admin API to apply; start simulation offline        | High       | High        | High        | High            |
| Patterned Load Generator              | Validates perf; great for demos            | Add sine/burst/ramp patterns; save/load profiles; chart overlay             | Build on bench; add guardrails (limits/jitter)               | Low        | Medium      | Medium      | Medium          |
| Anomaly Radar + SLO Budget            | At‑a‑glance health; actionable signals     | Compute backlog growth, p95, failure rate; thresholds; status widget        | Define SLO; calibrate thresholds; integrate metrics          | Medium     | Medium      | Medium‑High | Medium‑High     |

---
## Appendix C: Codex Ideas in Detail

The detailed mini design specs have been moved to separate documents under `docs/ideas/`:

- HTTP/gRPC Admin API: `docs/ideas/admin-api.md`
- DLQ Remediation UI: `docs/ideas/dlq-remediation-ui.md`
- Trace Drill‑down + Log Tail: `docs/ideas/trace-drilldown-log-tail.md`
- Interactive Policy Tuning + Simulator: `docs/ideas/policy-simulator.md`
- Patterned Load Generator: `docs/ideas/patterned-load-generator.md`
- Anomaly Radar + SLO Budget: `docs/ideas/anomaly-radar-slo-budget.md`

Keep the “Codex’s Top Picks” table above in sync with these docs.
