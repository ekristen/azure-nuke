package resources

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"
	"github.com/hashicorp/go-azure-sdk/resource-manager/resources/2022-09-01/resourcegroups"
	"github.com/hashicorp/go-azure-sdk/sdk/environments"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/nuke"
)

const ResourceGroupResource = "ResourceGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:   ResourceGroupResource,
		Lister: &ResourceGroupLister{},
		// Scope is set to ResourceGroup because we want to be able to query based on location and resource group
		// 	which is not possible if we treat this as part of subscription because that's considered global.
		Scope: nuke.Subscription,
	})
}

type ResourceGroup struct {
	client         *resourcegroups.ResourceGroupsClient
	name           *string
	location       string
	subscriptionId string
	listerOpts     *nuke.ListerOpts
}

func (r *ResourceGroup) Filter() error {
	if !slices.Contains(r.listerOpts.Locations, r.location) {
		return fmt.Errorf("resource not in enabled region/location")
	}

	return nil
}

func (r *ResourceGroup) Remove(ctx context.Context) error {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancel()

	_, err := r.client.Delete(ctx, commonids.NewResourceGroupID(r.subscriptionId, *r.name), resourcegroups.DefaultDeleteOperationOptions())
	return err
}

func (r *ResourceGroup) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", r.name)
	properties.Set("Location", r.location)
	properties.Set("SubscriptionId", r.subscriptionId)

	return properties
}

func (r *ResourceGroup) String() string {
	return *r.name
}

// -------------------

type ResourceGroupLister struct {
}

func (l ResourceGroupLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancel()

	log := logrus.WithField("r", ResourceGroupResource).WithField("s", opts.SubscriptionId)

	client, err := resourcegroups.NewResourceGroupsClientWithBaseURI(environments.AzurePublic().ResourceManager)
	if err != nil {
		return nil, err
	}
	client.Client.Authorizer = opts.Authorizers.Management

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list groups")

	list, err := client.List(ctx, commonids.NewSubscriptionID(opts.SubscriptionId), resourcegroups.ListOperationOptions{})
	if err != nil {
		return nil, err
	}

	log.Trace("listing ....")

	for _, entity := range *list.Model {
		resources = append(resources, &ResourceGroup{
			client:         client,
			name:           entity.Name,
			location:       entity.Location,
			subscriptionId: opts.SubscriptionId,
			listerOpts:     opts,
		})
	}

	log.Trace("done")

	return resources, nil
}
