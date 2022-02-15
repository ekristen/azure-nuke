# Azure Nuke

**This is potentially very destructive! Use at your own risk!**

**Status:** This is late alpha, early beta. Expect some behaviors around safeguarding, delays, and prompts to change. Likely will change CLI behavior a bit as well.

Originally based on the source code from [aws-nuke fork](https://github.com/ekristen/aws-nuke) and [aws-nuke original](https://github.com/rebuy-de/aws-nuke)

## Overview

This tool is designed to target an Azure Tenant and all subscriptions within the tenant and remove all resources from that tenant.

## Usage

By default no destructive actions will be taken.

```bash
azure-nuke nuke --tenant-id=00000000-0000-0000-0000-000000000000 --resource-id=api://11111111-1111-1111-1111-111111111111 --config=./config.yaml
```

To actually destroy you must add the `--no-dry-run` cli parameter.

```bash
azure-nuke nuke --tenant-id=00000000-0000-0000-0000-000000000000 --resource-id=api://11111111-1111-1111-1111-111111111111 --config=./config.yaml --no-dry-run
```

## Authentication

**Subject to Change**

Authentication is only supported via a Service Principal and you can authenticate via a `shared secret` or `certificate`.

### Shared Secret

```bash
export AZURE_CLIENT_ID=00000000-0000-0000-0000-000000000000
export AZURE_CLIENT_SECRET=000000000000
```

### Certificate

```bash
export AZURE_CLIENT_ID=00000000-0000-0000-0000-000000000000
export AZURE_CLIENT_CERTIFICATE_FILE=./path/to/cert
```

## Configuring

The entire configuration of the tool is done via a single YAML file.

### Example Configuration

**Note:** you must add at least one entry to the blocklist.

```yaml
tenant-blocklist:
  - 00001111-2222-3333-4444-555566667777

tenants:
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

## Azure Locations

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
