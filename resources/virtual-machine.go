package resources

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2021-04-01/compute" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/nuke"
)

const VirtualMachineResource = "VirtualMachine"

func init() {
	registry.Register(&registry.Registration{
		Name:   VirtualMachineResource,
		Scope:  nuke.ResourceGroup,
		Lister: &VirtualMachineLister{},
	})
}

type VirtualMachine struct {
	client compute.VirtualMachinesClient

	Region        *string
	ResourceGroup *string
	Name          *string
	Tags          map[string]*string
}

func (r *VirtualMachine) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.ResourceGroup, *r.Name, &[]bool{true}[0])
	return err
}

func (r *VirtualMachine) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *VirtualMachine) String() string {
	return *r.Name
}

// -----------------------------------------

type VirtualMachineLister struct {
}

func (l VirtualMachineLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	log := logrus.WithField("r", VirtualMachineResource).WithField("s", opts.SubscriptionID)

	client := compute.NewVirtualMachinesClient(opts.SubscriptionID)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list virtual machines")

	list, err := client.List(ctx, opts.ResourceGroup)
	if err != nil {
		return nil, err
	}

	log.Trace("listing resources")

	for list.NotDone() {
		log.Trace("list not done")
		for _, g := range list.Values() {
			resources = append(resources, &VirtualMachine{
				client:        client,
				Name:          g.Name,
				ResourceGroup: &opts.ResourceGroup,
				Region:        g.Location,
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
