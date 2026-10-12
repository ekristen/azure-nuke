package resources

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault/v2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const KeyVaultResource = "KeyVault"

func init() {
	registry.Register(&registry.Registration{
		Name:     KeyVaultResource,
		Scope:    azure.SubscriptionScope,
		Resource: &KeyVault{},
		Lister:   &KeyVaultLister{},
	})
}

type KeyVaultLister struct{}

func (l KeyVaultLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", KeyVaultResource).WithField("s", opts.SubscriptionID)

	client, err := armkeyvault.NewVaultsClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list key vaults")

	pager := client.NewListBySubscriptionPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, entity := range page.Value {
			resources = append(resources, &KeyVault{
				BaseResource: &BaseResource{
					Region:         entity.Location,
					ResourceGroup:  azure.GetResourceGroupFromID(*entity.ID),
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

type KeyVault struct {
	*BaseResource `property:",inline"`

	client *armkeyvault.VaultsClient
	Name   *string
	Tags   map[string]*string
}

func (r *KeyVault) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.ResourceGroup, *r.Name, nil)
	return err
}

func (r *KeyVault) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *KeyVault) String() string {
	return *r.Name
}
