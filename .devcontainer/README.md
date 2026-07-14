# Azure-Nuke DevContainer

Lean development environment for azure-nuke using Go 1.26-trixie with Docker-in-Docker and linting.

## Quick Start

### Build and Open DevContainer

```bash
# Build the devcontainer image
devcontainer build --workspace-folder /path/to/azure-nuke

# Open in VS Code (if using Remote - Containers extension)
devcontainer open /path/to/azure-nuke

# Or from VS Code command palette: Remote-Containers: Reopen in Container
```

### First Steps in DevContainer

After container opens, you'll see initialization output with quick-start commands. Dependencies are cached automatically.

## Available Tools

| Tool | Purpose | Command |
|------|---------|---------|
| **Go 1.26** | Language runtime | `go version` |
| **golangci-lint v8+** | Code linting | `golangci-lint run` |
| **pre-commit** | Git hook framework | `pre-commit --version` |
| **Delve** | Debugger | F5 in VS Code |
| **Docker socket** | Container builds | `docker ps` |
| **Git** | Version control | `git --version` |

## Common Development Tasks

### Run Tests

```bash
go test ./...
go test -v ./...          # Verbose
go test -race ./...       # With race detector
go test -cover ./...      # With coverage
```

### Lint Code

```bash
golangci-lint run              # Full lint
golangci-lint run --fast       # Fast mode
golangci-lint cache clean      # Clear cache
```

### Build Locally

Use repo make targets for local builds, for example:

```bash
make build
```

### Build Container Images

```bash
# Build and test a Docker image locally
docker build -t azure-nuke:dev .
docker run --rm azure-nuke:dev --help
```

### Serve Documentation

```bash
# Build and serve docs locally (port 8080)
make docs-serve

# Then open http://localhost:8080 in browser
```

### Format and Fix Issues

```bash
# Format Go code
go fmt ./...

# Auto-fix some linting issues
golangci-lint run --fix

# Check for security issues
gosec ./...
```

## Git Hooks

Git hooks are optional. `pre-commit` is installed in this devcontainer via feature. Enable hooks with:

```bash
pre-commit install
```

## Debugging

### Step-Through Debugging in VS Code

1. **Set breakpoint**: Click left margin in `main.go` or any file
2. **Start debugger**: Press `F5` or use Run menu → Start Debugging
3. **Inspect variables**: Hover over variables or use Debug Console
4. **Continue**: Press `F5` (or F10 step, F11 step into)

### Debug a Specific Test

```bash
# Run test with debugger
dlv test ./pkg/azure -- -test.run TestXXX

# Or from VS Code: Click CodeLens "Debug Test" above test function
```

### Debug a Specific Main Invocation

```bash
# Run azure-nuke with debugger
dlv debug ./... -- --help
```

### View Logs/Traces

```bash
# Enable verbose logging in go run
LOGLEVEL=debug go run main.go ...

# Or inspect via dlv console
dlv debug
(dlv) break main.main
(dlv) continue
```

## Azure SDK Development

Azure-nuke uses multiple Azure SDKs intentionally (see [CLAUDE.md](../CLAUDE.md)). When testing with Azure:

### Option A: Environment Variables (Recommended for Local Dev)

```bash
# Client Secret (non-interactive)
export AZURE_TENANT_ID="<tenant-id>"
export AZURE_CLIENT_ID="<client-id>"
export AZURE_CLIENT_SECRET="<client-secret>"

# Then run
azure-nuke --config config.yaml --dry-run
```

### Option B: Mount Azure CLI Credentials

```bash
# From host machine, assuming ~/.azure/config exists:
# VSCode Remote: Add to devcontainer.json:
"mounts": [
  "source=${localEnv:HOME}/.azure,target=/home/vscode/.azure,type=bind,readonly"
]
```

### Option C: Workload Identity Federation (OIDC)

```bash
export AZURE_TENANT_ID="<tenant-id>"
export AZURE_CLIENT_ID="<client-id>"
export AZURE_FEDERATED_TOKEN_FILE="/path/to/token"
```

## Environment Variables

Pre-set in devcontainer:

| Variable | Value | Purpose |
|----------|-------|---------|
| `CGO_ENABLED` | `0` | Static linking (required for azure-nuke) |
| `GOFLAGS` | `-v -mod=readonly` | Verbose Go builds, prevent mod mutations |

Additional options:

```bash
export LOGLEVEL=debug          # Enable debug logging
export GITHUB_TOKEN=...        # For goreleaser GitHub integration
```

## Troubleshooting

### Issue: Docker socket not working

```bash
# Verify socket is accessible
docker ps

# If fails, check mount in devcontainer.json
# Ensure /var/run/docker.sock is mounted
```

### Host Runtime Setup (Docker Desktop, Colima, Podman, Windows/WSL)

This devcontainer uses Docker-in-Docker by default, so it no longer requires a host socket bind in normal use.

