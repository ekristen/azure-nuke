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
		Name:   "Application",
		Scope:  nuke.Tenant,
		Lister: ApplicationLister{},
	})
}

type ApplicationLister struct {
}

func (l ApplicationLister) List(o interface{}) ([]resource.Resource, error) {
	opts := o.(nuke.ListerOpts)

	logrus.Tracef("subscription id: %s", opts.SubscriptionId)

	client := msgraph.NewApplicationsClient()
	client.BaseClient.Authorizer = opts.Authorizers.Graph
	client.BaseClient.DisableRetries = true

	resources := make([]resource.Resource, 0)

	logrus.Trace("attempting to list service principals")

	ctx := context.Background()

	entities, _, err := client.List(ctx, odata.Query{})
	if err != nil {
		return nil, err
	}

	logrus.Trace("listing ....")

	for _, entity := range *entities {
		resources = append(resources, &Application{
			client: client,
			id:     entity.ID(),
			name:   entity.DisplayName,
		})
	}

	return resources, nil
}

type Application struct {
	client *msgraph.ApplicationsClient
	id     *string
	name   *string
}

func (r *Application) Filter() error {
	return nil
}

func (r *Application) Remove() error {
	if _, err := r.client.Delete(context.TODO(), *r.id); err != nil {
		return err
	}

	if _, err := r.client.DeletePermanently(context.TODO(), *r.id); err != nil {
		return err
	}

	return nil
}

func (r *Application) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", *r.name)

	return properties
}

func (r *Application) String() string {
	return *r.id
}
