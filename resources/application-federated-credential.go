package resources

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/hashicorp/go-azure-sdk/sdk/odata"
	"github.com/manicminer/hamilton/msgraph"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/nuke"
)

const ApplicationFederatedCredentialResource = "ApplicationFederatedCredential"

func init() {
	registry.Register(&registry.Registration{
		Name:   ApplicationFederatedCredentialResource,
		Scope:  nuke.Tenant,
		Lister: &ApplicationFederatedCredentialLister{},
	})
}

type ApplicationFederatedCredential struct {
	client        *msgraph.ApplicationsClient
	ID            *string
	Name          *string
	AppID         *string
	AppUniqueName *string
}

func (r *ApplicationFederatedCredential) Filter() error {
	return nil
}

func (r *ApplicationFederatedCredential) Remove(ctx context.Context) error {
	_, err := r.client.DeleteFederatedIdentityCredential(ctx, *r.AppID, *r.ID)
	return err
}

func (r *ApplicationFederatedCredential) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ApplicationFederatedCredential) String() string {
	return *r.Name
}

type ApplicationFederatedCredentialLister struct {
}

func (l ApplicationFederatedCredentialLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	log := logrus.WithField("r", ApplicationFederatedCredentialResource).WithField("s", opts.SubscriptionID)

	client := msgraph.NewApplicationsClient()
	client.BaseClient.Authorizer = opts.Authorizers.Graph
	client.BaseClient.DisableRetries = true

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list application federated creds")

	entities, _, err := client.List(ctx, odata.Query{})
	if err != nil {
		return nil, err
	}

	log.Trace("listing application federated creds")

	for i := range *entities {
		entity := &(*entities)[i]

		creds, _, err := client.ListFederatedIdentityCredentials(ctx, *entity.ID(), odata.Query{})
		if err != nil {
			return nil, err
		}

		for _, cred := range *creds {
			resources = append(resources, &ApplicationFederatedCredential{
				client:        client,
				ID:            cred.ID,
				Name:          cred.Name,
				AppID:         entity.ID(),
				AppUniqueName: entity.UniqueName,
			})
		}
	}

	log.Trace("done")

	return resources, nil
}