If your workflow needs host runtime access instead of Docker-in-Docker, provide a Docker-compatible socket at `/var/run/docker.sock`.

#### Docker Desktop (macOS)

```bash
open -a Docker
docker version
```

#### Colima (macOS)

```bash
colima start
docker context use colima
docker version
```

If `docker version` works on the host, reopen the devcontainer.

#### Podman (macOS)

Podman does not always expose Docker's default socket path automatically. Start Podman machine and locate the Podman API socket:

```bash
podman machine start
podman machine inspect --format '{{.ConnectionInfo.PodmanSocket.Path}}'
```

If you want to use Podman as the host daemon (instead of Docker-in-Docker), expose that socket at `/var/run/docker.sock` so VS Code Dev Containers can bind it.

Example (host-side):

```bash
PODMAN_SOCK="$(podman machine inspect --format '{{.ConnectionInfo.PodmanSocket.Path}}')"
sudo mkdir -p /var/run
sudo ln -sf "$PODMAN_SOCK" /var/run/docker.sock
docker version
```

Notes:
- This symlink is host-level and may need to be re-created after Podman machine restarts.
- If you use a different Podman socket strategy, keep the end state the same: `/var/run/docker.sock` must be a working Docker-compatible socket when host-socket mode is enabled.
- Inside the container, `.devcontainer/container-runtime-preflight.sh` checks this and prints guidance.

#### Docker Desktop (Windows + WSL)

Use Docker Desktop as the runtime and enable WSL integration.

1. Start Docker Desktop.
2. Open Settings > Resources > WSL Integration.
3. Enable integration for your WSL distro.
4. In WSL, verify:

```bash
docker version
docker context ls
```

Then reopen the devcontainer from VS Code.

#### Podman (WSL/Linux)

Use Podman host daemon mode only if you provide a Docker-compatible API socket to Dev Containers.

```bash
podman system service -t 0 unix:///tmp/podman.sock
export DOCKER_HOST=unix:///tmp/podman.sock
docker version
```

If your VS Code host process uses `/var/run/docker.sock`, create a compatible host socket path or symlink that resolves to your Podman socket before reopening the devcontainer.

Notes:
- Podman support in Dev Containers is improving, but Docker Desktop remains the most predictable option on Windows.
- Keep the required end state the same when using host-socket mode: a working Docker-compatible socket path at `/var/run/docker.sock`.

### Issue: golangci-lint not available

```bash
# Install manually
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint --version
```

### Issue: Pre-commit hooks failing on commit

```bash
# Temporarily skip hooks (not recommended)
git commit --no-verify -m "..."

# Or fix and retry
git add .
git commit -m "..."
```

### Issue: Tests hanging or timeout

```bash
# Increase timeout
go test -timeout 5m ./...

# Or run specific test with verbose output
go test -timeout 30s -v -run TestXXX ./pkg/azure
```

### Issue: Module download slow

```bash
# Clear module cache and re-download
go clean -modcache
go mod download

# Or use proxy override (in .devcontainer post-create)
export GOPROXY=https://proxy.golang.org,direct
```

## Port Forwarding

The devcontainer forwards two ports:

| Port | Service | Auto-Forward |
|------|---------|--------------|
| **8080** | Documentation (make docs-serve) | Notify |
| **6060** | Profiling (pprof) | Notify |

When accessed, VS Code will notify with browser links.

## Advanced: Modifying the DevContainer

### Add a New Feature

Edit `.devcontainer/devcontainer.json`:

```json
"features": {
  "ghcr.io/devcontainers/features/node:latest": {
    "version": "20"
  }
}
```

Rebuild: `devcontainer build --workspace-folder .`

### Add a Tool

This setup uses the upstream devcontainer image directly (no local Dockerfile).

Install ad hoc tools in the running container, for example:

```bash
apt-get update && apt-get install -y --no-install-recommends my-tool
```

For repeatable setup, prefer adding an appropriate Dev Container Feature in `.devcontainer/devcontainer.json`.

### Install Go Tools

Most Go CLI tools can be installed in post-create:

```bash
# Edit .devcontainer/post-create-command.sh
go install github.com/some/tool@latest
```

## Tips & Best Practices

- **Use `--fast` flag**: `golangci-lint run --fast` for quick feedback during development
- **Keep .azure mounted readonly**: Security best practice for credentials
- **Check CLAUDE.md**: Reference for SDK decisions and architecture patterns
- **Run full lint before PR**: `golangci-lint run` (without --fast) before pushing
- **Test in CI**: Always verify tests pass in GitHub Actions before merge
- **Profile if slow**: Use `pprof` on port 6060 for performance analysis

## References

- [Go DevContainer Docs](https://github.com/devcontainers/images/tree/main/src/go)
- [azure-nuke CLAUDE.md](../CLAUDE.md)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Pre-commit Framework](https://pre-commit.com/)
- [goreleaser Docs](https://goreleaser.com/)
- [Delve Debugger](https://github.com/go-delve/delve)
