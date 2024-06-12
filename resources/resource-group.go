package resources

import (
	"context"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"
	"github.com/hashicorp/go-azure-sdk/resource-manager/resources/2022-09-01/resourcegroups"
	"github.com/hashicorp/go-azure-sdk/sdk/environments"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const ResourceGroupResource = "ResourceGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:     ResourceGroupResource,
		Scope:    azure.SubscriptionScope,
		Resource: &ResourceGroup{},
		Lister:   &ResourceGroupLister{},
	})
}

// ResourceGroup represents an Azure Resource Group.
type ResourceGroup struct {
	*BaseResource `property:",inline"`

	client *resourcegroups.ResourceGroupsClient
	Name   *string            `description:"The Name of the resource group."`
	Tags   *map[string]string `description:"The tags assigned to the resource group."`
}

func (r *ResourceGroup) Remove(ctx context.Context) error {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancel()

	_, err := r.client.Delete(ctx,
		commonids.NewResourceGroupID(*r.SubscriptionID, *r.Name),
		resourcegroups.DefaultDeleteOperationOptions())
	return err
}

func (r *ResourceGroup) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ResourceGroup) String() string {
	return *r.Name
}

// -------------------

type ResourceGroupLister struct {
}

func (l ResourceGroupLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancel()

	log := logrus.WithField("r", ResourceGroupResource).WithField("s", opts.SubscriptionID)

	client, err := resourcegroups.NewResourceGroupsClientWithBaseURI(environments.AzurePublic().ResourceManager)
	if err != nil {
		return nil, err
	}
	client.Client.Authorizer = opts.Authorizers.Management

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list groups")

	list, err := client.List(ctx, commonids.NewSubscriptionID(opts.SubscriptionID), resourcegroups.ListOperationOptions{})
	if err != nil {
		return nil, err
	}

	log.Trace("listing ....")

	for _, entity := range *list.Model {
		resources = append(resources, &ResourceGroup{
			BaseResource: &BaseResource{
				Region:         ptr.String(entity.Location),
				SubscriptionID: ptr.String(opts.SubscriptionID),
			},
			client: client,
			Name:   entity.Name,
			Tags:   entity.Tags,
		})
	}

	log.Trace("done")

	return resources, nil
}
