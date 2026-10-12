# Resource Group

## Details

- **Type:** `ResourceGroup`
- **Scope:** subscription

## Properties

- **`BaseResource`**: No description provided
- **`ManagedBy`**: The ID of the resource that manages the resource group.
- **`Name`**: The Name of the resource group.
- **`tag:<key>:`**: This resource has tags with property `Tags`. These are key/value pairs that are
	added as their own property with the prefix of `tag:` (e.g. [tag:example: "value"]) 
## Depends On

!!! Experimental Feature
    This is an **experimental** feature, please read more about it here <>. This feature attempts to remove all resources in one resource type before moving onto the dependent resource type

- [Management Lock](management-lock.md)
## Settings

- `DeleteManaged`
- `DeleteLocks`
- `SoftDeleteKeyVaultKeys`
- `SoftDeleteKeyVaultKeysTag`
