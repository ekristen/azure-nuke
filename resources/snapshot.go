package resources

import (
	"context"
	"github.com/gotidy/ptr"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2021-04-01/compute" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const ComputeSnapshotResource = "ComputeSnapshot"

func init() {
	registry.Register(&registry.Registration{
		Name:     ComputeSnapshotResource,
		Scope:    azure.ResourceGroupScope,
		Lister:   &ComputeSnapshotLister{},
		Resource: &ComputeSnapshot{},
		DependsOn: []string{
			VirtualMachineResource,
		},
	})
}

type ComputeSnapshot struct {
	*BaseResource `property:",inline"`

	client       compute.SnapshotsClient
	Name         *string
	Tags         map[string]*string
	CreationDate *time.Time
}

func (r *ComputeSnapshot) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.ResourceGroup, *r.Name)
	return err
}

func (r *ComputeSnapshot) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ComputeSnapshot) String() string {
	return *r.Name
}

type ComputeSnapshotLister struct {
}

func (l ComputeSnapshotLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", ComputeSnapshotResource).WithField("s", opts.SubscriptionID)

	client := compute.NewSnapshotsClient(opts.SubscriptionID)
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
			resources = append(resources, &ComputeSnapshot{
				BaseResource: &BaseResource{
					Region:         r.Location,
					ResourceGroup:  &opts.ResourceGroup,
					SubscriptionID: &opts.SubscriptionID,
				},
				client:       client,
				Name:         r.Name,
				Tags:         r.Tags,
				CreationDate: ptr.Time(r.SnapshotProperties.TimeCreated.Time),
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	log.Trace("done")

	return resources, nil
}
