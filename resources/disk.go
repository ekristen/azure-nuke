package resources

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2021-04-01/compute"

	"github.com/ekristen/azure-nuke/pkg/resource"
	"github.com/ekristen/azure-nuke/pkg/types"
)

type Disk struct {
	client compute.DisksClient
	name   *string
}

func init() {
	resource.RegisterV2(resource.Registration{
		Name:      "Disk",
		Scope:     resource.ResourceGroup,
		Lister:    ListDisk,
		DependsOn: []string{"VirtualMachine"},
	})
}

func ListDisk(opts resource.ListerOpts) ([]resource.Resource, error) {
	logrus.Tracef("subscription id: %s", opts.SubscriptionId)

	client := compute.NewDisksClient(opts.SubscriptionId)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	logrus.Trace("attempting to list ssh key")

	ctx := context.Background()
	list, err := client.ListByResourceGroup(ctx, opts.ResourceGroup)
	if err != nil {
		return nil, err
	}

	logrus.Trace("listing ....")

	for list.NotDone() {
		logrus.Trace("list not done")
		for _, g := range list.Values() {
			resources = append(resources, &Disk{
				client: client,
				name:   g.Name,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	return resources, nil
}

func (r *Disk) Remove() error {
	_, err := r.client.Delete(context.TODO(), "Default", *r.name)
	return err
}

func (r *Disk) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", *r.name)

	return properties
}

func (r *Disk) String() string {
	return *r.name
}
