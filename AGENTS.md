# AGENTS

- Quick notes for working on this repo (Go Redis Work Queue) 
- Things learned / want to remember when iterating fast
- Activity log
- Tasklist
- Ideas

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

Feature status is tracked in `docs/features-ledger.md`.

### Notes

- Phase II: Green the Tests. Build is stable; prioritize getting every suite passing.
- Continue documenting fixes and triage outcomes as we burn down failing tests.

Near-term TODOs I’m targeting:

- Charts expand-on-click (toggle 2/3 vs 1/3), precise mouse hitboxes (bubblezone), table polish (striping, thresholds, selection glyph), enqueue actions (`e`/`E`), right-click peek.

---

## Working Tasklist

> [!NOTE]
> Phase II — Green the Tests. Run suites first, capture failures, and feed fixes back into the backlog.


(maintain and use this from now on)

Note: Whenever you update this tasklist, please also update the features ledger document at `docs/features-ledger.md`.

Progress automation
- To refresh the overall project progress bars (in the Features Ledger and README) after editing the features table, run: `python3 scripts/update_progress.py` and commit the changes.
- The script weights features by approximate Go LOC of the linked code paths and recomputes the overall percent, updating both docs in place between `<!-- progress:begin -->` and `<!-- progress:end -->` markers.
- When adding rows, use valid repo paths in the Code column (e.g., `[internal/admin-api](../internal/admin-api)`) so LOC can be computed. If no code path, the row gets a minimum weight.

Pre-commit hook
- A pre-commit hook runs the progress update script automatically and stages `docs/features-ledger.md` and `README.md` so bars/KLoC stay current.
- Enable hooks once per clone: `make hooks` (sets `core.hooksPath=.githooks`).
- Reminder: Whenever you touch `AGENTS.md`, also ensure the features ledger is current (the hook will do this, but run the script manually if needed).

#### Updating Backlog & Features Ledger
1. Edit the `Prioritized Backlog` section here in `AGENTS.md` with the new item status/notes.
2. Mirror the change in `docs/features-ledger.md` (same feature row or add a new one) so both artifacts stay aligned.
3. Run `python3 scripts/update_progress.py` to refresh the progress bars in `docs/features-ledger.md` and `README.md`.
4. Review the script output, then stage/commit the updated files together with your backlog changes.

CI auto-update
- On merges to `main`, a GitHub Actions workflow (`.github/workflows/update-progress.yml`) runs the progress updater and commits any changes to the ledger/README automatically.
- This provides a consistent source of truth even if local hooks are bypassed.

Use this checklist to track work. Keep it prioritized, update statuses, and reference it in PRs/commits. Add new items as they surface; close them when done. This is your backlog.

### Prioritized Backlog

- [x] TUI: Charts expand-on-click (Charts 2/3 vs Queues 1/3; toggle back on Queues click)
- [ ] TUI: Integrate `bubblezone` for precise mouse hitboxes (tabs, table rows, future context menus)
- [ ] Real green: capacity planning/forecasting/policy simulator suite
- [ ] Real green: distributed tracing integration suite
- [ ] Real green: job budgeting suite
- [ ] Real green: JSON payload studio suite
- [ ] Real green: long-term archives suite
- [ ] Real green: advanced rate limiting suite
- [ ] Real green: anomaly radar / SLO budget suite
- [ ] Real green: canary deployments suite
- [ ] Real green: chaos harness suite
- [ ] Real green: multi-cluster control suite
- [ ] Real green: multi-tenant isolation suite
- [ ] Real green: smart payload deduplication suite
- [ ] Real green: smart retry strategies suite
- [ ] Real green: storage backends suite
- [ ] Real green: tenant manager suite
- [ ] Real green: trace drilldown log tail suite
- [ ] Real green: worker suite
- [ ] Real green: patterned load generator suite
- [ ] Real green: plugin panel system suite
- [ ] Real green: e2e suite
- [ ] Real green: integration suite
- [ ] Real green: forecasting suite
- [ ] Real green: queue snapshot testing suite
- [ ] Real green: policy simulator suite
- [ ] Real green: kubernetes operator suite
- [ ] Real green: time travel debugger suite

- [x] TUI: Table polish — colorized counts by thresholds (green/yellow/red), selection glyph, alternating row striping
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
- [ ] Admin: Rename ExactlyOnce handler/tests to AtLeastOnce and implement missing AtLeastOnce admin API endpoints
- [x] Docs: Add TUI design README with SVG mockups
- [ ] Docs: Update README TUI section with tabs, screenshots, and new keybindings
- [ ] Release: Add changelog entries for TUI tabbed layout and overlays
- [x] Observability: Publish Anomaly Radar OpenAPI spec + client CI automation
- [x] Observability: Finalize Anomaly Radar auth/error/pagination contract and document endpoints
- [x] Observability: Inject scopes into Anomaly Radar HTTP handlers via Admin API gateway/context plumbing
- [x] Observability: Update dashboards/clients to follow Anomaly Radar pagination cursors and surface `next_cursor`
- [x] Observability: Revisit chunk_008 rejections with enhanced OpenAPI auth/error responses and close out review items
- [ ] Ops: Share port-forward helper across deployment scripts
- [ ] DevOps: Add policy-as-code checks for security contexts and secret mounts
- [x] Docs: Audit API references to ensure they document the standardized error envelope + request IDs
- [x] Tooling: Add automated checks that validate handlers emit/log `X-Request-ID`

