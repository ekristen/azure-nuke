package resources

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/keyvault/mgmt/2019-09-01/keyvault" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/nuke"
)

const KeyVaultResource = "KeyVault"

func init() {
	registry.Register(&registry.Registration{
		Name:   KeyVaultResource,
		Scope:  nuke.ResourceGroup,
		Lister: &KeyVaultLister{},
	})
}

type KeyVaultLister struct {
}

func (l KeyVaultLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	log := logrus.WithField("r", KeyVaultResource).WithField("s", opts.SubscriptionID)

	client := keyvault.NewVaultsClient(opts.SubscriptionID)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list key vaults")

	list, err := client.ListByResourceGroup(ctx, opts.ResourceGroup, nil)
	if err != nil {
		return nil, err
	}

	log.Trace("listing key vaults")

	for list.NotDone() {
		log.Trace("list not done")
		for _, g := range list.Values() {
			resources = append(resources, &KeyVault{
				client:         client,
				Region:         g.Location,
				ResourceGroup:  opts.ResourceGroup,
				SubscriptionID: opts.SubscriptionID,
				Name:           g.Name,
				Tags:           g.Tags,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	log.Trace("done")

	return resources, nil
}

type KeyVault struct {
	client         keyvault.VaultsClient
	Region         *string
	ResourceGroup  string
	SubscriptionID string
	Name           *string
	Tags           map[string]*string
}

func (r *KeyVault) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, r.ResourceGroup, *r.Name)

	return err
}

func (r *KeyVault) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *KeyVault) String() string {
	return *r.Name
}
