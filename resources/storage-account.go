package resources

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2021-09-01/storage"

	"github.com/ekristen/azure-nuke/pkg/resource"
	"github.com/ekristen/azure-nuke/pkg/types"
)

type StorageAccount struct {
	client storage.AccountsClient
	name   string
	rg     string
}

func init() {
	resource.RegisterV2(resource.Registration{
		Name:      "StorageAccount",
		Scope:     resource.ResourceGroup,
		Lister:    ListStorageAccount,
		DependsOn: []string{"VirtualMachine"},
	})
}

func ListStorageAccount(opts resource.ListerOpts) ([]resource.Resource, error) {
	logrus.Tracef("subscription id: %s", opts.SubscriptionId)

	client := storage.NewAccountsClient(opts.SubscriptionId)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	logrus.Trace("attempting to list ssh key")

	ctx := context.Background()
	list, err := client.ListByResourceGroup(ctx, opts.ResourceGroup)
	if err != nil {
		return nil, err
	}

	logrus.Trace("listing ....")

	for list.NotDone() {
		logrus.Trace("list not done")
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

	return resources, nil
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
