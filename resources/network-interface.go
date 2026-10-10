package resources

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v12"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const NetworkInterfaceResource = "NetworkInterface"

func init() {
	registry.Register(&registry.Registration{
		Name:     NetworkInterfaceResource,
		Scope:    azure.ResourceGroupScope,
		Resource: &NetworkInterface{},
		Lister:   &NetworkInterfaceLister{},
	})
}

type NetworkInterfaceLister struct{}

func (l NetworkInterfaceLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancel()

	log := logrus.WithField("r", NetworkInterfaceResource).WithField("s", opts.SubscriptionID)

	resources := make([]resource.Resource, 0)

	client, err := armnetwork.NewInterfacesClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return resources, err
	}

	log.Trace("attempting to list network interfaces")

	pager := client.NewListPager(opts.ResourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, entity := range page.Value {
			resources = append(resources, &NetworkInterface{
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

	log.Trace("done listing network interfaces")

	return resources, nil
}

type NetworkInterface struct {
	*BaseResource `property:",inline"`

	client   *armnetwork.InterfacesClient
	removeOp *runtime.Poller[armnetwork.InterfacesClientDeleteResponse]
	Name     *string
	Tags     map[string]*string
}

func (r *NetworkInterface) Remove(ctx context.Context) error {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancel()

	r.removeOp = nil

	poller, err := r.client.BeginDelete(ctx, *r.ResourceGroup, *r.Name, nil)
	if err != nil {
		return azure.IgnoreNotFound(err)
	}

	r.removeOp = poller
	return nil
}

func (r *NetworkInterface) HandleWait(ctx context.Context) error {
	return azure.WaitForPoller(ctx, r.removeOp)
}

func (r *NetworkInterface) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *NetworkInterface) String() string {
	return *r.Name
}
