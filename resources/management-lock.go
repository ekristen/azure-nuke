package resources

import (
	"context"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armlocks"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const ManagementLockResource = "ManagementLock"

func init() {
	registry.Register(&registry.Registration{
		Name:     ManagementLockResource,
		Scope:    azure.ResourceGroupScope,
		Resource: &ManagementLock{},
		Lister:   &ManagementLockLister{},
	})
}

type ManagementLock struct {
	*BaseResource `property:",inline"`

	client    *armlocks.ManagementLocksClient
	ID        *string `property:"-"`
	Name      *string
	LockLevel string
}

func (r *ManagementLock) Remove(ctx context.Context) error {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancel()

	_, err := r.client.DeleteAtResourceGroupLevel(ctx, *r.ResourceGroup, *r.Name, nil)
	return err
}

func (r *ManagementLock) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ManagementLock) String() string {
	return *r.Name
}

type ManagementLockLister struct{}

func (l ManagementLockLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancel()

	log := logrus.WithField("r", ManagementLockResource).WithField("s", opts.SubscriptionID)

	resources := make([]resource.Resource, 0)

	client, err := armlocks.NewManagementLocksClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return resources, err
	}

	log.Trace("attempting to list resources")

	pager := client.NewListAtResourceGroupLevelPager(opts.ResourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, lock := range page.Value {
			var lockLevel string
			if lock.Properties != nil && lock.Properties.Level != nil {
				lockLevel = string(*lock.Properties.Level)
			}

			resources = append(resources, &ManagementLock{
				BaseResource: &BaseResource{
					Region:         ptr.String("global"),
					ResourceGroup:  &opts.ResourceGroup,
					SubscriptionID: &opts.SubscriptionID,
				},
				client:    client,
				ID:        lock.ID,
				Name:      lock.Name,
				LockLevel: lockLevel,
			})
		}
	}

	log.Trace("done listing")

	return resources, nil
}
