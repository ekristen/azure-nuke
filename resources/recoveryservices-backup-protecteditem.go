package resources

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservices"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/aws/smithy-go/ptr"
	"github.com/ekristen/azure-nuke/pkg/nuke"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
	"github.com/sirupsen/logrus"
	"strings"
)

const RecoveryServicesBackupProtectedItemResource = "RecoveryServicesBackupProtectedItem"

func init() {
	resource.Register(resource.Registration{
		Name:   RecoveryServicesBackupProtectedItemResource,
		Scope:  nuke.ResourceGroup,
		Lister: RecoveryServicesBackupProtectedItemLister{},
	})
}

type RecoveryServicesBackupProtectedItem struct {
	client        *armrecoveryservicesbackup.BackupProtectedItemsClient
	itemClient    *armrecoveryservicesbackup.ProtectedItemsClient
	vaultName     *string
	id            *string
	name          *string
	location      *string
	resourceGroup *string
	containerName *string
	backupFabric  *string
}

func (r *RecoveryServicesBackupProtectedItem) Filter() error {
	return nil
}

func (r *RecoveryServicesBackupProtectedItem) Remove() error {
	_, err := r.itemClient.Delete(context.TODO(), to.String(r.vaultName), to.String(r.resourceGroup), to.String(r.backupFabric), to.String(r.containerName), to.String(r.name), nil)
	return err
}

func (r *RecoveryServicesBackupProtectedItem) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("ID", r.id)
	properties.Set("Name", r.name)
	properties.Set("Location", r.location)
	properties.Set("ResourceGroup", r.resourceGroup)
	properties.Set("VaultName", r.vaultName)
	properties.Set("ContainerName", r.containerName)

	return properties
}

func (r *RecoveryServicesBackupProtectedItem) String() string {
	return ptr.ToString(r.name)
}

type RecoveryServicesBackupProtectedItemLister struct {
}

func (l RecoveryServicesBackupProtectedItemLister) List(o interface{}) ([]resource.Resource, error) {
	opts := o.(nuke.ListerOpts)

	resources := make([]resource.Resource, 0)

	log := logrus.
		WithField("r", RecoveryServicesBackupProtectedItemResource).
		WithField("s", opts.SubscriptionId).
		WithField("rg", opts.ResourceGroup)

	log.Trace("creating client")
	vaultsClient, err := armrecoveryservices.NewVaultsClient(opts.SubscriptionId, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return resources, err
	}

	client, err := armrecoveryservicesbackup.NewBackupProtectedItemsClient(opts.SubscriptionId, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return resources, err
	}

	protectedItems, err := armrecoveryservicesbackup.NewProtectedItemsClient(opts.SubscriptionId, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return resources, err
	}

	log.Trace("listing resources")

	ctx := context.TODO()
	vaultsPager := vaultsClient.NewListByResourceGroupPager(opts.ResourceGroup, nil)
	for vaultsPager.More() {
		log.Trace("not done")
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
					containerName := to.String(i.Properties.GetProtectedItem().ContainerName)
					if to.String(i.Properties.GetProtectedItem().ProtectedItemType) == "Microsoft.Compute/virtualMachines" {
						if !strings.HasPrefix(containerName, "IaasVMContainer;") {
							containerName = fmt.Sprintf("IaasVMContainer;%s", containerName)
						}
					}

					resources = append(resources, &RecoveryServicesBackupProtectedItem{
						client:        client,
						itemClient:    protectedItems,
						vaultName:     v.Name,
						id:            i.ID,
						name:          i.Name,
						location:      i.Location,
						resourceGroup: to.StringPtr(opts.ResourceGroup),
						containerName: to.StringPtr(containerName),
						backupFabric:  to.StringPtr("Azure"), // TODO: this should be calculated
					})
				}
			}

		}
	}

	log.Trace("done")

	return resources, nil
}
