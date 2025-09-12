# Release and Versioning

- Last updated: 2025-09-12

## Executive Summary
Defines our versioning scheme, changelog process, and release checklist for alpha → beta → RC → GA.

## Versioning
- Semantic Versioning (SemVer): MAJOR.MINOR.PATCH
- Pre-releases: `-alpha`, `-beta`, `-rc`
- Branch: main is protected; release tags from main.

## Changelog
- Keep `CHANGELOG.md` in the repo.
- Use conventional commit messages; sections: Features, Fixes, Docs, CI, Refactor, Tests, Chore.
- On each release, summarize notable changes since last tag.

## Release Checklist
1) Ensure CI green; govulncheck passes; tests (unit/race/e2e) pass.
2) Update docs (README, PRD, performance baseline) if needed.
3) Bump version in `CHANGELOG.md` with date and summary.
4) Tag release
```bash
git tag vX.Y.Z[-pre]
git push --tags
```
5) Publish GitHub Release notes, attach Docker image reference.
