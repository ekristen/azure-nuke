package resources

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v8"

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

	client       *armcompute.SnapshotsClient
	Name         *string
	Tags         map[string]*string
	CreationDate *time.Time
}

func (r *ComputeSnapshot) Remove(ctx context.Context) error {
	poller, err := r.client.BeginDelete(ctx, *r.ResourceGroup, *r.Name, nil)
	if err != nil {
		return err
	}

	_, err = poller.PollUntilDone(ctx, nil)
	return err
}

func (r *ComputeSnapshot) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ComputeSnapshot) String() string {
	return *r.Name
}

type ComputeSnapshotLister struct{}

func (l ComputeSnapshotLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", ComputeSnapshotResource).WithField("s", opts.SubscriptionID)

	client, err := armcompute.NewSnapshotsClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list snapshots")

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

			resources = append(resources, &ComputeSnapshot{
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
