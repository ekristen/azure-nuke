# CLAUDE.md - Azure-Nuke Development Guide

This document provides comprehensive guidance for developing and maintaining azure-nuke. It documents the current design patterns, SDK choices, and architectural decisions.

## 1. Project Overview

**What azure-nuke does:** Azure-nuke is a CLI tool for automatically deleting Azure resources. It scans Azure tenants, subscriptions, and resource groups, then removes resources based on configurable filters. Built on the libnuke framework, it provides consistent resource cleanup across Azure services.

### Core Dependencies

| Dependency | Version | Purpose |
|------------|---------|---------|
| libnuke | v0.24.5 | Core nuke framework (registry, queue, filters) |
| urfave/cli/v2 | v2.27.7 | CLI framework |
| logrus | v1.9.3 | Structured logging |

### Directory Structure

```
azure-nuke/
├── main.go                 # Entry point
├── pkg/
│   ├── azure/              # Azure-specific utilities
│   │   ├── auth.go         # Authentication configuration
│   │   ├── types.go        # Authorizers struct definition
│   │   └── resource.go     # Scope definitions
│   ├── commands/           # CLI command implementations
│   ├── config/             # Configuration handling
│   └── common/             # Common utilities, version info
├── resources/              # Resource implementations (40+ resources)
│   ├── base-resource.go    # Base struct all resources embed
│   ├── disk.go             # Example: Track 1 SDK
│   ├── resource-group.go   # Example: HashiCorp SDK
│   ├── application.go      # Example: msgraph/hamilton
│   └── ...
├── docs/                   # Documentation
└── tools/                  # Utility tools
```

---

## 2. SDK Strategy & Decision Tree

### Why Multiple SDKs Exist

Azure-nuke intentionally uses multiple Azure SDKs. This is not technical debt—it's a deliberate design choice because:

1. **Different Azure services have different SDK maturity levels** - Not all services are available in all SDK generations
2. **Microsoft Graph resources require hamilton** - Azure AD/Entra ID resources use a separate API
3. **Track 1 is deprecated but still necessary** - Some services only exist in Track 1
4. **Consistency within service families** - Resources of the same Azure service should use the same SDK

### SDK Decision Tree

```
Adding or modifying a resource?
│
├── Is it an Azure AD/Entra ID resource (users, groups, apps, service principals)?
│   └── Use: manicminer/hamilton (msgraph)
│
├── Is there an existing resource of the same Azure service type?
│   └── Use: Same SDK as existing resource for consistency
│       Examples:
│       - New compute resource? Use Track 1 (like Disk, VM, Snapshot)
│       - New network resource? Check existing - some use Track 1, some HashiCorp
│
├── Does hashicorp/go-azure-sdk support this resource?
│   └── Use: HashiCorp SDK (preferred for new resources)
│       Import: github.com/hashicorp/go-azure-sdk/resource-manager/...
│
├── Does Track 2 arm* SDK support this resource?
│   └── Use: Track 2 SDK
│       Import: github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/...
│
└── Otherwise
    └── Use: Track 1 SDK (with //nolint:staticcheck)
        Import: github.com/Azure/azure-sdk-for-go/services/...
```

---

## 3. Per-Resource SDK Mapping Table

### Microsoft Graph Resources (manicminer/hamilton)

| Resource | Scope | Import |
|----------|-------|--------|
| AADGroup | Tenant | `hamilton/msgraph` |
| AADUser | Tenant | `hamilton/msgraph` |
| Application | Tenant | `hamilton/msgraph` |
| ApplicationCertificate | Tenant | `hamilton/msgraph` |
| ApplicationFederatedCredential | Tenant | `hamilton/msgraph` |
| ApplicationSecret | Tenant | `hamilton/msgraph` |
| ServicePrincipal | Tenant | `hamilton/msgraph` |

### HashiCorp SDK Resources (go-azure-sdk)

