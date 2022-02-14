package azure

import (
	"time"

	"github.com/Azure/go-autorest/autorest"
)

// authResult contains the subset of results from token acquisition operation in ConfidentialClientApplication
// For details see https://aka.ms/msal-net-authenticationresult
type authResult struct {
	accessToken    string
	expiresOn      time.Time
	grantedScopes  []string
	declinedScopes []string
}

// OAuthToken implements the OAuthTokenProvider interface.  It returns the current access token.
func (ar authResult) OAuthToken() string {
	return ar.accessToken
}

func (ar authResult) Token() string {
	return ar.accessToken
}

type Authorizers struct {
	Graph      autorest.Authorizer
	Management autorest.Authorizer
}
