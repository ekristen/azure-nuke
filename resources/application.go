package resources

import (
	"context"
	"github.com/ekristen/azure-nuke/pkg/azure"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/hashicorp/go-azure-sdk/sdk/odata"
	"github.com/manicminer/hamilton/msgraph"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
)

const ApplicationResource = "Application"

func init() {
	registry.Register(&registry.Registration{
		Name:     ApplicationResource,
		Scope:    azure.TenantScope,
		Resource: &Application{},
		Lister:   &ApplicationLister{},
	})
}

type ApplicationLister struct {
}

func (l ApplicationLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	var resources []resource.Resource
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", ApplicationResource).WithField("s", opts.SubscriptionID)

	client := msgraph.NewApplicationsClient()
	client.BaseClient.Authorizer = opts.Authorizers.Graph
	client.BaseClient.DisableRetries = true

	log.Trace("attempting to list applications")

	entities, _, err := client.List(ctx, odata.Query{})
	if err != nil {
		return nil, err
	}

	log.Trace("listing applications")

	for i := range *entities {
		entity := &(*entities)[i]

		resources = append(resources, &Application{
			BaseResource: &BaseResource{
				Region: ptr.String("global"),
			},
			client: client,
			ID:     entity.ID(),
			Name:   entity.DisplayName,
		})
	}

	log.Trace("done")

	return resources, nil
}

type Application struct {
	*BaseResource `property:",inline"`

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
