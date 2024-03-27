package resources

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-05-01/network" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/nuke"
)

const IPAllocationResource = "IPAllocation"

func init() {
	registry.Register(&registry.Registration{
		Name:   IPAllocationResource,
		Scope:  nuke.ResourceGroup,
		Lister: &IPAllocationLister{},
	})
}

type IPAllocationLister struct {
}

func (l IPAllocationLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	log := logrus.WithField("r", IPAllocationResource).WithField("s", opts.SubscriptionID)

	client := network.NewIPAllocationsClient(opts.SubscriptionID)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list virtual networks")

	list, err := client.ListByResourceGroup(ctx, opts.ResourceGroup)
	if err != nil {
		return nil, err
	}

	log.Trace("listing resources")

	for list.NotDone() {
		log.Trace("list not done")
		for _, g := range list.Values() {
			resources = append(resources, &IPAllocation{
				rg:     &opts.ResourceGroup,
				client: client,
				name:   g.Name,
				region: g.Location,
				tags:   g.Tags,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	log.Trace("done")

	return resources, nil
}

type IPAllocation struct {
	client network.IPAllocationsClient
	name   *string
	rg     *string
	region *string
	tags   map[string]*string
}

func (r *IPAllocation) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.rg, *r.name)
	return err
}

func (r *IPAllocation) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", r.name)
	properties.Set("ResourceGroup", r.rg)
	properties.Set("Region", r.region)

	for k, v := range r.tags {
		properties.SetTag(&k, v)
	}

	return properties
}

func (r *IPAllocation) String() string {
	return *r.name
}
