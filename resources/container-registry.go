package resources

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/containerregistry/mgmt/2019-05-01/containerregistry"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/nuke"
)

const ContainerRegistryResource = "ContainerRegistry"

func init() {
	resource.Register(&resource.Registration{
		Name:   ContainerRegistryResource,
		Scope:  nuke.ResourceGroup,
		Lister: &ContainerRegistryLister{},
	})
}

type ContainerRegistry struct {
	client        containerregistry.RegistriesClient
	name          *string
	resourceGroup *string
}

func (r *ContainerRegistry) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.resourceGroup, *r.name)
	return err
}

func (r *ContainerRegistry) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", *r.name)
	properties.Set("ResourceGroup", *r.resourceGroup)

	return properties
}

func (r *ContainerRegistry) String() string {
	return *r.name
}

type ContainerRegistryLister struct {
}

func (l ContainerRegistryLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	log := logrus.WithField("r", ContainerRegistryResource).WithField("s", opts.SubscriptionId)

	client := containerregistry.NewRegistriesClient(opts.SubscriptionId)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list container registries")

	list, err := client.List(ctx)
	if err != nil {
		return nil, err
	}

	log.Trace("listing resources")

	for list.NotDone() {
		log.Trace("list not done")
		for _, entity := range list.Values() {
			resources = append(resources, &ContainerRegistry{
				client:        client,
				name:          entity.Name,
				resourceGroup: &opts.ResourceGroup,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	log.Trace("done")

	return resources, nil
}
