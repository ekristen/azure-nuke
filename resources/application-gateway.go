package resources

import (
	"context"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"
	network "github.com/hashicorp/go-azure-sdk/resource-manager/network/2023-09-01"
	"github.com/hashicorp/go-azure-sdk/resource-manager/network/2023-09-01/applicationgateways"
	"github.com/hashicorp/go-azure-sdk/sdk/client/resourcemanager"
	"github.com/hashicorp/go-azure-sdk/sdk/environments"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const ApplicationGatewayResource = "ApplicationGateway"

func init() {
	registry.Register(&registry.Registration{
		Name:     ApplicationGatewayResource,
		Scope:    azure.ResourceGroupScope,
		Resource: &ApplicationGateway{},
		Lister:   &ApplicationGatewayLister{},
	})
}

type ApplicationGatewayLister struct{}

func (l ApplicationGatewayLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	var resources []resource.Resource
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", ApplicationGatewayResource).WithField("s", opts.SubscriptionID)

	client, err := network.NewClientWithBaseURI(environments.AzurePublic().ResourceManager, func(c *resourcemanager.Client) {
		c.Authorizer = opts.Authorizers.ResourceManager
	})
	if err != nil {
		return nil, err
	}

	log.Trace("attempting to list budgets for subscription")

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(10*time.Second))
	defer cancel()

	log.Trace("attempting to list applications")

	listing, err := client.ApplicationGateways.List(ctx, commonids.NewResourceGroupID(opts.SubscriptionID, opts.ResourceGroup))
	if err != nil {
		return nil, err
	}

	log.Trace("listing applications")

	for _, entry := range *listing.Model {
		resources = append(resources, &ApplicationGateway{
			BaseResource: &BaseResource{
				Region:         ptr.String("global"),
				SubscriptionID: ptr.String(opts.SubscriptionID), // note: this is just the guid
				ResourceGroup:  ptr.String(opts.ResourceGroup),
			},
			client: client,
			ID:     entry.Id,
			Name:   entry.Name,
		})
	}

	log.Trace("done")

	return resources, nil
}

type ApplicationGateway struct {
	*BaseResource `property:",inline"`

	client *network.Client
	ID     *string
	Name   *string
}

func (r *ApplicationGateway) CommonID() applicationgateways.ApplicationGatewayId {
	return applicationgateways.NewApplicationGatewayID(*r.SubscriptionID, *r.ResourceGroup, *r.Name)
}

func (r *ApplicationGateway) Filter() error {
	return nil
}

func (r *ApplicationGateway) Remove(ctx context.Context) error {
	if _, err := r.client.ApplicationGateways.Delete(ctx, r.CommonID()); err != nil {
		return err
	}

	return nil
}

func (r *ApplicationGateway) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ApplicationGateway) String() string {
	return *r.Name
}
