package resources

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservices"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/aws/smithy-go/ptr"
	"github.com/ekristen/azure-nuke/pkg/nuke"
	"github.com/ekristen/cloud-nuke-sdk/pkg/resource"
	"github.com/ekristen/cloud-nuke-sdk/pkg/types"
	"github.com/sirupsen/logrus"
)

func init() {
	resource.Register(resource.Registration{
		Name:   "RecoveryServicesBackupProtectionIntent",
		Scope:  nuke.ResourceGroup,
		Lister: RecoveryServicesBackupProtectionIntentLister{},
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

func (r *RecoveryServicesBackupProtectionIntent) Remove() error {
	_, err := r.pClient.Delete(context.TODO(), to.String(r.vaultName), to.String(r.resourceGroup), to.String(r.backupFabric), to.String(r.name), nil)
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
	opts nuke.ListerOpts
}

func (l RecoveryServicesBackupProtectionIntentLister) SetOptions(opts interface{}) {
	l.opts = opts.(nuke.ListerOpts)
}

func (l RecoveryServicesBackupProtectionIntentLister) List() ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)

	log := logrus.
		WithField("resource", "RecoveryServicesBackupProtectionIntent").
		WithField("scope", nuke.ResourceGroup).
		WithField("subscription", l.opts.SubscriptionId).
		WithField("rg", l.opts.ResourceGroup)

	log.Trace("creating client")

	vaultsClient, err := armrecoveryservices.NewVaultsClient(l.opts.SubscriptionId, l.opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return resources, err
	}

	client, err := armrecoveryservicesbackup.NewBackupProtectionIntentClient(l.opts.SubscriptionId, l.opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return resources, err
	}

	protectedContainers, err := armrecoveryservicesbackup.NewProtectionIntentClient(l.opts.SubscriptionId, l.opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return resources, err
	}

	log.Trace("listing resources")

	ctx := context.TODO()
	vaultsPager := vaultsClient.NewListByResourceGroupPager(l.opts.ResourceGroup, nil)
	for vaultsPager.More() {
		page, err := vaultsPager.NextPage(ctx)
		if err != nil {
			return resources, err
		}

		for _, v := range page.Value {

			itemPager := client.NewListPager(to.String(v.Name), l.opts.ResourceGroup, nil)
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
						resourceGroup: to.StringPtr(l.opts.ResourceGroup),
						backupFabric:  to.StringPtr("Azure"), // TODO: this should be calculated
					})
				}
			}

		}
	}

	return resources, nil
}
