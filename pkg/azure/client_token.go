package azure

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/Azure/go-autorest/autorest"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
)

func AcquireTokenFederatedToken(ctx context.Context, tenantID string) func(tenantID, resource string) (*autorest.BearerAuthorizer, error) {
	return func(tenantID, resource string) (*autorest.BearerAuthorizer, error) {
		// Azure AD Workload Identity webhook will inject the following env vars
		// 	AZURE_CLIENT_ID with the clientID set in the service account annotation
		// 	AZURE_TENANT_ID with the tenantID set in the service account annotation. If not defined, then
		// 	the tenantID provided via azure-wi-webhook-config for the webhook will be used.
		// 	AZURE_FEDERATED_TOKEN_FILE is the service account token path
		// 	AZURE_AUTHORITY_HOST is the AAD authority hostname
		clientID := os.Getenv("AZURE_CLIENT_ID")
		tokenFilePath := os.Getenv("AZURE_FEDERATED_TOKEN_FILE")
		authorityHost := "https://login.microsoftonline.com"
		if v := os.Getenv("AZURE_AUTHORITY_HOST"); v != "" {
			authorityHost = v
		}
		if v := os.Getenv("AZURE_TENANT_ID"); v != "" {
			tenantID = v
		}

		// trim the suffix / if exists
		resource = strings.TrimSuffix(resource, "/")
		// .default needs to be added to the scope
		if !strings.HasSuffix(resource, ".default") {
			resource += "/.default"
		}

		scopes := []string{resource}

		// generate a token using the msal confidential client
		// this will always generate a new token request to AAD
		// TODO (aramase) consider using acquire token silent (https://github.com/Azure/azure-workload-identity/issues/76)

		// read the service account token from the filesystem
		signedAssertion, err := readJWTFromFS(tokenFilePath)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read service account token")
		}
		cred, err := confidential.NewCredFromAssertion(signedAssertion)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create confidential creds")
		}

		// create the confidential client to request an AAD token
		confidentialClientApp, err := confidential.New(
			clientID,
			cred,
			confidential.WithAuthority(fmt.Sprintf("%s%s/oauth2/token", authorityHost, tenantID)))
		if err != nil {
			return nil, errors.Wrap(err, "failed to create confidential client app")
		}

		result, err := confidentialClientApp.AcquireTokenByCredential(ctx, scopes)
		if err != nil {
			return nil, errors.Wrap(err, "failed to acquire token")
		}

		return autorest.NewBearerAuthorizer(authResult{
			accessToken:    result.AccessToken,
			expiresOn:      result.ExpiresOn,
			grantedScopes:  result.GrantedScopes,
			declinedScopes: result.DeclinedScopes,
		}), nil
	}
}

// readJWTFromFS reads the jwt from file system
func readJWTFromFS(tokenFilePath string) (string, error) {
	token, err := os.ReadFile(tokenFilePath)
	if err != nil {
		return "", err
	}
	return string(token), nil
}
