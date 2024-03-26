package resources

import (
	"context"
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

	"github.com/ekristen/azure-nuke/pkg/nuke"
)

const RecoveryServicesBackupPolicyResource = "RecoveryServicesBackupPolicy"

func init() {
	registry.Register(&registry.Registration{
		Name:   RecoveryServicesBackupPolicyResource,
		Scope:  nuke.Subscription,
		Lister: &RecoveryServicesBackupPolicyLister{},
		DependsOn: []string{
			RecoveryServicesBackupProtectedItemResource,
		},
	})
}

type RecoveryServicesBackupPolicy struct {
	client            backuppolicies.BackupPoliciesClient
	protectionsClient protectionpolicies.ProtectionPoliciesClient
	id                *string
	name              *string
	location          *string
	rg                string
	backupPolicyId    protectionpolicies.BackupPolicyId
}

func (r *RecoveryServicesBackupPolicy) Filter() error {
	return nil
}

func (r *RecoveryServicesBackupPolicy) Remove(ctx context.Context) error {
	_, err := r.protectionsClient.Delete(ctx, r.backupPolicyId)
	return err
}

func (r *RecoveryServicesBackupPolicy) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", r.name)
	properties.Set("Location", r.location)
	properties.Set("ResourceGroup", r.rg)

	return properties
}

func (r *RecoveryServicesBackupPolicy) String() string {
	return ptr.ToString(r.name)
}

type RecoveryServicesBackupPolicyLister struct {
}

func (l RecoveryServicesBackupPolicyLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancel()

	log := logrus.
		WithField("r", RecoveryServicesBackupPolicyResource).
		WithField("s", opts.SubscriptionId).
		WithField("rg", opts.ResourceGroup)

	log.Trace("creating client")

	vaultsClient, err := vaults.NewVaultsClientWithBaseURI(environments.AzurePublic().ResourceManager) // TODO: pass in the endpoint
	if err != nil {
		return nil, err
	}
	vaultsClient.Client.Authorizer = opts.Authorizers.Management

	client := backuppolicies.NewBackupPoliciesClientWithBaseURI("https://management.azure.com") // TODO: pass in the endpoint
	client.Client.Authorizer = opts.Authorizers.Management
	client.Client.RetryAttempts = 1
	client.Client.RetryDuration = time.Second * 2

	protectionsClient := protectionpolicies.NewProtectionPoliciesClientWithBaseURI("https://management.azure.com") // TODO: pass in the endpoint
	protectionsClient.Client.Authorizer = opts.Authorizers.Management
	protectionsClient.Client.RetryAttempts = 1
	protectionsClient.Client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("listing resources")

	vaultsRes, err := vaultsClient.ListByResourceGroupComplete(ctx, commonids.NewResourceGroupID(opts.SubscriptionId, opts.ResourceGroup))
	if err != nil {
		return nil, err
	}

	for _, v := range vaultsRes.Items {
		vaultId := backuppolicies.NewVaultID(opts.SubscriptionId, opts.ResourceGroup, ptr.ToString(v.Name))
		items, err := client.ListComplete(ctx, vaultId, backuppolicies.DefaultListOperationOptions())
		if err != nil {
			return nil, err
		}

		for _, item := range items.Items {
			resources = append(resources, &RecoveryServicesBackupPolicy{
				client:            client,
				protectionsClient: protectionsClient,
				id:                item.Id,
				name:              item.Name,
				location:          item.Location,
				rg:                opts.ResourceGroup,
				backupPolicyId:    protectionpolicies.NewBackupPolicyID(opts.SubscriptionId, opts.ResourceGroup, ptr.ToString(v.Name), ptr.ToString(item.Name)),
			})
		}
	}

	log.Trace("done")

	return resources, nil
}
