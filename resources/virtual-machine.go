package resources

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2021-04-01/compute"

	"github.com/ekristen/azure-nuke/pkg/resource"
	"github.com/ekristen/azure-nuke/pkg/types"
)

type VirtualMachine struct {
	client        compute.VirtualMachinesClient
	name          *string
	resourceGroup *string
}

func init() {
	resource.RegisterV2(resource.Registration{
		Name:   "VirtualMachine",
		Lister: ListVirtualMachine,
		Scope:  resource.ResourceGroup,
	})
}

func ListVirtualMachine(opts resource.ListerOpts) ([]resource.Resource, error) {
	logrus.Tracef("subscription id: %s", opts.SubscriptionId)

	client := compute.NewVirtualMachinesClient(opts.SubscriptionId)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	logrus.Trace("attempting to list virtual machines")

	ctx := context.Background()

	list, err := client.List(ctx, opts.ResourceGroup)
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
				resourceGroup: &opts.ResourceGroup,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	return resources, nil
}

func (r *VirtualMachine) Remove() error {
	_, err := r.client.Delete(context.TODO(), "Default", *r.name, &[]bool{true}[0])
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
