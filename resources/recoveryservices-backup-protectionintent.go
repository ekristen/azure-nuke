package resources

import (
	"context"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservices"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup"
	"github.com/Azure/go-autorest/autorest/to"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/nuke"
)

const RecoveryServicesBackupProtectionIntentResource = "RecoveryServicesBackupProtectionIntent"

func init() {
	registry.Register(&registry.Registration{
		Name:   RecoveryServicesBackupProtectionIntentResource,
		Scope:  nuke.Subscription,
		Lister: &RecoveryServicesBackupProtectionIntentLister{},
	})
}

type RecoveryServicesBackupProtectionIntent struct {
	client        *armrecoveryservicesbackup.BackupProtectionIntentClient
	pClient       *armrecoveryservicesbackup.ProtectionIntentClient
	id            *string
	name          *string
	location      *string
	resourceGroup *string
	vaultName     *string
	backupFabric  *string
}

func (r *RecoveryServicesBackupProtectionIntent) Filter() error {
	return nil
}

func (r *RecoveryServicesBackupProtectionIntent) Remove(ctx context.Context) error {
	_, err := r.pClient.Delete(ctx, to.String(r.vaultName), to.String(r.resourceGroup), to.String(r.backupFabric), to.String(r.name), nil)
	return err
}

func (r *RecoveryServicesBackupProtectionIntent) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", r.name)
	properties.Set("Location", r.location)
	properties.Set("ResourceGroup", r.resourceGroup)
	properties.Set("VaultName", r.vaultName)

	return properties
}

func (r *RecoveryServicesBackupProtectionIntent) String() string {
	return ptr.ToString(r.name)
}

type RecoveryServicesBackupProtectionIntentLister struct {
}

func (l RecoveryServicesBackupProtectionIntentLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	resources := make([]resource.Resource, 0)

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancel()

	log := logrus.
		WithField("r", RecoveryServicesBackupProtectionIntentResource).
		WithField("s", opts.SubscriptionId).
		WithField("rg", opts.ResourceGroup)

	log.Trace("creating client")

	vaultsClient, err := armrecoveryservices.NewVaultsClient(opts.SubscriptionId, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return resources, err
	}

	client, err := armrecoveryservicesbackup.NewBackupProtectionIntentClient(opts.SubscriptionId, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return resources, err
	}

	protectedContainers, err := armrecoveryservicesbackup.NewProtectionIntentClient(opts.SubscriptionId, opts.Authorizers.IdentityCreds, nil)
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
					resources = append(resources, &RecoveryServicesBackupProtectionIntent{
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

	log.Trace("done")

	return resources, nil
}
