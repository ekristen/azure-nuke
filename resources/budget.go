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

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/nuke"
)

const BudgetResource = "Budget"

func init() {
	resource.Register(resource.Registration{
		Name:   BudgetResource,
		Scope:  nuke.Subscription,
		Lister: &BudgetLister{},
	})
}

type Budget struct {
	client *budgets.BudgetsClient
	name   *string
	id     *string
}

type BudgetLister struct{}

func (l BudgetLister) List(pctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	log := logrus.WithField("r", BudgetResource).WithField("s", opts.SubscriptionId)

	resources := make([]resource.Resource, 0)

	// TODO: move higher up in call stack
	env, err := environments.FromName("global")
	if err != nil {
		return nil, err
	}
	client, err := budgets.NewBudgetsClientWithBaseURI(env.ResourceManager)
	if err != nil {
		return nil, err
	}
	client.Client.Authorizer = opts.Authorizers.Management
	//client.Authorizer = opts.Authorizers.Management
	//client.RetryAttempts = 1
	//client.RetryDuration = time.Second * 2

	log.Trace("attempting to list budgets for subscription")

	ctx, cancel := context.WithDeadline(pctx, time.Now().Add(10*time.Second))
	defer cancel()

	list, err := client.List(ctx, commonids.ScopeId{
		Scope: fmt.Sprintf("/subscriptions/%s", opts.SubscriptionId),
	})
	if err != nil {
		return nil, err
	}

	log.Trace("listing budgets for subscription")

	for _, entry := range *list.Model {
		resources = append(resources, &Budget{
			client: client,
			name:   entry.Name,
			id:     entry.Id,
		})
	}

	log.Trace("done")

	return resources, nil
}

func (r *Budget) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, budgets.ScopedBudgetId{
		Scope:      "",
		BudgetName: ptr.ToString(r.name),
	})
	return err
}

func (r *Budget) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", r.name)

	return properties
}

func (r *Budget) String() string {
	return ptr.ToString(r.name)
}
