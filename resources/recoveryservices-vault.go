package resources

import (
	"context"
	"github.com/gotidy/ptr"
	"github.com/hashicorp/go-azure-sdk/sdk/environments"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"
	"github.com/hashicorp/go-azure-sdk/resource-manager/recoveryservices/2023-02-01/vaults"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/nuke"
)

const RecoveryServicesVaultResource = "RecoveryServicesVault"

func init() {
	resource.Register(&resource.Registration{
		Name:   RecoveryServicesVaultResource,
		Scope:  nuke.ResourceGroup,
		Lister: &RecoveryServicesVaultLister{},
		DependsOn: []string{
			RecoveryServicesBackupProtectedItemResource,
		},
	})
}

type RecoveryServicesVault struct {
	client   *vaults.VaultsClient
	vaultId  vaults.VaultId
	id       *string
	name     *string
	location string
	rg       string
}

func (r *RecoveryServicesVault) Filter() error {
	return nil
}

func (r *RecoveryServicesVault) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, r.vaultId)
	return err
}

func (r *RecoveryServicesVault) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", r.name)
	properties.Set("Location", r.location)
	properties.Set("ResourceGroup", r.rg)

	return properties
}

func (r *RecoveryServicesVault) String() string {
	return ptr.ToString(r.name)
}

type RecoveryServicesVaultLister struct {
}

func (l RecoveryServicesVaultLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancel()

	log := logrus.
		WithField("r", RecoveryServicesVaultResource).
		WithField("s", opts.SubscriptionId).
		WithField("rg", opts.ResourceGroup)

	log.Trace("creating client")

	client, err := vaults.NewVaultsClientWithBaseURI(environments.AzurePublic().ResourceManager) // TODO: pass in the endpoint
	if err != nil {
		return nil, err
	}
	client.Client.Authorizer = opts.Authorizers.Management

	resources := make([]resource.Resource, 0)

	log.Trace("listing resources")

	items, err := client.ListByResourceGroupComplete(ctx, commonids.NewResourceGroupID(opts.SubscriptionId, opts.ResourceGroup))
	if err != nil {
		return nil, err
	}

	for _, item := range items.Items {
		resources = append(resources, &RecoveryServicesVault{
			client:   client,
			id:       item.Id,
			name:     item.Name,
			location: item.Location,
			vaultId:  vaults.NewVaultID(opts.SubscriptionId, opts.ResourceGroup, ptr.ToString(item.Id)),
			rg:       opts.ResourceGroup,
		})
	}

	log.Trace("done")

	return resources, nil
}
