package resources

import (
	"context"
	"fmt"
	"github.com/ekristen/azure-nuke/pkg/nuke"
	"github.com/ekristen/cloud-nuke-sdk/pkg/resource"
	"github.com/ekristen/cloud-nuke-sdk/pkg/types"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/security/mgmt/v3.0/security"
)

func init() {
	resource.Register(resource.Registration{
		Name:   "SecurityPricing",
		Scope:  nuke.Subscription,
		Lister: SecurityPricingLister{},
		DependsOn: []string{
			"SecurityAlert",
		},
	})
}

type SecurityPricing struct {
	client security.PricingsClient
	id     string
	name   string
	tier   string
}

func (r *SecurityPricing) Filter() error {
	if r.tier == "Free" {
		return fmt.Errorf("already set to free tier")
	}
	return nil
}

func (r *SecurityPricing) Remove() error {
	_, err := r.client.Update(context.TODO(), r.name, security.Pricing{
		PricingProperties: &security.PricingProperties{
			PricingTier: "Free",
		},
	})
	return err
}

func (r *SecurityPricing) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("ID", r.id)
	properties.Set("Name", r.name)
	properties.Set("PricingTier", r.tier)

	return properties
}

func (r *SecurityPricing) String() string {
	return r.name
}

// -------------------------------------------------------------------

type SecurityPricingLister struct {
	opts nuke.ListerOpts
}

func (l SecurityPricingLister) SetOptions(opts interface{}) {
	l.opts = opts.(nuke.ListerOpts)
}

func (l SecurityPricingLister) List() ([]resource.Resource, error) {
	log := logrus.
		WithField("resource", "SecurityPricing").
		WithField("scope", nuke.Subscription).
		WithField("subscription", l.opts.SubscriptionId)

	log.Trace("creating client")

	client := security.NewPricingsClient(l.opts.SubscriptionId)
	client.Authorizer = l.opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("listing resources")

	ctx := context.TODO()
	list, err := client.List(ctx)
	if err != nil {
		return nil, err
	}

	for _, price := range *list.Value {
		resources = append(resources, &SecurityPricing{
			client: client,
			id:     *price.ID,
			name:   *price.Name,
			tier:   string(price.PricingTier),
		})
	}

	return resources, nil
}
