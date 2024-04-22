Remove all resources from an Azure Tenant and it's Subscriptions.

**azure-nuke** is stable, but it is likely that not all Azure resources are covered by it. Be encouraged to add missing
resources and create a Pull Request or to create an [Issue](https://github.com/ekristen/aws-nuke/issues/new).

!!! danger "Destructive Tool"
    Be aware that this is a very destructive tool, hence you have to be very careful while using it. Otherwise, 
    you might delete production data.

## What's New in Version 1

This is not a comprehensive list, but here are some of the highlights:

* New Feature: [Global Filters](features/global-filters.md)
* New Feature: [Run Against All Enabled Regions](features/enabled-regions.md)
* [Behavior of Filter for Regions](features/regions.md)
* Upcoming Feature: Filter Groups (**in progress**)
* Completely rewrote the core of the tool as a dedicated library [libnuke](https://github.com/ekristen/libnuke)
    * This library has over 95% test coverage which makes iteration and new features easier to implement.
* Semantic Releases with notifications on issues / pull requests
* Documentation for [all resources](resources/overview.md)
* New Resources

## Introducing libnuke

Officially over the Christmas break of 2023, I decided to create [libnuke](https://github.com/ekristen/libnuke) which
is a library that can be used to create similar tools for other cloud providers. This library is used by both this tool,
aws-nuke, and [azure-nuke](https://github.com/ekristen/azure-nuke) and soon [gcp-nuke](https://github.com/ekristen/gcp-nuke).

I also needed a version of this tool for Azure and GCP, and initially I just copied and altered the code I needed for
Azure, but I didn't want to have to maintain multiple copies of the same code, so I decided to create
[libnuke](https://github.com/ekristen/libnuke) to abstract all the code that was common between the two tools and write proper unit tests for it.

