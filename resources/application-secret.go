package resources

import (
	"context"

	"github.com/manicminer/hamilton/msgraph"
	"github.com/manicminer/hamilton/odata"
	"github.com/sirupsen/logrus"

	"github.com/ekristen/azure-nuke/pkg/resource"
	"github.com/ekristen/azure-nuke/pkg/types"
)

func init() {
	resource.RegisterV2(resource.Registration{
		Name:   "ApplicationSecret",
		Scope:  resource.Tenant,
		Lister: ListApplicationSecret,
	})
}

type ApplicationSecret struct {
	client  *msgraph.ApplicationsClient
	id      *string
	name    *string
	appId   *string
	appName *string
}

func (r *ApplicationSecret) Filter() error {
	return nil
}

func (r *ApplicationSecret) Remove() error {
	_, err := r.client.Delete(context.TODO(), *r.id)
	return err
}

func (r *ApplicationSecret) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", r.name)
	properties.Set("AppName", r.appName)

	return properties
}

func (r *ApplicationSecret) String() string {
	return *r.id
}

func ListApplicationSecret(opts resource.ListerOpts) ([]resource.Resource, error) {
	logrus.Tracef("subscription id: %s", opts.SubscriptionId)

	client := msgraph.NewApplicationsClient(opts.TenantId)
	client.BaseClient.Authorizer = opts.Authorizers.Graph
	client.BaseClient.DisableRetries = true

	resources := make([]resource.Resource, 0)

	logrus.Trace("attempting to list service principals")

	ctx := context.Background()

	entites, _, err := client.List(ctx, odata.Query{})
	if err != nil {
		return nil, err
	}

	logrus.Trace("listing ....")

	for _, entity := range *entites {
		for _, cred := range *entity.PasswordCredentials {
			resources = append(resources, &ApplicationSecret{
				client:  client,
				id:      cred.KeyId,
				name:    cred.DisplayName,
				appId:   entity.ID,
				appName: entity.DisplayName,
			})
		}
	}

	return resources, nil
}
