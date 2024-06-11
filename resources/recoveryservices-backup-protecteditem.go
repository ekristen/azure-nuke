package resources

import (
	"context"
	"fmt"
	"github.com/ekristen/azure-nuke/pkg/azure"
	"strings"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservices"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup"
	"github.com/Azure/go-autorest/autorest/to"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
)

const RecoveryServicesBackupProtectedItemResource = "RecoveryServicesBackupProtectedItem"

func init() {
	registry.Register(&registry.Registration{
		Name:     RecoveryServicesBackupProtectedItemResource,
		Scope:    azure.ResourceGroupScope,
		Resource: &RecoveryServicesBackupProtectedItem{},
		Lister:   &RecoveryServicesBackupProtectedItemLister{},
	})
}

type RecoveryServicesBackupProtectedItem struct {
	*BaseResource `property:",inline"`

	client     *armrecoveryservicesbackup.BackupProtectedItemsClient
	itemClient *armrecoveryservicesbackup.ProtectedItemsClient

	ID            *string
	Name          *string
	VaultName     *string
	ContainerName *string
	backupFabric  *string
}

func (r *RecoveryServicesBackupProtectedItem) Filter() error {
	return nil
}

func (r *RecoveryServicesBackupProtectedItem) Remove(ctx context.Context) error {
	_, err := r.itemClient.Delete(
		ctx, to.String(r.VaultName), to.String(r.ResourceGroup),
		to.String(r.backupFabric), to.String(r.ContainerName), to.String(r.Name), nil)
	return err
}

func (r *RecoveryServicesBackupProtectedItem) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *RecoveryServicesBackupProtectedItem) String() string {
	return ptr.ToString(r.Name)
}

type RecoveryServicesBackupProtectedItemLister struct {
}

func (l RecoveryServicesBackupProtectedItemLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancel()

	resources := make([]resource.Resource, 0)

	log := logrus.
		WithField("r", RecoveryServicesBackupProtectedItemResource).
		WithField("s", opts.SubscriptionID).
		WithField("rg", opts.ResourceGroup)

	log.Trace("creating client")
	vaultsClient, err := armrecoveryservices.NewVaultsClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return resources, err
	}

	client, err := armrecoveryservicesbackup.NewBackupProtectedItemsClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return resources, err
	}

	protectedItems, err := armrecoveryservicesbackup.NewProtectedItemsClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return resources, err
	}

	log.Trace("listing resources")

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
						BaseResource: &BaseResource{
							Region:         i.Location,
							ResourceGroup:  &opts.ResourceGroup,
							SubscriptionID: &opts.SubscriptionID,
						},
						client:     client,
						itemClient: protectedItems,
						VaultName:  v.Name,
						ID:         i.ID,
						Name:       i.Name,

						ContainerName: to.StringPtr(containerName),
						backupFabric:  to.StringPtr("Azure"), // TODO: this should be calculated
					})
				}
			}
		}
	}

	log.Trace("done")

	return resources, nil
}
