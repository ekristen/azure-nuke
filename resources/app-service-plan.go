package resources

import (
	"context"
	"github.com/ekristen/azure-nuke/pkg/nuke"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/web/mgmt/2021-03-01/web"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
)

const AppServicePlanResource = "AppServicePlan"

func init() {
	resource.Register(resource.Registration{
		Name:   AppServicePlanResource,
		Scope:  nuke.ResourceGroup,
		Lister: AppServicePlanLister{},
	})
}

type AppServicePlanLister struct {
}

func (l AppServicePlanLister) List(o interface{}) ([]resource.Resource, error) {
	opts := o.(nuke.ListerOpts)

	log := logrus.WithField("r", AppServicePlanResource).WithField("s", opts.SubscriptionId)

	client := web.NewAppServicePlansClient(opts.SubscriptionId)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list ssh key")

	ctx := context.Background()
	list, err := client.ListByResourceGroup(ctx, opts.ResourceGroup)
	if err != nil {
		return nil, err
	}

	log.Trace("listing resources")

	for list.NotDone() {
		log.Trace("list not done")
		for _, g := range list.Values() {
			resources = append(resources, &AppServicePlan{
				client: client,
				name:   *g.Name,
				rg:     opts.ResourceGroup,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	log.Trace("done")

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
