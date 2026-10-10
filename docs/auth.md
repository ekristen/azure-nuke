# Authentication

The tenant is set with `--tenant-id` (`AZURE_TENANT_ID`) and the subscriptions to nuke with `--subscription-id`
(`AZURE_SUBSCRIPTION_ID`, can be repeated; without it every subscription the identity can see, except blocklisted ones).

## Azure CLI

The Azure CLI login (`az login --tenant <tenant-id>`) is used when none of the service principal flags below is set. The
container image does not include the Azure CLI.

## CLI Flags (or environment variables)

The following service principal flags are available for authentication, the first one that is set is used:

- `--client-id` - the client-id to use for authentication 
- `--client-secret` - the client-secret to use for authentication
- `--client-certificate-file` - the client-certificate-file to use for authentication
- `--client-federated-token-file` - the client-federated-token-file to use for authentication

`--client-id` is required with any of the other three.

The client certificate file must be a PKCS#12 (`.pfx`) bundle with the private key, no password and legacy encryption,
e.g. `openssl pkcs12 -export -legacy -in cert.pem -inkey key.pem -out cert.pfx -passout pass:`.

## Environment Variables

- `AZURE_CLIENT_ID`
- `AZURE_CLIENT_SECRET`
- `AZURE_CLIENT_CERTIFICATE_FILE`
- `AZURE_FEDERATED_TOKEN_FILE`
