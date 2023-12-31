package resources

import (
	"context"
	"github.com/aws/smithy-go/ptr"
	"github.com/ekristen/azure-nuke/pkg/nuke"
	"github.com/ekristen/cloud-nuke-sdk/pkg/resource"
	"github.com/ekristen/cloud-nuke-sdk/pkg/types"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/hashicorp/go-azure-sdk/resource-manager/recoveryservices/2023-02-01/vaults"
)

func init() {
	resource.Register(resource.Registration{
		Name:   "RecoveryServicesVault",
		Scope:  nuke.ResourceGroup,
		Lister: RecoveryServicesVaultLister{},
		DependsOn: []string{
			"RecoveryServicesBackupProtectedItem",
		},
	})
}

type RecoveryServicesVault struct {
	client   vaults.VaultsClient
	vaultId  vaults.VaultId
	id       *string
	name     *string
	location string
	rg       string
}

func (r *RecoveryServicesVault) Filter() error {
	return nil
}

func (r *RecoveryServicesVault) Remove() error {
	_, err := r.client.Delete(context.TODO(), r.vaultId)
	return err
}

func (r *RecoveryServicesVault) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", r.name)
	properties.Set("Location", r.location)
	properties.Set("ResourceGroup", r.rg)

	return properties
}

func (r *RecoveryServicesVault) String() string {
	return ptr.ToString(r.name)
}

type RecoveryServicesVaultLister struct {
	opts nuke.ListerOpts
}

func (l RecoveryServicesVaultLister) SetOptions(opts interface{}) {
	l.opts = opts.(nuke.ListerOpts)
}

func (l RecoveryServicesVaultLister) List() ([]resource.Resource, error) {
	log := logrus.
		WithField("resource", "RecoveryServicesVault").
		WithField("scope", nuke.ResourceGroup).
		WithField("subscription", l.opts.SubscriptionId).
		WithField("rg", l.opts.ResourceGroup)

	log.Trace("creating client")

	client := vaults.NewVaultsClientWithBaseURI("https://management.azure.com") // TODO: pass in the endpoint
	client.Client.Authorizer = l.opts.Authorizers.Management
	client.Client.RetryAttempts = 1
	client.Client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("listing resources")

	ctx := context.TODO()
	items, err := client.ListByResourceGroupComplete(ctx, commonids.NewResourceGroupID(l.opts.SubscriptionId, l.opts.ResourceGroup))
	if err != nil {
		return nil, err
	}

	for _, item := range items.Items {
		resources = append(resources, &RecoveryServicesVault{
			client:   client,
			id:       item.Id,
			name:     item.Name,
			location: item.Location,
			vaultId:  vaults.NewVaultID(l.opts.SubscriptionId, l.opts.ResourceGroup, ptr.ToString(item.Id)),
			rg:       l.opts.ResourceGroup,
		})
	}

	return resources, nil
}
