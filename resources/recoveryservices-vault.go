package resources

import (
	"context"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"
	"github.com/hashicorp/go-azure-sdk/resource-manager/recoveryservices/2023-02-01/vaults"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/nuke"
)

const RecoveryServicesVaultResource = "RecoveryServicesVault"

func init() {
	resource.Register(resource.Registration{
		Name:   RecoveryServicesVaultResource,
		Scope:  nuke.ResourceGroup,
		Lister: RecoveryServicesVaultLister{},
		DependsOn: []string{
			RecoveryServicesBackupProtectedItemResource,
		},
	})
}

type RecoveryServicesVault struct {
	client   vaults.VaultsClient
	vaultId  vaults.VaultId
	id       *string
	name     *string
	location string
	rg       string
}

func (r *RecoveryServicesVault) Filter() error {
	return nil
}

func (r *RecoveryServicesVault) Remove() error {
	_, err := r.client.Delete(context.TODO(), r.vaultId)
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

func (l RecoveryServicesVaultLister) List(o interface{}) ([]resource.Resource, error) {
	opts := o.(nuke.ListerOpts)

	log := logrus.
		WithField("r", RecoveryServicesVaultResource).
		WithField("s", opts.SubscriptionId).
		WithField("rg", opts.ResourceGroup)

	log.Trace("creating client")

	client := vaults.NewVaultsClientWithBaseURI("https://management.azure.com") // TODO: pass in the endpoint
	client.Client.Authorizer = opts.Authorizers.Management
	client.Client.RetryAttempts = 1
	client.Client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("listing resources")

	ctx := context.TODO()
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
