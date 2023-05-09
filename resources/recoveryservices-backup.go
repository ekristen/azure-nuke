package resources

import (
	"context"
	"github.com/aws/smithy-go/ptr"
	"github.com/ekristen/azure-nuke/pkg/resource"
	"github.com/ekristen/azure-nuke/pkg/types"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"
	"github.com/hashicorp/go-azure-sdk/resource-manager/recoveryservices/2023-02-01/vaults"
	"github.com/hashicorp/go-azure-sdk/resource-manager/recoveryservicesbackup/2023-02-01/backupprotecteditems"
	"github.com/sirupsen/logrus"
	"time"
)

type RecoveryServicesBackupProtectedItem struct {
	client   backupprotecteditems.BackupProtectedItemsClient
	id       *string
	name     *string
	location string
	rg       string
}

func init() {
	resource.RegisterV2(resource.Registration{
		Name:   "RecoveryServicesBackupProtectedItem",
		Scope:  resource.ResourceGroup,
		Lister: ListRecoveryServicesBackupProtectedItem,
	})
}

func ListRecoveryServicesBackupProtectedItem(opts resource.ListerOpts) ([]resource.Resource, error) {
	log := logrus.
		WithField("resource", "RecoveryServicesBackupProtectedItem").
		WithField("scope", resource.ResourceGroup).
		WithField("subscription", opts.SubscriptionId).
		WithField("rg", opts.ResourceGroup)

	log.Trace("creating client")

	vaultsClient := vaults.NewVaultsClientWithBaseURI("https://management.azure.com") // TODO: pass in the endpoint
	vaultsClient.Client.Authorizer = opts.Authorizers.Management
	vaultsClient.Client.RetryAttempts = 1
	vaultsClient.Client.RetryDuration = time.Second * 2

	client := backupprotecteditems.NewBackupProtectedItemsClientWithBaseURI("https://management.azure.com") // TODO: pass in the endpoint
	client.Client.Authorizer = opts.Authorizers.Management
	client.Client.RetryAttempts = 1
	client.Client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("listing resources")

	ctx := context.TODO()
	vaultsRes, err := vaultsClient.ListByResourceGroupComplete(ctx, commonids.NewResourceGroupID(opts.SubscriptionId, opts.ResourceGroup))
	if err != nil {
		return nil, err
	}

	for _, v := range vaultsRes.Items {
		vaultId := backupprotecteditems.NewVaultID(opts.SubscriptionId, opts.ResourceGroup, ptr.ToString(v.Name))
		items, err := client.ListComplete(ctx, vaultId, backupprotecteditems.DefaultListOperationOptions())
		if err != nil {
			return nil, err
		}

		for _, item := range items.Items {
			resources = append(resources, &RecoveryServicesBackupProtectedItem{
				client: client,
				id:     item.Id,
				name:   item.Name,
				rg:     opts.ResourceGroup,
			})
		}
	}

	return resources, nil
}

func (r *RecoveryServicesBackupProtectedItem) Filter() error {
	return nil
}

func (r *RecoveryServicesBackupProtectedItem) Remove() error {
	return nil
}

func (r *RecoveryServicesBackupProtectedItem) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", r.name)
	properties.Set("Location", r.location)
	properties.Set("ResourceGroup", r.rg)

	return properties
}

func (r *RecoveryServicesBackupProtectedItem) String() string {
	return ptr.ToString(r.name)
}
