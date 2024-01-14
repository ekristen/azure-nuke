package resources

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/keyvault/mgmt/2019-09-01/keyvault"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/nuke"
)

const KeyVaultResource = "KeyVault"

func init() {
	resource.Register(resource.Registration{
		Name:   KeyVaultResource,
		Scope:  nuke.ResourceGroup,
		Lister: KeyVaultLister{},
	})
}

type KeyVaultLister struct {
}

func (l KeyVaultLister) List(o interface{}) ([]resource.Resource, error) {
	opts := o.(nuke.ListerOpts)

	log := logrus.WithField("r", KeyVaultResource).WithField("s", opts.SubscriptionId)

	client := keyvault.NewVaultsClient(opts.SubscriptionId)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list key vaults")

	ctx := context.TODO()
	list, err := client.ListByResourceGroup(ctx, opts.ResourceGroup, nil)
	if err != nil {
		return nil, err
	}

	log.Trace("listing key vaults")

	for list.NotDone() {
		log.Trace("list not done")
		for _, g := range list.Values() {
			resources = append(resources, &KeyVault{
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

type KeyVault struct {
	client keyvault.VaultsClient
	name   string
	rg     string
}

func (r *KeyVault) Remove() error {
	_, err := r.client.Delete(context.TODO(), r.rg, r.name)
	return err
}

func (r *KeyVault) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", r.name)

	return properties
}

func (r *KeyVault) String() string {
	return r.name
}
