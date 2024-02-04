package resources

import (
	"context"
	"github.com/hashicorp/go-azure-sdk/sdk/environments"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"
	"github.com/hashicorp/go-azure-sdk/resource-manager/resources/2022-09-01/resourcegroups"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/nuke"
)

const ResourceGroupResource = "ResourceGroup"

func init() {
	resource.Register(&resource.Registration{
		Name:   ResourceGroupResource,
		Lister: &ResourceGroupLister{},
		// Scope is set to ResourceGroup because we want to be able to query based on location and resource group
		// 	which is not possible if we treat this as part of subscription because that's considered global.
		Scope: nuke.ResourceGroup,
	})
}

type ResourceGroup struct {
	client         *resourcegroups.ResourceGroupsClient
	name           *string
	location       string
	subscriptionId *string
}

func (r *ResourceGroup) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, commonids.NewResourceGroupID(*r.subscriptionId, *r.name), resourcegroups.DefaultDeleteOperationOptions())
	return err
}

func (r *ResourceGroup) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", r.name)
	properties.Set("Location", r.location)

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
			client:   client,
			name:     entity.Name,
			location: entity.Location,
		})
	}

	log.Trace("done")

	return resources, nil
}
