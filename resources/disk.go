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
		Name:      "Disk",
		Scope:     nuke.ResourceGroup,
		Lister:    DiskLister{},
		DependsOn: []string{"VirtualMachine"},
	})
}

type Disk struct {
	client compute.DisksClient
	name   string
	rg     string
}

func (r *Disk) Remove() error {
	_, err := r.client.Delete(context.TODO(), r.rg, r.name)
	return err
}

func (r *Disk) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", r.name)

	return properties
}

func (r *Disk) String() string {
	return r.name
}

type DiskLister struct {
	opts nuke.ListerOpts
}

func (l DiskLister) SetOptions(opts interface{}) {
	l.opts = opts.(nuke.ListerOpts)
}

func (l DiskLister) List() ([]resource.Resource, error) {
	logrus.Tracef("subscription id: %s", l.opts.SubscriptionId)

	client := compute.NewDisksClient(l.opts.SubscriptionId)
	client.Authorizer = l.opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	logrus.Trace("attempting to list ssh key")

	ctx := context.Background()
	list, err := client.ListByResourceGroup(ctx, l.opts.ResourceGroup)
	if err != nil {
		return nil, err
	}

	logrus.Trace("listing ....")

	for list.NotDone() {
		logrus.Trace("list not done")
		for _, g := range list.Values() {
			resources = append(resources, &Disk{
				client: client,
				name:   *g.Name,
				rg:     l.opts.ResourceGroup,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	return resources, nil
}
