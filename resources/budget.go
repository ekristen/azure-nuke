package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"
	"github.com/hashicorp/go-azure-sdk/resource-manager/consumption/2021-10-01/budgets"
	"github.com/hashicorp/go-azure-sdk/sdk/environments"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/nuke"
)

const BudgetResource = "Budget"

func init() {
	registry.Register(&registry.Registration{
		Name:     BudgetResource,
		Scope:    nuke.Subscription,
		Resource: &Budget{},
		Lister:   &BudgetLister{},
	})
}

type Budget struct {
	client *budgets.BudgetsClient

	ID             *string
	Name           *string
	SubscriptionID *string
}

type BudgetLister struct{}

func (l BudgetLister) List(pctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	log := logrus.WithField("r", BudgetResource).WithField("s", opts.SubscriptionID)

	resources := make([]resource.Resource, 0)

	client, err := budgets.NewBudgetsClientWithBaseURI(environments.AzurePublic().ResourceManager)
	if err != nil {
		return nil, err
	}
	client.Client.Authorizer = opts.Authorizers.Management

	log.Trace("attempting to list budgets for subscription")

	ctx, cancel := context.WithDeadline(pctx, time.Now().Add(10*time.Second))
	defer cancel()

	list, err := client.List(ctx, commonids.ScopeId{
		Scope: fmt.Sprintf("/subscriptions/%s", opts.SubscriptionID),
	})
	if err != nil {
		return nil, err
	}

	log.Trace("listing budgets for subscription")

	for _, entry := range *list.Model {
		resources = append(resources, &Budget{
			client:         client,
			ID:             entry.Id,
			Name:           entry.Name,
			SubscriptionID: ptr.String(opts.SubscriptionID), // note: this is just the guid
		})
	}

	log.Trace("done")

	return resources, nil
}

func (r *Budget) Remove(ctx context.Context) error {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(10*time.Second))
	defer cancel()

	_, err := r.client.Delete(ctx, budgets.ScopedBudgetId{
		Scope:      fmt.Sprintf("/subscriptions/%s", ptr.ToString(r.SubscriptionID)),
		BudgetName: *r.Name,
	})
	return err
}

func (r *Budget) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *Budget) String() string {
	return *r.Name
}