| Resource | Scope | Import Path |
|----------|-------|-------------|
| ApplicationGateway | ResourceGroup | `resource-manager/network/2023-09-01/applicationgateways` |
| Budget | Subscription | `resource-manager/consumption/2021-10-01/budgets` |
| ManagementLock | Subscription | `resource-manager/resources/2020-05-01/managementlocks` |
| MonitorDiagnosticSetting | Subscription | `resource-manager/insights/2021-05-01-preview/diagnosticsettings` |
| NetworkInterface | ResourceGroup | `resource-manager/network/2023-09-01/networkinterfaces` |
| RecoveryServicesBackupPolicy | ResourceGroup | `resource-manager/recoveryservicesbackup/2023-02-01/...` |
| RecoveryServicesVault | ResourceGroup | `resource-manager/recoveryservices/2023-02-01/vaults` |
| ResourceGroup | Subscription | `resource-manager/resources/2022-09-01/resourcegroups` |

### Track 2 SDK Resources (arm* packages)

| Resource | Scope | Import Path |
|----------|-------|-------------|
| RecoveryServicesBackupProtectedItem | ResourceGroup | `sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup` |
| RecoveryServicesBackupProtectionContainer | ResourceGroup | `sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup` |
| RecoveryServicesBackupProtectionIntent | ResourceGroup | `sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup` |
| SecurityAssessment | Subscription | `sdk/resourcemanager/security/armsecurity` |
| SubscriptionRoleAssignment | Subscription | `sdk/resourcemanager/authorization/armauthorization` |

### Track 1 SDK Resources (azure-sdk-for-go/services)

| Resource | Scope | Import Path |
|----------|-------|-------------|
| AppServicePlan | ResourceGroup | `services/web/mgmt/2021-03-01/web` |
| ContainerRegistry | ResourceGroup | `services/containerregistry/mgmt/2019-05-01/containerregistry` |
| Disk | ResourceGroup | `services/compute/mgmt/2021-04-01/compute` |
| DNSZone | ResourceGroup | `services/dns/mgmt/2018-05-01/dns` |
| IPAllocation | ResourceGroup | `services/network/mgmt/2021-05-01/network` |
| KeyVault | ResourceGroup | `services/keyvault/mgmt/2019-09-01/keyvault` |
| NetworkSecurityGroup | ResourceGroup | `services/network/mgmt/2022-05-01/network` |
| PolicyAssignment | Subscription | `services/preview/resources/mgmt/2021-06-01-preview/policy` |
| PolicyDefinition | Subscription | `services/preview/resources/mgmt/2021-06-01-preview/policy` |
| PrivateDNSZone | ResourceGroup | `services/privatedns/mgmt/2018-09-01/privatedns` |
| PublicIPAddress | ResourceGroup | `services/network/mgmt/2022-05-01/network` |
| SecurityAlert | Subscription | `services/preview/security/mgmt/v3.0/security` |
| SecurityPricing | Subscription | `services/preview/security/mgmt/v3.0/security` |
| SecurityWorkspace | Subscription | `services/preview/security/mgmt/v3.0/security` |
| Snapshot | ResourceGroup | `services/compute/mgmt/2021-04-01/compute` |
| SSHKey | ResourceGroup | `services/compute/mgmt/2021-04-01/compute` |
| StorageAccount | ResourceGroup | `services/storage/mgmt/2021-09-01/storage` |
| VirtualMachine | ResourceGroup | `services/compute/mgmt/2021-04-01/compute` |
| VirtualNetwork | ResourceGroup | `services/network/mgmt/2021-05-01/network` |

---

## 4. Authentication Architecture

### Non-Interactive Only Design

Azure-nuke explicitly supports only non-interactive authentication methods. This is intentional—the tool is designed for automation scenarios (CI/CD, scheduled cleanup) where interactive login is not possible.

**Supported Methods:**
- Client Secret
- Client Certificate (unencrypted PEM only)
- Federated Token (OIDC/Workload Identity)

