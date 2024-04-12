# First Run

## First Configuration

First you need to create a config file for *aws-nuke*. This is a minimal one:

```yaml
regions:
  - global
  - us-east-1

blocklist:
- "00000000-0000-0000-0000-000000000000" # production

accounts:
  "11111111-1111-1111-1111-111111111111": {} # azure-nuke-example
```

## First Run (Dry Run)

With this config we can run *aws-nuke*:

```bash
$ azure-nuke run -c config/nuke-config.yaml
azure-nuke version v1.0.0-dev - Fri Jul 28 16:26:41 CEST 2017 - c2f318f37b7d2dec0e646da3d4d05ab5296d5bce

Would delete these resources. Provide --no-dry-run to actually destroy resources.
```

As we see, *aws-nuke* only lists all found resources and exits. This is because the `--no-dry-run` flag is missing.
Also, it wants to delete the administrator. We don't want to do this, because we use this user to access our account.
Therefore, we have to extend the config, so it ignores this user:

```yaml
regions:
  - global
  - us-east-1

blocklist:
  - "00000000-0000-0000-0000-000000000000" # production

accounts:
  "11111111-1111-1111-1111-111111111111": # azure-nuke-example
    filters:
      ResourceGroup:
        - Default
```

## Second Run (No Dry Run)

!!! warning
This will officially remove resources from your AWS account. Make sure you really want to do this!

```bash
$ azure-nuke nuke -c config/nuke-config.yml --no-dry-run
azure-nuke version v1.0.0-dev - Fri Jul 28 16:26:41 CEST 2017 - c2f318f37b7d2dec0e646da3d4d05ab5296d5bce

Do you really want to nuke the account with the ID 000000000000 and the alias 'azure-nuke-example'?
Do you want to continue? Enter account alias to continue.
> azure-nuke-example


Do you really want to nuke these resources on the account with the ID 000000000000 and the alias 'azure-nuke-example'?
Do you want to continue? Enter account alias to continue.
> aws-nuke-example


--- truncating long output ---
```

As you see *azure-nuke* now tries to delete all resources which aren't filtered, without caring about the dependencies
between them. This results in API errors which can be ignored. These errors are shown at the end of the *azure-nuke* run,
if they keep to appear.

*azure-nuke* retries deleting all resources until all specified ones are deleted or until there are only resources
with errors left.

