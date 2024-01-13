package resources

import (
	"context"
	"github.com/ekristen/azure-nuke/pkg/nuke"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2021-09-01/storage"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
)

const StorageAccountResource = "StorageAccount"

func init() {
	resource.Register(resource.Registration{
		Name:   StorageAccountResource,
		Scope:  nuke.ResourceGroup,
		Lister: StorageAccountLister{},
		DependsOn: []string{
			VirtualMachineResource,
		},
	})
}

type StorageAccount struct {
	client storage.AccountsClient
	name   string
	rg     string
}

func (r *StorageAccount) Remove() error {
	_, err := r.client.Delete(context.TODO(), r.rg, r.name)
	return err
}

func (r *StorageAccount) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", r.name)

	return properties
}

func (r *StorageAccount) String() string {
	return r.name
}

// --------------------------------------

type StorageAccountLister struct {
}

func (l StorageAccountLister) List(o interface{}) ([]resource.Resource, error) {
	opts := o.(nuke.ListerOpts)

	log := logrus.WithField("r", StorageAccountResource).WithField("s", opts.SubscriptionId)

	client := storage.NewAccountsClient(opts.SubscriptionId)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list ssh key")

	ctx := context.Background()
	list, err := client.ListByResourceGroup(ctx, opts.ResourceGroup)
	if err != nil {
		return nil, err
	}

	log.Trace("listing storage accounts")

	for list.NotDone() {
		log.Trace("list not done")
		for _, g := range list.Values() {
			resources = append(resources, &StorageAccount{
				client: client,
				name:   *g.Name,
				rg:     opts.ResourceGroup,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	log.Trace("done")

	return resources, nil
}