**Not Supported (by design):**
- Browser-based login
- Device code flow
- Azure CLI token
- Managed Identity (currently)

### Dual Authorizer Pattern

The `Authorizers` struct provides credentials for all SDK types:

```go
// pkg/azure/types.go
type Authorizers struct {
    // For Track 1 SDK and hamilton (autorest-compatible)
    Graph      *autorest.Authorizer    // Microsoft Graph API
    Management *autorest.Authorizer    // Azure Resource Manager

    // For HashiCorp go-azure-sdk
    MicrosoftGraph  auth.Authorizer    // Microsoft Graph API
    ResourceManager auth.Authorizer    // Azure Resource Manager

    // For Track 2 SDK (arm* packages)
    IdentityCreds azcore.TokenCredential
}
```

### Authentication Flow

```go
// pkg/azure/auth.go - Simplified flow

func ConfigureAuth(...) (*Authorizers, error) {
    authorizers := &Authorizers{}

    // 1. Create Track 2 credentials (azidentity)
    if clientSecret != "" {
        creds, _ := azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret, nil)
        authorizers.IdentityCreds = creds
    } else if clientCertFile != "" {
        // Parse cert, create credential
        creds, _ := azidentity.NewClientCertificateCredential(tenantID, clientID, certs, pkey, nil)
        authorizers.IdentityCreds = creds
    } else if clientFedTokenFile != "" {
        creds, _ := azidentity.NewWorkloadIdentityCredential(...)
        authorizers.IdentityCreds = creds
    }

    // 2. Create go-azure-sdk authorizers
    graphAuthorizer, _ := auth.NewAuthorizerFromCredentials(ctx, credentials, env.MicrosoftGraph)
    mgmtAuthorizer, _ := auth.NewAuthorizerFromCredentials(ctx, credentials, env.ResourceManager)

    // 3. Convert to autorest format for Track 1/hamilton compatibility
    authorizers.Management = autorest.AutorestAuthorizer(mgmtAuthorizer)
    authorizers.Graph = autorest.AutorestAuthorizer(graphAuthorizer)

    // 4. Store native go-azure-sdk authorizers
    authorizers.MicrosoftGraph = graphAuthorizer
    authorizers.ResourceManager = mgmtAuthorizer

    return authorizers, nil
}
```

---

## 5. Code Patterns by SDK Type

### Track 1 Pattern

Track 1 uses `autorest.Authorizer` and iterator-based pagination.

```go
import (
    "github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2021-04-01/compute" //nolint:staticcheck
)

func (l DiskLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
    opts := o.(*azure.ListerOpts)

    // Create client with subscription
    client := compute.NewDisksClient(opts.SubscriptionID)
    client.Authorizer = opts.Authorizers.Management
    client.RetryAttempts = 1
    client.RetryDuration = time.Second * 2

    resources := make([]resource.Resource, 0)

    // List returns an iterator
    list, err := client.ListByResourceGroup(ctx, opts.ResourceGroup)
    if err != nil {
        return nil, err
    }

    // Iterate with NotDone() pattern
    for list.NotDone() {
        for _, r := range list.Values() {
            resources = append(resources, &Disk{
                BaseResource: &BaseResource{
                    Region:         r.Location,
                    ResourceGroup:  &opts.ResourceGroup,
                    SubscriptionID: &opts.SubscriptionID,
                },
                client: client,
                Name:   r.Name,
            })
        }

        if err := list.NextWithContext(ctx); err != nil {
            return nil, err
        }
    }

    return resources, nil
}
```

**Key Points:**
- Import requires `//nolint:staticcheck` (Track 1 is deprecated)
- Use `NewXxxClient(subscriptionID)` constructor
- Set `client.Authorizer = opts.Authorizers.Management`
- Pagination via `list.NotDone()` / `list.NextWithContext()`

### Track 2 Pattern

