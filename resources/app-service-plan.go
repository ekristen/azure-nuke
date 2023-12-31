package resources

import (
	"context"
	"github.com/ekristen/azure-nuke/pkg/nuke"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/web/mgmt/2021-03-01/web"

	"github.com/ekristen/cloud-nuke-sdk/pkg/resource"
	"github.com/ekristen/cloud-nuke-sdk/pkg/types"
)

func init() {
	resource.Register(resource.Registration{
		Name:   "AppServicePlan",
		Scope:  nuke.ResourceGroup,
		Lister: AppServicePlanLister{},
	})
}

type AppServicePlanLister struct {
	opts nuke.ListerOpts
}

func (l AppServicePlanLister) SetOptions(opts interface{}) {
	l.opts = opts.(nuke.ListerOpts)
}

func (l AppServicePlanLister) List() ([]resource.Resource, error) {
	logrus.Tracef("subscription id: %s", l.opts.SubscriptionId)

	client := web.NewAppServicePlansClient(l.opts.SubscriptionId)
	client.Authorizer = l.opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	logrus.Trace("attempting to list ssh key")

	ctx := context.Background()
	list, err := client.ListByResourceGroup(ctx, l.opts.ResourceGroup)
	if err != nil {
		return nil, err
	}

	logrus.Trace("listing ....")

	for list.NotDone() {
		logrus.Trace("list not done")
		for _, g := range list.Values() {
			resources = append(resources, &AppServicePlan{
				client: client,
				name:   *g.Name,
				rg:     l.opts.ResourceGroup,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	return resources, nil
}

type AppServicePlan struct {
	client web.AppServicePlansClient
	name   string
	rg     string
}

func (r *AppServicePlan) Remove() error {
	_, err := r.client.Delete(context.TODO(), r.rg, r.name)
	return err
}

func (r *AppServicePlan) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", r.name)

	return properties
}

func (r *AppServicePlan) String() string {
	return r.name
}
