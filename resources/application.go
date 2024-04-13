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

const ApplicationResource = "Application"

func init() {
	registry.Register(&registry.Registration{
		Name:   ApplicationResource,
		Scope:  nuke.Tenant,
		Lister: &ApplicationLister{},
	})
}

type ApplicationLister struct {
}

func (l ApplicationLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	log := logrus.WithField("r", ApplicationResource).WithField("s", opts.SubscriptionID)

	client := msgraph.NewApplicationsClient()
	client.BaseClient.Authorizer = opts.Authorizers.Graph
	client.BaseClient.DisableRetries = true

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list applications")

	entities, _, err := client.List(ctx, odata.Query{})
	if err != nil {
		return nil, err
	}

	log.Trace("listing applications")

	for i := range *entities {
		entity := &(*entities)[i]

		resources = append(resources, &Application{
			client: client,
			ID:     entity.ID(),
			Name:   entity.DisplayName,
		})
	}

	log.Trace("done")

	return resources, nil
}

type Application struct {
	client *msgraph.ApplicationsClient
	ID     *string
	Name   *string
}

func (r *Application) Filter() error {
	return nil
}

func (r *Application) Remove(ctx context.Context) error {
	if _, err := r.client.Delete(ctx, *r.ID); err != nil {
		return err
	}

	if _, err := r.client.DeletePermanently(context.TODO(), *r.ID); err != nil {
		return err
	}

	return nil
}

func (r *Application) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *Application) String() string {
	return *r.Name
}
