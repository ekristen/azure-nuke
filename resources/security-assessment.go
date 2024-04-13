package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
	"github.com/Azure/go-autorest/autorest/to"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/nuke"
)

const SecurityAssessmentResource = "SecurityAssessment"

func init() {
	registry.Register(&registry.Registration{
		Name:   SecurityAssessmentResource,
		Scope:  nuke.Subscription,
		Lister: &SecurityAssessmentLister{},
	})
}

type SecurityAssessment struct {
	client *armsecurity.AssessmentsClient

	ID         *string
	ResourceID *string
	Name       *string
	Status     *string
}

func (r *SecurityAssessment) Filter() error {
	return nil
}

func (r *SecurityAssessment) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, strings.TrimLeft(to.String(r.ResourceID), "/"), to.String(r.Name), nil)
	return err
}

func (r *SecurityAssessment) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *SecurityAssessment) String() string {
	return *r.Name
}

// -------------------------------------------------------------

type SecurityAssessmentLister struct {
}

func (l SecurityAssessmentLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	log := logrus.
		WithField("r", SecurityAssessmentResource).
		WithField("s", opts.SubscriptionID)

	log.Trace("creating client")

	clientFactory, err := armsecurity.NewClientFactory(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	client := clientFactory.NewAssessmentsClient()

	resources := make([]resource.Resource, 0)

	log.Trace("listing resources")

	pager := client.NewListPager(fmt.Sprintf("/subscriptions/%s", opts.SubscriptionID), nil)
	for pager.More() {
		log.Trace("listing not done")
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Fatalf("failed to advance page: %v", err)
		}
		for _, v := range page.Value {
			// TODO: this might not be right -- but without it it wants to delete things it cannot delete
			d := v.Properties.ResourceDetails.GetResourceDetails()
			if d.Source == nil {
				continue
			}

			parts := strings.Split(to.String(v.ID), "/providers/Microsoft.Security")
			resources = append(resources, &SecurityAssessment{
				client:     client,
				ResourceID: to.StringPtr(parts[0]),
				ID:         v.ID,
				Name:       v.Name,
			})
		}
	}

	log.WithField("total", len(resources)).Trace("done")

	return resources, nil
}
