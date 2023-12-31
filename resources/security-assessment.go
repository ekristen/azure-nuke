package resources

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/ekristen/azure-nuke/pkg/nuke"
	"github.com/ekristen/cloud-nuke-sdk/pkg/resource"
	"github.com/ekristen/cloud-nuke-sdk/pkg/types"
	"github.com/sirupsen/logrus"
	"strings"
)

func init() {
	resource.Register(resource.Registration{
		Name:   "SecurityAssessment",
		Scope:  nuke.Subscription,
		Lister: SecurityAssessmentLister{},
	})
}

type SecurityAssessment struct {
	client     *armsecurity.AssessmentsClient
	id         *string
	resourceId *string
	name       *string
	status     *string
}

func (r *SecurityAssessment) Filter() error {
	return nil
}

func (r *SecurityAssessment) Remove() error {
	_, err := r.client.Delete(context.TODO(), strings.TrimLeft(to.String(r.resourceId), "/"), to.String(r.name), nil)
	return err
}

func (r *SecurityAssessment) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("ID", r.id)
	properties.Set("ResourceID", r.resourceId)
	properties.Set("Name", r.name)
	properties.Set("StatusCode", r.status)

	return properties
}

func (r *SecurityAssessment) String() string {
	return to.String(r.name)
}

// -------------------------------------------------------------

type SecurityAssessmentLister struct {
}

func (l SecurityAssessmentLister) List(o interface{}) ([]resource.Resource, error) {
	opts := o.(nuke.ListerOpts)

	log := logrus.
		WithField("resource", "SecurityAssessment").
		WithField("scope", nuke.Subscription).
		WithField("subscription", opts.SubscriptionId)

	log.Trace("creating client")

	clientFactory, err := armsecurity.NewClientFactory(opts.SubscriptionId, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	client := clientFactory.NewAssessmentsClient()

	resources := make([]resource.Resource, 0)

	log.Trace("listing resources")

	ctx := context.TODO()
	pager := client.NewListPager(fmt.Sprintf("/subscriptions/%s", opts.SubscriptionId), nil)
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
				resourceId: to.StringPtr(parts[0]),
				id:         v.ID,
				name:       v.Name,
			})
		}
	}

	return resources, nil
}
