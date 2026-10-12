package resources

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v12"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const VirtualNetworkResource = "VirtualNetwork"

func init() {
	registry.Register(&registry.Registration{
		Name:     VirtualNetworkResource,
		Scope:    azure.ResourceGroupScope,
		Resource: &VirtualNetwork{},
		Lister:   &VirtualNetworkLister{},
	})
}

type VirtualNetworkLister struct{}

func (l VirtualNetworkLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", VirtualNetworkResource).WithField("s", opts.SubscriptionID)

	client, err := armnetwork.NewVirtualNetworksClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list virtual networks")

	pager := client.NewListPager(opts.ResourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, entity := range page.Value {
			resources = append(resources, &VirtualNetwork{
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

// ---------------------------------------------

type VirtualNetwork struct {
	*BaseResource `property:",inline"`

	client   *armnetwork.VirtualNetworksClient
	removeOp *runtime.Poller[armnetwork.VirtualNetworksClientDeleteResponse]
	Name     *string
	Tags     map[string]*string
}

func (r *VirtualNetwork) Remove(ctx context.Context) error {
	r.removeOp = nil

	poller, err := r.client.BeginDelete(ctx, *r.ResourceGroup, *r.Name, nil)
	if err != nil {
		return azure.IgnoreNotFound(err)
	}

	r.removeOp = poller
	return nil
}

func (r *VirtualNetwork) HandleWait(ctx context.Context) error {
	return azure.WaitForPoller(ctx, r.removeOp)
}

func (r *VirtualNetwork) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *VirtualNetwork) String() string {
	return *r.Name
}
