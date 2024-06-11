package resources

import (
	"context"
	"github.com/ekristen/azure-nuke/pkg/azure"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"
	"github.com/hashicorp/go-azure-sdk/resource-manager/network/2023-09-01/networkinterfaces"
	"github.com/hashicorp/go-azure-sdk/sdk/environments"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
)

const NetworkInterfaceResource = "NetworkInterface"

func init() {
	registry.Register(&registry.Registration{
		Name:     NetworkInterfaceResource,
		Scope:    azure.ResourceGroupScope,
		Resource: &NetworkInterface{},
		Lister:   &NetworkInterfaceLister{},
	})
}

type NetworkInterfaceLister struct {
}

func (l NetworkInterfaceLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

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
			BaseResource: &BaseResource{
				Region:         g.Location,
				ResourceGroup:  &opts.ResourceGroup,
				SubscriptionID: &opts.SubscriptionID,
			},
			client: client,
			Name:   g.Name,
			Tags:   g.Tags,
		})
	}

	log.Trace("done listing network interfaces")

	return resources, nil
}

type NetworkInterface struct {
	*BaseResource `property:",inline"`

	client *networkinterfaces.NetworkInterfacesClient
	Name   *string
	Tags   *map[string]string
}

func (r *NetworkInterface) Remove(ctx context.Context) error {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(30*time.Second))
	defer cancel()

	_, err := r.client.Delete(ctx, commonids.NewNetworkInterfaceID(*r.SubscriptionID, *r.ResourceGroup, *r.Name))
	return err
}

func (r *NetworkInterface) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *NetworkInterface) String() string {
	return *r.Name
}
