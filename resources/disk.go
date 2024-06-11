package resources

import (
	"context"
	"github.com/ekristen/azure-nuke/pkg/azure"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2021-04-01/compute" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
)

const DiskResource = "Disk"

func init() {
	registry.Register(&registry.Registration{
		Name:     DiskResource,
		Scope:    azure.ResourceGroupScope,
		Lister:   &DiskLister{},
		Resource: &Disk{},
		DependsOn: []string{
			VirtualMachineResource,
		},
	})
}

type Disk struct {
	*BaseResource `property:",inline"`

	client compute.DisksClient
	Name   *string
	Tags   map[string]*string
}

func (r *Disk) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.ResourceGroup, *r.Name)
	return err
}

func (r *Disk) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *Disk) String() string {
	return *r.Name
}

type DiskLister struct {
}

func (l DiskLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", DiskResource).WithField("s", opts.SubscriptionID)

	client := compute.NewDisksClient(opts.SubscriptionID)
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
		for _, r := range list.Values() {
			resources = append(resources, &Disk{
				BaseResource: &BaseResource{
					Region:         r.Location,
					ResourceGroup:  &opts.ResourceGroup,
					SubscriptionID: &opts.SubscriptionID,
				},
				client: client,
				Name:   r.Name,
				Tags:   r.Tags,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	log.Trace("done")

	return resources, nil
}