Track 2 uses `azcore.TokenCredential` and pager-based pagination.

```go
import (
    "github.com/Azure/azure-sdk-for-go/sdk/azcore"
    "github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
    "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"
)

func (l *SubscriptionRoleAssignmentLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
    opts := o.(*azure.ListerOpts)

    // Create client with TokenCredential
    client, err := armauthorization.NewRoleAssignmentsClient(
        opts.SubscriptionID,
        opts.Authorizers.IdentityCreds,  // azcore.TokenCredential
        &arm.ClientOptions{
            ClientOptions: azcore.ClientOptions{
                APIVersion: "2022-04-01",
            },
        })
    if err != nil {
        return nil, err
    }

    resources := make([]resource.Resource, 0)

    // Pagination via Pager pattern
    pager := client.NewListPager(&armauthorization.RoleAssignmentsClientListOptions{})

    for pager.More() {
        page, err := pager.NextPage(ctx)
        if err != nil {
            return nil, err
        }

        for _, item := range page.Value {
            resources = append(resources, &SubscriptionRoleAssignment{
                client: client,
                Name:   item.Name,
                // ...
            })
        }
    }

    return resources, nil
}
```

**Key Points:**
- Use `opts.Authorizers.IdentityCreds` (azcore.TokenCredential)
- Pagination via `NewXxxPager()` / `pager.More()` / `pager.NextPage()`
- Can specify API version in ClientOptions

### HashiCorp SDK Pattern

HashiCorp SDK uses `auth.Authorizer` and returns complete lists or provides `Complete` methods.

```go
import (
    "github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"
    "github.com/hashicorp/go-azure-sdk/resource-manager/resources/2022-09-01/resourcegroups"
    "github.com/hashicorp/go-azure-sdk/sdk/environments"
)

func (l ResourceGroupLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
    opts := o.(*azure.ListerOpts)

    // Create client with base URI
    client, err := resourcegroups.NewResourceGroupsClientWithBaseURI(environments.AzurePublic().ResourceManager)
    if err != nil {
        return nil, err
    }
    client.Client.Authorizer = opts.Authorizers.Management

    resources := make([]resource.Resource, 0)

    // List returns Model directly (no pager)
    list, err := client.List(ctx,
        commonids.NewSubscriptionID(opts.SubscriptionID),
        resourcegroups.ListOperationOptions{})
    if err != nil {
        return nil, err
    }

    // Iterate over Model directly
    for _, entity := range *list.Model {
        resources = append(resources, &ResourceGroup{
            BaseResource: &BaseResource{
                Region:         ptr.String(entity.Location),
                SubscriptionID: ptr.String(opts.SubscriptionID),
            },
            client: client,
            Name:   entity.Name,
        })
    }

    return resources, nil
}
```

**Alternative - Complete pagination:**
```go
// For resources with pagination, use ListComplete
items, err := client.ListByResourceGroupComplete(ctx,
    commonids.NewResourceGroupID(opts.SubscriptionID, opts.ResourceGroup))

for _, item := range items.Items {
    // ...
}
```

**Key Points:**
- Use `NewXxxClientWithBaseURI(environments.AzurePublic().ResourceManager)`
- Set `client.Client.Authorizer = opts.Authorizers.Management`
- Use `commonids` for typed resource IDs
- Access results via `*list.Model` or `ListComplete().Items`

### msgraph (Hamilton) Pattern

Hamilton is used for all Azure AD/Entra ID resources.

