package resources

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v8"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
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

	client       *armcompute.DisksClient
	removeOp     *runtime.Poller[armcompute.DisksClientDeleteResponse]
	Name         *string
	Tags         map[string]*string
	CreationDate *time.Time
}

func (r *Disk) Remove(ctx context.Context) error {
	r.removeOp = nil

	poller, err := r.client.BeginDelete(ctx, *r.ResourceGroup, *r.Name, nil)
	if err != nil {
		return azure.IgnoreNotFound(err)
	}

	r.removeOp = poller
	return nil
}

func (r *Disk) HandleWait(ctx context.Context) error {
	return azure.WaitForPoller(ctx, r.removeOp)
}

func (r *Disk) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *Disk) String() string {
	return *r.Name
}

type DiskLister struct{}

func (l DiskLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", DiskResource).WithField("s", opts.SubscriptionID)

	client, err := armcompute.NewDisksClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list disks")

	pager := client.NewListByResourceGroupPager(opts.ResourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, entity := range page.Value {
			var creationDate *time.Time
			if entity.Properties != nil {
				creationDate = entity.Properties.TimeCreated
			}

			resources = append(resources, &Disk{
				BaseResource: &BaseResource{
					Region:         entity.Location,
					ResourceGroup:  &opts.ResourceGroup,
					SubscriptionID: &opts.SubscriptionID,
				},
				client:       client,
				Name:         entity.Name,
				Tags:         entity.Tags,
				CreationDate: creationDate,
			})
		}
	}

	log.Trace("done")

	return resources, nil
}
