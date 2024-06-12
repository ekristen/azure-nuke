package azure

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/hashicorp/go-azure-sdk/sdk/auth"
	"github.com/hashicorp/go-azure-sdk/sdk/auth/autorest"
)

type Authorizers struct {
	Graph      *autorest.Authorizer
	Management *autorest.Authorizer

	MicrosoftGraph  auth.Authorizer
	ResourceManager auth.Authorizer

	IdentityCreds azcore.TokenCredential
}
