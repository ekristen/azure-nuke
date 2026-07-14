# Azure-Nuke Repository Memory

This directory contains the institutional knowledge for azure-nuke development, organized into three separate memory banks.

**All memory is version-controlled and shared upstream.** See `.copilot-instructions.md` for guidance.

## Memory Banks

### 📋 Sessions (`sessions/`)
Session-specific working notes, temporary discoveries, and in-progress research. These are scoped to individual development sessions.

**Naming**: `session-YYYY-MM-DD.md` or `session-{feature-name}.md`
**Lifecycle**: Can be archived/deleted after session completes
**Examples**:
- `session-2026-07-14-devcontainer.md` - DevContainer implementation work
- `session-bugfix-nil-pointer.md` - Bug investigation notes

### 🏗️ Planning (`planning/`)
Architectural planning, roadmap items, design discussions, and feature planning documents. These are mid-to-long-term planning artifacts.

**Naming**: Descriptive names like `roadmap.md`, `resource-expansion-plan.md`
**Lifecycle**: Persist across sessions, update as plans evolve
**Examples**:
- `roadmap.md` - Feature pipeline and priorities
- `resource-expansion-plan.md` - Plan to add new Azure resources
- `multi-sdk-strategy.md` - SDK selection and migration planning

### 🎯 Decisions (`decisions/`)
**Architecture Decision Records (ADRs)** that capture "why" behind technical choices. These are the highest-level decisions that shape the project.

**Naming**: `ADR-{number:03d}-{title}.md`
**Basis**: Each ADR should reference and be consistent with principles in the root `README.md`
**Lifecycle**: Permanent; immutable once decided (but can be superseded by new ADRs)
**Examples**:
- `ADR-001-project-foundation-from-readme.md` - Project purpose, scope, and origin extracted from README
- `ADR-002-devcontainer-architecture.md` - Why Go 1.26-trixie, published features, etc.
- `ADR-003-multi-sdk-strategy.md` - Why we use 4 different Azure SDKs

## Relationship to README.md

The root [README.md](../../README.md) is the **source of truth** for project principles:
- Project purpose and scope
- Constraints (static linking, non-interactive auth)
- Core values and philosophy

**Decisions** should:
1. ✅ Reference relevant README sections
2. ✅ Explain how they align with or justify deviations from principles
3. ✅ Link to other ADRs when there are dependencies
4. ✅ Include rationale, alternatives considered, and consequences

## Template: Creating a New ADR

```markdown
# ADR-{number:03d}: {Title}

**Status**: Decided (or Proposed/Superseded)
**Date**: YYYY-MM-DD
**Context**: [Link to related README section if applicable]

## Problem Statement
What challenge or decision did this address?

## Decision
What was decided?

## Rationale
Why this choice?

## Alternatives Considered
What else was evaluated and why wasn't it chosen?

## Consequences
What are the implications?

## References
- Root README section: [section link]
- Related ADRs: [ADR links]
- Decision file: .github/memory/decisions/ADR-{number:03d}-{title}.md
```

## Quick Links

- **Project README**: [README.md](../../README.md)
- **Copilot Instructions**: [copilot-instructions.md](../../.github/copilot-instructions.md)
- **Developer Guide**: [.devcontainer/README.md](../../.devcontainer/README.md)
- **Core SDK Reference**: [CLAUDE.md](../../CLAUDE.md)

## Workflow

### During Development
1. **Session work** → `.github/memory/sessions/`
   - Keep notes on what you're learning
   - Record temporary decisions and research

2. **Planning an initiative** → `.github/memory/planning/`
   - Outline the approach
   - Link to any relevant ADRs

3. **Making a significant decision** → Create an ADR in `.github/memory/decisions/`
   - Only for decisions that will persist long-term
   - Reference the README
   - Document alternatives and rationale

### When Creating a PR
- Reference memory files in PR description if they provide context
- Link to ADRs that your changes implement or follow
- Archive completed session notes (optional)

### When Reviewing
- Check if decisions align with existing ADRs
- If a new pattern emerges, consider capturing it as an ADR
- Reference memory files to explain project context

## Examples in This Repository

- `decisions/ADR-001-project-foundation-from-readme.md` - Project purpose, scope, and origin extracted from README
- `decisions/ADR-002-devcontainer-architecture.md` - Why Go 1.26-trixie, published features, custom tooling
- `planning/roadmap.md` (future) - What resources or features are coming
- `sessions/session-2026-07-14-devcontainer.md` (future) - Session notes from devcontainer implementation
