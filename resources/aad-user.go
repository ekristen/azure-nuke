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

type AzureADUser struct {
	client *msgraph.UsersClient
	id     *string
	name   *string
}

func init() {
	resource.RegisterV2(resource.Registration{
		Name:      "AzureADUser",
		Scope:     resource.Tenant,
		Lister:    ListAzureADUser,
		DependsOn: []string{"AzureADGroup"},
	})
}

func ListAzureADUser(opts resource.ListerOpts) ([]resource.Resource, error) {
	logrus.Tracef("subscription id: %s", opts.SubscriptionId)

	wrappedAuth, err := auth.NewAutorestAuthorizerWrapper(opts.Authorizers.Graph)
	if err != nil {
		return nil, err
	}

	client := msgraph.NewUsersClient(opts.TenantId)
	client.BaseClient.Authorizer = wrappedAuth
	client.BaseClient.DisableRetries = true

	resources := make([]resource.Resource, 0)

	logrus.Trace("attempting to list ssh key")

	ctx := context.Background()

	users, _, err := client.List(ctx, odata.Query{})
	if err != nil {
		return nil, err
	}

	logrus.Trace("listing ....")

	for _, user := range *users {
		resources = append(resources, &AzureADUser{
			client: client,
			id:     user.ID,
			name:   user.DisplayName,
		})
	}

	return resources, nil
}

func (r *AzureADUser) Remove() error {
	_, err := r.client.Delete(context.TODO(), *r.id)
	return err
}

func (r *AzureADUser) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", *r.name)

	return properties
}

func (r *AzureADUser) String() string {
	return *r.id
}
