package resources

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry/v3"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const ContainerRegistryResource = "ContainerRegistry"

func init() {
	registry.Register(&registry.Registration{
		Name:     ContainerRegistryResource,
		Scope:    azure.ResourceGroupScope,
		Resource: &ContainerRegistry{},
		Lister:   &ContainerRegistryLister{},
	})
}

type ContainerRegistry struct {
	*BaseResource `property:",inline"`

	client   *armcontainerregistry.RegistriesClient
	removeOp *runtime.Poller[armcontainerregistry.RegistriesClientDeleteResponse]
	Name     *string
	Tags     map[string]*string
}

func (r *ContainerRegistry) Remove(ctx context.Context) error {
	poller, err := r.client.BeginDelete(ctx, *r.ResourceGroup, *r.Name, nil)
	if err != nil {
		return err
	}

	r.removeOp = poller
	return nil
}

func (r *ContainerRegistry) HandleWait(ctx context.Context) error {
	return azure.WaitForPoller(ctx, r.removeOp)
}

func (r *ContainerRegistry) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ContainerRegistry) String() string {
	return *r.Name
}

type ContainerRegistryLister struct{}

func (l ContainerRegistryLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)
	var resources []resource.Resource

	log := logrus.WithField("r", ContainerRegistryResource).WithField("s", opts.SubscriptionID)

	client, err := armcontainerregistry.NewRegistriesClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	log.Trace("attempting to list container registries")

	pager := client.NewListByResourceGroupPager(opts.ResourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, entity := range page.Value {
			resources = append(resources, &ContainerRegistry{
				BaseResource: &BaseResource{
					Region:         entity.Location,
					ResourceGroup:  &opts.ResourceGroup,
					SubscriptionID: &opts.SubscriptionID,
				},
				client: client,
				Name:   entity.Name,
				Tags:   entity.Tags,
			})
		}
	}

	log.Trace("done")

	return resources, nil
}
