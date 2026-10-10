package resources

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v6"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const AppServicePlanResource = "AppServicePlan"

func init() {
	registry.Register(&registry.Registration{
		Name:     AppServicePlanResource,
		Scope:    azure.ResourceGroupScope,
		Resource: &AppServicePlan{},
		Lister:   &AppServicePlanLister{},
	})
}

type AppServicePlanLister struct{}

func (l AppServicePlanLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", AppServicePlanResource).WithField("s", opts.SubscriptionID)

	client, err := armappservice.NewPlansClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list app service plans")

	pager := client.NewListByResourceGroupPager(opts.ResourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, g := range page.Value {
			resources = append(resources, &AppServicePlan{
				BaseResource: &BaseResource{
					ResourceGroup: &opts.ResourceGroup,
				},
				client: client,
				Name:   *g.Name,
			})
		}
	}

	log.Trace("done")

	return resources, nil
}

type AppServicePlan struct {
	*BaseResource `property:",inline"`

	client *armappservice.PlansClient
	Name   string
}

func (r *AppServicePlan) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, r.GetResourceGroup(), r.Name, nil)
	return err
}

func (r *AppServicePlan) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *AppServicePlan) String() string {
	return r.Name
}
