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

const RecoveryServicesBackupPolicyResource = "RecoveryServicesBackupPolicy"

func init() {
	registry.Register(&registry.Registration{
		Name:     RecoveryServicesBackupPolicyResource,
		Scope:    azure.ResourceGroupScope,
		Resource: &RecoveryServicesBackupPolicy{},
		Lister:   &RecoveryServicesBackupPolicyLister{},
		DependsOn: []string{
			RecoveryServicesBackupProtectedItemResource,
		},
	})
}

type RecoveryServicesBackupPolicy struct {
	*BaseResource `property:",inline"`

	protectionsClient *armrecoveryservicesbackup.ProtectionPoliciesClient

	ID        *string
	Name      *string
	VaultName string
}

func (r *RecoveryServicesBackupPolicy) Filter() error {
	return nil
}

func (r *RecoveryServicesBackupPolicy) Remove(ctx context.Context) error {
	poller, err := r.protectionsClient.BeginDelete(ctx, r.VaultName, *r.ResourceGroup, *r.Name, nil)
	if err != nil {
		return err
	}

	_, err = poller.PollUntilDone(ctx, nil)
	return err
}

func (r *RecoveryServicesBackupPolicy) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *RecoveryServicesBackupPolicy) String() string {
	return ptr.ToString(r.Name)
}

type RecoveryServicesBackupPolicyLister struct{}

func (l RecoveryServicesBackupPolicyLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancel()

	log := logrus.
		WithField("r", RecoveryServicesBackupPolicyResource).
		WithField("s", opts.SubscriptionID).
		WithField("rg", opts.ResourceGroup)

	log.Trace("creating client")

	vaultsClient, err := armrecoveryservices.NewVaultsClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	backupClient, err := armrecoveryservicesbackup.NewBackupPoliciesClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	protectionsClient, err := armrecoveryservicesbackup.NewProtectionPoliciesClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)

	log.Trace("listing resources")

	vaultsPager := vaultsClient.NewListByResourceGroupPager(opts.ResourceGroup, nil)
	for vaultsPager.More() {
		vaultsPage, err := vaultsPager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range vaultsPage.Value {
			vaultName := ptr.ToString(v.Name)

			policyPager := backupClient.NewListPager(vaultName, opts.ResourceGroup, nil)
			for policyPager.More() {
				policyPage, err := policyPager.NextPage(ctx)
				if err != nil {
					return nil, err
				}

				for _, item := range policyPage.Value {
					resources = append(resources, &RecoveryServicesBackupPolicy{
						BaseResource: &BaseResource{
							Region:         item.Location,
							ResourceGroup:  &opts.ResourceGroup,
							SubscriptionID: &opts.SubscriptionID,
						},
						protectionsClient: protectionsClient,
						ID:                item.ID,
						Name:              item.Name,
						VaultName:         vaultName,
					})
				}
			}
		}
	}

	log.Trace("done")

	return resources, nil
}
