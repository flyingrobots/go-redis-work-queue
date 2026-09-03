# Developer Tools and Automation

This repo includes a few helper tools and automations to keep project metadata and reviews tidy.

## Request ID Enforcement Check

A lightweight static analysis pass (`tools/requestidlint`) keeps HTTP handlers on the happy path so every error response emits/logs the `X-Request-ID` from the shared error envelope.

- **Analyzer coverage**: Flags direct calls to `http.Error` or `ResponseWriter.WriteHeader` (for status codes ≥400) inside `internal/admin-api` packages unless they go through the shared helpers. This nudges new handlers toward `writeError`/`writeErrorWithDetails`.
- **Usage**: `go run ./tools/requestidlint/cmd/requestidlint ./internal/admin-api` (the CI / pre-commit hook can pass multiple packages).
- **Tests**: `go test ./tools/requestidlint/...` exercises analyzer fixtures, and `go test ./internal/admin-api` now executes the analyzer against real sources via `requestidlint_test.go`.
- **Integration**: `make lint` and the pre-commit hook both run `go run ./tools/requestidlint/cmd/requestidlint ./internal/admin-api`, so commits/CI fail fast if handlers bypass the helpers.

## Features Ledger + Progress Automation

- Canonical ledger: `docs/features-ledger.md`
- Overall progress bar also mirrored in `README.md`

Status model (with emoji mapping)

- 📋 Planned → ⏳ In Progress → 🚼 MVP → 🅰️ Alpha → 🅱️ Beta → ✅ V1
- Definitions are documented near the top of `docs/features-ledger.md`.

Weighted progress (how it’s computed)

- Weight per feature ≈ `1 + log10(KLoC + 10)/3` (min 0.5 if no code path)
- KLoC is computed from the directories linked in the Code column
- Progress is a weighted average across features and per-group tables

Updater script

- Path: `scripts/update_progress.py`
- What it does:
  - Recomputes weighted overall progress and per-group progress
  - Updates the fixed-width progress bars in both the ledger and README
  - Ensures a `KLoC (approx)` column exists and fills it per row
  - Normalizes Emoji to match the Status column (📋/⏳/🚼/🅰️/🅱️/✅)
- Usage:
  - `python3 scripts/update_progress.py`
  - The script uses the `<!-- progress:begin --> ... <!-- progress:end -->` markers to replace bars

Local pre-commit hook

- A pre-commit hook runs `update_progress.py` and stages the ledger/README changes
- Enable once per clone: `make hooks` (sets `core.hooksPath=.githooks`)

CI auto-update

- Workflow: `.github/workflows/update-progress.yml`
- On merges to `main`, CI runs the updater and commits any changes (skip-ci tagged)

Authoring tips

- Always put valid repo links in the Code column (e.g., `[internal/tui](../internal/tui)`) so KLoC can be computed
- If a feature needs a manual weight tweak, you can add an HTML comment in the row, e.g. `<!-- weight: 2.0 -->` (optional)

## Extraction CLI Reference

We often want a local, searchable copy of review feedback from CodeRabbit — and specifically the “Prompt for AI Agents” sections.

- Path: `scripts/extract_pr_comments.py`
- Requirements: GitHub CLI (`gh`) authenticated to this repo
- Usage examples:
  - Full comments from a PR by CodeRabbit:
    - `python3 scripts/extract_pr_comments.py --pr 123 --author coderabbit --out docs/audits/coderabbit-pr123-comments.md`
  - Prompts-only into a separate file:
    - `python3 scripts/extract_pr_comments.py --pr 123 --author coderabbit --prompts-only --prompts-out docs/audits/coderabbit-pr123-prompts.md`
- Notes
  - `--author` is a substring match (case-insensitive); defaults to `coderabbit`.
  - Issue comments, review comments (inline), and review bodies authored by CodeRabbit are included.

## Generate Code Review Worksheet (from CodeRabbit prompts)

Create a comprehensive, fill-in-place worksheet from a PR’s CodeRabbit “Prompt for AI Agents” items.

- Path: `scripts/generate_cr_worksheet.py`
- Output: `docs/audits/code-reviews/PR{pr}/{head_sha}.md`
- Requirements: GitHub CLI (`gh`) authenticated
- What it does:
  - Fetches PR metadata and comments with `gh`
  - Extracts only the fenced blocks under “Prompt for AI Agents” (skips HTML-only artifacts like `</summary>`)
  - Emits a worksheet with:
    - Front matter and header table (Date | Agent | SHA | Branch | PR)
    - Accepted/Rejected templates
    - For each prompt: a heading (`### path:line`), checkboxes, the prompt text in a fenced block, and `{response}` placeholder
    - A `---` horizontal rule before the Conclusion section
- Usage examples:
  - Current repo: `python3 scripts/generate_cr_worksheet.py --pr 3`
  - Any repo: `python3 scripts/generate_cr_worksheet.py --repo owner/name --pr 123`
  - Options:
    - `--author coderabbit` (default, substring match)
    - `--agent CodeRabbit` (label in the header)
    - `--out-root docs/audits/code-reviews` (destination root)

## Extracting CodeRabbit PR Comments and Prompts

We often want a local, searchable copy of review feedback from CodeRabbit — and specifically the “Prompt for AI Agents” sections.

Script

- Path: `scripts/extract_pr_comments.py`
- Requirements: `gh` CLI authenticated to GitHub

Examples

- Full comments from a PR by CodeRabbit:
  - `python3 scripts/extract_pr_comments.py --pr 123 --author coderabbit --out docs/audits/coderabbit-pr123-comments.md`
- Prompts-only extraction (finds “Prompt for AI Agents” blocks):
  - `python3 scripts/extract_pr_comments.py --pr 123 --author coderabbit --prompts-only --prompts-out docs/audits/coderabbit-pr123-prompts.md`
- If `--pr` is omitted, the script uses the PR associated with the current branch via `gh pr view`.
- `--author` is a substring match (case-insensitive); defaults to `coderabbit`.

What gets extracted

- Issue comments, review comments (inline), and review bodies authored by CodeRabbit
- Prompts-only mode searches for fenced code following the heading and for inline text near the “Prompt for AI Agents” header
- Output is chronological, with file/line and direct links to the comments

## Makefile Targets

- `make build` — build the main binary
- `make hooks` — enable local git hooks (pre-commit progress updater)

## Notes

- The progress updater and comment extractor are intentionally lightweight and have no extra Python dependencies (use `gh` for GitHub API).
- If you add or rename feature groups/tables in the ledger, the updater will scan all tables that include `Emoji` and `Progress %` columns, so it’s robust to grouping changes.
