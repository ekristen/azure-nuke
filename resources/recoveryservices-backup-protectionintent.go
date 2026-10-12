package resources

import (
	"context"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservices/v3"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v5"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const RecoveryServicesBackupProtectionIntentResource = "RecoveryServicesBackupProtectionIntent"

func init() {
	registry.Register(&registry.Registration{
		Name:     RecoveryServicesBackupProtectionIntentResource,
		Scope:    azure.ResourceGroupScope,
		Resource: &RecoveryServicesBackupProtectionIntent{},
		Lister:   &RecoveryServicesBackupProtectionIntentLister{},
	})
}

type RecoveryServicesBackupProtectionIntent struct {
	*BaseResource `property:",inline"`

	client       *armrecoveryservicesbackup.BackupProtectionIntentClient
	pClient      *armrecoveryservicesbackup.ProtectionIntentClient
	ID           *string
	Name         *string
	VaultName    *string
	backupFabric *string
}

func (r *RecoveryServicesBackupProtectionIntent) Filter() error {
	return nil
}

func (r *RecoveryServicesBackupProtectionIntent) Remove(ctx context.Context) error {
	_, err := r.pClient.Delete(ctx,
		ptr.ToString(r.VaultName), ptr.ToString(r.ResourceGroup), ptr.ToString(r.backupFabric), ptr.ToString(r.Name), nil)
	return err
}

func (r *RecoveryServicesBackupProtectionIntent) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *RecoveryServicesBackupProtectionIntent) String() string {
	return ptr.ToString(r.Name)
}

type RecoveryServicesBackupProtectionIntentLister struct{}

func (l RecoveryServicesBackupProtectionIntentLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)
	resources := make([]resource.Resource, 0)

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancel()

	log := logrus.
		WithField("r", RecoveryServicesBackupProtectionIntentResource).
		WithField("s", opts.SubscriptionID).
		WithField("rg", opts.ResourceGroup)

	log.Trace("creating client")

	vaultsClient, err := armrecoveryservices.NewVaultsClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return resources, err
	}

	client, err := armrecoveryservicesbackup.NewBackupProtectionIntentClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return resources, err
	}

	protectedContainers, err := armrecoveryservicesbackup.NewProtectionIntentClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
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
			itemPager := client.NewListPager(ptr.ToString(v.Name), opts.ResourceGroup, nil)
			for itemPager.More() {
				page, err := itemPager.NextPage(ctx)
				if err != nil {
					return resources, err
				}

				for _, i := range page.Value {
					resources = append(resources, &RecoveryServicesBackupProtectionIntent{
						BaseResource: &BaseResource{
							Region:        i.Location,
							ResourceGroup: ptr.String(opts.ResourceGroup),
						},
						client:       client,
						pClient:      protectedContainers,
						VaultName:    v.Name,
						ID:           i.ID,
						Name:         i.Name,
						backupFabric: ptr.String("Azure"), // TODO: this should be calculated
					})
				}
			}
		}
	}

	log.Trace("done")

	return resources, nil
}
