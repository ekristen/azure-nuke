package azure

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources/v4"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
)

type Tenant struct {
	Authorizers *Authorizers

	ID              string
	SubscriptionIds []string
	TenantIds       []string

	Regions        map[string][]string
	ResourceGroups map[string][]string
}

func NewTenant( //nolint:gocyclo
	pctx context.Context, authorizers *Authorizers,
	tenantID string, subscriptionIDs, regions []string,
) (*Tenant, error) {
	ctx, cancel := context.WithTimeout(pctx, time.Second*15)
	defer cancel()

	log := logrus.WithField("handler", "NewTenant")
	log.Trace("start: NewTenant")

	tenant := &Tenant{
		Authorizers:     authorizers,
		ID:              tenantID,
		TenantIds:       make([]string, 0),
		SubscriptionIds: make([]string, 0),
		Regions:         make(map[string][]string),
		ResourceGroups:  make(map[string][]string),
	}

	tenantClient, err := armsubscription.NewTenantsClient(authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	log.Trace("attempting to list tenants")
	tenantPager := tenantClient.NewListPager(nil)
	for tenantPager.More() {
		page, err := tenantPager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, t := range page.Value {
			log.Tracef("adding tenant: %s", *t.TenantID)
			tenant.TenantIds = append(tenant.TenantIds, *t.TenantID)
		}
	}

	subClient, err := armsubscription.NewSubscriptionsClient(authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	log.Trace("listing subscriptions")
	subPager := subClient.NewListPager(nil)
	for subPager.More() {
		page, err := subPager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, s := range page.Value {
			slog := log.WithField("subscription_id", *s.SubscriptionID)
			if len(subscriptionIDs) > 0 && !slices.Contains(subscriptionIDs, *s.SubscriptionID) {
				slog.Warnf("skipping subscription id: %s (reason: not requested)", *s.SubscriptionID)
				continue
			}

			slog.Trace("adding subscription")
			tenant.SubscriptionIds = append(tenant.SubscriptionIds, *s.SubscriptionID)

			slog.Trace("listing resource groups")
			groupsClient, err := armresources.NewResourceGroupsClient(*s.SubscriptionID, authorizers.IdentityCreds, nil)
			if err != nil {
				return nil, err
			}

			slog.Debugf("configured regions: %v", regions)
			groupsPager := groupsClient.NewListPager(nil)
			for groupsPager.More() {
				groupsPage, err := groupsPager.NextPage(ctx)
				if err != nil {
					return nil, err
				}

				for _, g := range groupsPage.Value {
					// If the region isn't in the list of regions we want to include, skip it
					if !slices.Contains(regions, ptr.ToString(g.Location)) && !slices.Contains(regions, "all") {
						continue
					}

					slog.Debugf("resource group name: %s", *g.Name)
					tenant.ResourceGroups[*s.SubscriptionID] = append(tenant.ResourceGroups[*s.SubscriptionID], *g.Name)
				}
			}
		}
	}

	for _, subscriptionID := range subscriptionIDs {
		if !slices.Contains(tenant.SubscriptionIds, subscriptionID) {
			return nil, fmt.Errorf("subscription not found: %s", subscriptionID)
		}
	}

	if len(tenant.TenantIds) == 0 {
		return nil, fmt.Errorf("tenant not found: %s", tenant.ID)
	}

	if tenant.TenantIds[0] != tenant.ID {
		return nil, fmt.Errorf("tenant ids do not match")
	}

	return tenant, nil
}
