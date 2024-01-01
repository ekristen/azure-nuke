package resources

import (
	"context"
	"github.com/ekristen/azure-nuke/pkg/nuke"
	"github.com/ekristen/cloud-nuke-sdk/pkg/resource"
	"github.com/ekristen/cloud-nuke-sdk/pkg/types"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/resources/mgmt/2021-06-01-preview/policy"
)

const PolicyAssignmentResource = "PolicyAssignment"

func init() {
	resource.Register(resource.Registration{
		Name:   PolicyAssignmentResource,
		Scope:  nuke.Subscription,
		Lister: PolicyAssignmentLister{},
	})
}

type PolicyAssignment struct {
	client          policy.AssignmentsClient
	name            string
	scope           string
	enforcementMode string
}

func (r *PolicyAssignment) Remove() error {
	_, err := r.client.Delete(context.TODO(), r.scope, r.name)
	return err
}

func (r *PolicyAssignment) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", r.name)
	properties.Set("Scope", r.scope)
	properties.Set("EnforcementMode", r.enforcementMode)

	return properties
}

func (r *PolicyAssignment) String() string {
	return r.name
}

type PolicyAssignmentLister struct {
}

func (l PolicyAssignmentLister) List(o interface{}) ([]resource.Resource, error) {
	opts := o.(nuke.ListerOpts)

	log := logrus.WithField("r", PolicyAssignmentResource).WithField("s", opts.SubscriptionId)

	client := policy.NewAssignmentsClient(opts.SubscriptionId)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list policy assignments")

	ctx := context.TODO()
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
				name:            *g.Name,
				scope:           *g.Scope,
				enforcementMode: string(g.EnforcementMode),
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	log.Trace("done")

	return resources, nil
}
