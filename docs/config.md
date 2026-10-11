# Config

The configuration is the user supplied configuration that is used to drive the nuke process. The configuration is a
YAML file that is loaded from the path specified by the `--config` flag.

## Sections

The configuration is broken down into the following sections:

- [blocklist](#blocklist)
- [regions](#regions)
- [accounts](#accounts)
    - [presets](#presets)
    - [filters](#filters)
    - [resource-types](#resource-types)
        - [includes](#includes)
        - [excludes](#excludes)
        - [cloud-control](#cloud-control)
- [resource-types](#resource-types)
    - [includes](#includes)
    - [excludes](#excludes)
    - [cloud-control](#cloud-control)
- [settings](#settings)
- [presets](#global-presets)

## Simple Example

```yaml
blocklist:
  - 000000000-0000-0000-0000-000000000000

regions:
  - global
  - eastus

accounts:
  11111111-1111-1111-1111-111111111111:
    filters:
      ResourceGroup:
        - Default

resource-types:
  includes:
    - ResourceGroup
```

## Blocklist

The blocklist is simply a list of Accounts that the tool cannot run against. This is to protect the user from accidentally
running the tool against the wrong account. The blocklist must always be populated with at least one entry.
The blocklist can hold tenant and subscription IDs. The run aborts if the tenant ID or any of the `--subscription-id`
values is in the blocklist; without `--subscription-id`, blocklisted subscriptions of the tenant are skipped. Subscription
IDs are compared case-insensitively.

## Regions

The regions is a list of Azure regions that the tool will run against. The tool will run against all regions specified in the
configuration. If no regions are listed, then the tool will **NOT** run against any region. Regions must be explicitly
provided.

### All Enabled Regions

You may specify the special region `all` to run against all enabled regions. This will run against all regions that are
enabled in the account. It will not run against regions that are disabled. It will also automatically include the 
special region `global` which is for specific global resources.

!!! important
    The use of `all` will ignore all other regions specified in the configuration. It will only run against regions
    that are enabled in the account.

## Accounts

The accounts section is a map of Azure Tenants to their configuration. The account ID is the key and the value is the
configuration for that tenant.

The configuration for each tenant is broken down into the following sections:

- presets
- filters
- resource-types
    - includes
    - excludes

### Presets

Presets under an account entry is a list of strings that must map to a globally defined preset in the configuration.

### Filters

Filters is a map of filters against resource types. To learn more about filters, see the [Filtering](./config-filtering.md)

**Note:** filters can be defined at the account level and at the preset level.

## Resource Types

Resource types is a map of resource types to their configuration. The resource type is the key and the value is the
configuration for that resource type.

The configuration for each resource type is broken down into the following sections:

- includes
- excludes

### Includes

Includes are a list of resource types the tool will run against. If no includes are specified, then the tool will run against
all resource types.

### Excludes

Excludes are a list of resource types the tool will not run against. If no excludes are specified, then the tool will run
against all resource types unless Includes is specified.

## Settings

Settings are a map of resource types to their settings configuration. This allows for alternative behavior when removing
resources. If a resource has a setting alternative, and you'd like to use its behavior, then you can specify the resource
type in the `settings` section.

```yaml
settings:
  ResourceGroup:
    DeleteManaged: true
    DeleteLocks: true
    SoftDeleteKeyVaultKeys: true
    SoftDeleteKeyVaultKeysTag: created-by=entigo-infralib
```

`ResourceGroup` settings:

- `DeleteManaged`: also delete resource groups managed by another resource (AKS node resource groups, Container Apps
  infrastructure); by default they are filtered.
- `SoftDeleteKeyVaultKeys`: before the resource group is deleted, soft-delete (never purge) the keys of its Key Vaults
  and wait until they are deleted. A purge-protected vault with a fixed name that is recovered later then comes back
  without live keys. `SoftDeleteKeyVaultKeysTag` limits this to vaults with that tag: `key=value`, or `key` for any
  value. Keys backing a certificate and vaults without soft delete are skipped. The data plane must be reachable with
  key delete rights (a firewall or a missing role makes the resource group fail on every attempt, it is not deleted
  without its keys).
- `DeleteLocks`: before the resource group is deleted, delete its management locks, including locks on its resources.
  Needs `Microsoft.Authorization/locks/delete` (Owner or User Access Administrator; Contributor does not have it).

`RoleAssignment` settings:

- `IncludeSubscriptionScope`: also delete assignments on the subscription itself (by default only those within its
  resource groups are deleted; inherited ones never). Keep owners and service principals with filters, e.g. keep
  everything but `PrincipalType` `ServicePrincipal` and keep a list of `PrincipalId`s.

## Global Presets

To read more on global presets, see the [Presets](./config-presets.md) documentation.