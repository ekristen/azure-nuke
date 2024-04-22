# Resources

!!! Important
    The `Type` and the `Properties` are all **case-sensitive**, you must match case when filtering on them.


This is a list of resources that this tool currently supports. It is generated from the `resources` directory in the
project root. The resources are listed in alphabetical order.

The primary use-case for the resource documentation is to help users in writing filters. The documentation provides
information about the properties that are available for filtering.

## Scope

The `Scope` field in the resource registration defines how the resource is scoped within Azure. This information is ued 
to determine how and when it's queried during the discovery phase of the tool. The following scopes are supported:

- `ResourceGroup` - The resource is scoped to a resource group.
- `Subscription` - The resource is scoped to a subscription.
- `Tenant` - The resource is scoped to a tenant (aka Entra ID / Azure AD).
