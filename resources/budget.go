package resources

import (
	"context"
	"fmt"
	"github.com/ekristen/azure-nuke/pkg/nuke"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"
	"github.com/hashicorp/go-azure-sdk/resource-manager/consumption/2021-10-01/budgets"
	"github.com/hashicorp/go-azure-sdk/sdk/environments"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/ekristen/cloud-nuke-sdk/pkg/resource"
	"github.com/ekristen/cloud-nuke-sdk/pkg/types"
)

func init() {
	resource.Register(resource.Registration{
		Name:   "Budget",
		Scope:  nuke.Subscription,
		Lister: BudgetLister{},
	})
}

type Budget struct {
	client *budgets.BudgetsClient
	name   string
	rg     string
}

type BudgetLister struct {
	opts nuke.ListerOpts
}

func (l BudgetLister) SetOptions(opts interface{}) {
	l.opts = opts.(nuke.ListerOpts)
}

func (l BudgetLister) List(opts interface{}) ([]resource.Resource, error) {
	opts = opts.(nuke.ListerOpts)

	logrus.Tracef("subscription id: %s", l.opts.SubscriptionId)

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
	client.Client.Authorizer = l.opts.Authorizers.Management
	//client.Authorizer = opts.Authorizers.Management
	//client.RetryAttempts = 1
	//client.RetryDuration = time.Second * 2

	logrus.Trace("attempting to list budgets for subscription")

	pctx := context.Background()
	ctx, cancel := context.WithDeadline(pctx, time.Now().Add(10*time.Second))
	defer cancel()

	list, err := client.List(ctx, commonids.ScopeId{
		Scope: fmt.Sprintf("/subscriptions/%s", l.opts.SubscriptionId),
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
