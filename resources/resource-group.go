package resources

import (
	"context"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources/v4"

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

	client *armresources.ResourceGroupsClient
	Name   *string            `description:"The Name of the resource group."`
	Tags   map[string]*string `description:"The tags assigned to the resource group."`
}

func (r *ResourceGroup) Remove(ctx context.Context) error {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancel()

	poller, err := r.client.BeginDelete(ctx, *r.Name, nil)
	if err != nil {
		return err
	}

	_, err = poller.PollUntilDone(ctx, nil)
	return err
}

func (r *ResourceGroup) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ResourceGroup) String() string {
	return *r.Name
}

// -------------------

type ResourceGroupLister struct{}

func (l ResourceGroupLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancel()

	log := logrus.WithField("r", ResourceGroupResource).WithField("s", opts.SubscriptionID)

	client, err := armresources.NewResourceGroupsClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list groups")

	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, entity := range page.Value {
			resources = append(resources, &ResourceGroup{
				BaseResource: &BaseResource{
					Region:         entity.Location,
					SubscriptionID: ptr.String(opts.SubscriptionID),
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
