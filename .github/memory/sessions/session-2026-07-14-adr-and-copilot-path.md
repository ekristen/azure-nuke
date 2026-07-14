# Session Notes - 2026-07-14

## Summary
- Moved repository Copilot instructions to `.github/copilot-instructions.md` so the repo-level location matches expected discovery.
- Renumbered ADRs in `.github/memory/decisions/`:
  - `ADR-001-devcontainer-architecture.md` -> `ADR-002-devcontainer-architecture.md`
  - Added new `ADR-001-project-foundation-from-readme.md` derived from README project purpose/history.
- Updated references in `.github/memory/README.md` to match new ADR numbering and Copilot instructions path.

## Rationale Captured
- ADR-001 should represent project foundation from README, with implementation details (devcontainer architecture) following as ADR-002.

## Follow-up
- If additional ADRs are added, continue sequence from ADR-003.

## Additional Update
- Updated `.github/copilot-instructions.md` to make `CLAUDE.md` the primary source of truth for architecture and implementation guidance.
- Added explicit model preference policy: prefer Claude-family models in Copilot Chat when available.
- Corrected stale links after moving instructions into `.github/` and synchronized ADR references with current numbering (ADR-001 foundation, ADR-002 devcontainer).

## Pre-commit Hook Fix
- `pre-commit` failed installing hook env for `golangci-lint` at `v1.60.3` with Go 1.26 due to `golang.org/x/tools@v0.24.0` compile incompatibility (`tokeninternal` negative array length assertion).
- Updated `.pre-commit-config.yaml` to use a local/system `golangci-lint` hook instead of building from the old upstream hook repo.
- Adjusted CLI flag for golangci-lint v2 (`--fast-only` replaces `--fast`).
- Validation: `pre-commit run golangci-lint --all-files` passes.
