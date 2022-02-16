package resources

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/containerregistry/mgmt/2019-05-01/containerregistry"

	"github.com/ekristen/azure-nuke/pkg/resource"
	"github.com/ekristen/azure-nuke/pkg/types"
)

func init() {
	resource.RegisterV2(resource.Registration{
		Name:   "ContainerRegistry",
		Lister: ListContainerRegistry,
		Scope:  resource.ResourceGroup,
	})
}

type ContainerRegistry struct {
	client        containerregistry.RegistriesClient
	name          *string
	resourceGroup *string
}

func (r *ContainerRegistry) Remove() error {
	_, err := r.client.Delete(context.TODO(), *r.resourceGroup, *r.name)
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

func ListContainerRegistry(opts resource.ListerOpts) ([]resource.Resource, error) {
	logrus.Tracef("subscription id: %s", opts.SubscriptionId)

	client := containerregistry.NewRegistriesClient(opts.SubscriptionId)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	logrus.Trace("attempting to list container registries")

	ctx := context.Background()

	list, err := client.List(ctx)
	if err != nil {
		return nil, err
	}

	logrus.Trace("listing ....")

	for list.NotDone() {
		logrus.Trace("list not done")
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

	return resources, nil
}
