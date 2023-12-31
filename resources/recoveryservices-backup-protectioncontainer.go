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
		Name:   "RecoveryServicesBackupProtectionContainers",
		Scope:  nuke.ResourceGroup,
		Lister: RecoveryServicesBackupProtectionContainersLister{},
	})
}

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

type RecoveryServicesBackupProtectionContainersLister struct {
	opts nuke.ListerOpts
}

func (l RecoveryServicesBackupProtectionContainersLister) SetOptions(opts interface{}) {
	l.opts = opts.(nuke.ListerOpts)
}

func (l RecoveryServicesBackupProtectionContainersLister) List() ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)

	log := logrus.
		WithField("resource", "RecoveryServicesBackupProtectionContainers").
		WithField("scope", nuke.ResourceGroup).
		WithField("subscription", l.opts.SubscriptionId).
		WithField("rg", l.opts.ResourceGroup)

	log.Trace("creating client")

	vaultsClient, err := armrecoveryservices.NewVaultsClient(l.opts.SubscriptionId, l.opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return resources, err
	}

	client, err := armrecoveryservicesbackup.NewBackupProtectionContainersClient(l.opts.SubscriptionId, l.opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return resources, err
	}

	protectedContainers, err := armrecoveryservicesbackup.NewProtectionContainersClient(l.opts.SubscriptionId, l.opts.Authorizers.IdentityCreds, nil)
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
					resources = append(resources, &RecoveryServicesBackupProtectionContainers{
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
