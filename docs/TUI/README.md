# TUI Design and Layout

This document describes the intended user experience and layout for the Bubble Tea TUI. It includes color-coded SVG mockups of each screen to help visualize the structure and resizing behavior.

- Top to bottom structure:
  - Tab Bar
  - Header (title), Subheader (live stats)
  - Body (flexbox layout)
  - Status Bar

The Body is composed with a flexbox-like grid (stickers/flexbox) that stretches and squishes with the terminal size. Borders are applied at the cell level, and content width/height are sized to the cell’s inner dimensions to avoid clipping.

## Job Queue — Balanced

![Job Queue (Balanced)](images/job-balanced.svg)

- Left: Queues table (with optional filter row)
- Right: Charts (time-series per key)
- Bottom: Info panel (summaries, peek results, bench form, progress)

## Job Queue — Charts Expanded

Clicking in the right (Charts) region animates the top row split via a spring easing (Harmonica), expanding Charts from 1:1 to roughly 1:2. Clicking the left side returns to balanced.

![Job Queue (Expanded)](images/job-expanded.svg)

## Workers

![Workers](images/workers.svg)

- Placeholder for now: summary of heartbeats and processing lists; later a table of workers with sort/filter, and details pane.

## Dead Letter Queue (DLQ)

![Dead Letter Queue](images/dlq.svg)

- Placeholder for now: count and key; later a paginated list with actions (peek/requeue/purge) and search.

## Settings

![Settings](images/settings.svg)

- Read-only snapshot of key configuration values (Redis, queues, defaults). Future: theme toggle and shortcuts.

## Overlays — Confirm

![Confirm Overlay](images/overlay-confirm.svg)

- Full-screen scrim dims background; confirm modal centered. ESC cancels; Y/Enter confirms.

## Overlays — Help

![Help Overlay](images/overlay-help.svg)

- Full-screen scrim dims background; help centered; ESC closes.

---

## Layout Notes

- Flex ratios (top row): left:right transitions between 1:1 (balanced) and 1:2 (expanded).
- Gutters: a fixed-width gutter separates left and right panels for readability.
- Borders are applied to the flex cells; content (table, charts, info viewport) is sized to the cell’s inner width/height to avoid clipped corners/edges.
- Tiny terminals: panels clamp to minimum inner widths/heights; filter row and other content may reduce available table height; we compute table height dynamically as (inner cell height − title − filter lines).

## Tab Bar

- Four tabs: Job Queue, Workers, Dead Letter, Settings.
- Click to switch; numbers 1–4 (future) will also switch.
- Styling kept compact to preserve horizontal space; we avoid heavy borders on the tab bar.

## Future Enhancements

- Bubblezone for precise mouse hitboxes (tabs, panels, rows)
- Keyboard toggle for expand (e.g., `c`)
- Small terminal fallback: stack panels vertically when the width drops below a threshold; hide Charts if extremely narrow.

