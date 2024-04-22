package resources

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"
	"github.com/hashicorp/go-azure-sdk/resource-manager/resources/2020-05-01/managementlocks"
	"github.com/hashicorp/go-azure-sdk/sdk/environments"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/nuke"
)

const ManagementLockResource = "ManagementLock"

func init() {
	registry.Register(&registry.Registration{
		Name:     ManagementLockResource,
		Scope:    nuke.ResourceGroup,
		Resource: &ManagementLock{},
		Lister:   &ManagementLockLister{},
	})
}

type ManagementLock struct {
	client         *managementlocks.ManagementLocksClient
	ID             *string `property:"-"`
	Name           *string
	LockLevel      string
	Region         *string
	ResourceGroup  *string
	SubscriptionID *string
}

func (r *ManagementLock) Remove(ctx context.Context) error {
	_, err := r.client.DeleteAtResourceGroupLevel(ctx,
		managementlocks.NewProviderLockID(*r.SubscriptionID, *r.ResourceGroup, *r.Name))
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
	opts := o.(*nuke.ListerOpts)

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancel()

	log := logrus.WithField("r", ManagementLockResource).WithField("s", opts.SubscriptionID)

	resources := make([]resource.Resource, 0)

	client, err := managementlocks.NewManagementLocksClientWithBaseURI(environments.AzurePublic().ResourceManager)
	if err != nil {
		return resources, err
	}
	client.Client.Authorizer = opts.Authorizers.Management

	log.Trace("attempting to list resources")

	list, err := client.ListAtResourceGroupLevelComplete(ctx,
		commonids.NewResourceGroupID(opts.SubscriptionID, opts.ResourceGroup),
		managementlocks.ListAtResourceGroupLevelOperationOptions{})
	if err != nil {
		return nil, err
	}

	for _, lock := range list.Items {
		resources = append(resources, &ManagementLock{
			client:         client,
			ID:             lock.Id,
			Name:           lock.Name,
			LockLevel:      string(lock.Properties.Level),
			ResourceGroup:  &opts.ResourceGroup,
			SubscriptionID: &opts.SubscriptionID,
		})
	}

	log.Trace("done listing")

	return resources, nil
}
