package azure

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Azure/go-autorest/autorest"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
)

func AcquireTokenClientCertificate(ctx context.Context, tenantID, resource string) func(tenantID, resource string) (*autorest.BearerAuthorizer, error) {
	return func(tenantID, resource string) (*autorest.BearerAuthorizer, error) {
		clientID := os.Getenv("AZURE_CLIENT_ID")
		certFilePath := os.Getenv("AZURE_CLIENT_CERTIFICATE_FILE")
		certFilePwd := os.Getenv("AZURE_CLIENT_CERTIFICATE_PASSWORD")
		authorityHost := os.Getenv("AZURE_AUTHORITY_HOST")

		// trim the suffix / if exists
		resource = strings.TrimSuffix(resource, "/")
		// .default needs to be added to the scope
		if !strings.HasSuffix(resource, ".default") {
			resource += "/.default"
		}

		scopes := []string{resource}

		pemData, err := ioutil.ReadFile(certFilePath)
		if err != nil {
			return nil, err
		}

		// This extracts our public certificates and private key from the PEM file.
		// The private key must be in PKCS8 format. If it is encrypted, the second argument
		// must be password to decode.
		certs, privateKey, err := confidential.CertFromPEM(pemData, certFilePwd)
		if err != nil {
			return nil, err
		}

		// PEM files can have multiple certs. This is usually for certificate chaining where roots
		// sign to leafs. Useful for TLS, not for this use case.
		if len(certs) > 1 {
			return nil, fmt.Errorf("too many certificates in PEM file")
		}

		cred := confidential.NewCredFromCert(certs[0], privateKey)
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
		}

		return autorest.NewBearerAuthorizer(authResult{
			accessToken:    result.AccessToken,
			expiresOn:      result.ExpiresOn,
			grantedScopes:  result.GrantedScopes,
			declinedScopes: result.DeclinedScopes,
		}), nil
	}

}