```go
import (
    "github.com/hashicorp/go-azure-sdk/sdk/odata"
    "github.com/manicminer/hamilton/msgraph"
)

func (l ApplicationLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
    opts := o.(*azure.ListerOpts)

    // Create client (no subscription needed - tenant-scoped)
    client := msgraph.NewApplicationsClient()
    client.BaseClient.Authorizer = opts.Authorizers.Graph
    client.BaseClient.DisableRetries = true

    resources := make([]resource.Resource, 0)

    // List with odata.Query for filtering
    entities, _, err := client.List(ctx, odata.Query{})
    if err != nil {
        return nil, err
    }

    for i := range *entities {
        entity := &(*entities)[i]
        resources = append(resources, &Application{
            BaseResource: &BaseResource{
                Region: ptr.String("global"),
            },
            client: client,
            ID:     entity.ID(),
            Name:   entity.DisplayName,
        })
    }

    return resources, nil
}

// Remove pattern for msgraph
func (r *Application) Remove(ctx context.Context) error {
    // Soft delete
    if _, err := r.client.Delete(ctx, *r.ID); err != nil {
        return err
    }
    // Permanent delete (if applicable)
    if _, err := r.client.DeletePermanently(ctx, *r.ID); err != nil {
        return err
    }
    return nil
}
```

**Key Points:**
- Use `opts.Authorizers.Graph` (autorest.Authorizer)
- Client is tenant-scoped (no subscription ID)
- Use `odata.Query{}` for filtering
- Region is typically "global" for AAD resources
- Some resources require both Delete and DeletePermanently

---

## 6. Resource Development Guide

### Resource Registration Pattern

Every resource must register itself in an `init()` function:

```go
const MyResourceResource = "MyResource"

func init() {
    registry.Register(&registry.Registration{
        Name:     MyResourceResource,
        Scope:    azure.ResourceGroupScope,  // or SubscriptionScope, TenantScope
        Resource: &MyResource{},
        Lister:   &MyResourceLister{},
        DependsOn: []string{
            // Resources that must be deleted before this one
            OtherResourceResource,
        },
    })
}
```

### Scope Selection

| Scope | When to Use | Example Resources |
|-------|-------------|-------------------|
| `azure.TenantScope` | Azure AD/Entra ID resources | Application, AADUser, ServicePrincipal |
| `azure.SubscriptionScope` | Subscription-wide resources | ResourceGroup, Budget, PolicyAssignment |
| `azure.ResourceGroupScope` | Resources inside resource groups | VM, Disk, StorageAccount, VNet |

### Implementing the Lister Interface

```go
type MyResourceLister struct {
    // Optional: caches for expensive lookups
}

func (l MyResourceLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
    opts := o.(*azure.ListerOpts)

    // opts provides:
    // - opts.SubscriptionID
    // - opts.ResourceGroup (if ResourceGroupScope)
    // - opts.Authorizers (all auth types)

    // 1. Create SDK client
    // 2. List resources
    // 3. Convert to []resource.Resource
    // 4. Return

    return resources, nil
}
```

### Implementing the Resource Interface

```go
type MyResource struct {
    *BaseResource `property:",inline"`  // REQUIRED: embed BaseResource

    client SomeClient  // Store client for Remove()
    Name   *string
    Tags   map[string]*string
    // Add filterable properties
}

// Remove deletes the resource
func (r *MyResource) Remove(ctx context.Context) error {
    _, err := r.client.Delete(ctx, *r.ResourceGroup, *r.Name)
    return err
}

// Filter returns error if resource should be excluded from deletion
func (r *MyResource) Filter() error {
    // Return an error to exclude this resource
    // Return nil to include it
    return nil
}

// Properties returns filterable properties
func (r *MyResource) Properties() types.Properties {
    return types.NewPropertiesFromStruct(r)
}

// String returns the display name
func (r *MyResource) String() string {
    return *r.Name
}
```

### BaseResource Embedding

All resources MUST embed `*BaseResource`:

```go
// resources/base-resource.go
type BaseResource struct {
    Region         *string `description:"The region of the resource."`
    SubscriptionID *string `description:"The subscription ID."`
    ResourceGroup  *string `description:"The resource group name."`
}

// BeforeEnqueue sets Owner to Region for filtering
func (r *BaseResource) BeforeEnqueue(item interface{}) {
    i := item.(*queue.Item)
    i.Owner = ptr.ToString(r.Region)
}
```

