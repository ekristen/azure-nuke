package azure

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
	"github.com/Azure/azure-sdk-for-go/services/subscription/mgmt/2020-09-01/subscription"
)

type Tenant struct {
	Authorizers *Authorizers

	ID              string
	SubscriptionIds []string
	TenantIds       []string

	Locations      map[string][]string
	ResourceGroups map[string][]string
}

func NewTenant(pctx context.Context, authorizers *Authorizers, tenantId string, subscriptionIds []string) (*Tenant, error) {
	ctx, cancel := context.WithTimeout(pctx, time.Second*15)
	defer cancel()

	log := logrus.WithField("handler", "NewTenant")
	log.Trace("start: NewTenant")

	tenant := &Tenant{
		Authorizers:     authorizers,
		ID:              tenantId,
		TenantIds:       make([]string, 0),
		SubscriptionIds: make([]string, 0),
		Locations:       make(map[string][]string),
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
			tenant.TenantIds = append(tenant.TenantIds, *t.TenantID)
		}
	}

	client := subscription.NewSubscriptionsClient()
	client.Authorizer = authorizers.Management

	logrus.Trace("listing subscriptions")
	for list, err := client.List(ctx); list.NotDone(); err = list.NextWithContext(ctx) {
		if err != nil {
			return nil, err
		}
		for _, s := range list.Values() {
			if len(subscriptionIds) > 0 && !slices.Contains(subscriptionIds, *s.SubscriptionID) {
				logrus.Warnf("skipping subscription id: %s (reason: not requested)", *s.SubscriptionID)
				continue
			}

			logrus.Tracef("adding subscriptions id: %s", *s.SubscriptionID)
			tenant.SubscriptionIds = append(tenant.SubscriptionIds, *s.SubscriptionID)

			logrus.Trace("listing locations")
			res, err := client.ListLocations(ctx, *s.SubscriptionID)
			if err != nil {
				return nil, err
			}
			for _, loc := range *res.Value {
				logrus.Tracef("adding location: %s", *loc.Name)
				tenant.Locations[*s.SubscriptionID] = append(tenant.Locations[*s.SubscriptionID], *loc.Name)
			}

			logrus.Trace("listing resource groups")
			groupsClient := resources.NewGroupsClient(*s.SubscriptionID)
			groupsClient.Authorizer = authorizers.Management

			for list, err := groupsClient.List(ctx, "", nil); list.NotDone(); err = list.NextWithContext(ctx) {
				if err != nil {
					return nil, err
				}

				for _, g := range list.Values() {
					logrus.Tracef("resource group name: %s", *g.Name)
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
