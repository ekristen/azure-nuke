package resources

import (
	"context"
	"github.com/ekristen/azure-nuke/pkg/azure"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2021-09-01/storage" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
)

const StorageAccountResource = "StorageAccount"

func init() {
	registry.Register(&registry.Registration{
		Name:     StorageAccountResource,
		Scope:    azure.ResourceGroupScope,
		Resource: &StorageAccount{},
		Lister:   &StorageAccountLister{},
		DependsOn: []string{
			VirtualMachineResource,
		},
	})
}

type StorageAccount struct {
	*BaseResource `property:",inline"`

	client storage.AccountsClient
	Name   *string
	Tags   map[string]*string
}

func (r *StorageAccount) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.ResourceGroup, *r.Name)
	return err
}

func (r *StorageAccount) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *StorageAccount) String() string {
	return *r.Name
}

// --------------------------------------

type StorageAccountLister struct {
}

func (l StorageAccountLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", StorageAccountResource).WithField("s", opts.SubscriptionID)

	client := storage.NewAccountsClient(opts.SubscriptionID)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list ssh key")

	list, err := client.ListByResourceGroup(ctx, opts.ResourceGroup)
	if err != nil {
		return nil, err
	}

	log.Trace("listing storage accounts")

	for list.NotDone() {
		log.Trace("list not done")
		for _, g := range list.Values() {
			resources = append(resources, &StorageAccount{
				BaseResource: &BaseResource{
					Region:         g.Location,
					ResourceGroup:  &opts.ResourceGroup,
					SubscriptionID: &opts.SubscriptionID,
				},
				client: client,
				Name:   g.Name,
				Tags:   g.Tags,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	log.Trace("done")

	return resources, nil
}
