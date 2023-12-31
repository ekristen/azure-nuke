package resources

import (
	"context"
	"github.com/ekristen/azure-nuke/pkg/nuke"

	"github.com/hashicorp/go-azure-sdk/sdk/odata"
	"github.com/manicminer/hamilton/msgraph"
	"github.com/sirupsen/logrus"

	"github.com/ekristen/cloud-nuke-sdk/pkg/resource"
	"github.com/ekristen/cloud-nuke-sdk/pkg/types"
)

func init() {
	resource.Register(resource.Registration{
		Name:   "AzureADGroup",
		Scope:  nuke.Tenant,
		Lister: AzureAdGroupLister{},
	})
}

type AzureAdGroup struct {
	client *msgraph.GroupsClient
	id     *string
	name   *string
}

type AzureAdGroupLister struct {
}

func (l AzureAdGroupLister) List(o interface{}) ([]resource.Resource, error) {
	opts := o.(nuke.ListerOpts)

	logrus.Tracef("subscription id: %s", opts.SubscriptionId)

	client := msgraph.NewGroupsClient()
	client.BaseClient.Authorizer = opts.Authorizers.MicrosoftGraph
	client.BaseClient.DisableRetries = true

	resources := make([]resource.Resource, 0)

	logrus.Trace("attempting to list ssh key")

	ctx := context.Background()

	groups, _, err := client.List(ctx, odata.Query{})
	if err != nil {
		return nil, err
	}

	logrus.Trace("listing ....")

	for _, group := range *groups {
		resources = append(resources, &AzureAdGroup{
			client: client,
			id:     group.ID(),
			name:   group.DisplayName,
		})
	}

	return resources, nil
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
