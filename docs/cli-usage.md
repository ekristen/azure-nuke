# Usage

## azure-nuke

```console
NAME:
   azure-nuke - remove everything from an azure tenant

USAGE:
   azure-nuke [global options] command [command options] [arguments...]

VERSION:
   1.0.0

AUTHOR:
   Erik Kristensen <erik@erikkristensen.com>

COMMANDS:
   run, nuke                       run nuke against an azure tenant to remove all configured resources
   resource-types, list-resources  list available resources to nuke
   help, h                         Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)

```

## aws-nuke run

```console
NAME:
   azure-nuke run - run nuke against an azure tenant to remove all configured resources

USAGE:
   azure-nuke run [command options] [arguments...]

OPTIONS:
   --config value                             path to config file (default: "config.yaml")
   --include value                            only include this specific resource
   --exclude value                            exclude this specific resource (this overrides everything)
   --quiet, -q                                hide filtered messages (default: false)
   --no-dry-run                               actually run the removal of the resources after discovery (default: false)
   --no-prompt, --force                       disable prompting for verification to run (default: false)
   --prompt-delay value, --force-sleep value  seconds to delay after prompt before running (minimum: 3 seconds) (default: 10)
   --feature-flag value                       enable experimental behaviors that may not be fully tested or supported
   --environment value                        Azure Environment (default: "global") [$AZURE_ENVIRONMENT]
   --tenant-id value                          the tenant-id to nuke [$AZURE_TENANT_ID]
   --subscription-id value                    the subscription-id to nuke (this filters to 1 or more subscription ids) [$AZURE_SUBSCRIPTION_ID]
   --client-id value                          the client-id to use for authentication [$AZURE_CLIENT_ID]
   --client-secret value                      the client-secret to use for authentication [$AZURE_CLIENT_SECRET]
   --client-certificate-file value            the client-certificate-file to use for authentication [$AZURE_CLIENT_CERTIFICATE_FILE]
   --client-federated-token-file value        the client-federated-token-file to use for authentication [$AZURE_FEDERATED_TOKEN_FILE]
   --log-level value, -l value                Log Level (default: "info") [$LOGLEVEL]
   --log-caller                               log the caller (aka line number and file) (default: false)
   --log-disable-color                        disable log coloring (default: false)
   --log-full-timestamp                       force log output to always show full timestamp (default: false)
   --help, -h                                 show help (default: false)
```
