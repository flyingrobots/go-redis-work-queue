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

