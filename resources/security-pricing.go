package resources

import (
	"context"
	"fmt"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"

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

	client         *armsecurity.PricingsClient
	subscriptionID string
	Name           *string
	PricingTier    string
}

func (r *SecurityPricing) Filter() error {
	if r.PricingTier == "Free" {
		return fmt.Errorf("already set to default, free tier")
	}

	if r.PricingTier == "Standard" && ptr.ToString(r.Name) == "Discovery" || ptr.ToString(r.Name) == "FoundationalCspm" {
		return fmt.Errorf("already set to default, standard tier")
	}

	return nil
}

func (r *SecurityPricing) Remove(ctx context.Context) error {
	pricingTier := armsecurity.PricingTier("Free")
	if ptr.ToString(r.Name) == "Discovery" || ptr.ToString(r.Name) == "FoundationalCspm" {
		pricingTier = armsecurity.PricingTier("Standard")
	}

	scopeID := "subscriptions/" + r.subscriptionID
	_, err := r.client.Update(ctx, scopeID, *r.Name, armsecurity.Pricing{
		Properties: &armsecurity.PricingProperties{
			PricingTier: &pricingTier,
		},
	}, nil)
	return err
}

func (r *SecurityPricing) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *SecurityPricing) String() string {
	return *r.Name
}

// -------------------------------------------------------------------

type SecurityPricingLister struct{}

func (l SecurityPricingLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	log := logrus.
		WithField("r", SecurityPricingResource).
		WithField("s", opts.SubscriptionID)

	log.Trace("creating client")

	client, err := armsecurity.NewPricingsClient(opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)

	log.Trace("listing resources")

	scopeID := "subscriptions/" + opts.SubscriptionID
	list, err := client.List(ctx, scopeID, nil)
	if err != nil {
		return nil, err
	}

	for _, price := range list.Value {
		var pricingTier string
		if price.Properties != nil && price.Properties.PricingTier != nil {
			pricingTier = string(*price.Properties.PricingTier)
		}

		resources = append(resources, &SecurityPricing{
			BaseResource: &BaseResource{
				Region: ptr.String("global"),
			},
			client:         client,
			subscriptionID: opts.SubscriptionID,
			Name:           price.Name,
			PricingTier:    pricingTier,
		})
	}

	log.Trace("done")

	return resources, nil
}
