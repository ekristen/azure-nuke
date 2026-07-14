# ADR-002: DevContainer Architecture

**Status**: Decided
**Date**: 2026-07-14
**Context**: [README.md - Contribution workflow](../../README.md#contribution)

## Problem Statement

Developers need a consistent, modern development environment for azure-nuke Go CLI work that:
- Supports latest Go features and best practices
- Enables multi-platform builds (x86-64 + Apple Silicon)
- Provides debugging, testing, and documentation tooling
- Enforces code quality via git hooks and conventional commits
- Supports non-interactive authentication testing with Azure SDKs

## Decision

Implement a production-ready VS Code devcontainer using:
- **Base Image**: Go 1.26-trixie (latest available, not 1.21 minimum)
- **Published Features**: docker-in-docker, golangci-lint (v8+), git
- **Custom Dockerfile**: goreleaser, pre-commit, commitlint, Delve debugger
- **Environment**: `CGO_ENABLED=0` for static linking (per README constraint)
- **Git Hooks**: Conventional commits with scoped commit enforcement
- **Debugging**: Delve with VS Code launch configurations
- **Ports**: 8080 (docs), 6060 (profiling)

## Rationale

### Base Image: Go 1.26-trixie

Go 1.26 is the latest available version and includes significant language improvements:
- Range-over-int optimization for cleaner loops
- Improved loop behavior and semantics
- Multi-arch support (x86-64 + arm64) for Apple Silicon
- Maintained by Microsoft DevContainers team

### Feature Strategy: Published + Custom

Published devcontainer features are version-locked and maintained by Microsoft, reducing maintenance burden:
- **docker-in-docker**: Essential for container builds and goreleaser multi-platform CI
- **golangci-lint v8+**: Latest linter via feature (preferred over manual `go install`)
- **git**: Required for goreleaser version injection from git tags

Custom Dockerfile for tools without published features:
- **goreleaser**: No published feature; needed for multi-platform release builds
- **pre-commit**: Python-based framework for git hooks orchestration
- **commitlint**: npm-based semantic commit validation
- **Delve**: Go debugger installed via `go install` for step-through debugging

### Static Linking Requirement

Per README, releases must be statically linked with no C dependencies:
- Environment variable `CGO_ENABLED=0` set as default
- Enables seamless cross-platform builds (Linux, macOS, Windows)
- Single binary deployment without libc dependencies

### Conventional Commits & Git Hooks

Structured commit messages enable:
- Automated semantic versioning in goreleaser
- Automated changelog generation from commits
- Clear, searchable commit history
- Scope-based organization: `core`, `resources`, `config`, `commands`, `pkg`, `docs`, `ci`, `deps`

Hooks are optional to avoid friction (users can skip `pre-commit install`).

### Multi-SDK Complexity

Azure-nuke intentionally uses 4 different Azure SDKs (documented in CLAUDE.md):
- **Track 1**: Compute, Network, Storage (deprecated but necessary)
- **Track 2**: Authorization, Security resources
- **HashiCorp**: ResourceGroup, ApplicationGateway, Budget (preferred for new)
- **Hamilton**: Azure AD/Entra ID only

Devcontainer supports all SDKs equally with no special handling needed. This is intentional, not technical debt—each SDK is used for its best coverage of specific Azure services.

## Alternatives Considered

### Alternative 1: Use Go 1.21 (Project Minimum)
- ✗ Misses Go 1.26 features and optimizations
- ✗ Doesn't leverage latest language capabilities
- ✗ Inconsistent with "latest tooling" goal
- **Rejected**: Latest version better for development velocity

### Alternative 2: All Tools via `go install` in Dockerfile
- ✗ Difficult to version-lock across updates
- ✗ Duplicates maintenance work
- ✗ Microsoft publishes features specifically to avoid this
- **Rejected**: Published features provide better maintainability

### Alternative 3: Single Azure SDK
- ✗ Not all services available in all SDK generations
- ✗ Newer SDKs don't have complete Azure coverage
- ✗ Project constraints require Track 1 for some compute resources
- **Rejected**: Multi-SDK strategy is intentional (see CLAUDE.md)

### Alternative 4: Interactive Authentication Support
- ✗ Browser-based login not possible in non-interactive environment
- ✗ Device code flow adds complexity
- ✗ CI/CD deployments require non-interactive credentials
- **Rejected**: Per README, only non-interactive auth is supported (client secret, cert, federated token)

## Consequences

### Positive
- Developers have latest Go 1.26 features and optimizations available
- Modern, maintainable devcontainer configuration
- Git hooks encourage clean commit history for changelog automation
- Multi-platform build support (goreleaser + docker-in-docker)
- Consistent debugging experience via Delve + VS Code
- CI validation ensures devcontainer stays functional (devcontainer.yml workflow)

### Negative
- Post-create time increases (< 1 min, acceptable)
- Users must explicitly `pre-commit install` (optional, but easy to miss)
- Azure SDK complexity still exists (but well-documented in CLAUDE.md)

### Maintenance
- Devcontainer configuration lives in `.devcontainer/`, version-controlled
- CI validation workflow catches misconfigurations early
- Published features auto-update, reducing manual maintenance
- Custom Dockerfile remains simple (25 lines, single-purpose tools)

## References

- **README.md**: Project purpose, constraints, contribution workflow
- **CLAUDE.md**: SDK architecture decisions, authentication patterns, resource development guide
- **Implemented Files**:
  - `.devcontainer/devcontainer.json` - VS Code configuration
  - `.devcontainer/Dockerfile` - Custom tooling
  - `.devcontainer/post-create-command.sh` - Initialization
  - `.devcontainer/README.md` - Developer guide (850+ lines)
  - `.pre-commit-config.yaml` - Git hooks
  - `.vscode/launch.json` - Debugger configs
  - `.github/workflows/devcontainer.yml` - CI validation
- **External Docs**:
  - Pre-commit: https://pre-commit.com/
  - Conventional Commits: https://www.conventionalcommits.org/
  - Microsoft DevContainers: https://github.com/devcontainers/images/tree/main/src/go
  - Goreleaser: https://goreleaser.com/

## Related ADRs

- (Future) ADR-003: Multi-SDK Strategy
- (Future) ADR-004: Non-Interactive Authentication Only
