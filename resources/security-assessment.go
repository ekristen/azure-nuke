package resources

import (
	"context"
	"fmt"
	"github.com/ekristen/azure-nuke/pkg/resource"
	"github.com/ekristen/azure-nuke/pkg/types"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/security/mgmt/v3.0/security"
)

type SecurityAssessment struct {
	client security.AssessmentsClient
	id     string
	name   string
}

func init() {
	resource.RegisterV2(resource.Registration{
		Name:   "SecurityAssessment",
		Scope:  resource.Subscription,
		Lister: ListSecurityAssessment,
	})
}

func ListSecurityAssessment(opts resource.ListerOpts) ([]resource.Resource, error) {
	log := logrus.
		WithField("resource", "SecurityAssessment").
		WithField("scope", resource.Subscription).
		WithField("subscription", opts.SubscriptionId)

	log.Trace("creating client")

	client := security.NewAssessmentsClient(opts.SubscriptionId)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("listing resources")

	ctx := context.TODO()
	list, err := client.List(ctx, fmt.Sprintf("/subscriptions/%s", opts.SubscriptionId))
	if err != nil {
		return nil, err
	}

	for list.NotDone() {
		log.Trace("listing not done")
		for _, g := range list.Values() {
			resources = append(resources, &SecurityAssessment{
				client: client,
				id:     *g.ID,
				name:   *g.Name,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	return resources, nil
}

func (r *SecurityAssessment) Remove() error {
	_, err := r.client.Delete(context.TODO(), r.id, r.name)
	return err
}

func (r *SecurityAssessment) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("ID", r.id)
	properties.Set("Name", r.name)

	return properties
}

func (r *SecurityAssessment) String() string {
	return r.name
}
