package resources

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/preview/security/mgmt/v3.0/security"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/nuke"
)

const SecurityWorkspaceResource = "SecurityWorkspace"

func init() {
	resource.Register(resource.Registration{
		Name:   SecurityWorkspaceResource,
		Scope:  nuke.Subscription,
		Lister: SecurityWorkspaceLister{},
	})
}

type SecurityWorkspace struct {
	client security.WorkspaceSettingsClient
	name   string
	scope  string
}

func (r *SecurityWorkspace) Remove() error {
	_, err := r.client.Delete(context.TODO(), r.name)
	return err
}

func (r *SecurityWorkspace) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", r.name)
	properties.Set("Scope", r.scope)

	return properties
}

func (r *SecurityWorkspace) String() string {
	return r.name
}

// -------------------------------------------------------------

type SecurityWorkspaceLister struct {
}

func (l SecurityWorkspaceLister) List(o interface{}) ([]resource.Resource, error) {
	opts := o.(nuke.ListerOpts)

	log := logrus.
		WithField("r", SecurityWorkspaceResource).
		WithField("s", opts.SubscriptionId)

	log.Trace("creating client")

	client := security.NewWorkspaceSettingsClient(opts.SubscriptionId)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("listing resources")

	ctx := context.TODO()
	list, err := client.List(ctx)
	if err != nil {
		return nil, err
	}

	for list.NotDone() {
		log.Trace("listing not done")
		for _, g := range list.Values() {
			resources = append(resources, &SecurityWorkspace{
				client: client,
				name:   *g.Name,
				scope:  *g.Scope,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	log.Trace("done")

	return resources, nil
}
