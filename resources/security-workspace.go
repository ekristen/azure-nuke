package resources

import (
	"context"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const SecurityWorkspaceResource = "SecurityWorkspace"

func init() {
	registry.Register(&registry.Registration{
		Name:     SecurityWorkspaceResource,
		Scope:    azure.SubscriptionScope,
		Resource: &SecurityWorkspace{},
		Lister:   &SecurityWorkspaceLister{},
	})
}

type SecurityWorkspace struct {
	*BaseResource `property:",inline"`

	client *armsecurity.WorkspaceSettingsClient
	Name   *string `description:"The name of the workspace"`
	Scope  *string `description:"The scope of the workspace"`
}

func (r *SecurityWorkspace) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.Name, nil)
	return err
}

func (r *SecurityWorkspace) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *SecurityWorkspace) String() string {
	return *r.Name
}

// -------------------------------------------------------------

type SecurityWorkspaceLister struct{}

func (l SecurityWorkspaceLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	log := logrus.
		WithField("r", SecurityWorkspaceResource).
		WithField("s", opts.SubscriptionID)

	log.Trace("creating client")

	client, err := armsecurity.NewWorkspaceSettingsClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)

	log.Trace("listing resources")

	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, entity := range page.Value {
			var scope *string
			if entity.Properties != nil {
				scope = entity.Properties.Scope
			}

			resources = append(resources, &SecurityWorkspace{
				BaseResource: &BaseResource{
					Region: ptr.String("global"),
				},
				client: client,
				Name:   entity.Name,
				Scope:  scope,
			})
		}
	}

	log.Trace("done")

	return resources, nil
}
