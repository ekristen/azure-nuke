# DevContainer Implementation Decisions

**Status**: ✅ COMPLETE
**Branch**: feature/add-devcontainer
**Date**: 2026-07-14

## Summary

Production-ready devcontainer for azure-nuke Go CLI development using Go 1.26-trixie (latest) with multi-arch support (x86-64 + arm64), published devcontainer features, and custom tooling.

## Architecture Decisions

### Base Image: Go 1.26-trixie

**Decision**: Use `mcr.microsoft.com/devcontainers/go:1.26-trixie`

**Rationale**:
- Latest available Go version (1.26, not 1.21 project minimum)
- Go 1.26 features: range-over-int optimization, improved loop behavior
- Debian Trixie (latest stable)
- Multi-arch: x86-64 and arm64 (Apple Silicon support)
- Maintained by Microsoft DevContainers team

**Alternative Considered**: Go 1.21 (project minimum) — rejected to leverage latest language features.

### Feature Strategy: Published Features + Custom Dockerfile

**Decision**: Use published devcontainer features for docker-in-docker, golangci-lint, git; add custom Dockerfile for goreleaser, pre-commit, commitlint, Delve.

**Rationale**:
- Published features are version-locked and maintained by Microsoft
- docker-in-docker: Essential for container builds during development
- golangci-lint: Latest v8+ via published feature (easier than `go install`)
- git: Required for goreleaser version injection from git tags
- Custom Dockerfile for: goreleaser (no published feature), pre-commit (Python-based), commitlint (npm-based), Delve (debugger)

**Published Features Used**:
```json
"features": {
  "ghcr.io/devcontainers/features/docker-in-docker:latest": {},
  "ghcr.io/devcontainers/features/golangci-lint:latest": {"version": "latest"},
  "ghcr.io/devcontainers/features/git:latest": {}
}
```

### Environment Configuration

**Decision**: Set `CGO_ENABLED=0` and `GOFLAGS="-v -mod=readonly"` as default environment variables.

**Rationale**:
- `CGO_ENABLED=0`: Required for static linking (project constraint per CLAUDE.md)
- `GOFLAGS="-v -mod=readonly"`: Verbose builds, prevent accidental module mutations
- Users can override via `.env` or command-line if needed

### Docker Socket Binding

**Decision**: Mount `/var/run/docker.sock` for Docker-in-Docker support.

**Rationale**:
- Enables building and testing container images during development
- Essential for goreleaser multi-platform builds
- Read-write binding allows full Docker CLI access

### Git Hooks: Conventional Commits

**Decision**: Pre-commit framework with golangci-lint, conventional commits validation, and file cleanup.

**Rationale**:
- Conventional commits enforce structured commit messages for:
  - Automated changelog generation
  - Semantic versioning in goreleaser
  - Clear commit history
- Scope enforcement: `core`, `resources`, `config`, `commands`, `pkg`, `docs`, `ci`, `deps`
- Hooks are optional (users can skip `pre-commit install`)

**Scopes Rationale**:
- `core` - Main CLI logic, framework code
- `resources` - Individual Azure resource implementations
- `config` - Configuration parsing, file handling
- `commands` - CLI command implementations
- `pkg` - Utility packages
- `docs` - Documentation and examples
- `ci` - CI/CD workflows
- `deps` - Dependency upgrades

### Debugging Setup

**Decision**: Delve debugger with VS Code launch configurations.

**Rationale**:
- Delve is the standard Go debugger
- VS Code configurations support:
  - Attach to running debugger
  - Debug main program (dry-run mode for testing)
  - Debug specific tests
  - Debug entire package tests
- Installed via `go install` in Dockerfile

### Port Forwarding

**Decision**: Forward ports 8080 (docs) and 6060 (profiling/pprof).

**Rationale**:
- Port 8080: `make docs-serve` for local documentation viewing
- Port 6060: pprof profiling server for performance analysis
- Auto-notification via VS Code when services are running

## Files Created

| File | Lines | Purpose |
|------|-------|---------|
| `.devcontainer/devcontainer.json` | 89 | Primary VS Code devcontainer configuration |
| `.devcontainer/Dockerfile` | 25 | Custom tooling installation |
| `.devcontainer/post-create-command.sh` | 42 | Initialization and quick-start |
| `.pre-commit-config.yaml` | 42 | Git hooks configuration |
| `.devcontainer/README.md` | 850+ | Developer guide and troubleshooting |
| `.vscode/launch.json` | 60 | Delve debugger configurations |
| `.github/workflows/devcontainer.yml` | 60 | CI validation workflow |

## Known Considerations

### Azure SDK Authentication

Users must provide credentials via:
- **Environment variables**: `AZURE_TENANT_ID`, `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`
- **Mount .azure**: Bind mount `~/.azure` directory (read-only)
- **Workload Identity**: OIDC token exchange for CI/CD

Documented in `.devcontainer/README.md` with examples.

### Post-Create Time

- Minimal (< 1 min typical)
- `go mod download` caches dependencies
- `pre-commit install` is optional
- User-triggered for `go test`, builds, etc.

### Git Hooks Optional

- Pre-commit hooks can be skipped if users just want devcontainer
- Conventional commits helpful for goreleaser changelog generation
- Can be installed manually: `pre-commit install`

### Multi-SDK Complexity

Azure-nuke uses 4 different Azure SDKs (per CLAUDE.md):
- Track 1 (deprecated but necessary for some services)
- Track 2 (`azure-sdk-for-go/sdk/resourcemanager/...`)
- HashiCorp go-azure-sdk
- Hamilton (msgraph for Azure AD/Entra ID)

Devcontainer supports all SDKs equally; no special handling needed.

## Testing Verification Steps

**Before merging to main:**

```bash
# Test devcontainer builds successfully
devcontainer build --workspace-folder .

# Verify post-create script is valid shell
bash -n .devcontainer/post-create-command.sh

# Validate devcontainer.json
python3 -m json.tool .devcontainer/devcontainer.json

# Check all files exist
ls -la .devcontainer/
ls -la .vscode/launch.json
ls -la .pre-commit-config.yaml
ls -la .github/workflows/devcontainer.yml
```

## Future Extensions

### Adding New Tools

Edit `.devcontainer/Dockerfile` and rebuild:

```dockerfile
RUN go install github.com/some/tool@latest
```

Or add published features to `devcontainer.json`:

```json
"features": {
  "ghcr.io/devcontainers/features/node:latest": {"version": "20"}
}
```

### Modifying Git Hooks

Edit `.pre-commit-config.yaml`:
- Update repo versions
- Add new hooks (e.g., `shellcheck` for shell scripts)
- Adjust scope enforcement

### Adding Environment Variables

Edit `devcontainer.json` `remoteEnv` section or create `.devcontainer/.env` file.

### Supporting New Resource Scopes

If azure-nuke adds new logical scopes (e.g., `networking`, `identity`), update:
- `.pre-commit-config.yaml` scope regex
- `.devcontainer/README.md` commit examples

## References

- **CLAUDE.md**: SDK decisions, authentication, resource patterns
- **.devcontainer/README.md**: User-facing developer guide
- **devcontainer.json**: VS Code configuration
- **Pre-commit docs**: https://pre-commit.com/
- **Conventional Commits**: https://www.conventionalcommits.org/
- **Go DevContainer**: https://github.com/devcontainers/images/tree/main/src/go
