# azure-nuke

[![license](https://img.shields.io/github/license/ekristen/azure-nuke.svg)](https://github.com/ekristen/azure-nuke/blob/main/LICENSE)
[![release](https://img.shields.io/github/release/ekristen/azure-nuke.svg)](https://github.com/ekristen/azure-nuke/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/ekristen/azure-nuke)](https://goreportcard.com/report/github.com/ekristen/azure-nuke)
[![Maintainability](https://api.codeclimate.com/v1/badges/51b67f545bfb93ecab2f/maintainability)](https://codeclimate.com/github/ekristen/azure-nuke/maintainability)

**This is potentially very destructive! Use at your own risk!**

**Status:** This is early beta. Expect some behaviors around safeguarding, delays, and prompts to change.
Likely will change CLI behavior a bit as well.

## v1 (beta) is out

Please test, get started right now.

```console
brew install ekristen/tap/azure-nuke@1
```

Or grab from the [next releases](https://github.com/ekristen/azure-nuke/releases?q=next&expanded=true)

## Overview

Remove all resources from an Azure Tenant and it's Subscriptions and Resource Groups.

**azure-nuke** is stable, but it is likely that not all Azure resources are covered by it. Be encouraged to add missing
resources and create a Pull Request or to create an [Issue](https://github.com/ekristen/azure-nuke/issues/new).

## Documentation

All documentation is in the [docs/](docs) directory and is built using [Material for Mkdocs](https://squidfunk.github.io/mkdocs-material/).

It is hosted at [https://ekristen.github.io/azure-nuke/](https://ekristen.github.io/azure-nuke/).

## History

This was originally put together by taking various components of [rebuy-de/aws-nuke](https://github.com/rebuy-de/aws-nuke)
and rewriting them to work with Azure. I forked the original aws-nuke after attempting to make contributions and respond
to issues to learn that the current maintainers only have time to work on the project about once a month and while
receptive to bringing in other people to help maintain, made it clear it would take time. Considering the feedback cycle
was already weeks on initial communications, I had to make the hard decision to fork and maintain it.

I then decided in December 2023 to write [libnuke](https://github.com/ekristen/libnuke) so that I could rewrite
my fork of aws-nuke to use it and then allow rewriting azure-nuke and gcp-nuke to use it as well.

### libnuke

I also needed a version of this tool for Azure and GCP, and initially I just copied and altered the code I needed for
Azure, but I didn't want to have to maintain multiple copies of the same code, so I decided to create
[libnuke](https://github.com/ekristen/libnuke) to abstract all the code that was common between the two tools and write
proper unit tests for it.

## Attribution, License, and Copyright

The rewrite of this tool to use [libnuke](https://github.com/ekristen/libnuke) would not have been possible without the hard work that came before me
on the original tool by the team and contributors over at [rebuy-de](https://github.com/rebuy-de) and their original work on [rebuy-de/aws-nuke](https://github.com/rebuy-de/aws-nuke).

This tool is licensed under the MIT license. See the [LICENSE](LICENSE) file for more information. The bulk of this
tool was rewritten to use [libnuke](https://github.com/ekristen/libnuke) which was in part originally sourced from [rebuy-de/aws-nuke](https://github.com/rebuy-de/aws-nuke).

## Usage

**Note:** all cli flags can also be expressed as environment variables.

**By default, no destructive actions will be taken.**

Due to how Azure Authentication works, there's no way to determine the tenant ID and must be explicitly given. This is
done via the `--tenant-id` cli flag.

### Example - Dry Run only

```bash
azure-nuke run \
  --tenant-id=00000000-0000-0000-0000-000000000000 \
  --config=test-config.yaml
```

### Example - No Dry Run (DESTRUCTIVE)

To actually destroy you must add the `--no-dry-run` cli parameter.

```bash
azure-nuke run \
  --tenant-id=00000000-0000-0000-0000-000000000000 \
  --config=test-config.yaml \
  --no-dry-run
```

## Authentication

Authentication is only supported via a Service Principal and you can authenticate via a `shared secret`, `certificate`, or `federated token (kubernetes)`

### Shared Secret

```bash
export AZURE_CLIENT_ID=00000000-0000-0000-0000-000000000000
export AZURE_CLIENT_SECRET=000000000000
```

### Certificate

```bash
export AZURE_CLIENT_ID=00000000-0000-0000-0000-000000000000
export AZURE_CLIENT_CERTIFICATE=""
export AZURE_CLIENT_PRIVATE_KEY=""
```

### Federated Token (Kubernetes)

You can also authenticate using Federated Tokens with Kubernetes and the Azure Workload Identity.

To make this work you'll need to deploy azure-nuke with a Service Account that's configured to do federation with the Service Principal.

## Configuring

The entire configuration of the tool is done via a single YAML file.

### Example Configuration

**Note:** you must add at least one entry to the blocklist.

```yaml
regions:
  - global
  - eastus

blocklist:
  - 00001111-2222-3333-4444-555566667777

accounts:
  77776666-5555-4444-3333-222211110000:
    presets:
      - common
    filters:
      AzureADUser:
        - property: Name
          type: contains
          value: ImportantUser
      ServicePrincipal:
        - type: contains
          property: Name
          value: testing-azure-nuke

presets:
  common:
    filters:
      ResourceGroup:
        - Default
        - NetworkWatcherRG
```

## Azure Locations (aka Regions in the config)

- global **this is not an actual location but represents the tenant, as in global resources**
- eastus
- eastus2
- southcentralus
- westus2
- westus3
- australiaeast
- southeastasia
- northeurope
- swedencentral
- uksouth
- westeurope
- centralus
- northcentralus
- westus
- southafricanorth
- centralindia
- eastasia
- japaneast
- jioindiawest
- koreacentral
- canadacentral
- francecentral
- germanywestcentral
- norwayeast
- switzerlandnorth
- uaenorth
- brazilsouth
- centralusstage
- eastusstage
- eastus2stage
- northcentralusstage
- southcentralusstage
- westusstage
- westus2stage
- asia
- asiapacific
- australia
- brazil
- canada
- europe
- france
- germany
- global
- india
- japan
- korea
- norway
- southafrica
- switzerland
- uae
- uk
- unitedstates
- unitedstateseuap
- eastasiastage
- southeastasiastage
- centraluseuap
- eastus2euap
- westcentralus
- southafricawest
- australiacentral
- australiacentral2
- australiasoutheast
- japanwest
- jioindiacentral
- koreasouth
- southindia
- westindia
- canadaeast
- francesouth
- germanynorth
- norwaywest
- switzerlandwest
- ukwest
- uaecentral
- brazilsoutheast
