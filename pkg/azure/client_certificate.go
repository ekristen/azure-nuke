package azure

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/go-autorest/autorest"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
)

func AcquireTokenClientCertificate(ctx context.Context, tenantID, resource string) func(tenantID, resource string) (*autorest.BearerAuthorizer, error) {
	return func(tenantID, resource string) (*autorest.BearerAuthorizer, error) {
		clientID := os.Getenv("AZURE_CLIENT_ID")
		certPem := os.Getenv("AZURE_CLIENT_CERTIFICATE")
		keyPem := os.Getenv("AZURE_CLIENT_PRIVATE_KEY")
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

		certBlock, _ := pem.Decode([]byte(certPem))
		if certBlock == nil {
			return nil, fmt.Errorf("failed to parse certificate PEM")
		}
		cert, err := x509.ParseCertificate(certBlock.Bytes)
		if err != nil {
			return nil, err
		}

		keyBlock, _ := pem.Decode([]byte(keyPem))
		if keyBlock == nil {
			return nil, fmt.Errorf("failed to parse certificate PEM")
		}
		privateKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
		if err != nil {
			return nil, err
		}

		cred := confidential.NewCredFromCert(cert, privateKey)
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
