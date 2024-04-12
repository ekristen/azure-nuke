# Examples

## Basic usage

```bash
azure-nuke run --config config.yml
```

## Using the force flags

!!! danger
    Running without prompts can be dangerous. Make sure you understand what you are doing before using these flags.

The following is an example of how you automate the command to run without any prompts of the user. This is useful
for running in a CI/CD pipeline.

```bash
azure-nuke run --config config.yml --force --force-delay 5
```