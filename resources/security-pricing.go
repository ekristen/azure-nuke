package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/preview/security/mgmt/v3.0/security"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/nuke"
)

const SecurityPricingResource = "SecurityPricing"

func init() {
	registry.Register(&registry.Registration{
		Name:   SecurityPricingResource,
		Scope:  nuke.Subscription,
		Lister: &SecurityPricingLister{},
		DependsOn: []string{
			SecurityAlertResource,
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

func (r *SecurityPricing) Remove(ctx context.Context) error {
	_, err := r.client.Update(ctx, r.name, security.Pricing{
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
}

func (l SecurityPricingLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	log := logrus.
		WithField("r", SecurityPricingResource).
		WithField("s", opts.SubscriptionId)

	log.Trace("creating client")

	client := security.NewPricingsClient(opts.SubscriptionId)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("listing resources")

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

	log.Trace("done")

	return resources, nil
}
