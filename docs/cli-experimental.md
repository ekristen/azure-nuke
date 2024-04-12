# Experimental Features

## Overview

These are the experimental features hidden behind feature flags that are currently available in aws-nuke. They are all
disabled by default. These are switches that changes the actual behavior of the tool itself. Changing the behavior of
a resource is done via resource settings.

!!! note
    The original tool had configuration options called `feature-flags` which were used to enable/disable certain
    behaviors with resources, those are now called settings and `feature-flags` have been deprecated in the config.

## Usage

```console
azure-nuke run --feature-flag "some-feature"
```

**Note:** other CLI arguments are omitted for brevity.

## Available Feature Flags

No available features at this time.