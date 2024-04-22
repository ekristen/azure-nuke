package azure

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"

	"github.com/hashicorp/go-azure-sdk/sdk/auth"
	"github.com/hashicorp/go-azure-sdk/sdk/auth/autorest"
	"github.com/hashicorp/go-azure-sdk/sdk/environments"
)

func ConfigureAuth(
	ctx context.Context,
	environment, tenantID, clientID, clientSecret, clientCertFile, clientFedTokenFile string) (*Authorizers, error) {
	env, err := environments.FromName(environment)
	if err != nil {
		return nil, err
	}

	authorizers := &Authorizers{}

	credentials := auth.Credentials{
		Environment: *env,
		TenantID:    tenantID,
		ClientID:    clientID,

		EnableAuthenticatingUsingClientSecret: true,
	}

	if clientSecret != "" {
		logrus.Debug("authentication type: client secret")
		credentials.EnableAuthenticatingUsingClientSecret = true
		credentials.ClientSecret = clientSecret

		creds, err := azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret, &azidentity.ClientSecretCredentialOptions{})
		if err != nil {
			return nil, err
		}
		authorizers.IdentityCreds = creds
	} else if clientCertFile != "" {
		logrus.Debug("authentication type: client certificate")
		credentials.EnableAuthenticatingUsingClientCertificate = true
		credentials.ClientCertificatePath = clientCertFile

		certData, err := os.ReadFile(clientCertFile)
		if err != nil {
			return nil, err
		}

		certs, pkey, err := azidentity.ParseCertificates(certData, nil)
		if err != nil {
			return nil, err
		}

		creds, err := azidentity.NewClientCertificateCredential(tenantID, clientID, certs, pkey, &azidentity.ClientCertificateCredentialOptions{})
		if err != nil {
			return nil, err
		}
		authorizers.IdentityCreds = creds
	} else if clientFedTokenFile != "" {
		logrus.Debug("authentication type: federated token")
		token, err := os.ReadFile(clientFedTokenFile)
		if err != nil {
			return nil, err
		}
		credentials.EnableAuthenticationUsingOIDC = true
		credentials.OIDCAssertionToken = string(token)

		creds, err := azidentity.NewWorkloadIdentityCredential(&azidentity.WorkloadIdentityCredentialOptions{
			ClientID:      clientID,
			TenantID:      tenantID,
			TokenFilePath: clientFedTokenFile,
		})
		if err != nil {
			return nil, err
		}
		authorizers.IdentityCreds = creds
	}

	graphAuthorizer, err := auth.NewAuthorizerFromCredentials(ctx, credentials, env.MicrosoftGraph)
	if err != nil {
		return nil, err
	}

	mgmtAuthorizer, err := auth.NewAuthorizerFromCredentials(ctx, credentials, env.ResourceManager)
	if err != nil {
		return nil, err
	}

	authorizers.Management = autorest.AutorestAuthorizer(mgmtAuthorizer)
	authorizers.Graph = autorest.AutorestAuthorizer(graphAuthorizer)

	authorizers.MicrosoftGraph = graphAuthorizer
	authorizers.ResourceManager = mgmtAuthorizer

	return authorizers, nil
}
