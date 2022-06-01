package azure

import (
	"github.com/Azure/go-autorest/autorest"
	"github.com/manicminer/hamilton/auth"
)

type Authorizers struct {
	Graph      auth.Authorizer
	Management autorest.Authorizer
}
