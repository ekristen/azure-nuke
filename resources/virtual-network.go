package resources

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-05-01/network"

	"github.com/ekristen/azure-nuke/pkg/resource"
	"github.com/ekristen/azure-nuke/pkg/types"
)

type VirtualNetwork struct {
	client network.VirtualNetworksClient
	name   *string
	rg     *string
}

func init() {
	resource.RegisterV2(resource.Registration{
		Name:   "VirtualNetwork",
		Lister: ListVirtualNetwork,
		Scope:  resource.ResourceGroup,
	})
}

func ListVirtualNetwork(opts resource.ListerOpts) ([]resource.Resource, error) {
	log := logrus.WithField("handler", "ListVirtualNetwork").WithField("subscription", opts.SubscriptionId)

	client := network.NewVirtualNetworksClient(opts.SubscriptionId)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list virtual networks")

	ctx := context.Background()

	list, err := client.List(ctx, opts.ResourceGroup)
	if err != nil {
		return nil, err
	}

	log.Trace("listing virtual networks")

	for list.NotDone() {
		for _, g := range list.Values() {
			resources = append(resources, &VirtualNetwork{
				client: client,
				name:   g.Name,
				rg:     &opts.ResourceGroup,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	return resources, nil
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
