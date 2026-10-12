package resources

import (
	"context"
	"errors"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources/v4"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/settings"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const ResourceGroupResource = "ResourceGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:     ResourceGroupResource,
		Scope:    azure.SubscriptionScope,
		Resource: &ResourceGroup{},
		Lister:   &ResourceGroupLister{},
		DependsOn: []string{
			ManagementLockResource,
		},
		Settings: []string{
			"DeleteManaged",
			"DeleteLocks",
			"SoftDeleteKeyVaultKeys",
			"SoftDeleteKeyVaultKeysTag",
		},
	})
}

// ResourceGroup represents an Azure Resource Group.
type ResourceGroup struct {
	*BaseResource `property:",inline"`

	client   *armresources.ResourceGroupsClient
	creds    azcore.TokenCredential
	removeOp *runtime.Poller[armresources.ResourceGroupsClientDeleteResponse]
	settings *settings.Setting

	Name      *string            `description:"The Name of the resource group."`
	ManagedBy *string            `description:"The ID of the resource that manages the resource group."`
	Tags      map[string]*string `description:"The tags assigned to the resource group."`
}

func (r *ResourceGroup) Filter() error {
	if ptr.ToString(r.ManagedBy) != "" && !r.setting("DeleteManaged") {
		return errors.New("resource group is managed by another resource")
	}

	return nil
}

func (r *ResourceGroup) Remove(ctx context.Context) error {
	r.removeOp = nil

	// Keys before locks: a failed key delete leaves the locks in place, and the next round retries both
	if r.setting("SoftDeleteKeyVaultKeys") || r.setting("DeleteLocks") {
		cleanupCtx, cancel := context.WithTimeout(ctx, resourceGroupCleanupTimeout)
		defer cancel()

		deleting, err := r.deleting(cleanupCtx)
		if err != nil {
			return azure.IgnoreNotFound(err)
		}

		// a retry after a failed poll: the delete already runs, its vaults may be half gone
		if !deleting && r.setting("SoftDeleteKeyVaultKeys") {
			if err := r.softDeleteKeys(cleanupCtx); err != nil {
				return err
			}
		}
		if !deleting && r.setting("DeleteLocks") {
			if err := r.deleteLocks(cleanupCtx); err != nil {
				return err
			}
		}
	}

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancel()

	poller, err := r.client.BeginDelete(ctx, *r.Name, nil)
	if err != nil {
		return azure.IgnoreNotFound(err)
	}

	r.removeOp = poller
	return nil
}

func (r *ResourceGroup) HandleWait(ctx context.Context) error {
	return azure.WaitForPoller(ctx, r.removeOp)
}

func (r *ResourceGroup) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ResourceGroup) String() string {
	return *r.Name
}

func (r *ResourceGroup) Settings(setting *settings.Setting) {
	r.settings = setting
}

func (r *ResourceGroup) setting(key string) bool {
	return r.settings != nil && r.settings.GetBool(key)
}

// -------------------

type ResourceGroupLister struct{}

func (l ResourceGroupLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancel()

	log := logrus.WithField("r", ResourceGroupResource).WithField("s", opts.SubscriptionID)

	client, err := armresources.NewResourceGroupsClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list groups")

	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, entity := range page.Value {
			resources = append(resources, &ResourceGroup{
				BaseResource: &BaseResource{
					Region:         entity.Location,
					SubscriptionID: ptr.String(opts.SubscriptionID),
				},
				client:    client,
				creds:     opts.Authorizers.IdentityCreds,
				Name:      entity.Name,
				ManagedBy: entity.ManagedBy,
				Tags:      entity.Tags,
			})
		}
	}

	log.Trace("done")

	return resources, nil
}
