package resources

import (
	"context"
	"github.com/ekristen/azure-nuke/pkg/azure"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"
	"github.com/hashicorp/go-azure-sdk/resource-manager/recoveryservices/2023-02-01/vaults"
	"github.com/hashicorp/go-azure-sdk/resource-manager/recoveryservicesbackup/2023-02-01/backuppolicies"
	"github.com/hashicorp/go-azure-sdk/resource-manager/recoveryservicesbackup/2023-02-01/protectionpolicies"
	"github.com/hashicorp/go-azure-sdk/sdk/environments"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
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

	client            backuppolicies.BackupPoliciesClient
	protectionsClient protectionpolicies.ProtectionPoliciesClient

	ID   *string
	Name *string

	backupPolicyID protectionpolicies.BackupPolicyId
}

func (r *RecoveryServicesBackupPolicy) Filter() error {
	return nil
}

func (r *RecoveryServicesBackupPolicy) Remove(ctx context.Context) error {
	_, err := r.protectionsClient.Delete(ctx, r.backupPolicyID)
	return err
}

func (r *RecoveryServicesBackupPolicy) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *RecoveryServicesBackupPolicy) String() string {
	return ptr.ToString(r.Name)
}

type RecoveryServicesBackupPolicyLister struct {
}

func (l RecoveryServicesBackupPolicyLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancel()

	log := logrus.
		WithField("r", RecoveryServicesBackupPolicyResource).
		WithField("s", opts.SubscriptionID).
		WithField("rg", opts.ResourceGroup)

	log.Trace("creating client")

	vaultsClient, err := vaults.NewVaultsClientWithBaseURI(environments.AzurePublic().ResourceManager) // TODO: pass in the endpoint
	if err != nil {
		return nil, err
	}
	vaultsClient.Client.Authorizer = opts.Authorizers.Management

	// TODO: pass in the endpoint
	client :=
		backuppolicies.NewBackupPoliciesClientWithBaseURI("https://management.azure.com")
	client.Client.Authorizer = opts.Authorizers.Management
	client.Client.RetryAttempts = 1
	client.Client.RetryDuration = time.Second * 2

	// TODO: pass in the endpoint
	protectionsClient :=
		protectionpolicies.NewProtectionPoliciesClientWithBaseURI("https://management.azure.com")
	protectionsClient.Client.Authorizer = opts.Authorizers.Management
	protectionsClient.Client.RetryAttempts = 1
	protectionsClient.Client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("listing resources")

	vaultsRes, err :=
		vaultsClient.ListByResourceGroupComplete(
			ctx, commonids.NewResourceGroupID(opts.SubscriptionID, opts.ResourceGroup))
	if err != nil {
		return nil, err
	}

	for _, v := range vaultsRes.Items {
		vaultID := backuppolicies.NewVaultID(opts.SubscriptionID, opts.ResourceGroup, ptr.ToString(v.Name))
		items, err := client.ListComplete(ctx, vaultID, backuppolicies.DefaultListOperationOptions())
		if err != nil {
			return nil, err
		}

		for _, item := range items.Items {
			resources = append(resources, &RecoveryServicesBackupPolicy{
				BaseResource: &BaseResource{
					Region:         item.Location,
					ResourceGroup:  &opts.ResourceGroup,
					SubscriptionID: &opts.SubscriptionID,
				},
				client:            client,
				protectionsClient: protectionsClient,
				ID:                item.Id,
				Name:              item.Name,
				backupPolicyID: protectionpolicies.NewBackupPolicyID(
					opts.SubscriptionID, opts.ResourceGroup, ptr.ToString(v.Name), ptr.ToString(item.Name)),
			})
		}
	}

	log.Trace("done")

	return resources, nil
}
