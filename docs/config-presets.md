# Presets

It might be the case that some filters are the same across multiple accounts. This especially could happen, if
provisioning tools like Terraform are used or if IAM resources follow the same pattern.

For this case *aws-nuke* supports presets of filters, that can applied on multiple accounts.

`Presets` are defined globally. They can then be referenced in the `accounts` section of the configuration.

A preset configuration could look like this:

```yaml
presets:
  common:
    filters:
      ResourceGroup:
        - Default
```

An account referencing a preset would then look something like this:

```yaml
accounts:
  11111111-1111-1111-1111-111111111111:
    presets:
      - common
```

Putting it all together it would look something like this:

```yaml
blocklist:
  - 00000000-0000-0000-0000-000000000000

regions:
  - global
  - us-east-1

accounts:
  11111111-1111-1111-1111-111111111111:
    presets:
      - common

presets:
  common:
    filters:
      ResourceGroup:
        - Default
```