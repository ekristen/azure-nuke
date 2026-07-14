# ADR-001: Project Foundation from README

**Status**: Decided
**Date**: 2026-07-14
**Context**: [README.md - Overview](../../README.md#overview), [README.md - History](../../README.md#history)

## Problem Statement

The repository needs a first, canonical ADR that captures project identity and baseline constraints from the root README so later architectural ADRs have a shared foundation.

## Decision

Adopt the root README as the foundational source for project purpose, scope, and origin, with the following extracted principles:
- azure-nuke exists to remove resources from an Azure tenant and its subscriptions.
- The tool is potentially destructive and must be treated with caution.
- Resource coverage is intentionally incomplete and expanded over time through contributions.
- The project is built on libnuke to share cross-cloud cleanup behavior and testing patterns.
- Documentation is maintained in docs/ and published via MkDocs.

## Rationale

The README is the most visible and stable project artifact. Anchoring ADR numbering to a README-derived baseline:
- Preserves the original project intent in a durable, decision-oriented format.
- Gives future ADRs a clear source of truth for consistency checks.
- Reduces drift between contributor-facing documentation and implementation decisions.

## Alternatives Considered

### Alternative 1: Keep ADR-001 as DevContainer Architecture
- Makes implementation detail the first project decision instead of purpose.
- Leaves no formal ADR that captures the repository's baseline identity.
- **Rejected**: Foundational intent should precede implementation details.

### Alternative 2: Do not create a README-based ADR
- Keeps context only in prose documentation.
- Makes it harder to reference immutable decision history from later ADRs.
- **Rejected**: Core project principles should be captured in ADR form.

## Consequences

### Positive
- Establishes a stable ADR baseline for future decisions.
- Clarifies that README principles are normative inputs to architecture choices.
- Makes ADR numbering chronological and easier to reason about.

### Negative
- Introduces light duplication of README content.
- Requires occasional updates if README foundational statements materially change.

## References

- [README.md](../../README.md)
- [README.md - Overview](../../README.md#overview)
- [README.md - History](../../README.md#history)
- [README.md - Documentation](../../README.md#documentation)
- [README.md - libnuke](../../README.md#libnuke)

## Related ADRs

- ADR-002: DevContainer Architecture
- (Future) ADR-003: Multi-SDK Strategy
- (Future) ADR-004: Non-Interactive Authentication Only
