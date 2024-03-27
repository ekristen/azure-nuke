package resources

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"
	"github.com/hashicorp/go-azure-sdk/resource-manager/network/2023-09-01/networkinterfaces"
	"github.com/hashicorp/go-azure-sdk/sdk/environments"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/nuke"
)

const NetworkInterfaceResource = "NetworkInterface"

func init() {
	registry.Register(&registry.Registration{
		Name:   NetworkInterfaceResource,
		Scope:  nuke.ResourceGroup,
		Lister: &NetworkInterfaceLister{},
	})
}

type NetworkInterfaceLister struct {
}

func (l NetworkInterfaceLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancel()

	log := logrus.WithField("r", NetworkInterfaceResource).WithField("s", opts.SubscriptionID)

	resources := make([]resource.Resource, 0)

	client, err := networkinterfaces.NewNetworkInterfacesClientWithBaseURI(environments.AzurePublic().ResourceManager)
	if err != nil {
		return resources, err
	}
	client.Client.Authorizer = opts.Authorizers.Management

	log.Trace("attempting to list network interfaces")

	list, err := client.ListComplete(ctx, commonids.NewResourceGroupID(opts.SubscriptionID, opts.ResourceGroup))
	if err != nil {
		return nil, err
	}

	log.Trace("listing ....")

	for _, g := range list.Items {
		resources = append(resources, &NetworkInterface{
			client:         client,
			name:           g.Name,
			region:         g.Location,
			tags:           g.Tags,
			rg:             &opts.ResourceGroup,
			subscriptionID: &opts.SubscriptionID,
		})
	}

	log.Trace("done listing network interfaces")

	return resources, nil
}

type NetworkInterface struct {
	client         *networkinterfaces.NetworkInterfacesClient
	name           *string
	rg             *string
	region         *string
	subscriptionID *string
	tags           *map[string]string
}

func (r *NetworkInterface) Remove(ctx context.Context) error {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancel()

	_, err := r.client.Delete(ctx, commonids.NewNetworkInterfaceID(*r.subscriptionID, *r.rg, *r.name))
	return err
}

func (r *NetworkInterface) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", r.name)
	properties.Set("ResourceGroup", r.rg)
	properties.Set("Region", r.region)

	for k, v := range *r.tags {
		properties.SetTag(&k, v)
	}

	return properties
}

func (r *NetworkInterface) String() string {
	return *r.name
}
