package resources

import (
	"context"
	"github.com/aws/smithy-go/ptr"
	"github.com/ekristen/azure-nuke/pkg/nuke"
	"github.com/ekristen/cloud-nuke-sdk/pkg/resource"
	"github.com/ekristen/cloud-nuke-sdk/pkg/types"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"
	"github.com/hashicorp/go-azure-sdk/resource-manager/recoveryservices/2023-02-01/vaults"
	"github.com/hashicorp/go-azure-sdk/resource-manager/recoveryservicesbackup/2023-02-01/backuppolicies"
	"github.com/hashicorp/go-azure-sdk/resource-manager/recoveryservicesbackup/2023-02-01/protectionpolicies"
	"github.com/sirupsen/logrus"
	"time"
)

func init() {
	resource.Register(resource.Registration{
		Name:   "RecoveryServicesBackupPolicy",
		Scope:  nuke.ResourceGroup,
		Lister: RecoveryServicesBackupPolicyLister{},
		DependsOn: []string{
			"RecoveryServicesBackupProtectedItem",
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

func (r *RecoveryServicesBackupPolicy) Remove() error {
	_, err := r.protectionsClient.Delete(context.TODO(), r.backupPolicyId)
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

func (l RecoveryServicesBackupPolicyLister) List(o interface{}) ([]resource.Resource, error) {
	opts := o.(nuke.ListerOpts)

	log := logrus.
		WithField("resource", "RecoveryServicesBackupPolicy").
		WithField("scope", nuke.ResourceGroup).
		WithField("subscription", opts.SubscriptionId).
		WithField("rg", opts.ResourceGroup)

	log.Trace("creating client")

	vaultsClient := vaults.NewVaultsClientWithBaseURI("https://management.azure.com") // TODO: pass in the endpoint
	vaultsClient.Client.Authorizer = opts.Authorizers.Management
	vaultsClient.Client.RetryAttempts = 1
	vaultsClient.Client.RetryDuration = time.Second * 2

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

	ctx := context.TODO()

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

	return resources, nil
}
