package resources

import (
	"context"
	"github.com/ekristen/azure-nuke/pkg/resource"
	"github.com/ekristen/azure-nuke/pkg/types"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/resources/mgmt/2021-06-01-preview/policy"
)

type PolicyAssignment struct {
	client policy.AssignmentsClient
	name   string
	scope  string
}

func init() {
	resource.RegisterV2(resource.Registration{
		Name:   "PolicyAssignment",
		Scope:  resource.Subscription,
		Lister: ListPolicyAssignment,
	})
}

func ListPolicyAssignment(opts resource.ListerOpts) ([]resource.Resource, error) {
	logrus.Tracef("subscription id: %s", opts.SubscriptionId)

	client := policy.NewAssignmentsClient(opts.SubscriptionId)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	logrus.Trace("attempting to list policy assignments")

	ctx := context.Background()
	list, err := client.List(ctx, "", nil)
	if err != nil {
		return nil, err
	}

	logrus.Trace("listing ....")

	for list.NotDone() {
		logrus.Trace("list not done")
		for _, g := range list.Values() {
			resources = append(resources, &PolicyAssignment{
				client: client,
				name:   *g.Name,
				scope:  *g.Scope,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	return resources, nil
}

func (r *PolicyAssignment) Remove() error {
	_, err := r.client.Delete(context.TODO(), r.scope, r.name)
	return err
}

func (r *PolicyAssignment) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", r.name)
	properties.Set("Scope", r.scope)

	return properties
}

func (r *PolicyAssignment) String() string {
	return r.name
}
