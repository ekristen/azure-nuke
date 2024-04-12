# Install

## Install the pre-compiled binary 

### Homebrew Tap (MacOS/Linux)

```console
brew install ekristen/tap/azure-nuke@1
```

!!! warning "Brew Warning"
    azure-nuke is NOT on homebrew due to their new ridiculous policy of requiring 75 stars or more.

## Releases

You can download pre-compiled binaries from the [releases](https://github.com/ekristen/azure-nuke/releases) page.

## Docker

Registries:

- [ghcr.io/ekristen/azure-nuke](https://github.com/ekristen/azure-nuke/pkgs/container/azure-nuke)

You can run **azure-nuke** with Docker by using a command like this:

## Source

To compile **azure-nuke** from source you need a working [Golang](https://golang.org/doc/install) development environment and [goreleaser](https://goreleaser.com/install/).

**azure-nuke** uses go modules and so the clone path should not matter. Then simply change directory into the clone and run:

```bash
goreleaser build --clean --snapshot --single-target
```

