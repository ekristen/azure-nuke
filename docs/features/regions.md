# Regions as Global Filters

Azure is broken down into regions, these regions are where resources are created. Like most things Azure there is a mix
in naming nomenclature, sometimes these are referred to as "locations" and sometimes they are referred to as "regions".
To keep things consistent with [libnuke](https://github.com/ekristen/libnuke) we will refer to them as regions.

## Behavior

Due to how Azure APIs are designed and unlike AWS, there is no easy or simple way to limit a query to a specific region.
Some APIs support it, some state they do, but do not, therefore this tool uses the [regions configuration](../config.md#regions)
to filter resources automatically based on region. 