### Finished Log
- [x] Rewrite `AGENTS.md` **2025-09-13 07:18** [Link to PR #123](https://github.com/flyingrobots/go-redis-work-queue/pull/123)

- [x] TUI layout revamp (flexbox + animation) **2025-09-13**
  - Integrated `github.com/76creates/stickers/flexbox` for panel layout across all tabs
  - Applied borders at cell level; sized content to cell inner dims (fixed clipped corners/edges)
  - Added Harmonica spring animation for Charts expansion (animated 1:1 → 1:2)
  - Implemented narrow-width stacking (Queues → Charts → Info) to avoid overflow
  - Compact tab bar styling with visible borders; active tab colored; added 1–4 keybindings
  - Fixed Queues table bottom-border clipping by clamping table height from cell height
  - Rendered charts using cell width; resized Info viewport from cell dims
  - Added TUI design doc with SVG mockups under `docs/TUI/`

- [x] Theme Playground MVP **2025-09-14**
  - Added persistence (`internal/theme-playground/persistence.go`) and tests
  - Implemented playground + types; integrated with theme system
  - Extended docs and examples

- [x] Terminal Voice Commands MVP **2025-09-14–15** *(archived 2025-09-20)*
  - Implemented recognizer/processor/config with full test suite (module removed during cleanup)
  - Added API docs (`docs/api/terminal-voice-commands.md`) — retained for archival reference
  - Prepared for TUI integration (feedback loop and config) before feature pause

- [x] Admin API v1 (HTTP) **2025-09-15**
  - Endpoints: Stats, StatsKeys, Peek, PurgeDLQ, PurgeAll, Bench
  - Middleware chain: Auth (deny-by-default), Rate Limit, Audit, CORS, Recovery
  - OpenAPI served at `/api/v1/openapi.yaml`; integration tests green
  - TUI switchover for Stats pending

## TUI Tasks

Step-by-step task list to build the TUI up to design spec.

- [ ] `TUI001`
> [!NOTE] Launch CLI + Config Discovery
> Implement flags/env/config bootstrap for smooth first-run.
> - Add flags in `cmd/tui/main.go`: `--config`, `--redis-url`, `--cluster`, `--namespace`, `--read-only`, `--refresh`, `--metrics-addr`, `--log-level`, `--theme`, `--fps`, `--no-mouse`.
> - Read env overrides: `GRQ_*` (see docs/design/TUI2-design.md Launch section).
> - Config discovery precedence: flag > env > XDG (`~/.config/grq/config.yaml`) > `./config/config.yaml` > defaults.
> - Validate config; on error return a dedicated message to the model to show an error modal (don’t exit hard).
> - Persist last-good config path and last cluster (no secrets) under XDG data dir.

- [ ] `TUI002`
> [!NOTE] First‑Run Welcome Overlay
> Fullscreen scrim with: Quick Connect, Demo Mode, Init Config.
> - Add `internal/tui/overlays.go` view+model for Welcome; trigger when no config/URL or connection fails.
> - Inputs: Redis URL (with auth), cluster name, namespace; test button runs a lightweight `PING`.
> - Actions: `Enter` connect, `d` Demo Mode, `i` Init Config, `esc` to Help.
> - On success, dismiss and hydrate dashboard; store connection metadata in state.

- [ ] `TUI003`
> [!NOTE] Demo Mode (Seeded, Read‑only)
> Provide instant value safely.
> - Add `tui demo` subcommand and Welcome shortcut; seed queues/workers using existing `admin.Bench` with low rate and cap.
> - Mark UI as read‑only (status bar badge + theme accent); gate destructive ops.
> - Auto-clean on exit or keep ephemeral keys under `rq:demo:*` namespace.

- [ ] `TUI004`
> [!NOTE] Doctor Subcommand
> Connectivity and permissions diagnostics for support and CI.
> - Implement `tui doctor` in `cmd/tui`: DNS/TCP, TLS/mTLS, Redis PING, role (INFO), latency sample.
> - ACL probe for required commands (LLEN, XRANGE, HGETALL, etc.); print table + exit code.
> - Optional `--redis-url`/`--cluster`; respect config discovery.

- [ ] `TUI005`
> [!NOTE] Global Read‑only Mode + Guardrails
> Enforce safe defaults everywhere.
> - Add `readOnly bool` to `internal/tui/model.go`; render status bar indicator and toggle (if allowed).
> - Wrap dangerous actions: Purge, Requeue, Enqueue > confirm modal; block when read‑only.
> - `--read-only` flag overrides persisted state; cannot be disabled from UI.

- [ ] `TUI006`
> [!NOTE] Responsive Breakpoints (Mobile/Tablet/Desktop/Ultrawide)
> Finish breakpoint-aware layouts using stickers flexbox.
> - Implement thresholds (≤40, 41–80, 81–120, 121+) in `WindowSizeMsg` handler.
> - Define per-breakpoint cell grids and hide/reflow panels accordingly.
> - Adaptive tab bar: bottom (mobile), top (tablet/desktop), sidebar (ultrawide).
> - Ensure charts/tables clamp to cell inner dims; test very small widths.

- [ ] `TUI007`
> [!NOTE] Precise Mouse Hitboxes (bubblezone)
> Map clickable regions for tabs, rows, splitters, context menus.
> - Integrate `github.com/lrstanley/bubblezone` (or maintained fork) in `internal/tui/view.go`.
> - Register zones for: tab labels, tables rows, charts panel, splitter.
> - Update `MouseMsg` handling to resolve zone hits deterministically.

- [ ] `TUI008`
> [!NOTE] Table Polish (Threshold Colors, Glyphs, Striping)
> Improve readability and status signaling in Queues table.
> - Add thresholds (green/yellow/red) for backlog, latency; centralize in `internal/tui/styles.go`.
> - Selection glyph and alternating row background; clamp height to cell.
> - Ensure color accessibility in high-contrast theme.

- [ ] `TUI009`
> [!NOTE] Persist UI State
> Restore previous session: tab, focus, filters, theme, split ratio.
> - XDG data file (JSON/YAML) with `activeTab`, `focus`, `filter`, `theme`, `split`, `lastCluster`.
> - Load at start, save on change (debounced); allow `--no-state` flag to disable.

- [ ] `TUI010`
> [!NOTE] Adjustable Panel Split (Keys + Mouse)
> User-controlled left/right ratio within bounds.
> - Keys `[`/`]` adjust ratio; mouse drag on splitter via bubblezone.
> - Persist split ratio; constrain to 25–75% range; animate with Harmonica.

- [ ] `TUI011`
> [!NOTE] Enqueue Actions (`e`/`E`)
> Quick enqueue to selected queue from Job Queue tab.
> - `e` enqueue 1 with default payload; `E` opens inline form for count, payload size, jitter.
> - Use admin enqueue/bench helper; show toasts on success/failure; respect read‑only.

- [ ] `TUI012`
> [!NOTE] Right‑click Peek on Queues
> Contextual mouse shortcut for discoverability.
> - Right‑click selected row to trigger Peek; fallback long‑press on touch terminals.
> - Reuse existing peek panel; add bubblezone hit region per row.

- [ ] `TUI013`
> [!NOTE] Bench UX Enhancements
> Cancel, ETA, live throughput, payload size/jitter, concurrency.
> - Convert bench to non-blocking; ESC cancels (context cancel).
> - Compute ETA from baseline + current rate; show progress + rate sparkline.
> - Inputs for payload size, jitter %, concurrency; validate before run.

- [ ] `TUI014`
> [!NOTE] Bench Progress Baseline
> Avoid overcount when Completed list pre-populated.
> - On bench start, record `LLEN` of `cfg.Worker.CompletedList` as baseline; subtract from subsequent counts.
> - Handle overflow/reset; guard against large lists by capping scan.

- [ ] `TUI015` [blocked by Admin: Requeue-from-DLQ]
> [!NOTE] DLQ Tab (List, Peek, Requeue, Purge, Search)
> Full remediation workflow.
> - Paginate DLQ items (cursor-based); peek full payload with pretty JSON.
> - Actions: requeue selected, purge selected, bulk operations with confirm.
> - Fuzzy search/filter, sort by age/queue; respect read‑only; show counts.
>
> Unblockers (Backend API Contract)
> - Admin function: `admin.DLQList(ctx, ns string, cursor string, limit int) (items []DLQItem, next string, err error)`
>   - DLQItem: `{ID string, Queue string, Payload []byte, Reason string, Attempts int, FirstSeen time.Time, LastSeen time.Time}`
>   - Backed by Redis list/stream; stable cursor (opaque string) with upper bound on `limit` (e.g., 200)
> - Admin function: `admin.DLQRequeue(ctx, ns string, ids []string, destQueue string) (requeued int, err error)`
>   - If `destQueue==""`, requeue to original queue; idempotent on missing IDs
> - Admin function: `admin.DLQPurge(ctx, ns string, ids []string) (purged int, err error)`
> - HTTP mapping (Admin API v1):
>   - `GET /api/v1/dlq?ns=NS&cursor=C&limit=N`
>   - `POST /api/v1/dlq/requeue` `{ns, ids, destQueue}` → `{requeued}`
>   - `POST /api/v1/dlq/purge` `{ns, ids}` → `{purged}`
> - ACL: ensure endpoints honor read-only mode (reject with 403)
> - Code stubs: see `internal/admin/tui_contracts.go` (DLQList, DLQRequeue, DLQPurge)

- [ ] `TUI016` [blocked by Admin: Workers-list API]
> [!NOTE] Workers Tab (Live View)
> Worker IDs, last heartbeat, active job/queue; sort/filter.
> - Admin call for workers list (IDs, timestamps, active item); poll periodically.
> - Table with status coloring and sort; drill-in to worker details/log tail.
>
> Unblockers (Backend API Contract)
> - Admin function: `admin.Workers(ctx, ns string) ([]WorkerInfo, error)`
>   - WorkerInfo: `{ID string, LastHeartbeat time.Time, Queue string, JobID string, StartedAt *time.Time, Version string, Host string}`
>   - Consider TTL (e.g., 15s) to mark workers stale/offline
> - Optional details: `admin.Worker(ctx, ns, id string) (WorkerDetail, error)` including recent logs/metrics
> - HTTP mapping:
>   - `GET /api/v1/workers?ns=NS` → list
>   - `GET /api/v1/workers/{id}?ns=NS` → detail (optional)
> - ACL: read-only sufficient
> - Code stubs: see `internal/admin/tui_contracts.go` (Workers)

- [ ] `TUI017`
> [!NOTE] Settings Tab (Interactive)
> Theme toggle, config path, copy, open config.
> - Implement theme chooser; display current config path; add "copy value" actions.
> - Shortcut to open config in `$EDITOR` when available.

- [ ] `TUI018`
> [!NOTE] Theme System (Centralized + High Contrast)
> Consistent styling with adaptive colors and playground integration.
> - Centralize styles in `internal/tui/theme/*.go`; add dark/light/high-contrast palettes.
> - Respect `NO_COLOR` and terminal truecolor detection; expose `--theme` and UI toggle.
> - Surface Theme Playground under Settings.

- [ ] `TUI019`
> [!NOTE] Help Overlay Expansion
> List all shortcuts (tabs, enqueue, right-click), mouse hints, README link.
> - Context-aware help per tab and breakpoint; `?` or `esc` toggles.
> - Include numeric tab shortcuts (1–4) and new commands.

- [ ] `TUI020`
> [!NOTE] Mouse UX Extras
> Double-click row to peek; header click to sort (if supported).
> - Implement double-click detection with time threshold; use bubblezone for headers.
> - Sort toggles per column where data supports it; visual sort glyphs.

- [ ] `TUI021`
> [!NOTE] Non‑blocking Toasts / Status Area
> Transient error/info messages without stealing focus.
> - Top-right stack with auto-dismiss timers; queue messages; log tail in Info panel.
> - Provide API `tui.toast(level, msg)` for internal use.

- [ ] `TUI022`
> [!NOTE] Command Palette (`Ctrl+P`)
> Fuzzy action launcher with context-aware suggestions.
> - Action registry with IDs, labels, shortcuts; integrate fzf-like filter.
> - Invoke enqueue, peek, switch tabs, toggle theme, open docs, etc.

- [ ] `TUI023`
> [!NOTE] Frame‑Gated Batching (60/30fps)
> Smooth updates without over-rendering.
> - Switch `Update` to pointer receiver; implement coalescing buffer + `tea.Tick` frame messages.
> - Cap FPS via `--fps` and auto-drop to 30fps if render >12ms.
> - Enable `viewport.HighPerformanceRendering` and incremental content updates.

- [ ] `TUI024`
> [!NOTE] Cross‑Platform Compatibility
> Terminal quirks and fallbacks.
> - Windows/WSL truecolor detection; degrade gracefully; default 30fps on slow terms.
> - tmux mouse escape hatch `--no-mouse`; document recommended settings.
> - Verify selection/clipboard friendliness; avoid clearing screen unnecessarily.

- [ ] `TUI025`
> [!NOTE] TUI Runtime Metrics
> Optional instrumentation for debugging.
> - Export FPS, render time, RPC latency histograms at `:9090/metrics`.
> - Add on/off flag `--metrics`; annotate frames with dropped/merged counts.

- [ ] `TUI026`
> [!NOTE] Tests (Helpers + Layout)
> Unit tests for pure helpers; snapshot-ish rendering checks.
> - Test filtering, formatting, thresholds, clamp functions.
> - Add layout tests that render at key widths and assert presence of key strings.

- [ ] `TUI027`
> [!NOTE] Docs Update (README TUI)
> Screenshots, tabs description, new keybindings, launch flags.
> - Update README TUI section; include SVG mockups; cross-link design doc.
> - Add short “Quickstart” with Demo Mode and Doctor.

- [ ] `TUI028`
> [!NOTE] Release Notes
> Changelog entries for tabbed layout, overlays, and new UX.
> - Summarize user-visible changes; include safety guardrails and flags.

- [ ] `TUI029`
> [!NOTE] Multi‑Cluster Foundation (Ultrawide)
> Basic cluster selector + compare view scaffold.
> - Add cluster switcher UI; persist recent clusters; wire to config.
> - In ultrawide, render two clusters side-by-side (read-only compare to start).

- [ ] `TUI030` [blocked by Admin: Job events stream]
> [!NOTE] Time Travel Debugger (MVP)
> Minimal timeline + state viewer for a single job.
> - Open by ID; fetch event stream; scrub timeline with left/right.
> - Show state deltas; no step-in code view yet; export link placeholder.
>
> Unblockers (Backend API Contract)
> - Admin function: `admin.JobTimeline(ctx, ns, jobID string, start, end *time.Time, limit int) ([]JobEvent, error)`
>   - JobEvent: `{TS time.Time, Type string, Data map[string]any}` with canonical types: enqueued, dequeued, started, heartbeat, completed, failed, retried, moved_to_dlq, requeued
>   - Ordering ascending; `limit` cap (e.g., 1000); filterable by time range
> - Optional streaming: `admin.SubscribeJob(ctx, ns, jobID string) (<-chan JobEvent, func(), error)` for live updates
> - HTTP mapping:
>   - `GET /api/v1/jobs/{id}/timeline?ns=NS&start=&end=&limit=`
>   - `GET /api/v1/jobs/{id}/events` (SSE/WebSocket) for live follow (later)
> - Storage: Redis Streams per job or namespaced global stream with jobID index
> - ACL: read-only sufficient
> - Code stubs: see `internal/admin/tui_contracts.go` (JobTimeline, SubscribeJob)

- [ ] `TUI031`
> [!NOTE] Voice Command Integration (MVP)
> Hook Terminal Voice Commands to simple actions.
> - Map phrases to palette actions (peek queue, switch tab, run bench).
> - Provide visual feedback toast when recognized; allow disable in Settings.

- [ ] `TUI032`
> [!NOTE] Error & Offline Degrade
> Friendly failure modes, no dead ends.
> - On connection loss, show non-blocking banner + retry loop with backoff.
> - Offer "Edit connection" and "Switch to Demo" inline; keep UI responsive.

### TUI Task Chains

Sequenced dependencies across tasks. Use these to plan workstreams; items in parentheses are non-TUI prerequisites from the backlog.

- Launch & First‑Run
  - TUI001 → TUI002 → TUI003 → TUI005 (with TUI004 parallel/optional)

- Stats Data Source & Performance
  - (Admin API v1 Stats/StatsKeys switchover) → TUI023 → TUI025

- Responsive Layout & Interaction Polish
  - TUI006 → TUI010 → TUI008 → TUI019

- Mouse/Hitboxes & Contextual Actions
  - TUI007 → TUI020 → TUI012

- State, Theme, Settings
  - TUI001 → TUI009 → TUI018 → TUI017

- Queue Operations UX
  - TUI011 → TUI021 → TUI022

- Bench Improvements
  - TUI013 → TUI014

- DLQ Remediation
  - (Admin: Requeue‑from‑DLQ) → TUI015

- Workers Live View
  - (Admin: Workers‑list API) → TUI016

- Multi‑Cluster (Ultrawide foundation)
  - TUI001 → TUI029

- Advanced Features
  - TUI022 → TUI031
  - (Admin: Job events stream) → TUI030

- Docs & Release
  - TUI027 → TUI028

### TUI Parallelization & Priorities

Suggested execution order with safe parallel tracks. Parentheses note external prerequisites.

**Start‑Here Priority (Suggested Order)**
- P1 — Foundation: TUI001 (Launch/Config), TUI002 (Welcome), TUI004 (Doctor, parallel), TUI003 (Demo), TUI005 (Read‑only)
- P2 — Performance Baseline: TUI023 (Frame‑gated batching), TUI025 (Runtime metrics, optional)
- P3 — Responsive Layout: TUI006 (Breakpoints/adaptive tabs)
- P4 — State & Theme & Settings: TUI009 (Persist UI), TUI018 (Theme System), TUI017 (Settings Tab)
- P5 — Interaction Polish: TUI007 (Hitboxes), TUI010 (Split ratio), TUI008 (Table polish), TUI019 (Help), TUI021 (Toasts), TUI020 (Mouse extras)
- P6 — Core Queue Ops: TUI011 (Enqueue actions), TUI022 (Command palette)
- P7 — Bench Improvements: TUI013 (Bench UX), TUI014 (Bench baseline)
- P8 — Error/Offline: TUI032 (Degrade & recovery)
- P9 — DLQ & Workers: TUI015 (DLQ Remediation) [requires Admin Requeue], TUI016 (Workers View) [requires Admin Workers‑list]
- P10 — Advanced Features: TUI029 (Multi‑cluster), TUI030 (Time Travel) [(Admin: job events stream)], TUI031 (Voice)
- P11 — Docs & Release: TUI027 (Docs Update), TUI028 (Release Notes)

**Parallelization Groups**
- Group A — Boot Flow: TUI001, TUI002, TUI004 (small team can run these in parallel; coordinate shared config/types)
- Group B — Safety/Demo: TUI003, TUI005 (can proceed after TUI001)
- Group C — Performance: TUI023, TUI025 (unblocks UI smoothness; safe to start after skeleton renders)
- Group D — Layout & State: TUI006, TUI009, TUI010 (shared view/layout code; sync on sizes/ratios)
- Group E — Theme & Settings: TUI018, TUI017 (depends lightly on TUI009 for persistence hooks)
- Group F — Mouse & Polish: TUI007, TUI008, TUI019, TUI021, TUI020 (independent polish tasks; coordinate on keymaps)
- Group G — Queue Ops: TUI011, TUI022 (touches actions registry and overlays; align UX)
- Group H — Bench: TUI013, TUI014 (bench pipeline; verify telemetry)
- Group I — DLQ/Workers: TUI015, TUI016 [(Admin endpoints required)]
- Group J — Advanced: TUI029, TUI030, TUI031 [(Admin job events for TUI030)]
- Group K — Error/Offline: TUI032 (can be developed earlier if environment is flaky)
- Group L — Docs/Release: TUI027, TUI028 (finalization; can draft alongside feature work)

Notes
- If switching TUI data source to Admin API v1 for stats, perform that switchover before P2 for realistic performance testing.
- Where tasks overlap on input handling and overlays, designate a single owner to avoid keybinding conflicts.

- [x] Storage Backends — core adapters **2025-09**
  - Core adapters implemented with tests; API docs present
  - Conformance matrix and migration docs ongoing

- [x] RBAC & Tokens (v1; hardened) **2025-09-14**
  - JWT manager + middleware; revoke/cache; security fixes merged
  - Add’l e2e coverage planned

---
## Daily Activity Logs
> [!NOTE]
> ### 2025-09-20 – `make clean` Fix
> Resolved permission denials when cleaning the repo.
>
> Changes
> - Added a pre-removal `chmod` step in `Makefile` so `.gocache` modules become writable before deletion.
>
> Follow-ups
> - None — `make clean` now succeeds even with read-only module cache entries.

> [!NOTE]
> ### 2025-09-20 – Request ID Linter Docs
> Added documentation so teammates can discover and run the internal analyzer.
>
> Changes
> - Wrote `tools/requestidlint/README.md` covering purpose, invocation, and tests.
>
> Follow-ups
> - Consider wiring the analyzer back into a `make lint` target or CI job.

> [!NOTE]
> ### 2025-09-20 – `test/` Map-of-Contents
> Documented every test and script under `test/` so the directory is navigable.
>
> Changes
> - Replaced `test/README.md` with a MoC summarizing purpose, execution notes, quality, and dependencies for each file.
>
> Follow-ups
> - Clean up unused fixtures and consider parameterizing shell scripts that hard-code absolute paths.

> [!NOTE]
> ### 2025-09-20 – `test/` Cleanup
> Removed dead scaffolding so only real integration/E2E suites remain.
>
> Changes
> - Deleted unused synthetic tests, fixtures, and stray macOS metadata from `test/`.
> - Moved the package-specific integration suites next to their source (`internal/exactly_once`, `internal/multi-cluster-control`, `internal/storage-backends`, `internal/obs`).
> - Relocated the Event Hooks test plan to `docs/testing/event-hooks-test-plan.md` and refreshed `test/README.md` to match the slimmer directory.
>
> Follow-ups
> - Parameterize the acceptance shell scripts (they still assume `/Users/james/...`).

> [!NOTE]
> ### 2025-09-20 – Phase II Kickoff (Green the Tests)
> Lifted the testing freeze now that builds are stable.
>
> Changes
> - Updated project notes to emphasize Phase II goal (test pass) instead of the previous build-only stance.
> - Ran `make test` (race-enabled `go test ./...`) to get a fresh failure inventory.
>
> Follow-ups
> - `internal/worker-fleet-controls` suite currently fails: data race in `WorkerController_PauseResumeWorkers`, safety checker assertions expecting different error text, signal ordering assumptions, and noisy Redis signal handler teardown. Need to triage and add to backlog.

> [!NOTE]
> ### 2025-09-18 – GREEN MACHINE progress
> Focused on making the build green module-by-module.
>
> Changes
> - Documented the internal module dependency DAG and stabilization order in AGENTS.md/README.
> - Brought `internal/storage-backends`, `internal/trace-drilldown-log-tail`, and `cmd/tui` back to a clean `go build`; experimental TUI view now lives behind `tui_experimental`.
> - Recorded per-module build status READMEs so breakages and fixes stay visible.
>
> Follow-ups
> - JSON payload studio, job budgeting, kubernetes operator, long-term archives, collaborative session, worker fleet controls, and calendar view now build; next target is `cmd/job-queue-system`.
> - Keep updating the dependency map as we clean additional modules.
>

(maintain and use this from now on)

Please keep this document up-to-date with records of what you've worked on as you're working. When you start a task, write down what you're about to do. When you finish something, log that you've finished it. If it was an item off the backlog (see below), check it off. Build up a commit graph of the day's activity and keep it up-to-date as you make commits. Use this not only to record activity, but to capture ideas, make notes/observations/insights, and jot down bugs you don't have time to deal with in the moment.

> [!NOTE]
> ### 2025-09-13–Rewrote `AGENTS.md`
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
> - PR open at [PR #213](https://github.com/flyingrobots/go-redis-work-queue/pull/213)
> - Tests added: `/path/to/tests.whatever`
> ##### 13:42 – Bug Report
>   > [!WARNING] **Bug: Infinite Loop in Foo.bar**
>   > Repro steps:
>   > etc…

---
> [!NOTE]
> ### 2025-09-13 – TUI Layout Revamp (Flexbox + Animation + Docs)
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

> [!NOTE]
> ### 2025-09-15 – 24h Commit Review (Admin API hardening, TVC, Theme, Patterned Load)
> Summary of changes over the last 24 hours across core modules.
>
> Highlights
> - Admin API: hardened handlers/server; exactly-once integration; tests fixed (handlers, server, integration)
> - Terminal Voice Commands: added core module and full test suite; docs added under docs/api/
> - Theme Playground: added persistence/playground with tests; MVP complete
> - Patterned Load Generator: added handlers and generator; tests extended
> - Multi-Cluster Control: test suite reorganized; added types_basic tests; e2e/integration updated
> - Kubernetes Operator: controllers/main refined; integration tests updated; examples extended
> - RBAC & Tokens: critical security fixes merged; middleware tests updated
>
> Follow-ups
> - Wire TUI to Admin API for stats/ops where applicable
> - Add runtime Admin API for rate-limits (advanced RL); surface in TUI
> - Add DLQ pagination + filters; extend tests


> [!NOTE]
> ### 2025-09-15 – TUI Launch UX, Admin DLQ/Workers, Features Ledger
> Summary of today’s changes to align TUI with the redesign and extend Admin API.
>
> Changes
> - TUI2 design: added "Launch & First‑Run UX" section with flags/env mapping, config discovery, welcome wizard, error/offline flows, and guardrails (`docs/design/TUI2-design.md`).
> - AGENTS: added “TUI Tasks”, “TUI Task Chains”, and “TUI Parallelization & Priorities”; marked blockers and documented backend contracts.
> - Admin (library): added DLQ/Workers contracts and implementations (`internal/admin/tui_contracts.go`) — `DLQList`, `DLQRequeue`, `DLQPurge`, `Workers`.
> - Admin API (service): added handlers + OpenAPI + RBAC for DLQ list/requeue/purge and workers list (`internal/admin-api/{handlers.go,server.go,openapi.go,types.go}`, `internal/rbac-and-tokens/config.go`).
> - Endpoints now available: GET `/api/v1/dlq`, POST `/api/v1/dlq/requeue`, POST `/api/v1/dlq/purge`, GET `/api/v1/workers`.
> - TUI CLI: scaffolded flags `--redis-url`, `--cluster`, `--namespace`, `--read-only`, `--metrics-addr`, `--log-level`, `--theme`, `--fps`, `--no-mouse` (`cmd/tui/main.go`).
> - Docs: created Features Ledger tracking progress, domain tables, drift, and ported TUI task list (`docs/features-ledger.md`).
>
> Validation
> - Built `./internal/admin/...` and `./internal/admin-api/...` successfully; noted unrelated repo-wide build issues and left them unchanged.
>
> Follow-ups
> - Wire TUI DLQ and Workers tabs to new Admin API endpoints; decide DLQ requeue destination semantics.
> - Unify Redis client versions (v8→v9) or keep TUI over HTTP Admin API to avoid coupling.
> - Implement Job Timeline API and TUI Time Travel view.
> - Add integration tests for the new endpoints.

> [!NOTE]
> ### 2025-09-16 – PR#3 Review Chunk 005
> - Completed CodeRabbit chunk_005 (30/30 items) covering compose secrets, admin API manifests, PRD clarifications, and API/README updates.
> - Added `scripts/check_yaml_newlines.py` plus `make lint` to enforce YAML trailing newlines; pinned admin API deployment image and externalized JWT secret.
> - Logged dispositions in the chunk file and updated the progress bar to 100%.
> - Lessons: automate format safety (lint scripts, pinned images) instead of relying on manual edits; update review artifacts immediately to keep evidence aligned with commits.

> [!NOTE]
> ### 2025-09-16 – PR#3 Review Chunk 006
> - Addressed CodeRabbit chunk_006 (30/30 items) touching Docker/Kubernetes manifests, deployment scripts, and docs.
> - Introduced `deployments/scripts/lib/logging.sh`, parameterized Alertmanager SMTP settings, and derived runtime checks from live manifests.
> - Extracted the feature palette into `docs/colors.yml` and standardized template formatting; chunk log now reads 100%.
> - Summary: Compose/K8s manifests are now deterministic, RBAC catalog matches Admin API, and staging tooling (tests + monitoring scripts) fail fast on missing prerequisites.
> - Lessons: keep scripting helpers centralized, source production credentials from env/secret inputs, treat design tokens as structured data, and wrap optional integrations (SMTP) behind feature flags for safer defaults.
> - Follow-up: Alertmanager emails now gated by `ENABLE_ALERTMANAGER_SMTP`; defaults to webhook-only routing when disabled.


> [!NOTE]
> ### 2025-09-16 – CodeRabbit PR#3 chunk_007 sweep
> - Cleared CodeRabbit chunk_007 (30/30 items) with per-item commits and refreshed the chunk progress bar to 100%.
> - Hardened admin API docs/config (dedicated confirmation phrases, CORS guidance), refreshed webhook signing/idempotency guidance, and aligned RBAC monitoring with real metrics.
> - Standardized event-hooks testing docs around package-aware `go test` patterns and made the postmortem coordinator task derive dependencies dynamically.
> - Lessons: keep documentation commands module-aware (no filename globs) and generate orchestration dependencies from shared data instead of static lists.

> [!NOTE]
> ### 2025-09-16 – CodeRabbit PR#3 chunk_008 sweep
> - Closed CodeRabbit chunk_008 (30/30) with a bench payload-size flag, deterministic producer fixtures, and updated performance baseline guidance.
> - Hardened admin API build artifacts (trimpath, VERSION ldflags, non-root images) and expanded anomaly radar docs with versioning policy, metrics collector interface, units, and idempotent exporter examples.
> - Logged outstanding work (OpenAPI spec, full auth/error policy, pagination) as follow-ups after the API stabilises.
> - Lessons: keep docs wired to real entrypoints, prefer idempotent Prometheus patterns, and centralise policy sections for reuse across APIs.


> [!NOTE]
> ### 2025-09-16 – Redis v9 migration & chunk 004 wrap-up
> 
> Changes
> - Migrated the entire repository (and helper modules) to `github.com/redis/go-redis/v9`; removed v8 imports in favour of the consolidated client.
> - Hardened `cmd/tui` flag handling and plumbed runtime options (cluster/namespace/read-only/theme/FPS/metrics) into the TUI with fail-fast Redis health checks.
> - Standardised admin-api namespaces/docs and tightened deployment scripts (strict bash, quoting, docker compose pre-checks, JWT secret enforcement).
> - Finished PR#3 chunk_004 worksheet (30/30 items) and refreshed the progress bar.
> 
> Follow-ups
> - Run `go test ./...` once the existing suite failures (forecasting, exactly-once outbox, etc.) are resolved upstream.

> [!NOTE]
> ### 2025-09-17 – Anomaly Radar Scope Guardrails & Pagination UX
> Scoped the anomaly radar HTTP surface, made cursor pagination first-class, and refreshed docs/specs to match.
>
> Changes
> - Added scope-aware HTTP helpers, cursor utilities, and idempotent start/stop responses under `internal/anomaly-radar-slo-budget/`.
> - Updated handlers/tests to enforce scope checks, default/max pagination, and the standard JSON error envelope; new docs outline auth, error policy, and pagination flow.
> - Published `docs/api/anomaly-radar-openapi.yaml` with CI validation plus contract notes in `docs/design/anomaly-radar-api-contract.md`.
> - Exported a public wrapper at `pkg/anomaly-radar-slo-budget/` and refreshed the docs/OpenAPI spec with scope tables, error envelope notes, and a paginated metrics example.
>
> Important Learnings
> - Centralised error helpers keep CLI/UI clients aligned once scopes are enforced—no divergent envelopes during failures.
> - Treat cursors as opaque tokens in tests to avoid brittle assumptions; golden fixtures now flex with redis-backed pagination.
> - Writing the OpenAPI spec early flushes review gaps (auth scopes, error schema) before client integration work starts.
> - Stabilising the import path via a thin wrapper lets docs and clients converge without exposing internal packages prematurely.
>
> Next Steps
> - Implement SLO budget calculations/visuals so the TUI widget has real data to render.
> - Wire the TUI/Admin API integration to consume the new paginated endpoints and surface scope errors.
> - Expand integration coverage for the gateway scope propagation once end-to-end plumbing completes.

> [!NOTE]
> ### 2025-09-17 – PR#3 Review Chunk 009
> Cleared the next CodeRabbit batch (30/30) and hardened docs + tooling along the way.
>
> Changes
> - DLQ API docs now spell out auth scopes, server-enforced limits, purge-all schema/idempotency, and valid JSON responses (`docs/api/dlq-remediation-ui.md`).
> - Added public wrappers for anomaly radar and chaos harness packages (`pkg/anomaly-radar-slo-budget/`, `pkg/chaos-harness/`) and updated docs accordingly.
> - Improved auxiliary scripts (append_metadata, review task generators) with safer I/O, UTC timestamps, and accurate logging; trimmed slow sleeps from `demos/responsive-tui.tape`.
> - Hardened Redis deployment manifest and Docker image health checks; updated BUGS.md guidance for heartbeats/schedulers/ledgers.
>
> Important Learnings
> - Documentation needs to carry real contracts (auth, rate limits, JSON schemas) so downstream tooling stays in sync.
> - Wrapper packages are a low-friction way to expose internal modules without destabilising the tree.
> - Investing in script hygiene (UTC times, error handling) prevents subtle drift when other automation depends on them.
>
> Follow-ups
> - Chunk 010 (remaining CodeRabbit feedback) still open; expect similar breadth across docs + tooling.
> - Continue chipping away at backlog items once CodeRabbit sequence is complete (ExactlyOnce→AtLeastOnce rename already queued).

> [!NOTE]
> ### 2025-09-17 – CodeRabbit PR#3 chunk_010 sweep
> Closed the final CodeRabbit batch with deployment hardening, documentation polish, and secret handling fixes.
>
> Changes
> - Hardened admin API and RBAC Kubernetes manifests: pod/container security contexts, RuntimeDefault seccomp, disabled SA token mounts, distinct health/readiness probes, and corrected Grafana compose mounts.
> - Shifted the RBAC token service to file-backed secrets plus RS256 keys, updated token-service config/startup validation, and refreshed deployment docs (`deployments/docker/rbac-configs/token-service.yaml`, `deployments/README-RBAC-Deployment.md`).
> - Clarified key docs (release plan freeze policy, purge reason validation, DLQ pipeline guardrails, HTTPS defaults, hit_percent rename) and converted `AGENTS.md` TOC to standard anchors.
> - Completed the chunk_010 worksheet with accepted dispositions and a 100% progress bar (`docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md`).
>
> Validation
> - Updated shell scripts (`deploy-staging.sh`, `health-check-rbac.sh`, `setup-monitoring.sh`) to manage port-forward PIDs safely; existing Go test failures remain pre-existing and were not re-run.
>
> Follow-ups
> - Propagate the new port-forward helpers to other deployment scripts.
> - Add policy-as-code checks to enforce secret volume usage and security context drift.

> [!NOTE]
> ### 2025-09-17 – CodeRabbit PR#3 chunk_011 sweep
> Closed the remaining CodeRabbit review items with documentation polish and onboarding fixes.
>
> Changes
> - Standardized DLQ pipeline error envelopes (codes + request IDs), clarified rate-limit headers, and documented cursor pagination.
> - Updated DLQ UI purge-all example to the safe JSON POST form with idempotency and restructured the claude-008 reflection with front matter.
> - Added a `go mod download` preflight step to README so first-time TUI users fetch dependencies before running.
>
> Follow-ups
> - Verify other API docs reference the shared error envelope pattern.
> - Consider adding automated checks for missing `X-Request-ID` logging in new handlers.
>
> Important Learnings
> - Shared error envelope docs prevent API drift—keeping the pattern centralized avoids per-endpoint divergence.
> - Adding dependency preflight steps in README shortens new contributor setup loops and avoids common module errors.

> [!NOTE]
> ### 2025-09-17 – Error envelope harmonization
> Brought the remaining Admin API docs and handlers into alignment with the standardized error envelope.
>
> Changes
> - Updated Admin API, DLQ UI, Worker Fleet, Multi-tenant, and Exactly-Once docs to show `code`, `status`, `request_id`, and `timestamp` in error examples.
> - Extended `writeError` to propagate/generate request IDs and added tests validating the enriched envelope.
>
> Validation
> - `go test ./internal/admin-api` still hits pre-existing ExactlyOnce handler test failures (missing `PurgeDLQ`/`RunBenchmark`); no new regressions introduced.
>
> Important Learnings
> - Docs drift quickly when multiple teams own surfaces—central helpers/tests anchor expectations.
> - Request ID coverage is easiest to enforce at the helper level; tests give immediate feedback when handlers regress.


## Module Dependency Map
- Captured on 2025-09-18
- Keep this section updated as dependencies shift

### Dependency Graph (module → direct internal deps)
- admin → config, distributed-tracing-integration
- admin-api → admin, anomaly-radar-slo-budget, config, exactly-once-patterns, exactly_once
- anomaly-radar-slo-budget → (none)
- automatic-capacity-planning → (none)
- breaker → (none)
- calendar-view → (none)
- canary-deployments → (none)
- chaos-harness → (none)
- collaborative-session → (none)
- config → (none)
- distributed-tracing-integration → config
- dlq-remediation-pipeline → (none)
- event-hooks → (none)
- exactly-once-patterns → (none)
- exactly_once → (none)
- forecasting → (none)
- job-budgeting → (none)
- job-genealogy-navigator → (none)
- json-payload-studio → (none)
- kubernetes-operator → (none)
- long-term-archives → (none)
- multi-cluster-control → admin, config
- multi-tenant-isolation → (none)
- obs → config, queue
- patterned-load-generator → (none)
- plugin-panel-system → (none)
- policy-simulator → (none)
- producer → config, obs, queue
- producer-backpressure → (none)
- queue → (none)
- queue-snapshot-testing → (none)
- rbac-and-tokens → (none)
- reaper → config, obs, queue
- redisclient → config
- right-click-context-menus → (none)
- smart-payload-deduplication → (none)
- smart-retry-strategies → (none)
- storage-backends → (none)
- tenant → (none)
- terminal-voice-commands → (archived 2025-09-20)
- theme-playground → (none)
- time-travel-debugger → (none)
- trace-drilldown-log-tail → admin, config, distributed-tracing-integration
- tui → admin, config
- visual-dag-builder → (none)
- worker → breaker, config, obs, queue
- worker-fleet-controls → (none)

### Suggested Stabilization Order
advanced-rate-limiting → anomaly-radar-slo-budget → automatic-capacity-planning → breaker → calendar-view → canary-deployments → chaos-harness → collaborative-session → config → dlq-remediation-pipeline → event-hooks → exactly-once-patterns → exactly_once → forecasting → job-budgeting → job-genealogy-navigator → json-payload-studio → kubernetes-operator → long-term-archives → multi-tenant-isolation → patterned-load-generator → plugin-panel-system → policy-simulator → producer-backpressure → queue → queue-snapshot-testing → rbac-and-tokens → right-click-context-menus → smart-payload-deduplication → smart-retry-strategies → storage-backends → tenant → theme-playground → time-travel-debugger → visual-dag-builder → worker-fleet-controls → distributed-tracing-integration → redisclient → obs → admin → producer → reaper → worker → admin-api → multi-cluster-control → trace-drilldown-log-tail → tui
