# azure-nuke

[![license](https://img.shields.io/github/license/ekristen/azure-nuke.svg)](https://github.com/ekristen/azure-nuke/blob/main/LICENSE)
[![release](https://img.shields.io/github/release/ekristen/azure-nuke.svg)](https://github.com/ekristen/azure-nuke/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/ekristen/azure-nuke)](https://goreportcard.com/report/github.com/ekristen/azure-nuke)
[![Maintainability](https://api.codeclimate.com/v1/badges/51b67f545bfb93ecab2f/maintainability)](https://codeclimate.com/github/ekristen/azure-nuke/maintainability)

**This is potentially very destructive! Use at your own risk!**

## v1 (beta) is out

Please test, get started right now.

```console
brew install ekristen/tap/azure-nuke@1
```

Or grab from the [beta releases](https://github.com/ekristen/azure-nuke/releases?q=beta&expanded=true)

### Supported OS/Architectures

- Darwin AMD64/ARM64
- Linux AMD64/ARM64/ARM
- Windows AMD64

## Overview

Remove all resources from an Azure Tenant and it's Subscriptions.

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
