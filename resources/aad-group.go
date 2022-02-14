package resources

import (
	"context"

	"github.com/manicminer/hamilton/auth"
	"github.com/manicminer/hamilton/msgraph"
	"github.com/manicminer/hamilton/odata"
	"github.com/sirupsen/logrus"

	"github.com/ekristen/azure-nuke/pkg/resource"
	"github.com/ekristen/azure-nuke/pkg/types"
)

type AzureAdGroup struct {
	client *msgraph.GroupsClient
	id     *string
	name   *string
}

func init() {
	resource.RegisterV2(resource.Registration{
		Name:   "AzureADGroup",
		Scope:  resource.Tenant,
		Lister: ListAzureADGroup,
	})
}

func ListAzureADGroup(opts resource.ListerOpts) ([]resource.Resource, error) {
	logrus.Tracef("subscription id: %s", opts.SubscriptionId)

	wrappedAuth, err := auth.NewAutorestAuthorizerWrapper(opts.Authorizers.Graph)
	if err != nil {
		return nil, err
	}

	client := msgraph.NewGroupsClient(opts.TenantId)
	client.BaseClient.Authorizer = wrappedAuth
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
			id:     group.ID,
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
