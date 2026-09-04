# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added

- General-purpose Redis queue primitives with byte-exact payloads, optional
  payload schemas, a 1 MiB default payload limit, pluggable application
  handlers, retry-to-DLQ behavior, lease heartbeats, and stalled-job recovery.
- Public `pkg/queueclient` producer API and `grq enqueue` CLI command, backed by
  the same queue configuration and validation used by workers.
- Authenticated Admin API enqueue support with bounded request bodies and
  atomic batch admission.
- Optional per-key FIFO execution with an O(1) ready ring, active-key
  deduplication, expiring leases, renewal, recovery, and terminal transitions.
- Admin CLI commands for `stats`, `peek`, and `purge-dlq`.
- Health and readiness HTTP endpoints.
- Queue length and worker activity gauges exported through Prometheus.
- Strict configuration validation on startup.
- Tracing propagation from job IDs into spans.
- `govulncheck` execution and Redis-backed end-to-end coverage in CI.

### Changed

- Rate limiting sleeps using TTL and jitter for fairer admission.
- Queue aliases are preserved literally and Redis glob metacharacters are
  escaped when discovering namespaced keys.
- Runtime dependency versions were refreshed to clear all reachable Go
  vulnerability findings.

### Fixed

- Queue enqueue batches are all-or-nothing when any job is invalid.
- Ordered queue transitions validate every Redis key type and reject aliased
  control keys before mutating queue state.
- Ordered DLQ requeue and stalled-job recovery preserve per-key FIFO ordering.
- Queue statistics include ordered backlog, reject terminal priority aliases,
  and deduplicate worker heartbeat keys across Redis `SCAN` pages.
- Public retry counts are range-checked before conversion to the internal
  representation.

### TUI

- Introduced the Bubble Tea TUI with Queue, Workers, Dead Letter, and Settings
  tabs, colored borders, and keyboard shortcuts `1` through `4`.
- Added a flexbox-based responsive layout, narrow-terminal stacking, clamped
  panel dimensions, and animated queue/chart split expansion.
- Added mouse support for scrolling, hover, selection, queue peeking, and chart
  expansion; wheel movement now keeps the selected row decoration current.
- Added queue-length charts, benchmark controls and progress, queue filtering,
  payload peeking, and read-only settings summaries.
- Wrapped destructive actions in confirmation modals with a full-screen dimmed
  scrim, and made Escape handling modal-, input-, and filter-aware.
- Added threshold-colored queue counts, alternating row stripes, and a visible
  selection glyph.
- Added `build-tui` and `run-tui` Makefile targets plus TUI design documentation
  and SVG mockups.

[request_verification]: Replace provisional change descriptions with release
  references and a version/date when cutting the release.
