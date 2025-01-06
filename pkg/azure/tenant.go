package azure

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"       //nolint:staticcheck
	"github.com/Azure/azure-sdk-for-go/services/subscription/mgmt/2020-09-01/subscription" //nolint:staticcheck
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

	tenantClient := subscription.NewTenantsClient()
	tenantClient.Authorizer = authorizers.Management

	log.Trace("attempting to list tenants")
	for list, err := tenantClient.List(ctx); list.NotDone(); err = list.NextWithContext(ctx) {
		if err != nil {
			return nil, err
		}
		for _, t := range list.Values() {
			log.Tracef("adding tenant: %s", *t.TenantID)
			tenant.TenantIds = append(tenant.TenantIds, *t.TenantID)
		}
	}

	client := subscription.NewSubscriptionsClient()
	client.Authorizer = authorizers.Management

	log.Trace("listing subscriptions")
	for list, err := client.List(ctx); list.NotDone(); err = list.NextWithContext(ctx) {
		if err != nil {
			return nil, err
		}
		for _, s := range list.Values() {
			slog := log.WithField("subscription_id", *s.SubscriptionID)
			if len(subscriptionIDs) > 0 && !slices.Contains(subscriptionIDs, *s.SubscriptionID) {
				slog.Warnf("skipping subscription id: %s (reason: not requested)", *s.SubscriptionID)
				continue
			}

			slog.Trace("adding subscription")
			tenant.SubscriptionIds = append(tenant.SubscriptionIds, *s.SubscriptionID)

			slog.Trace("listing resource groups")
			groupsClient := resources.NewGroupsClient(*s.SubscriptionID)
			groupsClient.Authorizer = authorizers.Management

			slog.Debugf("configured regions: %v", regions)
			for list, err := groupsClient.List(ctx, "", nil); list.NotDone(); err = list.NextWithContext(ctx) {
				if err != nil {
					return nil, err
				}

				for _, g := range list.Values() {
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

	if len(tenant.TenantIds) == 0 {
		return nil, fmt.Errorf("tenant not found: %s", tenant.ID)
	}

	if tenant.TenantIds[0] != tenant.ID {
		return nil, fmt.Errorf("tenant ids do not match")
	}

	return tenant, nil
}
