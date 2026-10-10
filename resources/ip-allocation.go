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

const IPAllocationResource = "IPAllocation"

func init() {
	registry.Register(&registry.Registration{
		Name:     IPAllocationResource,
		Scope:    azure.ResourceGroupScope,
		Resource: &IPAllocation{},
		Lister:   &IPAllocationLister{},
	})
}

type IPAllocationLister struct{}

func (l IPAllocationLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", IPAllocationResource).WithField("s", opts.SubscriptionID)

	client, err := armnetwork.NewIPAllocationsClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list ip allocations")

	pager := client.NewListByResourceGroupPager(opts.ResourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, entity := range page.Value {
			resources = append(resources, &IPAllocation{
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

type IPAllocation struct {
	*BaseResource `property:",inline"`

	client   *armnetwork.IPAllocationsClient
	removeOp *runtime.Poller[armnetwork.IPAllocationsClientDeleteResponse]
	Name     *string
	Tags     map[string]*string
}

func (r *IPAllocation) Remove(ctx context.Context) error {
	poller, err := r.client.BeginDelete(ctx, *r.ResourceGroup, *r.Name, nil)
	if err != nil {
		return err
	}

	r.removeOp = poller
	return nil
}

func (r *IPAllocation) HandleWait(ctx context.Context) error {
	return azure.WaitForPoller(ctx, r.removeOp)
}

func (r *IPAllocation) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *IPAllocation) String() string {
	return *r.Name
}
