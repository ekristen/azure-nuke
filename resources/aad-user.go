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

const AzureADUserResource = "AzureADUser"

func init() {
	registry.Register(&registry.Registration{
		Name:   AzureADUserResource,
		Scope:  nuke.Tenant,
		Lister: &AzureADUserLister{},
		DependsOn: []string{
			AzureAdGroupResource,
		},
	})
}

type AzureADUserLister struct {
}

func (l AzureADUserLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	log := logrus.WithField("r", AzureADUserResource).WithField("s", opts.SubscriptionID)

	client := msgraph.NewUsersClient()
	client.BaseClient.Authorizer = opts.Authorizers.Graph
	client.BaseClient.DisableRetries = true

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list azure ad users")

	entities, _, err := client.List(ctx, odata.Query{})
	if err != nil {
		return nil, err
	}

	log.Trace("listing resources")

	for i := range *entities {
		entity := &(*entities)[i]

		resources = append(resources, &AzureADUser{
			client: client,
			id:     entity.ID(),
			name:   entity.DisplayName,
			upn:    entity.UserPrincipalName,
		})
	}

	return resources, nil
}

type AzureADUser struct {
	client *msgraph.UsersClient
	id     *string
	name   *string
	upn    *string
}

func (r *AzureADUser) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.id)
	return err
}

func (r *AzureADUser) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("ID", r.id)
	properties.Set("Name", r.name)
	properties.Set("UserPrincipalName", r.upn)

	return properties
}

func (r *AzureADUser) String() string {
	return *r.name
}
