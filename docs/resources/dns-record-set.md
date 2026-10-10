# DNS Record Set

## Details

- **Type:** `DNSRecordSet`
- **Scope:** subscription

## Properties

- **`BaseResource`**: No description provided
- **`ID`**: The ID of the record set.
- **`Name`**: The name of the record set relative to the zone, @ for the apex.
- **`TTL`**: The TTL of the records.
- **`Type`**: The record type, e.g. A, CNAME, NS.
- **`Zone`**: The DNS zone the record set belongs to.
- **`tag:<key>:`**: This resource has tags with property `Tags`. These are key/value pairs that are
	added as their own property with the prefix of `tag:` (e.g. [tag:example: "value"]) 
