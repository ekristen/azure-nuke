package resources

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservices"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/aws/smithy-go/ptr"
	"github.com/ekristen/azure-nuke/pkg/resource"
	"github.com/ekristen/azure-nuke/pkg/types"
	"github.com/sirupsen/logrus"
)

type RecoveryServicesBackupProtectionContainers struct {
	client        *armrecoveryservicesbackup.BackupProtectionContainersClient
	pClient       *armrecoveryservicesbackup.ProtectionContainersClient
	id            *string
	name          *string
	location      *string
	resourceGroup *string
	vaultName     *string
	backupFabric  *string
}

func init() {
	resource.RegisterV2(resource.Registration{
		Name:   "RecoveryServicesBackupProtectionContainers",
		Scope:  resource.ResourceGroup,
		Lister: ListRecoveryServicesBackupProtectionContainers,
	})
}

func ListRecoveryServicesBackupProtectionContainers(opts resource.ListerOpts) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)

	log := logrus.
		WithField("resource", "RecoveryServicesBackupProtectionContainers").
		WithField("scope", resource.ResourceGroup).
		WithField("subscription", opts.SubscriptionId).
		WithField("rg", opts.ResourceGroup)

	log.Trace("creating client")

	vaultsClient, err := armrecoveryservices.NewVaultsClient(opts.SubscriptionId, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return resources, err
	}

	client, err := armrecoveryservicesbackup.NewBackupProtectionContainersClient(opts.SubscriptionId, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return resources, err
	}

	protectedContainers, err := armrecoveryservicesbackup.NewProtectionContainersClient(opts.SubscriptionId, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return resources, err
	}

	log.Trace("listing resources")

	ctx := context.TODO()
	vaultsPager := vaultsClient.NewListByResourceGroupPager(opts.ResourceGroup, nil)
	for vaultsPager.More() {
		page, err := vaultsPager.NextPage(ctx)
		if err != nil {
			return resources, err
		}

		for _, v := range page.Value {

			itemPager := client.NewListPager(to.String(v.Name), opts.ResourceGroup, nil)
			for itemPager.More() {
				page, err := itemPager.NextPage(ctx)
				if err != nil {
					return resources, err
				}

				for _, i := range page.Value {
					resources = append(resources, &RecoveryServicesBackupProtectionContainers{
						client:        client,
						pClient:       protectedContainers,
						vaultName:     v.Name,
						id:            i.ID,
						name:          i.Name,
						location:      i.Location,
						resourceGroup: to.StringPtr(opts.ResourceGroup),
						backupFabric:  to.StringPtr("Azure"), // TODO: this should be calculated
					})
				}
			}

		}
	}

	return resources, nil
}

func (r *RecoveryServicesBackupProtectionContainers) Filter() error {
	return nil
}

func (r *RecoveryServicesBackupProtectionContainers) Remove() error {
	_, err := r.pClient.Unregister(context.TODO(), to.String(r.vaultName), to.String(r.resourceGroup), to.String(r.backupFabric), to.String(r.name), nil)
	return err
}

func (r *RecoveryServicesBackupProtectionContainers) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", r.name)
	properties.Set("Location", r.location)
	properties.Set("ResourceGroup", r.resourceGroup)
	properties.Set("VaultName", r.vaultName)

	return properties
}

func (r *RecoveryServicesBackupProtectionContainers) String() string {
	return ptr.ToString(r.name)
}
