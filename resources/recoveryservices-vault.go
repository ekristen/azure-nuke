package resources

import (
	"context"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"
	"github.com/hashicorp/go-azure-sdk/resource-manager/recoveryservices/2023-02-01/vaults"
	"github.com/hashicorp/go-azure-sdk/sdk/environments"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/nuke"
)

const RecoveryServicesVaultResource = "RecoveryServicesVault"

func init() {
	registry.Register(&registry.Registration{
		Name:   RecoveryServicesVaultResource,
		Scope:  nuke.ResourceGroup,
		Lister: &RecoveryServicesVaultLister{},
		DependsOn: []string{
			RecoveryServicesBackupProtectedItemResource,
		},
	})
}

type RecoveryServicesVault struct {
	client        *vaults.VaultsClient
	vaultID       vaults.VaultId
	Region        string
	ResourceGroup string
	ID            *string
	Name          *string
}

func (r *RecoveryServicesVault) Filter() error {
	return nil
}

func (r *RecoveryServicesVault) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, r.vaultID)
	return err
}

func (r *RecoveryServicesVault) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *RecoveryServicesVault) String() string {
	return ptr.ToString(r.Name)
}

type RecoveryServicesVaultLister struct {
}

func (l RecoveryServicesVaultLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancel()

	log := logrus.
		WithField("r", RecoveryServicesVaultResource).
		WithField("s", opts.SubscriptionID).
		WithField("rg", opts.ResourceGroup)

	log.Trace("creating client")

	client, err := vaults.NewVaultsClientWithBaseURI(environments.AzurePublic().ResourceManager) // TODO: pass in the endpoint
	if err != nil {
		return nil, err
	}
	client.Client.Authorizer = opts.Authorizers.Management

	resources := make([]resource.Resource, 0)

	log.Trace("listing resources")

	items, err := client.ListByResourceGroupComplete(ctx, commonids.NewResourceGroupID(opts.SubscriptionID, opts.ResourceGroup))
	if err != nil {
		return nil, err
	}

	for _, item := range items.Items {
		resources = append(resources, &RecoveryServicesVault{
			client:        client,
			vaultID:       vaults.NewVaultID(opts.SubscriptionID, opts.ResourceGroup, ptr.ToString(item.Id)),
			Region:        item.Location,
			ResourceGroup: opts.ResourceGroup,
			ID:            item.Id,
			Name:          item.Name,
		})
	}

	log.Trace("done")

	return resources, nil
}
