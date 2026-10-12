# azure-nuke

[![license](https://img.shields.io/github/license/ekristen/azure-nuke.svg)](https://github.com/ekristen/azure-nuke/blob/main/LICENSE)
[![release](https://img.shields.io/github/release/ekristen/azure-nuke.svg)](https://github.com/ekristen/azure-nuke/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/ekristen/azure-nuke)](https://goreportcard.com/report/github.com/ekristen/azure-nuke)
[![Maintainability](https://api.codeclimate.com/v1/badges/51b67f545bfb93ecab2f/maintainability)](https://codeclimate.com/github/ekristen/azure-nuke/maintainability)

**This is potentially very destructive! Use at your own risk!**

**Status:** Beta. Tool is stable, but could experience odd behaviors with some resources.

## Overview

Remove all resources from an Azure Tenant and its Subscriptions.

**azure-nuke** is in beta, but it is likely that not all Azure resources are covered by it. Be encouraged to add missing
resources and create a Pull Request or to create an [Issue](https://github.com/ekristen/azure-nuke/issues/new).

## Documentation

All documentation is in the [docs/](docs) directory and is built using [Material for Mkdocs](https://squidfunk.github.io/mkdocs-material/).

It is hosted at [https://ekristen.github.io/azure-nuke/](https://ekristen.github.io/azure-nuke/).

## Authentication

The tenant is set with `--tenant-id` (`AZURE_TENANT_ID`), the subscriptions to nuke with `--subscription-id`
(`AZURE_SUBSCRIPTION_ID`, can be repeated; without it every subscription the identity can see, except blocklisted ones).

### Azure CLI (Recommended)

Used when none of the service principal methods below is configured. Not available in the container image.

```bash
az login --tenant <tenant-id>
```

For the service principal methods, the first one that is configured is used:

### Service Principal with Client Secret

```bash
export AZURE_CLIENT_ID=<app-id>
export AZURE_CLIENT_SECRET=<secret>
```

### Service Principal with Client Certificate

The file must be a PKCS#12 (`.pfx`) bundle with the private key, no password and legacy encryption (PEM files and
OpenSSL 3 default bundles are rejected):

```bash
openssl pkcs12 -export -legacy -in cert.pem -inkey key.pem -out cert.pfx -passout pass:
export AZURE_CLIENT_ID=<app-id>
export AZURE_CLIENT_CERTIFICATE_FILE=/path/to/cert.pfx
```

### Workload Identity (GitHub Actions OIDC, AKS, etc.)

```bash
export AZURE_CLIENT_ID=<app-id>
export AZURE_FEDERATED_TOKEN_FILE=/path/to/token
```

## Quick Start (Docker)

**1. Create config file** (`nuke-config.yaml`):

```yaml
regions:
  - global # Nuke tenant and subscription scoped resources
  - eastus # Nuke resources in the eastus region

resource-types:
  excludes: # Keep Entra ID (tenant scoped) objects
    - AzureADUser
    - AzureADGroup
    - Application
    - ServicePrincipal

settings:
  ResourceGroup:
    DeleteManaged: true # Also delete resource groups managed by another resource (e.g. AKS node resource groups)

blocklist:
  - 00000000-0000-0000-0000-000000000000 # Never nuke this tenant or subscription

accounts: # i.e. Azure tenants to nuke
  11111111-1111-1111-1111-111111111111:
    filters:
      # Protect a resource group by name and every resource group tagged keep=true
      ResourceGroup:
        - 'shared-dns-eastus'
        - property: tag:keep
          value: 'true'

      # Protect locks in the protected resource group
      ManagementLock:
        - property: ResourceGroup
          value: 'shared-dns-eastus'

      # Protect key vaults with name containing "prod"
      KeyVault:
        - property: Name
          type: contains
          value: 'prod'

      # Protect virtual networks with prefix
      VirtualNetwork:
        - property: Name
          type: glob
          value: 'prod-*'
```

**2. Create a service principal** with access to the subscription and export its credentials:

```bash
export AZURE_TENANT_ID=11111111-1111-1111-1111-111111111111
export AZURE_SUBSCRIPTION_ID=22222222-2222-2222-2222-222222222222
export AZURE_CLIENT_ID=<app-id>
export AZURE_CLIENT_SECRET=<secret>
```

**3. Dry run (preview what will be deleted):**

```bash
docker run --rm -it \
  -e AZURE_TENANT_ID -e AZURE_SUBSCRIPTION_ID -e AZURE_CLIENT_ID -e AZURE_CLIENT_SECRET \
  -v "$(pwd)/nuke-config.yaml:/nuke-config.yaml:ro" \
  ghcr.io/ekristen/azure-nuke:v1.2.9 run \
  --config /nuke-config.yaml
```

**4. Nuke everything (DESTRUCTIVE):**

```bash
docker run --rm -it \
  -e AZURE_TENANT_ID -e AZURE_SUBSCRIPTION_ID -e AZURE_CLIENT_ID -e AZURE_CLIENT_SECRET \
  -v "$(pwd)/nuke-config.yaml:/nuke-config.yaml:ro" \
  ghcr.io/ekristen/azure-nuke:v1.2.9 run \
  --config /nuke-config.yaml \
  --no-dry-run
```

## Attribution, License, and Copyright

This tool was written using [libnuke](https://github.com/ekristen/libnuke) at it's core. It shares similarities and commonalities with [aws-nuke](https://github.com/ekristen/aws-nuke)
and [gcp-nuke](https://github.com/ekristen/gcp-nuke). It was originally put together by taking various components of
[rebuy-de/aws-nuke](https://github.com/rebuy-de/aws-nuke) and rewriting them to work with Azure. These tools would not have
been possible without the hard work that came before me on the original tool by the team and contributors over at
[rebuy-de](https://github.com/rebuy-de) and their original work on [rebuy-de/aws-nuke](https://github.com/rebuy-de/aws-nuke).

This tool is licensed under the MIT license. See the [LICENSE](LICENSE) file for more information. The bulk of this
tool was rewritten to use [libnuke](https://github.com/ekristen/libnuke) which was in part originally sourced from [rebuy-de/aws-nuke](https://github.com/rebuy-de/aws-nuke).
