package resources

import (
	"context"
	"github.com/ekristen/azure-nuke/pkg/nuke"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-05-01/network"

	"github.com/ekristen/cloud-nuke-sdk/pkg/resource"
	"github.com/ekristen/cloud-nuke-sdk/pkg/types"
)

func init() {
	resource.Register(resource.Registration{
		Name:   "VirtualNetwork",
		Lister: VirtualNetworkLister{},
		Scope:  nuke.ResourceGroup,
	})
}

type VirtualNetwork struct {
	client network.VirtualNetworksClient
	name   *string
	rg     *string
}

func (r *VirtualNetwork) Remove() error {
	_, err := r.client.Delete(context.TODO(), *r.rg, *r.name)
	return err
}

func (r *VirtualNetwork) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", *r.name)
	properties.Set("ResourceGroup", *r.rg)

	return properties
}

func (r *VirtualNetwork) String() string {
	return *r.name
}

// ---------------------------------------------

type VirtualNetworkLister struct {
	opts nuke.ListerOpts
}

func (l VirtualNetworkLister) SetOptions(opts interface{}) {
	l.opts = opts.(nuke.ListerOpts)
}

func (l VirtualNetworkLister) List() ([]resource.Resource, error) {
	log := logrus.WithField("handler", "ListVirtualNetwork").WithField("subscription", l.opts.SubscriptionId)

	client := network.NewVirtualNetworksClient(l.opts.SubscriptionId)
	client.Authorizer = l.opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list virtual networks")

	ctx := context.Background()

	list, err := client.List(ctx, l.opts.ResourceGroup)
	if err != nil {
		return nil, err
	}

	log.Trace("listing virtual networks")

	for list.NotDone() {
		for _, g := range list.Values() {
			resources = append(resources, &VirtualNetwork{
				client: client,
				name:   g.Name,
				rg:     &l.opts.ResourceGroup,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	return resources, nil
}
