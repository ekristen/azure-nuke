package resources

import (
	"context"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservices/v3"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const RecoveryServicesVaultResource = "RecoveryServicesVault"

func init() {
	registry.Register(&registry.Registration{
		Name:     RecoveryServicesVaultResource,
		Scope:    azure.ResourceGroupScope,
		Resource: &RecoveryServicesVault{},
		Lister:   &RecoveryServicesVaultLister{},
		DependsOn: []string{
			RecoveryServicesBackupProtectedItemResource,
		},
	})
}

type RecoveryServicesVault struct {
	*BaseResource `property:",inline"`

	client   *armrecoveryservices.VaultsClient
	removeOp *runtime.Poller[armrecoveryservices.VaultsClientDeleteResponse]
	ID       *string
	Name     *string
}

func (r *RecoveryServicesVault) Filter() error {
	return nil
}

func (r *RecoveryServicesVault) Remove(ctx context.Context) error {
	poller, err := r.client.BeginDelete(ctx, *r.ResourceGroup, *r.Name, nil)
	if err != nil {
		return err
	}

	r.removeOp = poller
	return nil
}

func (r *RecoveryServicesVault) HandleWait(ctx context.Context) error {
	return azure.WaitForPoller(ctx, r.removeOp)
}

func (r *RecoveryServicesVault) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *RecoveryServicesVault) String() string {
	return ptr.ToString(r.Name)
}

type RecoveryServicesVaultLister struct{}

func (l RecoveryServicesVaultLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancel()

	log := logrus.
		WithField("r", RecoveryServicesVaultResource).
		WithField("s", opts.SubscriptionID).
		WithField("rg", opts.ResourceGroup)

	log.Trace("creating client")

	client, err := armrecoveryservices.NewVaultsClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)

	log.Trace("listing resources")

	pager := client.NewListByResourceGroupPager(opts.ResourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, item := range page.Value {
			resources = append(resources, &RecoveryServicesVault{
				BaseResource: &BaseResource{
					Region:         item.Location,
					ResourceGroup:  ptr.String(opts.ResourceGroup),
					SubscriptionID: &opts.SubscriptionID,
				},
				client: client,
				ID:     item.ID,
				Name:   item.Name,
			})
		}
	}

	log.Trace("done")

	return resources, nil
}
