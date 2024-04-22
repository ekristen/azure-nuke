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

const AzureAdGroupResource = "AzureADGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:     AzureAdGroupResource,
		Scope:    nuke.Tenant,
		Resource: AzureAdGroup{},
		Lister:   &AzureAdGroupLister{},
	})
}

type AzureAdGroupLister struct {
}

func (l AzureAdGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	log := logrus.WithField("r", AzureAdGroupResource).WithField("s", opts.SubscriptionID)

	client := msgraph.NewGroupsClient()
	client.BaseClient.Authorizer = opts.Authorizers.MicrosoftGraph
	client.BaseClient.DisableRetries = true

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list azure ad groups")

	ctx := context.Background()

	entities, _, err := client.List(ctx, odata.Query{})
	if err != nil {
		return nil, err
	}

	log.Trace("listing resources")

	for i := range *entities {
		entity := &(*entities)[i]

		resources = append(resources, &AzureAdGroup{
			client: client,
			ID:     entity.ID(),
			Name:   entity.DisplayName,
		})
	}

	return resources, nil
}

type AzureAdGroup struct {
	client *msgraph.GroupsClient
	ID     *string `description:"The ID of the Entra ID Group"`
	Name   *string `description:"The name of the Entra ID Group"`
}

func (r *AzureAdGroup) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.ID)
	return err
}

func (r *AzureAdGroup) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *AzureAdGroup) String() string {
	return *r.Name
}
