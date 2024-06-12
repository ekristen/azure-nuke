package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/preview/security/mgmt/v3.0/security" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const SecurityPricingResource = "SecurityPricing"

func init() {
	registry.Register(&registry.Registration{
		Name:     SecurityPricingResource,
		Scope:    azure.SubscriptionScope,
		Resource: &SecurityPricing{},
		Lister:   &SecurityPricingLister{},
		DependsOn: []string{
			SecurityAlertResource,
		},
	})
}

type SecurityPricing struct {
	*BaseResource `property:",inline"`

	client      security.PricingsClient
	Name        *string
	PricingTier string
}

func (r *SecurityPricing) Filter() error {
	if r.PricingTier == "Free" {
		return fmt.Errorf("already set to free tier")
	}
	return nil
}

func (r *SecurityPricing) Remove(ctx context.Context) error {
	_, err := r.client.Update(ctx, *r.Name, security.Pricing{
		PricingProperties: &security.PricingProperties{
			PricingTier: "Free",
		},
	})
	return err
}

func (r *SecurityPricing) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *SecurityPricing) String() string {
	return *r.Name
}

// -------------------------------------------------------------------

type SecurityPricingLister struct {
}

func (l SecurityPricingLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	log := logrus.
		WithField("r", SecurityPricingResource).
		WithField("s", opts.SubscriptionID)

	log.Trace("creating client")

	client := security.NewPricingsClient(opts.SubscriptionID)
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
			BaseResource: &BaseResource{
				Region: ptr.String("global"),
			},
			client:      client,
			Name:        price.Name,
			PricingTier: string(price.PricingTier),
		})
	}

	log.Trace("done")

	return resources, nil
}
