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
		Name:     IPAllocationResource,
		Scope:    nuke.ResourceGroup,
		Resource: &IPAllocation{},
		Lister:   &IPAllocationLister{},
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
				client:        client,
				Region:        g.Location,
				ResourceGroup: &opts.ResourceGroup,
				Name:          g.Name,
				Tags:          g.Tags,
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

	Region        *string
	ResourceGroup *string
	Name          *string
	Tags          map[string]*string
}

func (r *IPAllocation) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.ResourceGroup, *r.Name)
	return err
}

func (r *IPAllocation) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *IPAllocation) String() string {
	return *r.Name
}
