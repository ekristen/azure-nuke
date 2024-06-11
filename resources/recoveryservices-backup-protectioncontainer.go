package resources

import (
	"context"
	"github.com/ekristen/azure-nuke/pkg/azure"
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

const RecoveryServicesBackupProtectionContainerResource = "RecoveryServicesBackupProtectionContainer"

func init() {
	registry.Register(&registry.Registration{
		Name:     RecoveryServicesBackupProtectionContainerResource,
		Scope:    azure.ResourceGroupScope,
		Resource: &RecoveryServicesBackupProtectionContainers{},
		Lister:   &RecoveryServicesBackupProtectionContainersLister{},
	})
}

type RecoveryServicesBackupProtectionContainers struct {
	*BaseResource `property:",inline"`

	client       *armrecoveryservicesbackup.BackupProtectionContainersClient
	pClient      *armrecoveryservicesbackup.ProtectionContainersClient
	ID           *string
	Name         *string
	VaultName    *string
	backupFabric *string
}

func (r *RecoveryServicesBackupProtectionContainers) Filter() error {
	return nil
}

func (r *RecoveryServicesBackupProtectionContainers) Remove(ctx context.Context) error {
	_, err := r.pClient.Unregister(ctx, to.String(r.VaultName), to.String(r.ResourceGroup), to.String(r.backupFabric), to.String(r.Name), nil)
	return err
}

func (r *RecoveryServicesBackupProtectionContainers) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *RecoveryServicesBackupProtectionContainers) String() string {
	return ptr.ToString(r.Name)
}

type RecoveryServicesBackupProtectionContainersLister struct {
}

func (l RecoveryServicesBackupProtectionContainersLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancel()

	resources := make([]resource.Resource, 0)

	log := logrus.
		WithField("r", RecoveryServicesBackupProtectionContainerResource).
		WithField("s", opts.SubscriptionID).
		WithField("rg", opts.ResourceGroup)

	log.Trace("creating client")

	vaultsClient, err :=
		armrecoveryservices.NewVaultsClient(
			opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return resources, err
	}

	client, err :=
		armrecoveryservicesbackup.NewBackupProtectionContainersClient(
			opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return resources, err
	}

	protectedContainers, err :=
		armrecoveryservicesbackup.NewProtectionContainersClient(
			opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return resources, err
	}

	log.Trace("listing resources")

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
						BaseResource: &BaseResource{
							Region:         i.Location,
							ResourceGroup:  &opts.ResourceGroup,
							SubscriptionID: &opts.SubscriptionID,
						},
						client:       client,
						pClient:      protectedContainers,
						VaultName:    v.Name,
						ID:           i.ID,
						Name:         i.Name,
						backupFabric: to.StringPtr("Azure"), // TODO: this should be calculated
					})
				}
			}
		}
	}

	log.Trace("done")

	return resources, nil
}