This provides:
- Consistent property exposure for filtering
- Region-based owner for consistent tool behavior
- Getter methods for common fields

### Dependency Declaration

Use `DependsOn` to ensure proper deletion order:

```go
registry.Register(&registry.Registration{
    Name:     RecoveryServicesVaultResource,
    DependsOn: []string{
        RecoveryServicesBackupProtectedItemResource,  // Delete items before vault
    },
})

registry.Register(&registry.Registration{
    Name:     DiskResource,
    DependsOn: []string{
        VirtualMachineResource,  // Delete VMs before their disks
    },
})
```

---

## 7. Gotchas and Known Quirks

### Track 1 Import Linting

Track 1 imports trigger staticcheck deprecation warnings. Always add the nolint directive:

```go
import (
    "github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2021-04-01/compute" //nolint:staticcheck
)
```

### Certificate Authentication

Only unencrypted PEM certificates are supported:

```go
// This works
certs, pkey, err := azidentity.ParseCertificates(certData, nil)  // nil = no password
```

Encrypted certificates will fail to parse.

### Graph vs MicrosoftGraph Authorizer

The `Authorizers` struct has two Graph authorizers:

```go
// For hamilton/msgraph (autorest-compatible)
opts.Authorizers.Graph  // *autorest.Authorizer

// For go-azure-sdk native Graph operations
opts.Authorizers.MicrosoftGraph  // auth.Authorizer
```

Most code uses `Graph` for hamilton clients. Use `MicrosoftGraph` only when using go-azure-sdk's Graph modules directly.

### Empty Results on Error

Some listers return empty results instead of errors for graceful degradation:

```go
client, err := someClient.New(...)
if err != nil {
    return resources, nil  // Return empty, not error
}
```

This is intentional—it allows the tool to continue processing other resources when one service is unavailable or unauthorized.

### Region Filtering via BeforeEnqueue

Resources set their region as the queue item "Owner" via `BaseResource.BeforeEnqueue()`. This enables region-based filtering in the config:

```yaml
regions:
  - eastus
  - westus2
```
Resources without a meaningful region (like AAD resources) typically use `"global"`.
### Context Deadlines
Some resources add explicit context deadlines to prevent hanging:
```go
ctx, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
defer cancel()
```

This is especially important for resources that make multiple API calls or have known slow responses.

### Soft Delete Resources

Some Azure resources (especially AAD) have a soft delete pattern:

```go
func (r *Application) Remove(ctx context.Context) error {
    // First: soft delete (moves to recycle bin)
    if _, err := r.client.Delete(ctx, *r.ID); err != nil {
        return err
    }
    // Then: permanent delete (removes from recycle bin)
    if _, err := r.client.DeletePermanently(ctx, *r.ID); err != nil {
        return err
    }
    return nil
}
```

### Property Struct Tags

Use struct tags to control property behavior:

```go
type MyResource struct {
    *BaseResource `property:",inline"`  // Inline base properties

    ID     *string `property:"-"`       // Exclude from properties
    Name   *string                       // Normal property
    Secret *string `property:"-"`       // Exclude sensitive data
}
```

---

## 8. Adding a New Resource Checklist

1. [ ] Determine the correct scope (Tenant/Subscription/ResourceGroup)
2. [ ] Check which SDK existing resources of the same service use
3. [ ] Create `resources/my-resource.go` with:
   - [ ] Resource const name
   - [ ] `init()` with `registry.Register()`
   - [ ] Lister struct and `List()` method
   - [ ] Resource struct embedding `*BaseResource`
   - [ ] `Remove()`, `Properties()`, `String()` methods
   - [ ] Optional: `Filter()` method
4. [ ] Add `//nolint:staticcheck` for Track 1 imports
5. [ ] Set appropriate `DependsOn` if this resource depends on others
6. [ ] Test with a dry run: `azure-nuke --config config.yaml --dry-run`