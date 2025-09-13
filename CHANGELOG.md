# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

- Admin CLI: stats, peek, purge-dlq
- Health/readiness endpoints
- Queue length gauges updater
- Config validation
- Tracing propagation from job IDs
- Smarter rate limiting (TTL-based sleep + jitter)
- Worker active gauge
- E2E tests with Redis service in CI
- Govulncheck in CI

- TUI (Bubble Tea):
  - Initial TUI with Queues, Keys, Peek, Bench views
  - Mouse support: wheel scroll, hover highlight, left-click select, right-click peek
  - Charts view: time-series graphs for queue lengths (asciigraph)
  - Modal confirmations for purge actions with dimmed background
  - Fuzzy filter on queues view (press 'f' to filter)
  - Full-screen scrim overlay for confirmations
  - Tabbed layout: Job Queue, Workers, Dead Letter, Settings (clickable tabs with lipgloss styling)
  - Per-tab colored panel borders; Workers/DLQ/Settings placeholders
  - Help overlay now uses a full-screen scrim; ESC toggles help when not in an input/modal
  - README keybindings updated for ESC behavior
  - AGENTS.md added with working tasklist and brainstorm (WILD IDEAS)
  - Makefile targets: `build-tui`, `run-tui`
  - README: added TUI usage and screenshots placeholders
