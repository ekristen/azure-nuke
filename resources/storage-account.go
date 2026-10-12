package resources

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage/v4"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
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

	client *armstorage.AccountsClient
	Name   *string
	Tags   map[string]*string
}

func (r *StorageAccount) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.ResourceGroup, *r.Name, nil)
	return err
}

func (r *StorageAccount) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *StorageAccount) String() string {
	return *r.Name
}

// --------------------------------------

type StorageAccountLister struct{}

func (l StorageAccountLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", StorageAccountResource).WithField("s", opts.SubscriptionID)

	client, err := armstorage.NewAccountsClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list storage accounts")

	pager := client.NewListByResourceGroupPager(opts.ResourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, entity := range page.Value {
			resources = append(resources, &StorageAccount{
				BaseResource: &BaseResource{
					Region:         entity.Location,
					ResourceGroup:  &opts.ResourceGroup,
					SubscriptionID: &opts.SubscriptionID,
				},
				client: client,
				Name:   entity.Name,
				Tags:   entity.Tags,
			})
		}
	}

	log.Trace("done")

	return resources, nil
}
