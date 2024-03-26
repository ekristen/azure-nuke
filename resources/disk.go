package resources

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2021-04-01/compute"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/nuke"
)

const DiskResource = "Disk"

func init() {
	registry.Register(&registry.Registration{
		Name:   DiskResource,
		Scope:  nuke.Subscription,
		Lister: &DiskLister{},
		DependsOn: []string{
			VirtualMachineResource,
		},
	})
}

type Disk struct {
	client compute.DisksClient
	name   string
	rg     string
}

func (r *Disk) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, r.rg, r.name)
	return err
}

func (r *Disk) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", r.name)
	properties.Set("ResourceGroup", r.rg)

	return properties
}

func (r *Disk) String() string {
	return r.name
}

type DiskLister struct {
}

func (l DiskLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	log := logrus.WithField("r", DiskResource).WithField("s", opts.SubscriptionId)

	client := compute.NewDisksClient(opts.SubscriptionId)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list disks")

	list, err := client.ListByResourceGroup(ctx, opts.ResourceGroup)
	if err != nil {
		return nil, err
	}

	log.Trace("listing ....")

	for list.NotDone() {
		log.Trace("list not done")
		for _, g := range list.Values() {
			resources = append(resources, &Disk{
				client: client,
				name:   *g.Name,
				rg:     opts.ResourceGroup,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	log.Trace("done")

	return resources, nil
}
