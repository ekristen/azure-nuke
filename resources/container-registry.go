package resources

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/containerregistry/mgmt/2019-05-01/containerregistry" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/nuke"
)

const ContainerRegistryResource = "ContainerRegistry"

func init() {
	registry.Register(&registry.Registration{
		Name:     ContainerRegistryResource,
		Scope:    nuke.ResourceGroup,
		Resource: &ContainerRegistry{},
		Lister:   &ContainerRegistryLister{},
	})
}

type ContainerRegistry struct {
	client containerregistry.RegistriesClient

	Region        *string
	ResourceGroup *string
	Name          *string
	Tags          map[string]*string
}

func (r *ContainerRegistry) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.ResourceGroup, *r.Name)
	return err
}

func (r *ContainerRegistry) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ContainerRegistry) String() string {
	return *r.Name
}

type ContainerRegistryLister struct {
}

func (l ContainerRegistryLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	log := logrus.WithField("r", ContainerRegistryResource).WithField("s", opts.SubscriptionID)

	client := containerregistry.NewRegistriesClient(opts.SubscriptionID)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list container registries")

	list, err := client.ListByResourceGroup(ctx, opts.ResourceGroup)
	if err != nil {
		return nil, err
	}

	log.Trace("listing resources")

	for list.NotDone() {
		log.Trace("list not done")
		for _, entity := range list.Values() {
			resources = append(resources, &ContainerRegistry{
				client:        client,
				Region:        entity.Location,
				ResourceGroup: &opts.ResourceGroup,
				Name:          entity.Name,
				Tags:          entity.Tags,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	log.Trace("done")

	return resources, nil
}
