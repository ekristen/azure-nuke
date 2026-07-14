# Copilot Instructions for azure-nuke

This file guides GitHub Copilot and other AI agents when working on azure-nuke. It captures architectural decisions, development constraints, and best practices for future work.

## Primary Source of Truth

Use [CLAUDE.md](../CLAUDE.md) as the default authority for architecture, SDK strategy, authentication patterns, and resource implementation details.

When this file and CLAUDE.md overlap, prefer CLAUDE.md unless this file states a stricter workflow requirement.

## Model Preference

For repository work, prefer Claude-family models in Copilot Chat whenever they are available in the model picker.

If Claude models are not available in the current environment, continue with the available model and still follow CLAUDE.md guidance.

## Project Overview

**azure-nuke** is a CLI tool for automatically deleting Azure resources across subscriptions and resource groups. Built on the libnuke framework with multiple Azure SDKs (Track 1, Track 2, HashiCorp, Hamilton/msgraph).

**Repository**: [ekristen/azure-nuke](https://github.com/ekristen/azure-nuke)
**Language**: Go (1.21+, latest 1.26)
**Key Dependencies**: libnuke v0.24.5, urfave/cli/v2, logrus, multiple Azure SDKs

For architecture details and SDK decisions, see [CLAUDE.md](../CLAUDE.md).

---

## Development Environment: DevContainer

**Status**: ✅ Complete (feature/add-devcontainer branch)

A production-ready devcontainer has been created for Go development.

### Quick Setup

```bash
# Build devcontainer
devcontainer build --workspace-folder .

# Open in VS Code (Remote - Containers extension)
# Command Palette: Remote-Containers: Reopen in Container
```

### Key Tools Included

- **Go 1.26-trixie**: Latest Go with multi-arch support (x86-64 + arm64)
- **Docker-in-Docker**: Build and test container images
- **golangci-lint v8+**: Code linting with fast mode
- **goreleaser**: Multi-platform builds with semantic versioning
- **pre-commit**: Git hooks with conventional commits
- **Delve**: Step-through debugger
- **VS Code Extensions**: golang.go, GitLens, Copilot, Pull Request review

### Environment Variables

- `CGO_ENABLED=0` - Static linking (required for azure-nuke)
- `GOFLAGS="-v -mod=readonly"` - Verbose builds, prevent mutations

### Git Hooks

Conventional commits with enforced scopes:
- `core`, `resources`, `config`, `commands`, `pkg`, `docs`, `ci`, `deps`

Example commits:
```bash
git commit -m "feat(resources): add storage account deletion"
git commit -m "fix(core): handle nil pointer in disk lister"
git commit -m "docs: update quick-start guide"
```

**Setup hooks** (optional):
```bash
pre-commit install
```

### Debugging

Debug main program:
- Press `F5` in VS Code or use Run menu
- Breakpoints work automatically
- Step through code with F10, F11

Debug tests:
- Use "Debug Test" CodeLens above test functions
- Or use launch config "Debug Package Tests"

### Documentation

See [.devcontainer/README.md](../.devcontainer/README.md) for:
- Common development tasks
- Azure SDK authentication options
- Port forwarding (docs at 8080, profiling at 6060)
- Troubleshooting

**Decisions logged**: [ADR-002-devcontainer-architecture.md](./memory/decisions/ADR-002-devcontainer-architecture.md)

---

## Code Patterns & Standards

### Multi-SDK Strategy

Azure-nuke intentionally uses 4 different Azure SDKs (NOT technical debt):

| SDK | Services | Examples |
|-----|----------|----------|
| **Track 1** (deprecated) | Fallback for older services | Disk, VM, NSG, KeyVault |
| **Track 2** (`armauthorization`) | Newer services | RoleAssignment, SecurityAssessment |
| **HashiCorp** (go-azure-sdk) | Preferred for new resources | ResourceGroup, ApplicationGateway, Budget |
| **Hamilton** (msgraph) | Azure AD/Entra ID only | AADUser, AADGroup, Application, ServicePrincipal |

**When adding resources**:
1. Check if same service type exists (use same SDK for consistency)
2. Check HashiCorp availability (preferred)
3. Fall back to Track 2, then Track 1
4. Use Hamilton only for Azure AD/Entra resources

See [CLAUDE.md](../CLAUDE.md) section 3 for detailed SDK mapping table.

### Resource Development

All resources must:
1. **Embed BaseResource**: Provides Region, SubscriptionID, ResourceGroup properties
2. **Implement Lister interface**: `List(ctx context.Context, o interface{}) ([]resource.Resource, error)`
3. **Implement Resource interface**: `Remove()`, `Properties()`, `Filter()`, `String()`
4. **Register in init()**: Add to libnuke registry with `DependsOn` declaration
5. **Handle Track 1 imports**: Add `//nolint:staticcheck` comment (Track 1 is deprecated)

### Authentication

**Non-interactive only** (no browser login):
- Client Secret ✅
- Client Certificate (unencrypted PEM) ✅
- Federated Token (OIDC/Workload Identity) ✅
- Managed Identity ❌ (currently)

**Credential locations**:
- Environment variables: `AZURE_TENANT_ID`, `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`
- Certificates: Unencrypted PEM format only
- Tokens: Via `.azure/` mount or environment variable

### Static Linking Requirement

`CGO_ENABLED=0` is mandatory for azure-nuke releases. This ensures:
- Single binary (no C dependencies)
- Cross-platform builds work seamlessly
- DevContainer enforces this via environment variable

---

## Common Development Tasks

### Running Tests

```bash
go test ./...              # Run all tests
go test -v ./...           # Verbose output
go test -race ./...        # With race detector
go test -cover ./...       # With coverage
go test -timeout 5m ./... # Increase timeout
```

### Linting

```bash
golangci-lint run          # Full lint (slow)
golangci-lint run --fast   # Fast mode (dev)
golangci-lint cache clean  # Clear cache
```

### Building

```bash
goreleaser build --snapshot --single-target  # Fast snapshot
goreleaser build --snapshot                  # Multi-platform
make docs-serve                              # Serve docs on :8080
```

### Debugging

```bash
# Via VS Code (F5)
# Or command line:
dlv debug ./... -- --config config.yaml --dry-run
```

### Azure SDK Testing

Mount credentials for testing:
```bash
export AZURE_TENANT_ID="<id>"
export AZURE_CLIENT_ID="<id>"
export AZURE_CLIENT_SECRET="<secret>"

go test ./... # Tests now have Azure access
```

---

## Gotchas & Common Issues

### Track 1 Import Warnings

Track 1 SDKs trigger deprecation warnings. Always add directive:

```go
import (
    "github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2021-04-01/compute" //nolint:staticcheck
)
```

### Certificate Authentication

Only unencrypted PEM certificates are supported:

```go
certs, pkey, err := azidentity.ParseCertificates(certData, nil)  // nil = no password
```

Encrypted certificates will fail to parse.

### Empty Results on Error

Some listers return empty results instead of errors for graceful degradation:

```go
if err != nil {
    return resources, nil  // Intentional: return empty, not error
}
```

This allows tool to continue when one service is unavailable.

### Soft Delete Resources

Some Azure resources (especially AAD) require both soft and permanent delete:

```go
func (r *Application) Remove(ctx context.Context) error {
    if _, err := r.client.Delete(ctx, *r.ID); err != nil {    // Soft delete
        return err
    }
    if _, err := r.client.DeletePermanently(ctx, *r.ID); err != nil {  // Permanent
        return err
    }
    return nil
}
```

---

## When Adding New Features

### Adding a New Azure Resource

**Checklist**:
- [ ] Determine scope (Tenant/Subscription/ResourceGroup)
- [ ] Check existing resources for SDK consistency
- [ ] Create `resources/my-resource.go`
- [ ] Register in `init()` with `registry.Register()`
- [ ] Implement Lister and Resource interfaces
- [ ] Add `//nolint:staticcheck` for Track 1 imports
- [ ] Declare `DependsOn` if applicable
- [ ] Test with `--dry-run` before committing

**Template**:
```go
const MyResourceResource = "MyResource"

func init() {
    registry.Register(&registry.Registration{
        Name:     MyResourceResource,
        Scope:    azure.ResourceGroupScope,
        Resource: &MyResource{},
        Lister:   &MyResourceLister{},
        DependsOn: []string{
            // Dependencies
        },
    })
}

type MyResourceLister struct{}

func (l MyResourceLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
    opts := o.(*azure.ListerOpts)
    // Implementation
    return resources, nil
}

type MyResource struct {
    *BaseResource `property:",inline"`
    client SomeClient
    Name   *string
}

func (r *MyResource) Remove(ctx context.Context) error {
    // Implementation
    return nil
}
```

### Updating Dependencies

Check for CVEs and update:
```bash
go get -u ./...           # Update all (with caution)
go list -u -m all         # Show available updates
go mod tidy               # Clean up
golangci-lint run         # Verify no new issues
```

### Updating DevContainer Tools

Edit `.devcontainer/Dockerfile` to add/update tools:
- Go tools: `go install github.com/org/tool@latest`
- Python packages: Add to pre-commit or pip
- System packages: `apt-get install`

Rebuild: `devcontainer build --workspace-folder .`

### Adding Configuration Options

1. Add flag to CLI (pkg/commands/)
2. Add to config struct (pkg/config/)
3. Document in docs/cli-options.md
4. Update config schema in docs/config.md

---

## Repository Memory System

### ⭐ THREE SEPARATE MEMORY BANKS (ALL Version-Controlled)

Azure-nuke uses **three separate knowledge banks** in `.github/memory/`, all shared via git. See [.github/memory/README.md](./memory/README.md) for complete structure.

### Required Policy

- For all repository work, use the repository memory bank under `.github/memory/`.
- Store session notes in `.github/memory/sessions/`, planning notes in `.github/memory/planning/`, and architecture decisions in `.github/memory/decisions/`.
- **Do not use private agent memory** (`/memories/`) for repository notes, plans, or decisions.
- If a note was accidentally written to private memory, move it into `.github/memory/` and remove the private copy.

#### 1️⃣ **Sessions** (`.github/memory/sessions/`)
Temporary, session-specific working notes and research.
- **Naming**: `session-YYYY-MM-DD.md` or `session-{feature}.md`
- **Lifecycle**: Can be archived/deleted after session
- **Examples**: Session notes on debugging, discovery work, research

#### 2️⃣ **Planning** (`.github/memory/planning/`)
Medium-to-long-term architectural planning and roadmaps.
- **Naming**: Descriptive like `roadmap.md`, `resource-expansion-plan.md`
- **Lifecycle**: Persist, updated as plans evolve
- **Examples**: Feature roadmap, multi-SDK migration strategy, performance improvements

#### 3️⃣ **Decisions** (`.github/memory/decisions/`)
**Architecture Decision Records (ADRs)** — the highest-level technical choices.
- **Naming**: `ADR-{number:03d}-{title}.md` (e.g., ADR-001-project-foundation-from-readme.md)
- **Basis**: Reference root `README.md` principles
- **Lifecycle**: Permanent; immutable once decided
- **Examples**:
    - `ADR-001-project-foundation-from-readme.md` ✅ (already created)
    - `ADR-002-devcontainer-architecture.md` ✅ (already created)
    - `ADR-003-multi-sdk-strategy.md` (future)

### ADR Template

Each ADR should follow this structure:

```markdown
# ADR-{number:03d}: {Title}

**Status**: Decided | Proposed | Superseded
**Date**: YYYY-MM-DD
**Context**: [Link to related README.md section]

## Problem Statement
What challenge or decision?

## Decision
What was decided?

## Rationale
Why this choice?

## Alternatives Considered
What else was evaluated?

## Consequences
What are the implications?

## References
- README.md section: [link]
- Related ADRs: [ADR links]
- Implementation: [code links]
```

### Relationship to README.md

The root [README.md](../README.md) is the source of truth for:
- Project purpose and scope
- Core constraints (static linking, non-interactive auth)
- Values and philosophy

**Every ADR should**:
1. ✅ Reference relevant README sections
2. ✅ Explain how it aligns with or justifies deviation from principles
3. ✅ Link to other ADRs with dependencies
4. ✅ Include full rationale and alternatives

### Workflow

**During Development**:
1. Session work → `sessions/`
2. Planning initiatives → `planning/`
3. Significant decisions → Create new ADR in `decisions/`

**When Creating PR**:
- Reference memory files for context
- Link to ADRs your changes implement
- Commit memory updates alongside code

**Example Session Work**:
```markdown
# session-2026-07-14-devcontainer.md

## Investigation: Go 1.26 vs 1.21
[Research notes and findings]

## Decision: Use Go 1.26-trixie
[Link to ADR-001 which captured this decision]
```

**Example Planning**:
```markdown
# planning/resource-expansion-plan.md

## Q3 Goals
- Add Container Registry support
- Modernize SecurityAlert to Track 2 SDK
[Outline and links to related ADRs]
```

### Current ADRs

| ID | Title | Status |
|----|-------|--------|
| ADR-001 | [Project Foundation from README](./memory/decisions/ADR-001-project-foundation-from-readme.md) | Decided |
| ADR-002 | [DevContainer Architecture](./memory/decisions/ADR-002-devcontainer-architecture.md) | Decided |
| ADR-003 | Multi-SDK Strategy | (Future) |
| ADR-004 | Non-Interactive Authentication Only | (Future) |

### Adding New ADRs

1. Create `ADR-{next-number:03d}-{title}.md` in `.github/memory/decisions/`
2. Reference README.md sections where applicable
3. Use template above
4. Update table above in this file
5. Commit and PR for team visibility

---

## References & Resources

| Resource | Location | Purpose |
|----------|----------|---------|
| Architecture decisions | [CLAUDE.md](../CLAUDE.md) | SDK strategy, auth, resource patterns |
| DevContainer guide | [.devcontainer/README.md](../.devcontainer/README.md) | Setup, debugging, Azure SDK auth |
| DevContainer decisions | [ADR-002-devcontainer-architecture.md](./memory/decisions/ADR-002-devcontainer-architecture.md) | Why Go 1.26, features, future extensions |
| Configuration docs | [docs/config.md](../docs/config.md) | Config file format and options |
| Resource list | [docs/resources/](../docs/resources/) | All 40+ resource documentation |
| CLI options | [docs/cli-options.md](../docs/cli-options.md) | Command-line flags |
| Contributing | [docs/contributing.md](../docs/contributing.md) | PR guidelines, commit standards |

---

## For Future Contributors & AI Agents

When working on azure-nuke:

1. **Read CLAUDE.md first**: Understand SDK decisions before adding resources
2. **Check .github/memory/**: Review previous architectural decisions
3. **Use DevContainer**: All development happens in the standardized environment
4. **Follow conventions**: Conventional commits, static linking, non-interactive auth
5. **Log decisions**: Add to .github/memory/ for future reference
6. **Run tests locally**: Before pushing, especially for resource changes
7. **Reference this file**: Update if you discover patterns not documented here

---

**Last Updated**: 2026-07-14
**Status**: Production Ready
**Next Review**: After feature/add-devcontainer merge to main
