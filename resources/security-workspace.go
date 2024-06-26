package resources

import (
	"context"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/preview/security/mgmt/v3.0/security" //nolint:staticcheck

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

	client security.WorkspaceSettingsClient
	Name   *string `description:"The name of the workspace"`
	Scope  *string `description:"The scope of the workspace"`
}

func (r *SecurityWorkspace) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.Name)
	return err
}

func (r *SecurityWorkspace) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *SecurityWorkspace) String() string {
	return *r.Name
}

// -------------------------------------------------------------

type SecurityWorkspaceLister struct {
}

func (l SecurityWorkspaceLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	log := logrus.
		WithField("r", SecurityWorkspaceResource).
		WithField("s", opts.SubscriptionID)

	log.Trace("creating client")

	client := security.NewWorkspaceSettingsClient(opts.SubscriptionID)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("listing resources")

	list, err := client.List(ctx)
	if err != nil {
		return nil, err
	}

	for list.NotDone() {
		log.Trace("listing not done")
		for _, g := range list.Values() {
			resources = append(resources, &SecurityWorkspace{
				BaseResource: &BaseResource{
					Region: ptr.String("global"),
				},
				client: client,
				Name:   g.Name,
				Scope:  g.Scope,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	log.Trace("done")

	return resources, nil
}
