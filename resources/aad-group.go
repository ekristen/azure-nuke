package resources

import (
	"context"
	"github.com/ekristen/azure-nuke/pkg/nuke"

	"github.com/hashicorp/go-azure-sdk/sdk/odata"
	"github.com/manicminer/hamilton/msgraph"
	"github.com/sirupsen/logrus"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
)

const AzureAdGroupResource = "AzureADGroup"

func init() {
	resource.Register(resource.Registration{
		Name:   AzureAdGroupResource,
		Scope:  nuke.Tenant,
		Lister: AzureAdGroupLister{},
	})
}

type AzureAdGroupLister struct {
}

func (l AzureAdGroupLister) List(o interface{}) ([]resource.Resource, error) {
	opts := o.(nuke.ListerOpts)

	log := logrus.WithField("r", AzureAdGroupResource).WithField("s", opts.SubscriptionId)

	client := msgraph.NewGroupsClient()
	client.BaseClient.Authorizer = opts.Authorizers.MicrosoftGraph
	client.BaseClient.DisableRetries = true

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list azure ad groups")

	ctx := context.Background()

	groups, _, err := client.List(ctx, odata.Query{})
	if err != nil {
		return nil, err
	}

	log.Trace("listing resources")

	for _, group := range *groups {
		resources = append(resources, &AzureAdGroup{
			client: client,
			id:     group.ID(),
			name:   group.DisplayName,
		})
	}

	return resources, nil
}

type AzureAdGroup struct {
	client *msgraph.GroupsClient
	id     *string
	name   *string
}

func (r *AzureAdGroup) Remove() error {
	_, err := r.client.Delete(context.TODO(), *r.id)
	return err
}

func (r *AzureAdGroup) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", *r.name)

	return properties
}

func (r *AzureAdGroup) String() string {
	return *r.id
}
