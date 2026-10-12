package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/consumption/armconsumption/v2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const BudgetResource = "Budget"

func init() {
	registry.Register(&registry.Registration{
		Name:     BudgetResource,
		Scope:    azure.SubscriptionScope,
		Resource: &Budget{},
		Lister:   &BudgetLister{},
	})
}

type Budget struct {
	*BaseResource `property:",inline"`

	client *armconsumption.BudgetsClient
	ID     *string
	Name   *string
}

type BudgetLister struct{}

func (l BudgetLister) List(pctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)
	var resources []resource.Resource

	log := logrus.WithField("r", BudgetResource).WithField("s", opts.SubscriptionID)

	client, err := armconsumption.NewBudgetsClient(opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	log.Trace("attempting to list budgets for subscription")

	ctx, cancel := context.WithDeadline(pctx, time.Now().Add(10*time.Second))
	defer cancel()

	scope := fmt.Sprintf("/subscriptions/%s", opts.SubscriptionID)

	log.Trace("listing budgets for subscription")

	pager := client.NewListPager(scope, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, entry := range page.Value {
			resources = append(resources, &Budget{
				BaseResource: &BaseResource{
					Region:         ptr.String("global"),
					SubscriptionID: ptr.String(opts.SubscriptionID),
				},
				client: client,
				ID:     entry.ID,
				Name:   entry.Name,
			})
		}
	}

	log.Trace("done")

	return resources, nil
}

func (r *Budget) Remove(ctx context.Context) error {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(10*time.Second))
	defer cancel()

	scope := fmt.Sprintf("/subscriptions/%s", ptr.ToString(r.SubscriptionID))
	_, err := r.client.Delete(ctx, scope, *r.Name, nil)
	return err
}

func (r *Budget) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *Budget) String() string {
	return *r.Name
}
