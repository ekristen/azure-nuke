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

const ApplicationSecretResource = "ApplicationSecret"

func init() {
	registry.Register(&registry.Registration{
		Name:   ApplicationSecretResource,
		Scope:  nuke.Tenant,
		Lister: &ApplicationSecretLister{},
	})
}

type ApplicationSecret struct {
	client  *msgraph.ApplicationsClient
	ID      *string
	Name    *string
	AppID   *string
	AppName *string
}

func (r *ApplicationSecret) Filter() error {
	return nil
}

func (r *ApplicationSecret) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.ID)
	return err
}

func (r *ApplicationSecret) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ApplicationSecret) String() string {
	return *r.Name
}

type ApplicationSecretLister struct {
}

func (l ApplicationSecretLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	log := logrus.WithField("r", ApplicationSecretResource).WithField("s", opts.SubscriptionID)

	client := msgraph.NewApplicationsClient()
	client.BaseClient.Authorizer = opts.Authorizers.Graph
	client.BaseClient.DisableRetries = true

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list application secrets")

	entities, _, err := client.List(ctx, odata.Query{})
	if err != nil {
		return nil, err
	}

	log.Trace("listing application secrets")

	for i := range *entities {
		entity := &(*entities)[i]

		for _, cred := range *entity.PasswordCredentials {
			resources = append(resources, &ApplicationSecret{
				client:  client,
				ID:      cred.KeyId,
				Name:    cred.DisplayName,
				AppID:   entity.ID(),
				AppName: entity.DisplayName,
			})
		}
	}

	log.Trace("done")

	return resources, nil
}
