package resources

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-05-01/network" //nolint:staticcheck

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

type VirtualNetworkLister struct {
}

func (l VirtualNetworkLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", VirtualNetworkResource).WithField("s", opts.SubscriptionID)

	client := network.NewVirtualNetworksClient(opts.SubscriptionID)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list virtual networks")

	list, err := client.List(ctx, opts.ResourceGroup)
	if err != nil {
		return nil, err
	}

	log.Trace("listing virtual networks")

	for list.NotDone() {
		for _, g := range list.Values() {
			resources = append(resources, &VirtualNetwork{
				BaseResource: &BaseResource{
					Region:         g.Location,
					ResourceGroup:  &opts.ResourceGroup,
					SubscriptionID: &opts.SubscriptionID,
				},
				client: client,
				Name:   g.Name,
				Tags:   g.Tags,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	log.Trace("debug")

	return resources, nil
}

// ---------------------------------------------

type VirtualNetwork struct {
	*BaseResource `property:",inline"`

	client network.VirtualNetworksClient
	Name   *string
	Tags   map[string]*string
}

func (r *VirtualNetwork) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.ResourceGroup, *r.Name)
	return err
}

func (r *VirtualNetwork) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *VirtualNetwork) String() string {
	return *r.Name
}
