package resources

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/web/mgmt/2021-03-01/web" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/nuke"
)

const AppServicePlanResource = "AppServicePlan"

func init() {
	registry.Register(&registry.Registration{
		Name:   AppServicePlanResource,
		Scope:  nuke.ResourceGroup,
		Lister: &AppServicePlanLister{},
	})
}

type AppServicePlanLister struct {
}

func (l AppServicePlanLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	log := logrus.WithField("r", AppServicePlanResource).WithField("s", opts.SubscriptionID)

	client := web.NewAppServicePlansClient(opts.SubscriptionID)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list ssh key")

	list, err := client.ListByResourceGroup(ctx, opts.ResourceGroup)
	if err != nil {
		return nil, err
	}

	log.Trace("listing resources")

	for list.NotDone() {
		log.Trace("list not done")
		for _, g := range list.Values() {
			resources = append(resources, &AppServicePlan{
				client:        client,
				Name:          *g.Name,
				ResourceGroup: opts.ResourceGroup,
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
	client        web.AppServicePlansClient
	Name          string
	ResourceGroup string
}

func (r *AppServicePlan) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, r.ResourceGroup, r.Name)
	return err
}

func (r *AppServicePlan) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *AppServicePlan) String() string {
	return r.Name
}
