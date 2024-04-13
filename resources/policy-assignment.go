package resources

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/preview/resources/mgmt/2021-06-01-preview/policy" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/nuke"
)

const PolicyAssignmentResource = "PolicyAssignment"

func init() {
	registry.Register(&registry.Registration{
		Name:   PolicyAssignmentResource,
		Scope:  nuke.Subscription,
		Lister: &PolicyAssignmentLister{},
	})
}

type PolicyAssignment struct {
	client          policy.AssignmentsClient
	Name            string
	Scope           string
	EnforcementMode string
}

func (r *PolicyAssignment) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, r.Scope, r.Name)
	return err
}

func (r *PolicyAssignment) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *PolicyAssignment) String() string {
	return r.Name
}

type PolicyAssignmentLister struct {
}

func (l PolicyAssignmentLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	log := logrus.WithField("r", PolicyAssignmentResource).WithField("s", opts.SubscriptionID)

	client := policy.NewAssignmentsClient(opts.SubscriptionID)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list policy assignments")

	list, err := client.List(ctx, "", nil)
	if err != nil {
		return nil, err
	}

	log.Trace("listing policy assignments")

	for list.NotDone() {
		log.Trace("list not done")
		for _, g := range list.Values() {
			resources = append(resources, &PolicyAssignment{
				client:          client,
				Name:            *g.Name,
				Scope:           *g.Scope,
				EnforcementMode: string(g.EnforcementMode),
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	log.Trace("done")

	return resources, nil
}
