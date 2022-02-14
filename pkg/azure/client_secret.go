package azure

import (
	"context"
	"os"
	"strings"

	"github.com/Azure/go-autorest/autorest"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
)

func AcquireTokenClientSecret(ctx context.Context, tenantID, resource string) func(tenantID, resource string) (*autorest.BearerAuthorizer, error) {
	return func(tenantID, resource string) (*autorest.BearerAuthorizer, error) {
		clientID := os.Getenv("AZURE_CLIENT_ID")
		clientSecret := os.Getenv("AZURE_CLIENT_SECRET")
		authorityHost := os.Getenv("AZURE_AUTHORITY_HOST")

		// trim the suffix / if exists
		resource = strings.TrimSuffix(resource, "/")
		// .default needs to be added to the scope
		if !strings.HasSuffix(resource, ".default") {
			resource += "/.default"
		}

		scopes := []string{resource}

		cred, err := confidential.NewCredFromSecret(clientSecret)
		if err != nil {
			return nil, err
		}

		// app, err := confidential.New(config.ClientID, cred, confidential.WithAuthority(config.Authority), confidential.WithAccessor(cacheAccessor))

		app, err := confidential.New(clientID, cred, confidential.WithAuthority(authorityHost+"/"+tenantID))
		if err != nil {
			return nil, err
		}

		result, err := app.AcquireTokenSilent(ctx, scopes)
		if err != nil {
			result, err = app.AcquireTokenByCredential(ctx, scopes)
			if err != nil {
				return nil, err
			}
			//fmt.Println("Access Token Is " + result.AccessToken)
		}
		//fmt.Println("Silently acquired token " + result.AccessToken)

		return autorest.NewBearerAuthorizer(authResult{
			accessToken:    result.AccessToken,
			expiresOn:      result.ExpiresOn,
			grantedScopes:  result.GrantedScopes,
			declinedScopes: result.DeclinedScopes,
		}), nil
	}
}
