# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added

- Admin CLI commands for `stats`, `peek`, and `purge-dlq` ([#PR?])
- Health and readiness HTTP endpoints ([#PR?])
- Queue length gauge updater to surface backlog metrics ([#PR?])
- Strict configuration validation on startup ([#PR?])
- Tracing propagation from job IDs into spans ([#PR?])
- Worker activity gauge exported via Prometheus ([#PR?])
- `govulncheck` execution in CI ([#PR?])
- E2E test coverage against the Redis service within CI ([#PR?])

### Changed

- Smarter rate limiting that sleeps using TTL and jitter for fairness ([#PR?])

### TUI

- Introduced the Bubble Tea TUI featuring Queues, Keys, Peek, and Bench views ([#PR?])
- Added mouse support for scroll, hover, selection, and context peek actions ([#PR?])
- Delivered Charts view with queue length time-series rendering via asciigraph ([#PR?])
- Wrapped destructive actions in modal confirmations with a dimmed scrim overlay ([#PR?])
- Enabled fuzzy queue filtering (press `f`) with ESC to clear ([#PR?])
- Applied full-screen scrim overlay for confirmations and help ([#PR?])
- Built tabbed layout (Job Queue, Workers, Dead Letter, Settings) with lipgloss styling ([#PR?])
- Highlighted per-tab colored borders and placeholder content for future expansion ([#PR?])
- Refined help overlay behavior to prioritise ESC handling outside modals/inputs ([#PR?])
- Updated README keybindings to reflect new ESC behavior ([#PR?])
- Added Makefile targets `build-tui` and `run-tui` for local workflows ([#PR?])
- Documented the TUI usage patterns and screenshots placeholders in README ([#PR?])
- Created `AGENTS.md` with the working tasklist and brainstorming backlog ([#PR?])

[request_verification]: Replace placeholder PR numbers with actual references post-merge.
