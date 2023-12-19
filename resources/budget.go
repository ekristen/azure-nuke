package resources

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"
	"github.com/hashicorp/go-azure-sdk/resource-manager/consumption/2021-10-01/budgets"
	"github.com/hashicorp/go-azure-sdk/sdk/environments"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/ekristen/azure-nuke/pkg/resource"
	"github.com/ekristen/azure-nuke/pkg/types"
)

type Budget struct {
	client *budgets.BudgetsClient
	name   string
	rg     string
}

func init() {
	resource.RegisterV2(resource.Registration{
		Name:   "Budget",
		Scope:  resource.Subscription,
		Lister: ListBudget,
	})
}

func ListBudget(opts resource.ListerOpts) ([]resource.Resource, error) {
	logrus.Tracef("subscription id: %s", opts.SubscriptionId)

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

	logrus.Trace("attempting to list budgets for subscription")

	pctx := context.Background()
	ctx, cancel := context.WithDeadline(pctx, time.Now().Add(10*time.Second))
	defer cancel()

	list, err := client.List(ctx, commonids.ScopeId{
		Scope: fmt.Sprintf("/subscriptions/%s", opts.SubscriptionId),
	})
	if err != nil {
		return nil, err
	}

	logrus.Trace("listing ....")

	for _, entry := range *list.Model {
		resources = append(resources, &Budget{
			client: client,
			name:   *entry.Name,
		})
	}

	return resources, nil
}

func (r *Budget) Remove() error {
	_, err := r.client.Delete(context.TODO(), budgets.ScopedBudgetId{
		Scope:      "",
		BudgetName: r.name,
	})
	return err
}

func (r *Budget) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", r.name)

	return properties
}

func (r *Budget) String() string {
	return r.name
}
