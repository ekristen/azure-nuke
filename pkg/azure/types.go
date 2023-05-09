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

/*
ref: https://github.com/Azure/go-autorest/issues/252
func LogRequestPreparer() autorest.PrepareDecorator {
	return func(p autorest.Preparer) autorest.Preparer {
		return autorest.PreparerFunc(func(r *http.Request) (*http.Request, error) {
			resDump, _ := httputil.DumpRequestOut(r, true)
			log.Println(string(resDump))
			return r, nil
		})
	}
}
*/
