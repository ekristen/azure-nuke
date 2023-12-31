package resources

import (
	"context"
	"github.com/ekristen/azure-nuke/pkg/nuke"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2021-04-01/compute"

	"github.com/ekristen/cloud-nuke-sdk/pkg/resource"
	"github.com/ekristen/cloud-nuke-sdk/pkg/types"
)

func init() {
	resource.Register(resource.Registration{
		Name:   "VirtualMachine",
		Lister: VirtualMachineLister{},
		Scope:  nuke.ResourceGroup,
	})
}

type VirtualMachine struct {
	client        compute.VirtualMachinesClient
	name          *string
	resourceGroup *string
}

func (r *VirtualMachine) Remove() error {
	_, err := r.client.Delete(context.TODO(), *r.resourceGroup, *r.name, &[]bool{true}[0])
	return err
}

func (r *VirtualMachine) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", *r.name)
	properties.Set("ResourceGroup", *r.resourceGroup)

	return properties
}

func (r *VirtualMachine) String() string {
	return *r.name
}

// -----------------------------------------

type VirtualMachineLister struct {
	opts nuke.ListerOpts
}

func (l VirtualMachineLister) SetOptions(opts interface{}) {
	l.opts = opts.(nuke.ListerOpts)
}

func (l VirtualMachineLister) List() ([]resource.Resource, error) {
	logrus.Tracef("subscription id: %s", l.opts.SubscriptionId)

	client := compute.NewVirtualMachinesClient(l.opts.SubscriptionId)
	client.Authorizer = l.opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	logrus.Trace("attempting to list virtual machines")

	ctx := context.Background()

	list, err := client.List(ctx, l.opts.ResourceGroup)
	if err != nil {
		return nil, err
	}

	logrus.Trace("listing ....")

	for list.NotDone() {
		logrus.Trace("list not done")
		for _, g := range list.Values() {
			resources = append(resources, &VirtualMachine{
				client:        client,
				name:          g.Name,
				resourceGroup: &l.opts.ResourceGroup,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	return resources, nil
}
